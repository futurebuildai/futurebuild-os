# Step 77: Magic Upload Trigger

**Phase:** 11 (The Conversational Hook)  
**Status:** READY FOR IMPLEMENTATION  
**Est. Duration:** 0.5 Days  
**Owner:** Frontend Developer

---

## 🎯 Objective

Implement drag-and-drop file upload in the onboarding wizard that automatically triggers document analysis when a blueprint PDF is dropped, providing visual feedback throughout the upload and extraction process.

---

## 📐 Architectural Guardrails

> [!IMPORTANT]
> These constraints are non-negotiable and align with the system architecture.

### File Handling Flow
```
User drops file → Upload to storage → Get URL → Call /agent/onboard with URL → Apply extraction
```

### Constraints
- **Max file size:** 50MB
- **Allowed types:** `.pdf`, `.png`, `.jpg`, `.jpeg`
- **Upload destination:** `api.documents.upload()` → returns `document_url`
- **No client-side processing:** OCR happens server-side via Vision API

### Reuse Existing Components
Reference: [FRONTEND_SCOPE.md](file:///home/colton/Desktop/FutureBuild_HQ/dev/specs/FRONTEND_SCOPE.md) Section 8.2

- Build on existing `fb-file-drop` patterns from Phase 10 (Step 73)
- Reuse upload API from `api.uploadDocument()`

---

## 📋 Implementation Checklist

### Files to Create/Modify

```
frontend/src/components/features/onboarding/
└── fb-onboarding-dropzone.ts       # New component
```

---

### Task 1: Create Dropzone Component (1 hour)

```typescript
// frontend/src/components/features/onboarding/fb-onboarding-dropzone.ts

import { LitElement, html, css } from 'lit';
import { customElement, state } from 'lit/decorators.js';
import { classMap } from 'lit/directives/class-map.js';

/**
 * Drag-and-drop upload zone for blueprints in onboarding wizard.
 * Triggers document analysis immediately upon file drop.
 * 
 * @fires file-dropped - When a valid file is dropped with { file, documentUrl }
 * @fires upload-error - When upload fails with { error }
 */
@customElement('fb-onboarding-dropzone')
export class FBOnboardingDropzone extends LitElement {
  
  static override styles = css`
    :host {
      display: block;
    }

    .dropzone {
      border: 2px dashed var(--fb-border);
      border-radius: var(--fb-radius-lg);
      padding: var(--fb-spacing-lg);
      text-align: center;
      transition: all 0.2s ease;
      cursor: pointer;
      background: var(--fb-bg-tertiary);
    }

    .dropzone:hover,
    .dropzone.drag-over {
      border-color: var(--fb-primary);
      background: rgba(102, 126, 234, 0.1);
    }

    .dropzone.drag-over {
      transform: scale(1.02);
    }

    .dropzone.uploading {
      border-color: var(--fb-primary);
      cursor: not-allowed;
    }

    .dropzone.error {
      border-color: var(--fb-error);
      background: rgba(198, 40, 40, 0.1);
    }

    .dropzone-icon {
      font-size: 2rem;
      margin-bottom: var(--fb-spacing-sm);
    }

    .dropzone-text {
      color: var(--fb-text-secondary);
      font-size: var(--fb-text-sm);
    }

    .dropzone-hint {
      color: var(--fb-text-muted);
      font-size: var(--fb-text-xs);
      margin-top: var(--fb-spacing-xs);
    }

    .progress-container {
      width: 100%;
      height: 4px;
      background: var(--fb-border);
      border-radius: 2px;
      margin-top: var(--fb-spacing-md);
      overflow: hidden;
    }

    .progress-bar {
      height: 100%;
      background: var(--fb-primary);
      transition: width 0.3s ease;
    }

    .error-message {
      color: var(--fb-error);
      font-size: var(--fb-text-sm);
      margin-top: var(--fb-spacing-sm);
    }

    /* Hidden file input */
    input[type="file"] {
      display: none;
    }
  `;

  @state() private _isDragOver = false;
  @state() private _isUploading = false;
  @state() private _uploadProgress = 0;
  @state() private _error: string | null = null;
  @state() private _fileName: string | null = null;

  private _fileInput?: HTMLInputElement;

  // Allowed file types and size
  private readonly ALLOWED_TYPES = [
    'application/pdf',
    'image/png',
    'image/jpeg',
    'image/jpg'
  ];
  private readonly MAX_SIZE_BYTES = 50 * 1024 * 1024; // 50MB

  render() {
    const classes = {
      dropzone: true,
      'drag-over': this._isDragOver,
      uploading: this._isUploading,
      error: !!this._error
    };

    return html`
      <div
        class=${classMap(classes)}
        @dragover=${this._handleDragOver}
        @dragleave=${this._handleDragLeave}
        @drop=${this._handleDrop}
        @click=${this._handleClick}
      >
        ${this._isUploading 
          ? this._renderUploading()
          : this._renderIdle()
        }
      </div>
      <input
        type="file"
        accept=".pdf,.png,.jpg,.jpeg"
        @change=${this._handleFileSelect}
      />
    `;
  }

  private _renderIdle() {
    return html`
      <div class="dropzone-icon">📄</div>
      <div class="dropzone-text">
        ${this._isDragOver 
          ? 'Drop to analyze' 
          : 'Drop blueprint here or click to upload'}
      </div>
      <div class="dropzone-hint">
        PDF or images up to 50MB
      </div>
      ${this._error ? html`
        <div class="error-message">${this._error}</div>
      ` : ''}
    `;
  }

  private _renderUploading() {
    return html`
      <div class="dropzone-icon">⏳</div>
      <div class="dropzone-text">
        ${this._uploadProgress < 100 
          ? `Uploading ${this._fileName}...` 
          : 'Analyzing blueprint...'}
      </div>
      <div class="progress-container">
        <div 
          class="progress-bar" 
          style="width: ${this._uploadProgress}%"
        ></div>
      </div>
    `;
  }

  // --- Event Handlers ---

  private _handleDragOver(e: DragEvent) {
    e.preventDefault();
    e.stopPropagation();
    this._isDragOver = true;
  }

  private _handleDragLeave(e: DragEvent) {
    e.preventDefault();
    e.stopPropagation();
    this._isDragOver = false;
  }

  private _handleDrop(e: DragEvent) {
    e.preventDefault();
    e.stopPropagation();
    this._isDragOver = false;

    const files = e.dataTransfer?.files;
    if (files && files.length > 0) {
      this._processFile(files[0]);
    }
  }

  private _handleClick() {
    if (this._isUploading) return;
    
    // Create and trigger file input
    if (!this._fileInput) {
      this._fileInput = this.shadowRoot?.querySelector('input[type="file"]') as HTMLInputElement;
    }
    this._fileInput?.click();
  }

  private _handleFileSelect(e: Event) {
    const input = e.target as HTMLInputElement;
    if (input.files && input.files.length > 0) {
      this._processFile(input.files[0]);
      // Reset input so same file can be selected again
      input.value = '';
    }
  }

  // --- File Processing ---

  private async _processFile(file: File) {
    // Validate file type
    if (!this.ALLOWED_TYPES.includes(file.type)) {
      this._error = 'Please upload a PDF or image file';
      return;
    }

    // Validate file size
    if (file.size > this.MAX_SIZE_BYTES) {
      this._error = 'File too large. Maximum size is 50MB.';
      return;
    }

    // Clear any previous error
    this._error = null;
    this._fileName = file.name;
    this._isUploading = true;
    this._uploadProgress = 0;

    try {
      // Stage 1: Upload file (0-50%)
      const documentUrl = await this._uploadFile(file);
      this._uploadProgress = 50;

      // Stage 2: Trigger analysis (50-100%)
      this._uploadProgress = 75;
      
      // Dispatch event for parent to call onboard API
      this.dispatchEvent(new CustomEvent('file-dropped', {
        bubbles: true,
        composed: true,
        detail: {
          file,
          documentUrl
        }
      }));

      this._uploadProgress = 100;
      
      // Reset after brief delay
      setTimeout(() => {
        this._isUploading = false;
        this._uploadProgress = 0;
        this._fileName = null;
      }, 500);

    } catch (error) {
      this._isUploading = false;
      this._error = error instanceof Error 
        ? error.message 
        : 'Upload failed. Please try again.';
      
      this.dispatchEvent(new CustomEvent('upload-error', {
        bubbles: true,
        composed: true,
        detail: { error }
      }));
    }
  }

  private async _uploadFile(file: File): Promise<string> {
    const formData = new FormData();
    formData.append('file', file);
    formData.append('type', 'blueprint');

    const response = await fetch('/api/v1/documents/upload', {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${localStorage.getItem('fb_token')}`
      },
      body: formData
    });

    if (!response.ok) {
      const error = await response.json().catch(() => ({}));
      throw new Error(error.message || 'Upload failed');
    }

    const result = await response.json();
    return result.document_url;
  }
}

