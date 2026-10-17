#!/bin/bash
# Converts all BTSE channels to AsyncAPI 3 client operations. The Go converter
# also repairs misplaced schemas and missing command payload declarations.
set -euo pipefail
trap 'on_error "$?" "$LINENO"' ERR

LOG_FILE="${LOG_FILE:-/tmp/$(basename "$0" .sh).log}"
exec > >(tee -a "$LOG_FILE") 2>&1

on_error() {
	log ERROR "command failed line=$2 exit_code=$1"
}

log() {
	jq -nc \
		--arg time "$(date -u '+%Y-%m-%dT%H:%M:%SZ')" \
		--arg level "$1" \
		--arg file "${BASH_SOURCE[1]##*/}" \
		--arg line "${BASH_LINENO[0]}" \
		--arg func "${FUNCNAME[1]:-main}" \
		--arg msg "$2" \
		'{time:$time,level:$level,file:$file,line:($line|tonumber),func:$func,msg:$msg}' >&2
}

product="${1:-}"
case "$product" in
spot | futures | otc) ;;
*)
	log ERROR 'Usage: convert-asyncapi.sh <spot|futures|otc>'
	exit 2
	;;
esac

cd "$(dirname "${BASH_SOURCE[0]}")/.."
go tool convert-asyncapi \
	"api/${product}-websocket.asyncapi.yaml" \
	"api/${product}-websocket.client.json"
log INFO "converted ${product} AsyncAPI"
