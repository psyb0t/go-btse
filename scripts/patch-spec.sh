#!/bin/bash
# Normalizes dash punctuation in provider descriptions before generation.
# Schema values, operation IDs, examples and wire payloads stay unchanged.
set -euo pipefail
trap 'on_error "$?" "$LINENO"' ERR

LOG_FILE="${LOG_FILE:-/tmp/$(basename "$0" .sh).log}"
exec > >(tee -a "$LOG_FILE") 2>&1

on_error() {
	printf '{"level":"ERROR","line":%d,"exit_code":%d}\n' "$2" "$1" >&2
}

log() {
	jq -nc \
		--arg time "$(date -u '+%Y-%m-%dT%H:%M:%SZ')" \
		--arg level "$1" --arg msg "$2" \
		--arg file "${BASH_SOURCE[1]##*/}" \
		--arg line "${BASH_LINENO[0]}" \
		--arg func "${FUNCNAME[1]:-main}" \
		'{time:$time,level:$level,file:$file,line:($line|tonumber),func:$func,msg:$msg}' >&2
}

product="${1:-}"
case "$product" in
spot | futures | markets | wallet | earn | otc) ;;
*)
	log ERROR 'Usage: patch-spec.sh <product>'
	exit 2
	;;
esac

cd "$(dirname "${BASH_SOURCE[0]}")/.."
go tool patch-spec \
	"api/${product}-rest.openapi.yaml" \
	"api/${product}-rest.openapi.fixed.yaml"
log INFO "normalized ${product} REST prose"
