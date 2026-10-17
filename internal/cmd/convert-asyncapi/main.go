// Command convert-asyncapi upgrades BTSE's AsyncAPI 2.6 documents for clients.
package main

import (
	"encoding/json"
	"log"
	"maps"
	"net/url"
	"os"
	"regexp"
	"strings"

	"github.com/psyb0t/ctxerrors"
	"github.com/psyb0t/ctxerrors/commerr"
	"github.com/psyb0t/go-btse/internal/specfile"
	"github.com/psyb0t/go-btse/internal/specprose"
	"gopkg.in/yaml.v3"
)

type object = map[string]any

const (
	schemasPrefix      = "#/components/schemas/"
	messagesPrefix     = "#/components/messages/"
	channelsPrefix     = "#/channels/"
	pathArguments      = 2
	subscribeDirection = "subscribe"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		log.Fatal(err)
	}
}

func run(args []string) error {
	if len(args) != pathArguments {
		return ctxerrors.Wrap(commerr.ErrInvalidValue, "want input and output")
	}

	data, err := specfile.Read(args[0])
	if err != nil {
		return ctxerrors.Wrap(err, "read AsyncAPI")
	}

	var document object
	if err := yaml.Unmarshal(data, &document); err != nil {
		return ctxerrors.Wrap(err, "decode AsyncAPI")
	}

	if document["asyncapi"] != "2.6.0" {
		return ctxerrors.Wrap(commerr.ErrInvalidValue, "want AsyncAPI 2.6.0")
	}

	convert(document)
	specprose.Normalize(document)

	encoded, err := json.MarshalIndent(document, "", "  ")
	if err != nil {
		return ctxerrors.Wrap(err, "encode AsyncAPI 3")
	}

	if err := specfile.Write(args[1], encoded); err != nil {
		return ctxerrors.Wrap(err, "write AsyncAPI 3")
	}

	return nil
}

func asObject(value any) object {
	result, _ := value.(map[string]any)

	return result
}

func text(value any) string {
	result, _ := value.(string)

	return result
}

func pointer(value string) string {
	return strings.NewReplacer("~", "~0", "/", "~1").Replace(value)
}

func reference(value string) object {
	return object{"$ref": value}
}

func convert(document object) {
	document["asyncapi"] = "3.0.0"
	components := asObject(document["components"])
	schemas := asObject(components["schemas"])
	messages := asObject(components["messages"])

	// BTSE places payload schemas in messages, but references schemas.
	for name, value := range messages {
		message := asObject(value)
		if message["type"] != nil {
			schemas[name] = message
			messages[name] = object{"payload": reference(schemasPrefix + name)}
		}
	}

	convertServers(asObject(document["servers"]))

	operations := object{}
	channels := asObject(document["channels"])

	for address, value := range channels {
		channel := asObject(value)
		channel["address"] = address
		convertChannelServers(channel)

		channelMessages := object{}

		for _, direction := range []string{"publish", subscribeDirection} {
			operation := asObject(channel[direction])
			if operation == nil {
				continue
			}

			convertOperation(address, direction, operation,
				channelMessages, messages, schemas)
			name := text(operation["operationId"])
			delete(operation, "operationId")
			operations[name] = operation

			delete(channel, direction)
		}

		channel["messages"] = channelMessages
	}

	document["operations"] = operations

	for _, value := range messages {
		message := asObject(value)
		if example, exists := message["example"]; exists {
			message["examples"] = []any{object{"payload": example}}
			delete(message, "example")
		}
	}
}

func convertServers(servers object) {
	for _, value := range servers {
		server := asObject(value)

		parsed, err := url.Parse(text(server["url"]))
		if err != nil {
			continue // The generator validates malformed server URLs.
		}

		server["host"] = parsed.Host
		server["pathname"] = parsed.Path
		delete(server, "url")
	}
}

func convertChannelServers(channel object) {
	servers, _ := channel["servers"].([]any)
	converted := make([]any, 0, len(servers))

	for _, server := range servers {
		converted = append(converted, reference("#/servers/"+text(server)))
	}

	channel["servers"] = converted
}

func convertOperation(
	address, direction string,
	operation, channelMessages, messages, schemas object,
) {
	name := text(operation["operationId"])
	message := asObject(operation["message"])

	variants, ok := message["oneOf"].([]any)
	if !ok {
		variants = []any{message}
	}

	refs := make([]any, 0, len(variants))
	for _, variant := range variants {
		current := asObject(variant)

		key := strings.TrimPrefix(text(current["$ref"]), messagesPrefix)
		if key == "" {
			key = name
		} else {
			current = asObject(messages[key])
		}

		repairPayload(current)

		if direction == subscribeDirection {
			// Each channel needs its own schema to separate identical shapes.
			key = name + key
			current = specialize(current, schemas, address)
		}

		messages[key] = current
		channelMessages[key] = reference(messagesPrefix + key)
		refs = append(refs, reference(channelsPrefix+pointer(address)+
			"/messages/"+pointer(key)))
	}

	action := "send"
	if direction == subscribeDirection {
		action = "receive"
	}

	operation["action"] = action
	operation["channel"] = reference(channelsPrefix + pointer(address))
	operation["messages"] = refs
	delete(operation, "message")
}

func repairPayload(message object) {
	if message["payload"] != nil {
		return
	}

	// L1/L2 commands only have examples; their wire schema is shared.
	if message["example"] != nil {
		message["payload"] = reference(schemasPrefix + "SubscriptionRequest")
	}
}

func specialize(message, schemas object, address string) object {
	payload := asObject(message["payload"])
	if ref := text(payload["$ref"]); ref != "" {
		payload = asObject(schemas[strings.TrimPrefix(ref, schemasPrefix)])
	}

	cloned := object{}
	maps.Copy(cloned, payload)

	properties := object{}
	maps.Copy(properties, asObject(payload["properties"]))

	if properties["topic"] != nil {
		// Optional envelope fields otherwise make unrelated messages overlap.
		pattern := regexp.QuoteMeta(address)
		pattern = regexp.MustCompile(`\\\{[^}]+\\\}`).ReplaceAllString(
			pattern, "[^:]+")
		// BTSE documents an omitted grouping suffix as the default level 0.
		if strings.HasSuffix(address, "_{grouping}") {
			pattern = strings.TrimSuffix(pattern, "_[^:]+") + "(?:_[^:]+)?"
		}

		properties["topic"] = object{"type": "string", "pattern": "^" +
			pattern + "$"}
		cloned["required"] = []any{"topic"}
	}

	if properties["event"] != nil {
		cloned["required"] = []any{"event", "channel"}
	}

	if len(properties) > 0 {
		cloned["properties"] = properties
	}

	result := object{}
	maps.Copy(result, message)

	result["payload"] = cloned

	return result
}
