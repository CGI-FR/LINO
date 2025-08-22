#!/bin/bash

# Manual script to rebase all open Dependabot PRs
# Usage: ./scripts/rebase-dependabot-prs.sh [--dry-run]

set -e

DRY_RUN=false
if [ "$1" = "--dry-run" ]; then
    DRY_RUN=true
    echo "Running in dry-run mode (no comments will be posted)"
fi

echo "Finding all open Dependabot PRs..."

# Check if gh CLI is available
if ! command -v gh &> /dev/null; then
    echo "Error: GitHub CLI (gh) is required but not installed."
    echo "Please install it from: https://cli.github.com/"
    exit 1
fi

# Get all open PRs created by dependabot[bot]
echo "Fetching open Dependabot PRs..."
dependabot_prs=$(gh pr list --author "dependabot[bot]" --state open --json number,title,url)

if [ "$(echo "$dependabot_prs" | jq '. | length')" -eq 0 ]; then
    echo "No open Dependabot PRs found."
    exit 0
fi

echo "Found open Dependabot PRs:"
echo "$dependabot_prs" | jq -r '.[] | "PR #\(.number): \(.title)"'
echo ""

if [ "$DRY_RUN" = "true" ]; then
    echo "Dry run mode - would request rebase for the above PRs but not posting comments."
    exit 0
fi

# Ask for confirmation
echo "Do you want to request rebase for all these PRs? (y/N)"
read -r confirmation
if [ "$confirmation" != "y" ] && [ "$confirmation" != "Y" ]; then
    echo "Cancelled."
    exit 0
fi

# Rebase each PR by posting a comment
echo "$dependabot_prs" | jq -r '.[] | .number' | while read -r pr_number; do
    if [ -n "$pr_number" ]; then
        echo "Requesting rebase for PR #$pr_number..."
        if gh pr comment "$pr_number" --body "@dependabot rebase"; then
            echo "✓ Rebase requested for PR #$pr_number"
        else
            echo "✗ Failed to request rebase for PR #$pr_number"
        fi
        # Add a small delay to avoid rate limiting
        sleep 2
    fi
done

echo "All Dependabot rebase requests completed."