declare global {
  interface HTMLElementTagNameMap {
    'fb-onboarding-dropzone': FBOnboardingDropzone;
  }
}
```

---

### Task 2: Wire Dropzone to Chat Panel (30 min)

```typescript
// In fb-onboarding-chat.ts

private async _handleFileDrop(e: CustomEvent<{ file: File; documentUrl: string }>) {
  const { file, documentUrl } = e.detail;
  
  // Add system message about upload
  addMessage({
    id: crypto.randomUUID(),
    role: 'system',
    content: `📄 Analyzing ${file.name}...`,
    timestamp: new Date()
  });

  // Call onboard API with document URL
  isProcessing.value = true;
  try {
    const response = await api.onboard.process({
      session_id: this._sessionId,
      document_url: documentUrl,
      current_state: onboardingValues.value
    });

    // Apply extractions
    if (response.extracted_values) {
      applyAIExtraction(
        response.extracted_values,
        response.confidence_scores || {}
      );
    }

    // Add assistant response
    addMessage({
      id: crypto.randomUUID(),
      role: 'assistant',
      content: response.reply,
      timestamp: new Date()
    });

  } finally {
    isProcessing.value = false;
  }
}
```

---

### Task 3: Add Keyboard Accessibility (15 min)

```typescript
// In fb-onboarding-dropzone.ts

