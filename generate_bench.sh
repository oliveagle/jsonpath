#!/bin/bash
# Generate benchmark report for the current version

VERSION=${1:-$(git rev-parse --abbrev-ref HEAD)}
DATE=$(date +%Y-%m-%d)
OUTPUT_DIR="benchmark_reports"

echo "Generating benchmark report for version: $VERSION"
echo "Date: $DATE"

# Run benchmarks and save output
REPORT_FILE="${OUTPUT_DIR}/report_${VERSION}_${DATE}.txt"

{
    echo "=========================================="
    echo "JSONPath Benchmark Report"
    echo "=========================================="
    echo ""
    echo "Version: $VERSION"
    echo "Date: $DATE"
    echo "Go Version: $(go version | awk '{print $3}')"
    echo ""
    echo "=========================================="
    echo "Benchmarks"
    echo "=========================================="
    echo ""
    go test -bench=. -benchmem ./...
} > "$REPORT_FILE" 2>&1

echo "Report saved to: $REPORT_FILE"

# Also print summary
echo ""
echo "Summary:"
echo "  Total benchmarks: $(grep -c "^Benchmark" "$REPORT_FILE" || echo "0")"
echo "  Report size: $(wc -c < "$REPORT_FILE") bytes"
