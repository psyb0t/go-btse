// Command patch-spec normalizes BTSE's specification prose for generation.
package main

import (
	"log"
	"os"

	"github.com/psyb0t/ctxerrors"
	"github.com/psyb0t/ctxerrors/commerr"
	"github.com/psyb0t/go-btse/internal/specfile"
	"github.com/psyb0t/go-btse/internal/specprose"
	"gopkg.in/yaml.v3"
)

const pathArguments = 2

func main() {
	if err := run(os.Args[1:]); err != nil {
		log.Fatal(err)
	}
}

func run(args []string) error {
	if len(args) != pathArguments {
		return ctxerrors.Wrap(commerr.ErrInvalidValue, "want input and output")
	}

	data, err := specfile.Read(args[0])
	if err != nil {
		return ctxerrors.Wrap(err, "read spec")
	}

	var document map[string]any
	if err := yaml.Unmarshal(data, &document); err != nil {
		return ctxerrors.Wrap(err, "decode spec")
	}

	specprose.Normalize(document)

	encoded, err := yaml.Marshal(document)
	if err != nil {
		return ctxerrors.Wrap(err, "encode spec")
	}

	if err := specfile.Write(args[1], encoded); err != nil {
		return ctxerrors.Wrap(err, "write spec")
	}

	return nil
}
