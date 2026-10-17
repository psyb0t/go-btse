package btseauth_test

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"io"
	"net/http"
	"net/http/httptest"
	"slices"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/psyb0t/ctxerrors/commerr"
	btseauth "github.com/psyb0t/go-btse/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testAPIKey    = "EXAMPLE-DO-NOT-USE"
	testAPISecret = "EXAMPLE-DO-NOT-USE-SECRET"
	testPath      = "/api/v3.3/order"
	testOrigin    = "https://example.invalid"
	testBody      = `{"symbol":"EXAMPLE","size":1}`
	decimalBase   = 10
	int64Bits     = 64
	nonceCount    = 128
)

func newSigner(t *testing.T) *btseauth.Signer {
	t.Helper()

	signer, err := btseauth.NewSigner(testAPIKey, testAPISecret)
	require.NoError(t, err)

	return signer
}

func assertSignature(
	t *testing.T,
	header http.Header,
	path string,
	body []byte,
) {
	t.Helper()

	mac := hmac.New(sha512.New384, []byte(testAPISecret))
	// hash.Hash.Write never returns an error.
	_, _ = mac.Write([]byte(path))
	_, _ = mac.Write([]byte(header.Get(btseauth.HeaderNonce)))
	_, _ = mac.Write(body)

	assert.Equal(t, testAPIKey, header.Get(btseauth.HeaderAPIKey))
	assert.Equal(
		t,
		hex.EncodeToString(mac.Sum(nil)),
		header.Get(btseauth.HeaderSignature),
	)
}

func TestRequestEditorSignsSpecPaths(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		path string
		want string
	}{
		{name: "spot", path: "/spot" + testPath, want: testPath},
		{
			name: "futures",
			path: "/futures/api/v2.3/order",
			want: "/api/v2.3/order",
		},
		{
			name: "wallet",
			path: "/public-api/wallet/v1/user/assets",
			want: "/public-api/wallet/v1/user/assets",
		},
		{name: "no prefix", path: testPath, want: testPath},
		{
			name: "escaped path",
			path: "/spot" + testPath + "/%2F",
			want: testPath + "/%2F",
		},
		{
			name: "similar prefix",
			path: "/spotlight" + testPath,
			want: "/spotlight" + testPath,
		},
		{
			name: "non API path",
			path: "/spot/ws",
			want: "/spot/ws",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequestWithContext(
				t.Context(),
				http.MethodPost,
				testOrigin+tc.path+"?symbol=EXAMPLE",
				bytes.NewBufferString(testBody),
			)
			req.Header = nil
			require.NoError(t, newSigner(t).RequestEditor(t.Context(), req))
			assertSignature(t, req.Header, tc.want, []byte(testBody))

			body, err := io.ReadAll(req.Body)
			require.NoError(t, err)
			require.NoError(t, req.Body.Close())
			assert.Equal(t, testBody, string(body))
			assert.Equal(t, "symbol=EXAMPLE", req.URL.RawQuery)
		})
	}
}

func TestHeadersUseFreshIncreasingMillisecondNonces(t *testing.T) {
	t.Parallel()

	signer := newSigner(t)
	previous := int64(0)

	for range nonceCount {
		header := signer.Headers(testPath, nil)
		nonce, err := strconv.ParseInt(
			header.Get(btseauth.HeaderNonce),
			decimalBase,
			int64Bits,
		)
		require.NoError(t, err)
		assert.Greater(t, nonce, previous)
		assert.WithinDuration(t, time.Now(), time.UnixMilli(nonce), time.Second)
		assertSignature(t, header, testPath, nil)

		previous = nonce
	}
}

func TestHeadersAllocateUniqueConcurrentNonces(t *testing.T) {
	t.Parallel()

	signer := newSigner(t)
	headers := make(chan http.Header, nonceCount)

	var workers sync.WaitGroup
	for range nonceCount {
		workers.Go(func() {
			headers <- signer.Headers(testPath, nil)
		})
	}

	workers.Wait()
	close(headers)

	nonces := make([]int64, 0, nonceCount)

	for header := range headers {
		assertSignature(t, header, testPath, nil)
		nonce, err := strconv.ParseInt(
			header.Get(btseauth.HeaderNonce),
			decimalBase,
			int64Bits,
		)
		require.NoError(t, err)

		nonces = append(nonces, nonce)
	}

	slices.Sort(nonces)
	assert.Len(t, slices.Compact(nonces), nonceCount)
}

