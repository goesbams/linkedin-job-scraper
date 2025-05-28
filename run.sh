#!/bin/bash

# LinkedIn Scraper Runner Script
# Usage: ./run.sh <country> <job_title> [limit]

if [ $# -lt 2 ]; then
    echo "Usage: $0 <country> <job_title> [limit]"
    echo "Examples:"
    echo "  $0 \"Germany\" \"golang developer\" 50"
    echo "  $0 \"Austria\" \"software engineer\" 25"
    echo "  $0 \"Estonia\" \"backend developer\" 30"
    exit 1
fi

COUNTRY="$1"
JOB_TITLE="$2"
LIMIT="${3:-25}"

echo "🔍 Searching for '$JOB_TITLE' jobs in $COUNTRY (limit: $LIMIT)"
echo "🇮🇩 Checking for Indonesian employees..."

# Run the scraper
go run main.go "$COUNTRY" "$JOB_TITLE" "$LIMIT"

# Check if results were generated
RESULT_FILE="linkedin_jobs_$(echo "$COUNTRY" | tr '[:upper:]' '[:lower:]' | tr ' ' '_')_$(echo "$JOB_TITLE" | tr '[:upper:]' '[:lower:]' | tr ' ' '_')_*.json"

if ls $RESULT_FILE 1> /dev/null 2>&1; then
    echo "📁 Moving results to results directory..."
    mv $RESULT_FILE results/
    echo "✅ Results saved in results directory"
else
    echo "⚠️  No results file found"
fi