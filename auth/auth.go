// Package btseauth signs private BTSE requests with HMAC-SHA384.
package btseauth

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"io"
	"maps"
	"net/http"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/psyb0t/ctxerrors"
	"github.com/psyb0t/ctxerrors/commerr"
)

// BTSE reads credentials and the lowercase hexadecimal signature here.
const (
	HeaderAPIKey    = "request-api"
	HeaderNonce     = "request-nonce"
	HeaderSignature = "request-sign"
)

// Signer allocates increasing millisecond nonces across concurrent calls.
// Share one signer per API key. Delivery order still depends on the transport.
// A Signer must not be copied after first use.
type Signer struct {
	apiKey    string
	secret    []byte
	lastNonce atomic.Int64
}

// NewSigner rejects missing credentials and API keys with header delimiters.
func NewSigner(apiKey, apiSecret string) (*Signer, error) {
	if strings.TrimSpace(apiKey) == "" || apiSecret == "" {
		return nil, ctxerrors.Wrap(
			commerr.ErrRequiredFieldNotSet,
			"API key and secret are required",
		)
	}

	if strings.ContainsAny(apiKey, "\r\n") {
		return nil, ctxerrors.Wrap(
			commerr.ErrValidationFailed,
			"API key contains a header delimiter",
		)
	}

	return &Signer{apiKey: apiKey, secret: []byte(apiSecret)}, nil
}

// Headers signs path + nonce + body. Supply the exact escaped API path without
// the query string. The body must contain the exact bytes sent on the wire.
func (s *Signer) Headers(path string, body []byte) http.Header {
	nonce := strconv.FormatInt(s.nextNonce(), 10)
	mac := hmac.New(sha512.New384, s.secret)
	// hash.Hash.Write never returns an error.
	_, _ = mac.Write([]byte(path + nonce))
	_, _ = mac.Write(body)

	header := make(http.Header)
	header.Set(HeaderAPIKey, s.apiKey)
	header.Set(HeaderNonce, nonce)
	header.Set(HeaderSignature, hex.EncodeToString(mac.Sum(nil)))

	return header
}

// RequestEditor matches every generated RequestEditorFn. It excludes the query
// string and preserves request body bytes and the original Close behavior.
// Register it after any editor that changes the URL or body.
func (s *Signer) RequestEditor(ctx context.Context, req *http.Request) error {
	if err := ctx.Err(); err != nil {
		return ctxerrors.Wrap(err, "sign canceled request")
	}

	if req == nil || req.URL == nil {
		return ctxerrors.Wrap(
			commerr.ErrRequiredFieldNotSet,
			"request URL is required",
		)
	}

	body, err := readBody(req)
	if err != nil {
		return ctxerrors.Wrap(err, "read request body for signing")
	}

	if req.Header == nil {
		req.Header = make(http.Header)
	}

	path := signingPath(req.URL.EscapedPath())
	maps.Copy(req.Header, s.Headers(path, body))

	return nil
}

func (s *Signer) nextNonce() int64 {
	for {
		previous := s.lastNonce.Load()
		nonce := max(time.Now().UnixMilli(), previous+1)

		if s.lastNonce.CompareAndSwap(previous, nonce) {
			return nonce
		}
	}
}

func signingPath(path string) string {
	// Server routing prefixes are absent from signed spec paths.
	for _, prefix := range []string{"/spot", "/futures"} {
		if strings.HasPrefix(path, prefix+"/api/") {
			return strings.TrimPrefix(path, prefix)
		}
	}

	return path
}

type preservedBody struct {
	io.Reader
	io.Closer
}

func readBody(req *http.Request) ([]byte, error) {
	if req.Body == nil || req.Body == http.NoBody {
		return nil, nil
	}

	original := req.Body

	body, err := io.ReadAll(original)
	if err != nil {
		req.Body = preservedBody{
			Reader: io.MultiReader(bytes.NewReader(body), original),
			Closer: original,
		}

		return nil, ctxerrors.Wrap(err, "buffer request body")
	}

	req.Body = preservedBody{
		Reader: bytes.NewReader(body),
		Closer: original,
	}

	return body, nil
}
