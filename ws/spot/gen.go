// Package spotsws is generated from BTSE's spot WebSocket specification.
package spotsws

//go:generate bash ../../scripts/fetch-spec.sh spot-websocket.asyncapi.yaml
//go:generate bash ../../scripts/convert-asyncapi.sh spot
//go:generate go tool gen-asyncapi --config=config.yaml ../../api/spot-websocket.client.json
