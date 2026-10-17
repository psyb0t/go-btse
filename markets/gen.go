// Package markets provides the BTSE markets REST API client.
package markets

//go:generate bash ../scripts/fetch-spec.sh markets-rest.openapi.yaml
//go:generate go tool oapi-codegen --config=config.yaml ../api/markets-rest.openapi.yaml
