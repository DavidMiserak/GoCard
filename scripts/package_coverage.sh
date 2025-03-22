#!/bin/bash
# scripts/package_coverage.sh

# Create output directory
mkdir -p coverage_analysis

# Get list of all packages
packages=$(go list ./...)

# Run tests for each package individually and generate report
echo "Package Coverage:"
echo "================="
echo "Package | Coverage | Files | Lines of Code | Covered Lines"
echo "--------|----------|-------|--------------|-------------"

for pkg in $packages; do
    # Skip test packages
    if [[ $pkg == *"_test" ]]; then
        continue
    fi

    echo "Testing $pkg..."

    # Run tests for this package only
    go test -coverprofile=coverage_analysis/temp.out $pkg >/dev/null 2>&1

    if [ $? -eq 0 ]; then
        # Get coverage percentage
        coverage=$(go tool cover -func=coverage_analysis/temp.out | grep total | awk '{print $3}')

        # Count files
        files=$(go list -f '{{len .GoFiles}}' $pkg)

        # Get lines of code (approximate)
        lines=$(find $(go list -f '{{.Dir}}' $pkg) -name "*.go" -not -path "*_test.go" | xargs wc -l 2>/dev/null | tail -n 1 | awk '{print $1}')

        # Calculate covered lines (approximate)
        if [[ $coverage == *"%" ]]; then
            coverage_num=${coverage//%/}
            covered_lines=$(echo "$lines * $coverage_num / 100" | bc)
        else
            covered_lines="N/A"
        fi

        echo "$pkg | $coverage | $files | $lines | $covered_lines" | tee -a coverage_analysis/package_report.txt
    else
        echo "$pkg | 0.0% | N/A | N/A | 0" | tee -a coverage_analysis/package_report.txt
    fi
done

# Sort packages by coverage (lowest first)
echo -e "\nRanked by Coverage (lowest first):"
echo "============================="
tail -n +3 coverage_analysis/package_report.txt | sort -t'|' -k2 -n | head -10

# Clean up temporary files
rm coverage_analysis/temp.out

echo -e "\nFull report saved to coverage_analysis/package_report.txt"
