/**
 * FBInterrogatorWizard - Split-Screen PDF Viewer + AI Chat Wizard
 * Sprint 2.2: The Interrogator Interface
 *
 * Two-column layout during onboarding:
 * - Left panel: PDF.js viewer with SVG bounding box overlays for extraction zones
 * - Right panel: AI chat connected to onboardingMessages signal with extraction cards
 * - Bottom: Action bar with "Generate Schedule" button
 *
 * Responsive: stacks vertically on mobile (< 768px).
 */
import { html, css, nothing, type TemplateResult } from 'lit';
import { customElement, state } from 'lit/decorators.js';
import { SignalWatcher } from '@lit-labs/preact-signals';
import { FBElement } from '../../base/FBElement';
import { api, type OnboardProcessResponse, type CreateProjectRequest } from '../../../services/api';
import {
    onboardingMessages,
    onboardingValues,
    onboardingConfidence,
    isProcessing,
    pdfHighlights,
    uploadedPdfUrl,
    extractedProcurement,
    interrogatorStatus,
    addMessage,
    applyAIExtraction,
    setExtractedProcurement,
    updateHighlights,
    setReadyToCreate,
    setInterrogatorStatus,
    type OnboardingMessage,
    type BoundingBox,
    type InterrogatorStatus,
} from '../../../store/onboarding-store';
import { clerkService } from '../../../services/clerk';
import '../../chat/fb-input-bar';
import './fb-onboarding-dropzone';
import type { FBOnboardingDropzone } from './fb-onboarding-dropzone';

// pdf.js dynamic import (ESM)
import * as pdfjsLib from 'pdfjs-dist';

// Vite-compatible worker: use the bundled worker path
pdfjsLib.GlobalWorkerOptions.workerSrc = new URL(
    'pdfjs-dist/build/pdf.worker.min.mjs',
    import.meta.url
).toString();

