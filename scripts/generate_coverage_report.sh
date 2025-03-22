#!/bin/bash
# scripts/generate_coverage_report.sh

# Create directory for reports
mkdir -p coverage/reports

# Run tests with coverage
go test -coverprofile=coverage/coverage.out ./...

# Generate summary
go tool cover -func=coverage/coverage.out > coverage/summary.txt
total=$(grep "total:" coverage/summary.txt | awk '{print $3}')
echo "Total coverage: $total"

# Generate HTML report
go tool cover -html=coverage/coverage.out -o coverage/report.html

# Generate per-package coverage information
echo "Package coverage:" > coverage/packages.txt
go tool cover -func=coverage/coverage.out | grep -v "total:" >> coverage/packages.txt

# Identify packages with low coverage (below 50%)
echo "Low coverage packages:" > coverage/low_coverage.txt
go tool cover -func=coverage/coverage.out | awk '$3 < "50.0%" && $1 !~ /total/ {print $1 ": " $3}' >> coverage/low_coverage.txt

# Identify files with no tests
echo "Files without tests:" > coverage/no_tests.txt
go list -f '{{.Dir}}' ./... | while read -r pkg; do
  go list -f '{{range .GoFiles}}{{$.Dir}}/{{.}}{{"\n"}}{{end}}' "$pkg" | while read -r file; do
    if ! grep -q "$(basename "$file" .go)" coverage/coverage.out; then
      echo "$file" >> coverage/no_tests.txt
    fi
  done
done

echo "Coverage reports generated in the coverage directory"
