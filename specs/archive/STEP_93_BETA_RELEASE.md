# Technical Specification: Beta Release (Step 93)

| Metadata | Details |
| :--- | :--- |
| **Step** | 93 |
| **Feature** | Beta Release Tagging |
| **Goal** | Final regression smoke test and version tagging. |
| **Related** | Phase 15, PRD Section 4.4 |

---

## 1. Feature Description

This step is the formal "seal of approval" for the Beta release. It involves bumping the version to `2.1.0-beta`, performing a final smoke test of critical paths, and tagging the release in git.

---

## 2. Implementation Steps

### Step 2.1: Version Bump
- Update `package.json`: `"version": "2.1.0-beta"`.
- Update `frontend/package.json` and `backend/package.json` (if applicable).
- Update the UI footer or "About" modal to display `v2.1.0-beta`.

### Step 2.2: Git Tagging
- Commit changes: `chore(release): bump version to 2.1.0-beta`.
- Create tag: `git tag v2.1.0-beta`.
- (Do not push in this step, simply prepare the local state).

---

## 3. Verification Plan

### 3.1 Automated Browser Smoke Test (Claude in Chrome)

**CRITICAL INSTRUCTION**: You must use the `/chome` extension (or equivalent Browser Tool) to execute this Smoke Test.

**Workflow**:
1. **Launch Browser**: Open `http://localhost:8080`.
2. **Version Check**:
    - Scroll to the footer or open the Settings/About menu.
    - **Verify**: text displays `v2.1.0-beta`.
3. **Critical Path Walkthrough**:
    - **Nav**: Click "Projects" → Load List.
    - **Detail**: Click a Project → Load Dashboard.
    - **Chat**: Open Chat → Type "Status check" → Send.
    - **Mobile**: Resize window to 375px → Verify Bottom Nav appears.
4. **Conclusion**:
    - If all steps pass without console errors or visual breakage, the release is deemed **STABLE**.

**Auto-Accept**:
- If using `/chome`, assume **Auto-Accept** permissions for localhost testing.

### 3.2 Manual Verification
- Review `git log -1` to confirm the version bump commit.
- Review `git tag --list` to confirm the tag exists.
