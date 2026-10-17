// Package otcws is generated from BTSE's OTC WebSocket specification.
package otcws

//go:generate bash ../../scripts/fetch-spec.sh otc-websocket.asyncapi.yaml
//go:generate bash ../../scripts/convert-asyncapi.sh otc
//go:generate go tool gen-asyncapi --config=config.yaml ../../api/otc-websocket.client.json
