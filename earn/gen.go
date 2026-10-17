// Package earn provides the BTSE earn REST API client.
package earn

//go:generate bash ../scripts/fetch-spec.sh earn-rest.openapi.yaml
//go:generate go tool oapi-codegen --config=config.yaml ../api/earn-rest.openapi.yaml
