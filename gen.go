// Package btse provides the BTSE spot REST API client.
package btse

//go:generate bash scripts/fetch-spec.sh spot-rest.openapi.yaml
//go:generate bash scripts/patch-spec.sh spot
//go:generate go tool oapi-codegen --config=config.yaml api/spot-rest.openapi.fixed.yaml
