// Package specfile confines generation inputs and outputs to api/.
package specfile

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/psyb0t/ctxerrors"
	"github.com/psyb0t/ctxerrors/commerr"
)

const (
	specDirectory   = "api"
	specPermissions = 0o600
)

func open(path string) (*os.Root, string, error) {
	relative, ok := strings.CutPrefix(path, specDirectory+"/")
	if !ok || !filepath.IsLocal(relative) {
		return nil, "", ctxerrors.Wrap(
			commerr.ErrInvalidValue, "spec path must stay inside api/")
	}

	root, err := os.OpenRoot(specDirectory)
	if err != nil {
		return nil, "", ctxerrors.Wrap(err, "open spec directory")
	}

	return root, relative, nil
}

// Read reads a spec without permitting traversal or symlink escapes from api/.
func Read(path string) ([]byte, error) {
	root, relative, err := open(path)
	if err != nil {
		return nil, ctxerrors.Wrap(err, "open spec input")
	}

	data, readErr := root.ReadFile(relative)
	if err := errors.Join(readErr, root.Close()); err != nil {
		return nil, ctxerrors.Wrap(err, "read spec input")
	}

	return data, nil
}

// Write writes a private generated spec beneath api/.
func Write(path string, data []byte) error {
	root, relative, err := open(path)
	if err != nil {
		return ctxerrors.Wrap(err, "open spec output")
	}

	writeErr := root.WriteFile(relative, data, specPermissions)
	if err := errors.Join(writeErr, root.Close()); err != nil {
		return ctxerrors.Wrap(err, "write spec output")
	}

	return nil
}
