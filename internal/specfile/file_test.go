package specfile_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/psyb0t/ctxerrors/commerr"
	"github.com/psyb0t/go-btse/internal/specfile"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPathsMustStayInsideAPI(t *testing.T) {
	t.Parallel()

	testCases := []string{"", "spec.yaml", "/spec.yaml", "api/../spec.yaml"}
	for _, tc := range testCases {
		t.Run(tc, func(t *testing.T) {
			t.Parallel()

			_, err := specfile.Read(tc)
			require.ErrorIs(t, err, commerr.ErrInvalidValue)
			require.ErrorIs(t, specfile.Write(tc, nil), commerr.ErrInvalidValue)
		})
	}
}

func TestRootConfinesSymlinks(t *testing.T) {
	const escapePath = "api/escape.yaml"

	t.Chdir(t.TempDir())
	require.NoError(t, os.Mkdir("api", 0o700))
	require.NoError(t, specfile.Write("api/spec.yaml", []byte("document")))

	data, err := specfile.Read("api/spec.yaml")
	require.NoError(t, err)
	assert.Equal(t, "document", string(data))

	outside := filepath.Join(t.TempDir(), "outside.yaml")
	require.NoError(t, os.WriteFile(outside, []byte("outside"), 0o600))
	require.NoError(t, os.Symlink(outside, escapePath))
	_, err = specfile.Read(escapePath)
	require.Error(t, err)
	require.Error(t, specfile.Write(escapePath, []byte("changed")))

	data, err = os.ReadFile(outside)
	require.NoError(t, err)
	assert.Equal(t, "outside", string(data))
}