@customElement('fb-interrogator-wizard')
export class FBInterrogatorWizard extends SignalWatcher(FBElement) {
    static override styles = [
        FBElement.styles,
        css`
            :host {
                display: flex;
                flex-direction: column;
                height: 100%;
                overflow: hidden;
                background: var(--md-sys-color-background);
                color: var(--md-sys-color-on-background);
                font-family: var(--md-ref-typeface-plain);
            }

            /* ========== Split-screen grid ========== */
            .split-screen {
                flex: 1;
                display: grid;
                grid-template-columns: 1fr 1fr;
                gap: 0;
                min-height: 0;
                overflow: hidden;
            }

            @media (max-width: 768px) {
                .split-screen {
                    grid-template-columns: 1fr;
                    grid-template-rows: 1fr 1fr;
                }
            }

            /* ========== Left panel: PDF Viewer ========== */
            .pdf-panel {
                display: flex;
                flex-direction: column;
                min-height: 0;
                overflow: hidden;
                border-right: 1px solid var(--md-sys-color-outline-variant);
                background: var(--md-sys-color-surface-container-low);
            }

            @media (max-width: 768px) {
                .pdf-panel {
                    border-right: none;
                    border-bottom: 1px solid var(--md-sys-color-outline-variant);
                }
            }

            .pdf-header {
                display: flex;
                align-items: center;
                justify-content: space-between;
                padding: 8px 16px;
                border-bottom: 1px solid var(--md-sys-color-outline-variant);
                background: var(--md-sys-color-surface-container);
                flex-shrink: 0;
            }

            .pdf-header-title {
                font: var(--md-sys-typescale-label-large);
                color: var(--md-sys-color-on-surface);
                display: flex;
                align-items: center;
                gap: 8px;
            }

            .pdf-header-title svg {
                width: 18px;
                height: 18px;
                fill: var(--md-sys-color-primary);
            }

            .pdf-page-nav {
                display: flex;
                align-items: center;
                gap: 8px;
            }

            .pdf-page-nav button {
                background: none;
                border: 1px solid var(--md-sys-color-outline-variant);
                border-radius: 6px;
                color: var(--md-sys-color-on-surface);
                width: 28px;
                height: 28px;
                display: flex;
                align-items: center;
                justify-content: center;
                cursor: pointer;
                transition: background 0.15s;
            }

            .pdf-page-nav button:hover:not(:disabled) {
                background: var(--md-sys-color-surface-container-highest);
            }

            .pdf-page-nav button:disabled {
                opacity: 0.4;
                cursor: not-allowed;
            }

            .pdf-page-nav span {
                font: var(--md-sys-typescale-label-medium);
                color: var(--md-sys-color-on-surface-variant);
            }

            .pdf-viewport {
                flex: 1;
                overflow: auto;
                position: relative;
                display: flex;
                align-items: flex-start;
                justify-content: center;
                padding: 16px;
            }

            .pdf-canvas-wrapper {
                position: relative;
                display: inline-block;
                box-shadow: var(--md-sys-elevation-2);
                border-radius: 4px;
                overflow: hidden;
            }

            .pdf-canvas-wrapper canvas {
                display: block;
            }

            .pdf-overlay {
                position: absolute;
                top: 0;
                left: 0;
                width: 100%;
                height: 100%;
                pointer-events: none;
            }

            .pdf-empty {
                flex: 1;
                display: flex;
                flex-direction: column;
                align-items: center;
                justify-content: center;
                gap: 12px;
                padding: 40px;
                text-align: center;
            }

            .pdf-empty svg {
                width: 64px;
                height: 64px;
                fill: var(--md-sys-color-outline);
                opacity: 0.5;
            }

            .pdf-empty p {
                font: var(--md-sys-typescale-body-medium);
                color: var(--md-sys-color-on-surface-variant);
                max-width: 280px;
            }

            /* ========== Right panel: AI Chat ========== */
            .chat-panel {
                display: flex;
                flex-direction: column;
                min-height: 0;
                overflow: hidden;
            }

            .chat-header {
                display: flex;
                align-items: center;
                padding: 8px 16px;
                border-bottom: 1px solid var(--md-sys-color-outline-variant);
                background: var(--md-sys-color-surface-container);
                flex-shrink: 0;
            }

            .chat-header-title {
                font: var(--md-sys-typescale-label-large);
                color: var(--md-sys-color-on-surface);
                display: flex;
                align-items: center;
                gap: 8px;
            }

            .chat-header-title svg {
                width: 18px;
                height: 18px;
                fill: var(--md-sys-color-tertiary);
            }

            .message-list {
                flex: 1;
                overflow-y: auto;
                padding: 16px;
                display: flex;
                flex-direction: column;
                gap: 12px;
                scroll-behavior: smooth;
            }

            .message-list::-webkit-scrollbar {
                width: 6px;
            }

            .message-list::-webkit-scrollbar-track {
                background: transparent;
            }

            .message-list::-webkit-scrollbar-thumb {
                background: var(--md-sys-color-outline-variant);
                border-radius: 3px;
            }

            .message {
                max-width: 85%;
                padding: 10px 14px;
                border-radius: var(--md-sys-shape-corner-large);
                font: var(--md-sys-typescale-body-medium);
                animation: fadeSlide 0.3s ease-out;
                word-wrap: break-word;
            }

            @keyframes fadeSlide {
                from { opacity: 0; transform: translateY(8px); }
                to { opacity: 1; transform: translateY(0); }
            }

            .message.user {
                align-self: flex-end;
                background: var(--md-sys-color-primary-container);
                color: var(--md-sys-color-on-primary-container);
                border-bottom-right-radius: 4px;
            }

            .message.assistant {
                align-self: flex-start;
                background: var(--md-sys-color-surface-container-high);
                color: var(--md-sys-color-on-surface);
                border-bottom-left-radius: 4px;
            }

            .message.system {
                align-self: center;
                background: transparent;
                color: var(--md-sys-color-outline);
                font: var(--md-sys-typescale-label-medium);
                text-align: center;
                max-width: 90%;
                padding: 4px 8px;
                border: 1px dashed var(--md-sys-color-outline-variant);
            }

            .processing-indicator {
                align-self: flex-start;
                padding: 6px 12px;
                color: var(--md-sys-color-secondary);
                font: var(--md-sys-typescale-body-small);
                font-style: italic;
                display: flex;
                align-items: center;
                gap: 8px;
            }

            .processing-indicator::before {
                content: '';
                display: inline-block;
                width: 8px;
                height: 8px;
                border-radius: 50%;
                background: var(--md-sys-color-secondary);
                animation: pulse 1.5s infinite;
            }

            @keyframes pulse {
                0% { transform: scale(0.8); opacity: 0.5; }
                50% { transform: scale(1.2); opacity: 1; }
                100% { transform: scale(0.8); opacity: 0.5; }
            }

            /* ========== Extraction cards ========== */
            .extraction-card {
                background: var(--md-sys-color-surface-container);
                border: 1px solid var(--md-sys-color-outline-variant);
                border-radius: var(--md-sys-shape-corner-medium);
                padding: 12px;
                margin: 4px 0;
                animation: fadeSlide 0.3s ease-out;
            }

            .extraction-card-header {
                font: var(--md-sys-typescale-label-large);
                color: var(--md-sys-color-primary);
                margin-bottom: 8px;
                display: flex;
                align-items: center;
                gap: 6px;
            }

            .extraction-card-header svg {
                width: 16px;
                height: 16px;
                fill: currentColor;
            }

            .extraction-fields {
                display: grid;
                grid-template-columns: repeat(auto-fit, minmax(130px, 1fr));
                gap: 8px;
            }

            .extraction-field {
                display: flex;
                flex-direction: column;
                gap: 2px;
            }

            .extraction-field-label {
                font: var(--md-sys-typescale-label-small);
                color: var(--md-sys-color-on-surface-variant);
                text-transform: uppercase;
                letter-spacing: 0.5px;
            }

            .extraction-field-value {
                font: var(--md-sys-typescale-body-medium);
                color: var(--md-sys-color-on-surface);
                display: flex;
                align-items: center;
                gap: 6px;
            }

            /* Confidence badges */
            .confidence-badge {
                display: inline-flex;
                align-items: center;
                padding: 2px 6px;
                border-radius: 10px;
                font: var(--md-sys-typescale-label-small);
                font-weight: 600;
                letter-spacing: 0.3px;
            }

            .confidence-high {
                background: rgba(34, 197, 94, 0.15);
                color: #00FFA3;
            }

            .confidence-medium {
                background: rgba(245, 158, 11, 0.15);
                color: #f59e0b;
            }

            .confidence-low {
                background: rgba(239, 68, 68, 0.15);
                color: #F43F5E;
            }

            /* ========== Chat input area ========== */
            .chat-input-area {
                flex-shrink: 0;
                border-top: 1px solid var(--md-sys-color-outline-variant);
            }

            .dropzone-compact {
                padding: 8px 16px;
                background: var(--md-sys-color-surface);
                border-top: 1px solid var(--md-sys-color-outline-variant);
            }

            fb-input-bar {
                flex-shrink: 0;
            }

            /* ========== Bottom action bar ========== */
            .action-bar {
                display: flex;
                align-items: center;
                justify-content: space-between;
                padding: 12px 20px;
                border-top: 1px solid var(--md-sys-color-outline-variant);
                background: var(--md-sys-color-surface-container-low);
                flex-shrink: 0;
                gap: 16px;
            }

            .action-summary {
                font: var(--md-sys-typescale-body-small);
                color: var(--md-sys-color-on-surface-variant);
                flex: 1;
                min-width: 0;
                overflow: hidden;
                text-overflow: ellipsis;
                white-space: nowrap;
            }

            .btn-generate {
                display: flex;
                align-items: center;
                gap: 8px;
                padding: 10px 24px;
                background: var(--md-sys-color-primary);
                color: var(--md-sys-color-on-primary);
                font: var(--md-sys-typescale-label-large);
                border: none;
                border-radius: var(--md-sys-shape-corner-full);
                cursor: pointer;
                transition: all 0.2s;
                box-shadow: var(--md-sys-elevation-2);
                white-space: nowrap;
            }

            .btn-generate:hover:not(:disabled) {
                box-shadow: var(--md-sys-elevation-3);
                transform: translateY(-1px);
            }

            .btn-generate:disabled {
                background: var(--md-sys-color-surface-variant);
                color: var(--md-sys-color-on-surface-variant);
                box-shadow: none;
                cursor: not-allowed;
            }

            .btn-generate svg {
                width: 18px;
                height: 18px;
                fill: currentColor;
            }

            .error-toast {
                color: var(--md-sys-color-error);
                font: var(--md-sys-typescale-body-small);
                text-align: center;
                padding: 4px;
            }

            /* ========== Interrogator Gate States ========== */
            .gate-status {
                display: flex;
                align-items: center;
                gap: 8px;
                font: var(--md-sys-typescale-label-medium);
                padding: 6px 14px;
                border-radius: var(--md-sys-shape-corner-full);
            }

            .gate-dot {
                width: 8px;
                height: 8px;
                border-radius: 50%;
                flex-shrink: 0;
            }

            .gate-gathering {
                color: var(--md-sys-color-on-surface-variant);
            }

            .gate-gathering .gate-dot {
                background: var(--md-sys-color-outline);
                animation: pulse 1.5s infinite;
            }

            .gate-clarifying {
                color: #f59e0b;
                background: rgba(245, 158, 11, 0.08);
            }

            .gate-clarifying .gate-dot {
                background: #f59e0b;
                animation: pulse 1.2s infinite;
            }

            .gate-error {
                color: var(--md-sys-color-error);
                background: rgba(239, 68, 68, 0.08);
            }

            .gate-error .gate-dot {
                background: var(--md-sys-color-error);
            }

            .btn-generate.gate-open {
                background: #00FFA3;
                color: #fff;
                box-shadow: 0 0 16px rgba(34, 197, 94, 0.35), var(--md-sys-elevation-2);
            }

            .btn-generate.gate-open:hover:not(:disabled) {
                box-shadow: 0 0 24px rgba(34, 197, 94, 0.5), var(--md-sys-elevation-3);
            }
        `
    ];

