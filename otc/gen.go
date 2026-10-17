// Package otc provides the BTSE otc REST API client.
package otc

//go:generate bash ../scripts/fetch-spec.sh otc-rest.openapi.yaml
//go:generate bash ../scripts/patch-spec.sh otc
//go:generate go tool oapi-codegen --config=config.yaml ../api/otc-rest.openapi.fixed.yaml
