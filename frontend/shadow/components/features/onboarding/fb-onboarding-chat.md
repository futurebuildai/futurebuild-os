# fb-onboarding-chat

## Intent
*   **High Level:** Chat interface for document-first project onboarding
*   **Business Value:** Enables users to create projects through natural conversation and document upload

## Responsibility
*   Display welcome message and document drop zone
*   Send user messages to `/api/v1/agent/onboard` endpoint
*   Apply AI-extracted values to onboarding store
*   Show procurement warnings for long-lead items
*   Display Create Project button when ready

## Key Logic
*   **_handleSend():** Sends text message, applies extracted_values to store via `applyAIExtraction()`
*   **_handleFileDrop():** Uploads document via multipart, shows progress, stores long-lead items
*   **_handleSubmit():** Creates project via `api.projects.create()`, emits 'project-created' event
*   **_renderProcurementWarnings():** Shows detected long-lead items with estimated lead times
*   **_renderCreateSection():** Shows summary and Create Project button when `isReadyToCreate` is true

## State Management
*   Uses signals from `onboarding-store.ts`:
  - `onboardingMessages` - Chat history
  - `onboardingValues` - Extracted project fields
  - `isProcessing` - Loading state
  - `isReadyToCreate` - Computed from required fields
  - `extractedProcurement` - Long-lead items

## Dependencies
*   **Upstream:** `fb-view-onboarding`
*   **Downstream:** `api.onboard.process`, `api.projects.create`, `fb-input-bar`, `fb-onboarding-dropzone`