func TestNewSignerRejectsInvalidCredentials(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name   string
		key    string
		secret string
		want   error
	}{
		{
			name: "missing key", secret: testAPISecret,
			want: commerr.ErrRequiredFieldNotSet,
		},
		{
			name: "blank key", key: "  ", secret: testAPISecret,
			want: commerr.ErrRequiredFieldNotSet,
		},
		{
			name: "missing secret", key: testAPIKey,
			want: commerr.ErrRequiredFieldNotSet,
		},
		{
			name: "header injection", key: "EXAMPLE\r\nInjected: value",
			secret: testAPISecret, want: commerr.ErrValidationFailed,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			signer, err := btseauth.NewSigner(tc.key, tc.secret)
			require.ErrorIs(t, err, tc.want)
			assert.Nil(t, signer)
			assert.NotContains(t, err.Error(), testAPISecret)
		})
	}
}

func TestRequestEditorRejectsMissingURL(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		req  *http.Request
	}{
		{name: "nil request"},
		{name: "nil URL", req: &http.Request{}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := newSigner(t).RequestEditor(t.Context(), tc.req)
			require.ErrorIs(t, err, commerr.ErrRequiredFieldNotSet)
		})
	}
}

func TestRequestEditorHonorsCancellation(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	req := httptest.NewRequestWithContext(
		t.Context(), http.MethodGet, testOrigin+testPath, nil,
	)
	err := newSigner(t).RequestEditor(ctx, req)
	require.ErrorIs(t, err, context.Canceled)
	assert.Empty(t, req.Header.Get(btseauth.HeaderSignature))
}

func TestRequestEditorPreservesEmptyBodies(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		body io.ReadCloser
	}{
		{name: "nil"},
		{name: "NoBody", body: http.NoBody},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequestWithContext(
				t.Context(), http.MethodGet, testOrigin+testPath, nil,
			)
			req.Body = tc.body
			require.NoError(t, newSigner(t).RequestEditor(t.Context(), req))
			assert.Equal(t, tc.body, req.Body)
			assertSignature(t, req.Header, testPath, nil)
		})
	}
}

type failingBody struct {
	reader *bytes.Reader
	failed bool
	closed bool
}

func (b *failingBody) Read(data []byte) (int, error) {
	if !b.failed {
		b.failed = true
		count, _ := b.reader.Read(data[:1]) // bytes.Reader has data here.

		return count, io.ErrUnexpectedEOF
	}

	count, err := b.reader.Read(data)
	if err != nil {
		// bytes.Reader only returns EOF, which io.Reader requires unwrapped.
		return count, io.EOF
	}

	return count, nil
}

func (b *failingBody) Close() error {
	b.closed = true

	return io.ErrClosedPipe
}

func TestRequestEditorRestoresBytesAfterReadFailure(t *testing.T) {
	t.Parallel()

	original := &failingBody{reader: bytes.NewReader([]byte(testBody))}
	req := httptest.NewRequestWithContext(
		t.Context(), http.MethodPost, testOrigin+testPath, nil,
	)
	req.Body = original
	err := newSigner(t).RequestEditor(t.Context(), req)
	require.ErrorIs(t, err, io.ErrUnexpectedEOF)
	assert.Empty(t, req.Header.Get(btseauth.HeaderSignature))
	assert.False(t, original.closed)

	body, err := io.ReadAll(req.Body)
	require.NoError(t, err)
	assert.Equal(t, testBody, string(body))
	require.ErrorIs(t, req.Body.Close(), io.ErrClosedPipe)
	assert.True(t, original.closed)
}

func TestRequestEditorPreservesCloseOnSuccess(t *testing.T) {
	t.Parallel()

	original := &failingBody{
		reader: bytes.NewReader([]byte(testBody)),
		failed: true,
	}
	req := httptest.NewRequestWithContext(
		t.Context(), http.MethodPost, testOrigin+testPath, nil,
	)
	req.Body = original
	require.NoError(t, newSigner(t).RequestEditor(t.Context(), req))
	assert.False(t, original.closed)
	assertSignature(t, req.Header, testPath, []byte(testBody))
	require.ErrorIs(t, req.Body.Close(), io.ErrClosedPipe)
	assert.True(t, original.closed)
}
