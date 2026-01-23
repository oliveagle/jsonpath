# Run tests
test:
    go test -v ./...

# Run benchmarks
bench:
    go test -bench=. -benchmem ./...

# Generate benchmark report for current version
# Usage: just bench-report [version]
bench-report version="v0.1.4":
    #!/bin/bash
    set -e
    date=$(date +%Y-%m-%d)
    report_file="benchmark_reports/report_{{version}}_${date}.txt"
    mkdir -p benchmark_reports
    echo "Generating benchmark report for version {{version}}..."
    {
        echo "=========================================="
        echo "JSONPath Benchmark Report"
        echo "=========================================="
        echo ""
        echo "Version: {{version}}"
        echo "Date: ${date}"
        echo "Go Version: $(go version | awk '{print $3}')"
        echo ""
        echo "=========================================="
        echo "Benchmarks"
        echo "=========================================="
        echo ""
        go test -bench=. -benchmem ./...
    } > "${report_file}" 2>&1
    echo "Report saved to: ${report_file}"
