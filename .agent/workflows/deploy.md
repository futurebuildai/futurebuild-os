---
description: Narrative-driven deployment guide based on current infrastructure.
---

# Deploy Workflow

This workflow is designed to ensure safe and intentional deployments by starting a conversation before taking action.

## 🏁 Phase 1: Context Gathering
1. Start by asking the user: "What is the goal of this deployment?"
2. Present the following options as categories:
    - **Stability/QA (Staging)**: Internal testing of new features.
    - **Client Review (Demo)**: Specific preview for stakeholders.
    - **Global Release (Production)**: Launching to the live audience.

## 🕵️ Phase 2: Architecture Review
1. Reference [DEPLOYMENT_ARCHITECTURE.md](file:///home/colton/Desktop/FutureBuild_HQ/dev/futurebuild-repo/docs/DEPLOYMENT_ARCHITECTURE.md).
2. Confirm the current branch the user is on.
3. If the user is on the wrong branch for their goal, suggest a merge:
    - Example: "You are on `build`, but we need to merge to `staging` to deploy. Should I do that now?"

## 🚀 Phase 3: Execution
1. Perform the necessary `git` operations (merge + push) to trigger the target branch.
2. If Production is the goal:
    - Offer to push a `v*` tag.
    - Or remind the user to click the "Run workflow" button in GitHub Actions.
3. Provide the user with the direct link to the GitHub Actions tab to monitor progress.

## 🏁 Phase 4: Verification
1. Once the deployment is triggered, prompt the user to check the "Activity" logs on DigitalOcean.
2. Stay "on call" until the user confirms the deployment is Green.