    @state() private _sessionId = crypto.randomUUID();
    @state() private _currentPage = 1;
    @state() private _totalPages = 0;
    @state() private _isSubmitting = false;
    @state() private _errorMessage = '';
    @state() private _pdfLoaded = false;

    private _pdfDoc: pdfjsLib.PDFDocumentProxy | null = null;

    override connectedCallback(): void {
        super.connectedCallback();
        // Add welcome message if none exist
        if (onboardingMessages.value.length === 0) {
            addMessage({
                id: `sys-${String(Date.now())}`,
                role: 'assistant',
                content: "I'm analyzing your document. I'll extract project details and ask clarifying questions. You can see highlighted extraction zones in the PDF on the left.",
                timestamp: new Date()
            });
        }
    }

    override disconnectedCallback(): void {
        super.disconnectedCallback();
        this._pdfDoc?.destroy();
        this._pdfDoc = null;

        // Revoke object URL to prevent memory leak
        const url = uploadedPdfUrl.value;
        if (url && url.startsWith('blob:')) {
            URL.revokeObjectURL(url);
        }
    }

    override updated(): void {
        // Auto-scroll chat
        const list = this.shadowRoot?.querySelector('.message-list');
        if (list) {
            list.scrollTop = list.scrollHeight;
        }

        // Load PDF when URL changes
        const url = uploadedPdfUrl.value;
        if (url && !this._pdfLoaded) {
            this._pdfLoaded = true;
            void this._loadPdf(url);
        }
    }

