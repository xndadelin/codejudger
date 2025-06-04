#!/bin/bash

if [ -z "$HACKACODE_API_KEY" ]; then
    echo "HACKACODE_API_KEY is not set."
    exit 1
fi

while [[ "$#" -gt 0 ]]; do
    case $1 in
        --file) file="$2"; shift ;;
        --challenge) challenge="$2"; shift ;;
        --language) language="$2"; shift ;;
        *) echo "Unknown parameter passed: $1"; exit 1 ;;
    esac
    shift
done

if [ -z "$file" ] || [ -z "$challenge" ] || [ -z "$language" ]; then
    echo "Error: --file, --challenge, and --language arguments are required"
    exit 1
fi

if [ ! -f "$file" ]; then
    echo "Error: File '$file' does not exist."
    exit 1
fi

code="$(cat "$file")"

request_json=$(jq -n \
    --arg code "$code" \
    --arg slug "$challenge" \
    --arg language "$language" \
    '{code: $code, slug: $slug, language: $language}')

response=$(curl -s -X POST "https://judger.hackacode.xyz/api/v1" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $HACKACODE_API_KEY" \
    -d "$request_json")

echo "$response"