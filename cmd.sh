#!/bin/bash

if [[ "$1" == "--help" || "$1" == "" ]]; then
    echo "Usage: hackacode <command> [OPTIONS]"
    echo ""
    echo "Commands:"
    echo "  submit               Submit code to a challenge"
    echo "  run                  Run code with custom input"
    echo ""
    echo "Global options:"
    echo "  --help               Show this help message and exit"
    echo ""
    echo "Submit opptions:"
    echo "  --file <path>        Path to the source code file to be submitted"
    echo "  --challenge <slug>   Slug of the challenge to submit the code for"
    echo "  --language <lang>    Programming language of the source code (e.g., Python, C++)"
    echo ""
    echo "Run options:"
    echo "  --file <path>        Path to the source code file to run"
    echo "  --language <lang>    Programming language of the source code (e.g., Python, C++)"
    echo "  --input <path>       Path to input file (optional)"
    echo ""
    echo "Environment variables:"
    echo "  HACKACODE_API_KEY    Your API key for authentication with the Hackacode Judger API"
    echo ""
    echo "Examples:"
    echo "  hackacode submit --file ./main.cpp --challenge sum --language C++"
    echo "  hackacode run --file ./solution.py --language Python --input ./input.txt"
    echo ""
    exit 0
fi

if [ -z "$HACKACODE_API_KEY" ]; then
    echo "Error: HACKACODE_API_KEY is not set."
    exit 1
fi

COMMAND=$1
shift

case "$COMMAND" in
    submit)
        while [[ "$#" -gt 0 ]]; do
            case $1 in
                --file) file="$2"; shift ;;
                --challenge) challenge="$2"; shift ;;
                --language) language="$2"; shift ;;
                --help) 
                    echo "Usage: hackacode submit [OPTIONS]"
                    echo ""
                    echo "Options:"
                    echo "  --file <path>        Path to the source code file to be submitted"
                    echo "  --challenge <slug>   Slug of the challenge to submit the code for"
                    echo "  --language <lang>    Programming language of the source code (e.g., Python, C++)"
                    exit 0 ;;
                *) echo "Error: Unknown parameter passed for submit: $1"; exit 1 ;;
            esac
            shift
        done

        if [ -z "$file" ] || [ -z "$challenge" ] || [ -z "$language" ]; then
            echo "Error: --file, --challenge, and --language arguments are required for submit."
            exit 1
        fi
        ;;
    run)
        while [[ "$#" -gt 0 ]]; do
            case $1 in
                --file) file="$2"; shift ;;
                --language) language="$2"; shift ;;
                --input) input_file="$2"; shift ;;
                --help)
                    echo "Usage: hackacode run [OPTIONS]"
                    echo ""
                    echo "Options:"
                    echo "  --file <path>        Path to the source code file to run"
                    echo "  --language <lang>    Programming language of the source code (e.g., Python, C++)"
                    echo "  --input <path>       Path to input file (optional)"
                    exit 0 ;;
                *) echo "Error: Unknown parameter passed for run: $1"; exit 1 ;;
            esac
            shift
        done

        if [ -z "$file" ] || [ -z "$language" ]; then
            echo "Error: --file and --language arguments are required for run."
            exit 1
        fi
        ;;
    *)
        echo "Error: Unknown command: $COMMAND"
        echo "Use 'hackacode --help' for usage information."
        exit 1
        ;;
esac

if [ ! -f "$file" ]; then
    echo "Error: File '$file' does not exist."
    exit 1
fi

code="$(cat "$file")"
input=""

if [ "$COMMAND" = "run" ] && [ -n "$input_file" ]; then
    if [ ! -f "$input_file" ]; then
        echo "Error: Input file '$input_file' does not exist."
        exit 1
    fi
    input="$(cat "$input_file")"
fi

if [ "$COMMAND" = "run" ]; then
    request_json=$(jq -n \
        --arg code "$code" \
        --arg language "$language" \
        --arg input "$input" \
        '{code: $code, language: $language, input: $input}')
    
    response=$(curl -s -X POST "https://judger.hackacode.xyz/api/v1/run" \
        -H "Content-Type: application/json" \
        -d "$request_json")
else
    request_json=$(jq -n \
        --arg code "$code" \
        --arg slug "$challenge" \
        --arg language "$language" \
        '{code: $code, slug: $slug, language: $language}')
    
    response=$(curl -s -X POST "https://judger.hackacode.xyz/api/v1" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $HACKACODE_API_KEY" \
        -d "$request_json")
fi

max_length=15

if [ "$COMMAND" = "run" ]; then
    error=$(echo "$response" | jq -r '.error // ""')
    if [ -n "$error" ]; then
        echo "❌ Error: $error"
        exit 1
    fi

    if echo "$response" | jq -e '.result' > /dev/null; then
        echo "✅ Execution successful"
        echo ""
        echo "⏱️ Time: $(echo "$response" | jq -r '.result.Time') seconds"
        echo "🧠 Memory: $(echo "$response" | jq -r '.result.Memory') KB"
        echo ""
        echo "🔤 Stdout:"
        echo "$(echo "$response" | jq -r '.result.Stdout')"
        
        stderr=$(echo "$response" | jq -r '.result.Stderr')
        if [ -n "$stderr" ]; then
            echo ""
            echo "⚠️ Stderr:"
            echo "$stderr"
        fi
    else
        echo "❌ Execution error"
    fi
else
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
fi

echo "$response" > /tmp/judger_response.json