    // ========== PDF Rendering ==========

    private async _loadPdf(url: string): Promise<void> {
        try {
            const loadingTask = pdfjsLib.getDocument(url);
            this._pdfDoc = await loadingTask.promise;
            this._totalPages = this._pdfDoc.numPages;
            this._currentPage = 1;
            await this._renderPage(1);
        } catch (err) {
            console.error('[FBInterrogatorWizard] PDF load failed:', err);
            this._pdfLoaded = false;
        }
    }

    private async _renderPage(pageNum: number): Promise<void> {
        if (!this._pdfDoc) return;
        try {
            const page = await this._pdfDoc.getPage(pageNum);
            const viewport = page.getViewport({ scale: 1.2 });

            // Wait for canvas to be in the DOM
            await this.updateComplete;
            const canvas = this.shadowRoot?.querySelector<HTMLCanvasElement>('.pdf-canvas');
            if (!canvas) return;

            const ctx = canvas.getContext('2d');
            if (!ctx) return;

            canvas.width = viewport.width;
            canvas.height = viewport.height;

            await page.render({ canvas, canvasContext: ctx, viewport }).promise;
            this.requestUpdate(); // trigger SVG overlay re-render
        } catch (err) {
            console.error('[FBInterrogatorWizard] Page render failed:', err);
        }
    }

    private _handlePrevPage(): void {
        if (this._currentPage > 1) {
            this._currentPage--;
            void this._renderPage(this._currentPage);
        }
    }

