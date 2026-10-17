// Package earn provides the BTSE earn REST API client.
package earn

//go:generate bash ../scripts/fetch-spec.sh earn-rest.openapi.yaml
//go:generate bash ../scripts/patch-spec.sh earn
//go:generate go tool oapi-codegen --config=config.yaml ../api/earn-rest.openapi.fixed.yaml
