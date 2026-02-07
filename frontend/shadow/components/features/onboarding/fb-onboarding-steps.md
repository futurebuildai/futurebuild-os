# fb-onboarding-steps

## Intent
*   **High Level:** Horizontal progress bar for project onboarding wizard
*   **Business Value:** Shows users their progress through the onboarding flow

## Responsibility
*   Display 4-step progress indicator: Upload → Extract → Details → Review
*   Show completed steps with checkmarks
*   Highlight active step with pulse animation
*   Respond to stage changes from onboarding store

## Key Logic
*   **_getStepState():** Compares step index to current stage to return 'completed', 'active', or 'pending'
*   **_renderIcon():** Shows checkmark for completed, step-specific icon for active/pending
*   **_renderStep():** Renders step indicator with connector line

## Visual States
*   **Completed:** Purple gradient background, white checkmark
*   **Active:** Purple border, pulse animation, highlighted label
*   **Pending:** Gray background, muted label

## Dependencies
*   **Upstream:** `fb-view-onboarding`
*   **Downstream:** Reads `currentStage` signal from `onboarding-store.ts`