    private _handleNextPage(): void {
        if (this._currentPage < this._totalPages) {
            this._currentPage++;
            void this._renderPage(this._currentPage);
        }
    }

    /** Navigate to a specific page. */
    public goToPage(page: number): void {
        if (page >= 1 && page <= this._totalPages && page !== this._currentPage) {
            this._currentPage = page;
            void this._renderPage(page);
        }
    }

    // ========== Chat & API ==========

    private async _handleSend(e: CustomEvent<{ content: string }>): Promise<void> {
        const content = e.detail.content.trim();
        if (!content || isProcessing.value) return;

        isProcessing.value = true;

        addMessage({
            id: `msg-${String(Date.now())}-user`,
            role: 'user',
            content,
            timestamp: new Date()
        });

        try {
            const response = await api.onboard.process({
                session_id: this._sessionId,
                message: content,
                current_state: onboardingValues.value
            });

            this._processAIResponse(response);
        } catch (err) {
            console.error('[FBInterrogatorWizard] Send failed:', err);
            addMessage({
                id: `sys-${String(Date.now())}-err`,
                role: 'system',
                content: 'Sorry, something went wrong. Please try again.',
                timestamp: new Date()
            });
        } finally {
            isProcessing.value = false;
        }
    }

    private async _handleFileDrop(e: CustomEvent<{ files: File[] }>): Promise<void> {
        const file = e.detail.files[0];
        if (!file || isProcessing.value) return;

        isProcessing.value = true;
        const dropzone = this.shadowRoot?.querySelector<FBOnboardingDropzone>('fb-onboarding-dropzone');
        dropzone?.setUploading(file.name);

        addMessage({
            id: `sys-${String(Date.now())}`,
            role: 'system',
            content: `Analyzing ${file.name}...`,
            timestamp: new Date()
        });

        try {
            const formData = new FormData();
            formData.append('file', file);
            formData.append('session_id', this._sessionId);
            formData.append('message', '');
            formData.append('current_state', JSON.stringify(onboardingValues.value));

            dropzone?.setProgress(30);

            const token = clerkService.getToken();
            const headers: Record<string, string> = {};
            if (token) headers['Authorization'] = `Bearer ${token}`;

            const rawResponse = await fetch('/api/v1/agent/onboard', {
                method: 'POST',
                headers,
                body: formData
            });

            dropzone?.setProgress(70);

            if (rawResponse.status === 401) {
                window.dispatchEvent(new CustomEvent('fb-unauthorized'));
                throw new Error('Session expired. Please log in again.');
            }

            if (!rawResponse.ok) {
                const errorText = await rawResponse.text();
                throw new Error(errorText || `Upload failed (${String(rawResponse.status)})`);
            }

            const response = await rawResponse.json() as OnboardProcessResponse;
            dropzone?.setProgress(100);
            dropzone?.setComplete();

            this._processAIResponse(response);
        } catch (err) {
            console.error('[FBInterrogatorWizard] File upload failed:', err);
            const message = err instanceof Error ? err.message : 'Upload failed';
            dropzone?.setError(message);
            addMessage({
                id: `sys-${String(Date.now())}-err`,
                role: 'system',
                content: 'Sorry, I had trouble reading that file. Please try again.',
                timestamp: new Date()
            });
        } finally {
            isProcessing.value = false;
        }
    }

