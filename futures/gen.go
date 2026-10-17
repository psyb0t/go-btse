// Package futures provides the BTSE futures REST API client.
package futures

//go:generate bash ../scripts/fetch-spec.sh futures-rest.openapi.yaml
//go:generate go tool oapi-codegen --config=config.yaml ../api/futures-rest.openapi.yaml
