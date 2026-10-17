// Package wallet provides the BTSE wallet REST API client.
package wallet

//go:generate bash ../scripts/fetch-spec.sh wallet-rest.openapi.yaml
//go:generate bash ../scripts/patch-spec.sh wallet
//go:generate go tool oapi-codegen --config=config.yaml ../api/wallet-rest.openapi.fixed.yaml