    private _processAIResponse(response: OnboardProcessResponse): void {
        // Apply extractions
        if (response.extracted_values) {
            applyAIExtraction(
                response.extracted_values,
                response.confidence_scores ?? {}
            );

            // Generate highlight boxes from extracted values for demonstration
            // In production, the API would return document_regions
            const boxes: BoundingBox[] = [];
            let idx = 0;
            for (const [field, value] of Object.entries(response.extracted_values)) {
                if (value !== undefined && value !== null) {
                    boxes.push({
                        page: 1,
                        x: 0.05,
                        y: 0.1 + idx * 0.1,
                        width: 0.5,
                        height: 0.06,
                        label: field.replace(/_/g, ' '),
                        field,
                        confidence: response.confidence_scores?.[field] ?? 0.5,
                    });
                    idx++;
                }
            }
            if (boxes.length > 0) {
                updateHighlights(boxes);
            }
        }

        // Update long-lead items
        if (response.long_lead_items && response.long_lead_items.length > 0) {
            setExtractedProcurement(response.long_lead_items);
        }

        // Track ready state
        if (response.ready_to_create) {
            setReadyToCreate(true);
        }

        const extractionCard = response.extracted_values
            ? {
                fields: Object.entries(response.extracted_values).map(([label, value]) => ({
                    label: label.replace(/_/g, ' '),
                    value: value as string | number,
                })),
                longLeadItems: response.long_lead_items ?? [],
            }
            : undefined;

        // Add reply message
        const msg: OnboardingMessage = {
            id: `msg-${String(Date.now())}-assistant`,
            role: 'assistant',
            content: response.reply,
            timestamp: new Date(),
        };
        if (extractionCard) {
            msg.extractionCard = extractionCard;
        }
        addMessage(msg);

        // Update interrogator gate status
        if (response.status) {
            setInterrogatorStatus(response.status);
        } else if (response.ready_to_create) {
            // Backward compat: map boolean to status
            setInterrogatorStatus('satisfied');
        } else if (response.clarifying_question) {
            setInterrogatorStatus('clarifying');
        }
    }

    private async _handleGenerateSchedule(): Promise<void> {
        if (this._isSubmitting || isProcessing.value) return;
        if (interrogatorStatus.value !== 'satisfied') return;

        this._isSubmitting = true;
        this._errorMessage = '';

        const values = onboardingValues.value;

        try {
            const startDate = values.start_date ?? new Date().toISOString().split('T')[0] ?? '';
            const projectData: CreateProjectRequest = {
                name: values.name ?? '',
                address: values.address ?? '',
                square_footage: values.square_footage ?? 0,
                bedrooms: values.bedrooms ?? 0,
                bathrooms: values.bathrooms ?? 0,
                start_date: startDate,
            };

            if (values.lot_size !== undefined) projectData.lot_size = values.lot_size;
            if (values.foundation_type !== undefined) projectData.foundation_type = values.foundation_type;
            if (values.stories !== undefined) projectData.stories = values.stories;
            if (values.topography !== undefined) projectData.topography = values.topography;
            if (values.soil_conditions !== undefined) projectData.soil_conditions = values.soil_conditions;

            const response = await api.projects.create(projectData);

            // Attempt to trigger schedule generation (may 404 if backend not implemented)
            try {
                await api.schedule.generate(response.id);
            } catch {
                // Schedule generation endpoint not available yet — non-fatal
                console.warn('[FBInterrogatorWizard] Schedule generation endpoint not available, skipping');
            }

            this.emit('project-created', { projectId: response.id });
        } catch (err) {
            console.error('[FBInterrogatorWizard] Project creation failed:', err);
            this._errorMessage = 'Failed to create project. Please try again.';
        } finally {
            this._isSubmitting = false;
        }
    }

    // ========== Rendering Helpers ==========

    private _getConfidenceClass(score: number): string {
        if (score >= 0.9) return 'confidence-high';
        if (score >= 0.7) return 'confidence-medium';
        return 'confidence-low';
    }

    private _getConfidenceLabel(score: number): string {
        if (score >= 0.9) return `${Math.round(score * 100)}%`;
        if (score >= 0.7) return `${Math.round(score * 100)}% Verify`;
        return `${Math.round(score * 100)}% Low`;
    }

