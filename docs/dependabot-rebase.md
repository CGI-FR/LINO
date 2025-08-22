# Dependabot Rebase Automation

This repository includes automation to help manage Dependabot PRs that have had automatic rebasing disabled due to being open for over 30 days.

## Overview

When Dependabot PRs remain open for more than 30 days, GitHub disables automatic rebasing. This automation provides two ways to request rebasing for all open Dependabot PRs:

1. **Automated GitHub Actions workflow** - Runs on a schedule and can be triggered manually
2. **Manual script** - For immediate one-time use

## Automated Workflow

### Schedule
The workflow runs automatically every Monday at 9 AM UTC to prevent PRs from accumulating.

### Manual Trigger
You can also trigger the workflow manually:

1. Go to the Actions tab in the GitHub repository
2. Select "Rebase Dependabot PRs" workflow
3. Click "Run workflow"
4. Optionally enable "Dry run" to see what would be done without making changes

### What it does
- Finds all open Dependabot PRs that are older than 7 days
- Posts `@dependabot rebase` comment on each PR
- Dependabot will then rebase the PR automatically

## Manual Script

For immediate use, you can run the manual script:

```bash
# Dry run to see what would be done
./scripts/rebase-dependabot-prs.sh --dry-run

# Actually request rebases (with confirmation prompt)
./scripts/rebase-dependabot-prs.sh
```

### Prerequisites
- GitHub CLI (`gh`) must be installed and authenticated
- Proper repository permissions (write access to pull requests)

## Current Status

As of this implementation, the following Dependabot PRs were identified as needing rebasing:
- PR #364: bump github.com/rs/zerolog from 1.33.0 to 1.34.0
- PR #359: bump github.com/docker/docker-credential-helpers from 0.9.2 to 0.9.3  
- PR #358: bump github.com/coder/websocket from 1.8.12 to 1.8.13

## Files Added

- `.github/workflows/dependabot-rebase.yml` - Automated workflow
- `scripts/rebase-dependabot-prs.sh` - Manual script
- `scripts/test-dependabot-logic.sh` - Logic validation script
- `docs/dependabot-rebase.md` - This documentation

## How Dependabot Rebase Works

When you comment `@dependabot rebase` on a Dependabot PR:
1. Dependabot receives the command
2. It rebases the PR branch against the latest main branch
3. Force-pushes the updated branch
4. Updates the PR with the new commits
5. Re-enables automatic rebasing for the PR

This solves the "Automatic rebases have been disabled on this pull request as it has been open for over 30 days" issue.