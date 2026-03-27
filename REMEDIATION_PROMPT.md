# Visual QA Remediation Prompt (Stage 4 Re-test)

**To: Claude Code**
Great work on the first pass! The Responsive Mobile layout now properly collapses to a bottom navigation, the console is clear of warnings, and CSS variables are correctly applied. 

However, we have one remaining **FAILED** check for the FutureBuild OS mobile standards:

## 1. Touch Targets Standards (Blocker)
- **Defect**: The intermediate action buttons (specifically the **"Approve Bid"** and **"Reject"** buttons found within the dashboard or project views) measure approximately `111x29px` and `71x29px`. This height fails the strict requirements of the Stage 1 Google Stitch intent.
- **Remediation**: Locate where these precise buttons are rendered (likely within a dashboard widget, action card, or list item in the Lit components). Update their Tailwind/CSS utility classes or Lit styles to enforce a minimum height of `44px` (`min-height: 44px; min-width: 44px;`). 
- **Note**: Ensure padding or layout logic accommodates the larger button heights without breaking the containing cards.

Please apply these final Touch Target updates and inform the Software Tester when ready for the Final Re-test.