    private _renderPdfPanel(): TemplateResult {
        const url = uploadedPdfUrl.value;
        const highlights = pdfHighlights.value;

        if (!url) {
            return html`
                <div class="pdf-panel">
                    <div class="pdf-header">
                        <span class="pdf-header-title">
                            <svg viewBox="0 0 24 24"><path d="M14 2H6c-1.1 0-1.99.9-1.99 2L4 20c0 1.1.89 2 1.99 2H18c1.1 0 2-.9 2-2V8l-6-6zm2 16H8v-2h8v2zm0-4H8v-2h8v2zm-3-5V3.5L18.5 9H13z"/></svg>
                            Document Viewer
                        </span>
                    </div>
                    <div class="pdf-empty">
                        <svg viewBox="0 0 24 24"><path d="M14 2H6c-1.1 0-1.99.9-1.99 2L4 20c0 1.1.89 2 1.99 2H18c1.1 0 2-.9 2-2V8l-6-6zm4 18H6V4h7v5h5v11z"/></svg>
                        <p>Upload a PDF document to see it here with AI-highlighted extraction zones</p>
                    </div>
                </div>
            `;
        }

        // Filter highlights for current page
        const pageHighlights = highlights.filter(h => h.page === this._currentPage);

        return html`
            <div class="pdf-panel">
                <div class="pdf-header">
                    <span class="pdf-header-title">
                        <svg viewBox="0 0 24 24"><path d="M14 2H6c-1.1 0-1.99.9-1.99 2L4 20c0 1.1.89 2 1.99 2H18c1.1 0 2-.9 2-2V8l-6-6zm2 16H8v-2h8v2zm0-4H8v-2h8v2zm-3-5V3.5L18.5 9H13z"/></svg>
                        Document Viewer
                    </span>
                    ${this._totalPages > 0 ? html`
                        <div class="pdf-page-nav">
                            <button
                                @click=${(): void => { this._handlePrevPage(); }}
                                ?disabled=${this._currentPage <= 1}
                                aria-label="Previous page"
                            >
                                <svg viewBox="0 0 24 24" width="16" height="16" fill="currentColor"><path d="M15.41 7.41L14 6l-6 6 6 6 1.41-1.41L10.83 12z"/></svg>
                            </button>
                            <span>${String(this._currentPage)} / ${String(this._totalPages)}</span>
                            <button
                                @click=${(): void => { this._handleNextPage(); }}
                                ?disabled=${this._currentPage >= this._totalPages}
                                aria-label="Next page"
                            >
                                <svg viewBox="0 0 24 24" width="16" height="16" fill="currentColor"><path d="M10 6L8.59 7.41 13.17 12l-4.58 4.59L10 18l6-6z"/></svg>
                            </button>
                        </div>
                    ` : nothing}
                </div>
                <div class="pdf-viewport">
                    <div class="pdf-canvas-wrapper">
                        <canvas class="pdf-canvas"></canvas>
                        ${pageHighlights.length > 0 ? html`
                            <svg class="pdf-overlay" viewBox="0 0 1 1" preserveAspectRatio="none">
                                ${pageHighlights.map(box => html`
                                    <rect
                                        x="${String(box.x)}"
                                        y="${String(box.y)}"
                                        width="${String(box.width)}"
                                        height="${String(box.height)}"
                                        fill="rgba(0, 255, 163, 0.12)"
                                        stroke="rgba(0, 255, 163, 0.7)"
                                        stroke-width="0.003"
                                        rx="0.004"
                                    />
                                `)}
                            </svg>
                        ` : nothing}
                    </div>
                </div>
            </div>
        `;
    }

    private _renderExtractionCard(card: OnboardingMessage['extractionCard']): TemplateResult | typeof nothing {
        if (!card || card.fields.length === 0) return nothing;

        const confidence = onboardingConfidence.value;

        return html`
            <div class="extraction-card">
                <div class="extraction-card-header">
                    <svg viewBox="0 0 24 24"><path d="M9 21c0 .5.4 1 1 1h4c.6 0 1-.5 1-1v-1H9v1zm3-19C8.1 2 5 5.1 5 9c0 2.4 1.2 4.5 3 5.7V17c0 .5.4 1 1 1h6c.6 0 1-.5 1-1v-2.3c1.8-1.3 3-3.4 3-5.7 0-3.9-3.1-7-7-7z"/></svg>
                    Extracted Data
                </div>
                <div class="extraction-fields">
                    ${card.fields.map(field => {
            const confScore = confidence[field.label.replace(/ /g, '_')] ?? 0;
            return html`
                            <div class="extraction-field">
                                <span class="extraction-field-label">${field.label}</span>
                                <span class="extraction-field-value">
                                    ${String(field.value)}
                                    ${confScore > 0 ? html`
                                        <span class="confidence-badge ${this._getConfidenceClass(confScore)}">
                                            ${this._getConfidenceLabel(confScore)}
                                        </span>
                                    ` : nothing}
                                </span>
                            </div>
                        `;
        })}
                </div>
            </div>
        `;
    }

    private _renderMessage(msg: OnboardingMessage): TemplateResult {
        return html`
            <div class="message ${msg.role}">${msg.content}</div>
            ${msg.extractionCard ? this._renderExtractionCard(msg.extractionCard) : nothing}
        `;
    }