render() {
  return html`
    <div
      class=${classMap(classes)}
      tabindex="0"
      role="button"
      aria-label="Upload blueprint file"
      @keydown=${this._handleKeyDown}
      ...
    >
  `;
}

private _handleKeyDown(e: KeyboardEvent) {
  if (e.key === 'Enter' || e.key === ' ') {
    e.preventDefault();
    this._handleClick();
  }
}
```

---

## 🔌 Interface Contracts

### Events Emitted

| Event | Payload | When |
|-------|---------|------|
| `file-dropped` | `{ file: File, documentUrl: string }` | Valid file uploaded successfully |
| `upload-error` | `{ error: Error }` | Upload or validation failed |

### CSS Custom Properties Used

| Property | Purpose |
|----------|---------|
| `--fb-primary` | Drag-over highlight, progress bar |
| `--fb-border` | Default border color |
| `--fb-error` | Error state styling |
| `--fb-bg-tertiary` | Drop zone background |

---

## ✅ Acceptance Criteria

- [ ] Dragging a file over the dropzone highlights it (border + scale)
- [ ] Dropping a valid PDF/image triggers upload
- [ ] Progress bar shows upload progress (0-50%) then analyzing (50-100%)
- [ ] Invalid file types show error: "Please upload a PDF or image file"
- [ ] Files over 50MB show error: "File too large. Maximum size is 50MB."
- [ ] Clicking the dropzone opens file picker dialog
- [ ] `file-dropped` event fires with document URL after successful upload
- [ ] Error state clears on next valid file attempt
- [ ] Dropzone is keyboard accessible (Enter/Space triggers file picker)
- [ ] Screen readers announce the dropzone purpose

---

## 🧪 Verification Plan

### Manual Testing
1. Drag PDF over dropzone → Verify visual highlight
2. Drop valid PDF → Verify progress bar and success
3. Drop .docx file → Verify "Please upload a PDF" error
4. Drop 60MB file → Verify size error message
5. Click dropzone → Verify file picker opens
6. Press Tab to dropzone + Enter → Verify file picker opens

### Integration Testing
```bash
# Test upload endpoint directly
curl -X POST http://localhost:8080/api/v1/documents/upload \
  -H "Authorization: Bearer $TOKEN" \
  -F "file=@test-blueprint.pdf" \
  -F "type=blueprint"
```

### Accessibility Testing
- Run axe-core on the component
- Test with VoiceOver/NVDA screen reader

---

## 📚 Reference Documents

- [PHASE_11_PRD.md](file:///home/colton/Desktop/FutureBuild_HQ/dev/planning/PHASE_11_PRD.md) Section 5.4 (or 5.6 after physics section added)
- [FRONTEND_SCOPE.md](file:///home/colton/Desktop/FutureBuild_HQ/dev/specs/FRONTEND_SCOPE.md) Section 8.2 (Invoice Ingestion Workflow)
- [fb-file-drop patterns from Phase 10]
- [API_AND_TYPES_SPEC.md](file:///home/colton/Desktop/FutureBuild_HQ/dev/specs/API_AND_TYPES_SPEC.md) - Document upload endpoint
