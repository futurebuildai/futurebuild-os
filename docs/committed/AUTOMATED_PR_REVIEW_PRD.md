# PRD: Automated PR Review (Step 68)

**Status**: Ready for Dev
**Task Name**: `AUTOMATED_PR_REVIEW`
**Owner**: Product Owner

---

## 1. Problem Statement & Goals
*See [PROBLEM_STATEMENT.md](file:///home/colton/.gemini/antigravity/brain/73f1b8c2-1517-4313-932f-ce76c4757bdd/PROBLEM_STATEMENT.md)*

**Goal**: Automate the feedback loop by connecting GitHub Pull Requests to The Tribunal.

---

## 2. User Stories
- **As a Developer**, I want the system to automatically review my PR so that I can fix security or architectural issues before a human reviewer even sees them.
- **As a Technical Architect**, I want the system to enforce project standards (defined in `specs/`) automatically across all PRs.
- **As an Ops Engineer**, I want all AI review decisions to be logged so that I can audit why a specific PR was flagged or approved.

---

## 3. Acceptance Criteria
1. **Verification**: Verify that the GitHub Webhook signature (HMAC-SHA256) is validated before processing.
2. **Persistence**: Verify that every PR review triggers a `TribunalDecision` record in the database.
3. **Visibility**: Verify that the final "Consensus Summary" is posted as a comment on the GitHub PR.
4. **Idempotency**: Verify that the system does not post duplicate comments for the same PR commit/event.
5. **Security**: Verify that NO reviews are performed if the webhook secret or GitHub token is missing (fail closed).

---

## 4. Functional Requirements

### 4.1 GitHub Webhook Integration
- **Endpoint**: `POST /api/v1/webhooks/github`
- **Events to Handle**: `pull_request` (actions: `opened`, `synchronize`, `reopened`).
- **Data Extraction**: Extract PR number, repository name, head SHA, and PR diff/context.

### 4.2 Integration with The Tribunal
- Transform the GitHub PR context into a `TribunalRequest`.
- **Intent**: "Review Pull Request #{PR_NUMBER}: {PR_TITLE}"
- **Context**: Include a summary of changed files and the PR diff (truncated if necessary).

### 4.3 GitHub Commenting
- Use the synthesized `Plan` or `Summary` from the `TribunalResponse` as the comment body.
- Post the comment to the PR via the GitHub REST API.

---

## 5. Data Models

### 5.1 Tribunal Request Context
```json
{
  "case_id": "GITHUB_PR_{REPO}_{PR_NUMBER}_{SHA}",
  "intent": "PR_AUDIT",
  "context": "Diff: ... (truncated to 10k chars)"
}
```

---

## 6. Edge Cases
- **Large PRs**: If the diff exceeds model token limits, the system should summarize the file list instead of providing the full diff, and VOTE: ABSTAIN if context is insufficient.
- **Rate Limiting**: GitHub API calls should include retry logic with exponential backoff.
- **No Change**: If a PR update doesn't change code (e.g., label change), skip the review.

---

## 7. Security & Compliance
- **Secrets Management**: `GITHUB_WEBHOOK_SECRET` and `GITHUB_PAT` must be provided via environment variables.
- **Audit Log**: Every review must be visible in the **Shadow Viewer** with a link back to the GitHub PR.

---

## 8. Handoff to DevTeam
PRD for `AUTOMATED_PR_REVIEW` is complete and ready for technical design.

**Next Step**: Invoke `/devteam AUTOMATED_PR_REVIEW` to generate the technical specifications.

**Input Artifact**: `docs/AUTOMATED_PR_REVIEW_PRD.md`
