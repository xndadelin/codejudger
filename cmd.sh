#!/bin/bash

if [[ "$1" == "--help" ]]; then
    echo "Usage: hackacode [OPTIONS]"
    echo ""
    echo "Options:"
    echo "  --file <path>         Path to the source code file to be submitted."
    echo "  --challenge <slug>    Slug of the challenge to submit the code for."
    echo "  --language <lang>     Programming language of the source code (e.g., python, cpp)."
    echo "  --help                Show this help message and exit."
    echo ""
    echo "Environment Variables:"
    echo "  HACKACODE_API_KEY     Your API key for authentication with the Hackacode Judger API."
    echo ""
    echo "Examples:"
    echo "  hackacode --file ./main.cpp --challenge sum --language cpp"
    echo "  hackacode --file ./solution.py --challenge max-array --language python"
    echo ""
    exit 0
fi

if [ -z "$HACKACODE_API_KEY" ]; then
    echo "Error: HACKACODE_API_KEY is not set."
    exit 1
fi

while [[ "$#" -gt 0 ]]; do
    case $1 in
        --file) file="$2"; shift ;;
        --challenge) challenge="$2"; shift ;;
        --language) language="$2"; shift ;;
        *) echo "Error: Unknown parameter passed: $1"; exit 1 ;;
    esac
    shift
done

if [ -z "$file" ] || [ -z "$challenge" ] || [ -z "$language" ]; then
    echo "Error: --file, --challenge, and --language arguments are required."
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

max_length=15

status=$(echo "$response" | jq -r '.status')
if [ "$status" == "FAILED" ]; then
    echo "❌ Status: FAILED"
elif [ "$status" == "ACCEPTED" ]; then
    echo "✅ Status: ACCEPTED"
elif [ "$status" == "comp-failed" ]; then
    message=$(echo "$response" | jq -r '.message')
    echo "❌ Compilation failed: $message"
    exit 1
else
    echo "⚠️ Status: Unknown"
fi


echo "🏆 Score: $(echo "$response" | jq -r '.score')"
echo "📚 Slug: $(echo "$response" | jq -r '.slug')"
echo ""

echo "+--------+----------+--------+--------+---------------+---------------+--------+"
echo "| Index  | ExitCode | Time   | Memory | Stdout        | Stderr        | Passed |"
echo "+--------+----------+--------+--------+---------------+---------------+--------+"
echo "$response" | jq -r '.results[] | [.ExitCode, .Time, .Memory, .Stdout, .Stderr, .Passed] | @tsv' | awk -F'\t' -v max_length="$max_length" '{printf "| %-6s | %-8s | %-6s | %-6s | %-13s | %-13s | %-6s |\n", NR, $1, $2, $3, substr($4, 1, max_length) (length($4) > max_length ? "..." : ""), substr($5, 1, max_length) (length($5) > max_length ? "..." : ""), ($6 == "true" ? "✅" : "❌")}'
echo "+--------+----------+--------+--------+---------------+---------------+--------+"

echo "$response" > /tmp/judger_response.json