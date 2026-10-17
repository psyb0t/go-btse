// Package futuresws is generated from BTSE's futures WebSocket specification.
package futuresws

//go:generate bash ../../scripts/fetch-spec.sh futures-websocket.asyncapi.yaml
//go:generate bash ../../scripts/convert-asyncapi.sh futures
//go:generate go tool gen-asyncapi --config=config.yaml ../../api/futures-websocket.client.json