    private _renderChatPanel(): TemplateResult {
        const messages = onboardingMessages.value;
        const processing = isProcessing.value;

        return html`
            <div class="chat-panel">
                <div class="chat-header">
                    <span class="chat-header-title">
                        <svg viewBox="0 0 24 24"><path d="M20 2H4c-1.1 0-1.99.9-1.99 2L2 22l4-4h14c1.1 0 2-.9 2-2V4c0-1.1-.9-2-2-2zm-2 12H6v-2h12v2zm0-3H6V9h12v2zm0-3H6V6h12v2z"/></svg>
                        AI Interrogator
                    </span>
                </div>
                <div class="message-list">
                    ${messages.map(msg => this._renderMessage(msg))}
                    ${processing ? html`
                        <div class="processing-indicator">Analyzing...</div>
                    ` : nothing}
                </div>
                <div class="chat-input-area">
                    <div class="dropzone-compact">
                        <fb-onboarding-dropzone
                            @file-dropped=${(e: CustomEvent<{ files: File[] }>): void => { void this._handleFileDrop(e); }}
                            ?disabled=${processing}
                        ></fb-onboarding-dropzone>
                    </div>
                    <fb-input-bar
                        @send=${(e: CustomEvent<{ content: string }>): void => { void this._handleSend(e); }}
                        ?disabled=${processing}
                    ></fb-input-bar>
                </div>
            </div>
        `;
    }

    private _renderGateStatus(status: InterrogatorStatus): TemplateResult | typeof nothing {
        switch (status) {
            case 'gathering':
                return html`
                    <span class="gate-status gate-gathering">
                        <span class="gate-dot"></span>
                        Collecting information...
                    </span>
                `;
            case 'clarifying':
                return html`
                    <span class="gate-status gate-clarifying">
                        <span class="gate-dot"></span>
                        AI has follow-up questions
                    </span>
                `;
            case 'error':
                return html`
                    <span class="gate-status gate-error">
                        <span class="gate-dot"></span>
                        Unable to validate — try again
                    </span>
                `;
            default:
                return nothing;
        }
    }

    private _renderActionBar(): TemplateResult {
        const values = onboardingValues.value;
        const gateStatus = interrogatorStatus.value;
        const procurement = extractedProcurement.value;
        const maxLeadWeeks = procurement.length > 0
            ? Math.max(...procurement.map(p => p.estimated_lead_weeks))
            : 0;

        const summaryParts: string[] = [];
        if (values.name) summaryParts.push(values.name);
        if (values.square_footage) summaryParts.push(`${String(values.square_footage)} sq ft`);
        if (values.foundation_type) summaryParts.push(`${values.foundation_type} foundation`);
        if (maxLeadWeeks > 0) summaryParts.push(`${String(maxLeadWeeks)}-week longest lead`);

        const isGateOpen = gateStatus === 'satisfied';

        return html`
            <div class="action-bar">
                <span class="action-summary">
                    ${summaryParts.length > 0 ? summaryParts.join(' · ') : 'Waiting for AI extraction...'}
                </span>
                ${!isGateOpen
                ? this._renderGateStatus(gateStatus)
                : html`
                        <button
                            class="btn-generate gate-open"
                            @click=${(): void => { void this._handleGenerateSchedule(); }}
                            ?disabled=${this._isSubmitting || isProcessing.value}
                        >
                            ${this._isSubmitting ? 'Creating...' : html`
                                <svg viewBox="0 0 24 24"><path d="M19 3H5c-1.1 0-2 .9-2 2v14c0 1.1.9 2 2 2h14c1.1 0 2-.9 2-2V5c0-1.1-.9-2-2-2zm-5 14H7v-2h7v2zm3-4H7v-2h10v2zm0-4H7V7h10v2z"/></svg>
                                Generate Schedule
                            `}
                        </button>
                    `
            }
            </div>
            ${this._errorMessage ? html`
                <div class="error-toast">${this._errorMessage}</div>
            ` : nothing}
        `;
    }

    override render(): TemplateResult {
        return html`
            <div class="split-screen">
                ${this._renderPdfPanel()}
                ${this._renderChatPanel()}
            </div>
            ${this._renderActionBar()}
        `;
    }
}

declare global {
    interface HTMLElementTagNameMap {
        'fb-interrogator-wizard': FBInterrogatorWizard;
    }
}
