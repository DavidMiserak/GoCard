#!/bin/bash
# scripts/visual_coverage_report.sh

# Ensure go-cover-treemap is installed
if ! command -v go-cover-treemap &> /dev/null; then
    echo "Installing go-cover-treemap..."
    go install github.com/nikolaydubina/go-cover-treemap@latest
fi

# Ensure gocovsh is installed
if ! command -v gocovsh &> /dev/null; then
    echo "Installing gocovsh..."
    go install github.com/orlangure/gocovsh@latest
fi

# Run tests with coverage
echo "Running tests with coverage..."
go test -coverprofile=coverage.out ./...

# Generate HTML report
echo "Generating standard HTML report..."
go tool cover -html=coverage.out -o coverage.html

# Generate treemap visualization
echo "Generating treemap visualization..."
go-cover-treemap -coverprofile coverage.out > coverage_treemap.svg

# Generate coverage heat map directory
mkdir -p coverage_report
echo "Generating coverage heat map..."
gocovsh covdata --profile coverage.out --html ./coverage_report

echo "Coverage reports generated:"
echo "1. Standard HTML: coverage.html"
echo "2. Treemap visualization: coverage_treemap.svg"
echo "3. Interactive heatmap: ./coverage_report/index.html"
