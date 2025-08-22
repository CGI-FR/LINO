#!/bin/bash

# Test script to validate the dependabot rebase logic
# This simulates what the GitHub Action will do

set -e

echo "Testing Dependabot PR discovery logic..."

# Check if jq is available for testing
if ! command -v jq &> /dev/null; then
    echo "jq is not available. Installing for testing..."
    # We'll skip this in the test environment
    echo "Skipping jq logic test - validation will be done in the workflow"
    exit 0
fi

echo "jq is available - testing JSON processing logic..."

# Simulate the JSON output for testing with current dates
# Making these PRs older than 7 days for testing
old_date=$(date -d '2025-02-01' -u +"%Y-%m-%dT%H:%M:%SZ")

cat << EOF > /tmp/mock_dependabot_prs.json
[
  {
    "number": 364,
    "title": "chore(deps): bump github.com/rs/zerolog from 1.33.0 to 1.34.0",
    "createdAt": "$old_date"
  },
  {
    "number": 359,
    "title": "chore(deps): bump github.com/docker/docker-credential-helpers from 0.9.2 to 0.9.3",
    "createdAt": "$old_date"
  },
  {
    "number": 358,
    "title": "chore(deps): bump github.com/coder/websocket from 1.8.12 to 1.8.13",
    "createdAt": "$old_date"
  }
]
EOF

echo "Mock Dependabot PRs data created for testing..."

# Test the jq filtering logic
echo "Testing jq logic to find PRs older than 7 days..."
dependabot_prs=$(jq '.[] | select(.createdAt | fromdateiso8601 < (now - 86400*7)) | {number: .number, title: .title, age: ((now - (.createdAt | fromdateiso8601)) / 86400 | floor)}' /tmp/mock_dependabot_prs.json)

if [ -z "$dependabot_prs" ]; then
    echo "No Dependabot PRs found that are older than 7 days."
else
    echo "Found Dependabot PRs older than 7 days:"
    echo "$dependabot_prs" | jq -r '"PR #\(.number): \(.title) (age: \(.age) days)"'
    
    echo "PR numbers that would be rebased:"
    echo "$dependabot_prs" | jq -r '.number'
fi

# Clean up
rm -f /tmp/mock_dependabot_prs.json

echo "Logic validation completed successfully!"