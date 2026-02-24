/**
 * FBOnboardingChat - Full-Width Chat Interface for Project Onboarding
 * See STEP_74_SPLIT_SCREEN_WIZARD.md, STEP_76_REALTIME_FORM_FILLING.md,
 * STEP_77_MAGIC_UPLOAD_TRIGGER.md
 *
 * Document-first, chat-only onboarding experience:
 * - Displays welcome message with document upload prompt
 * - Renders extraction cards with building specs and long-lead items
 * - Shows procurement warnings for items with long lead times
 * - Includes Create Project button when ready
 * - Uploads files via multipart to /agent/onboard
 */
import { html, css, TemplateResult, nothing } from 'lit';
import { customElement, state } from 'lit/decorators.js';
import { SignalWatcher } from '@lit-labs/preact-signals';
import { FBElement } from '../../base/FBElement';
import { api, type OnboardProcessResponse, type CreateProjectRequest } from '../../../services/api';
import {
    onboardingMessages,
    onboardingValues,
    isProcessing,
    isReadyToCreate,
    extractedProcurement,
    addMessage,
    applyAIExtraction,
    setExtractedProcurement,
    markDocumentUploaded,
    setUploadedPdfUrl,
    setReadyToCreate,
    type OnboardingMessage
} from '../../../store/onboarding-store';
import type { FBOnboardingDropzone } from './fb-onboarding-dropzone';

import { clerkService } from '../../../services/clerk';
import '../../chat/fb-input-bar';
import './fb-onboarding-dropzone';

@customElement('fb-onboarding-chat')
export class FBOnboardingChat extends SignalWatcher(FBElement) {
    static override styles = [
        FBElement.styles,
        css`
            :host {
                display: flex;
                flex-direction: column;
                height: 100%;
                overflow: hidden;
                background-color: var(--md-sys-color-background);
                color: var(--md-sys-color-on-background);
                font-family: var(--md-ref-typeface-plain);
            }

            .chat-container {
                flex: 1;
                display: flex;
                flex-direction: column;
                min-height: 0;
                position: relative;
            }

            /* Scrollbar styling */
            .message-list {
                flex: 1;
                overflow-y: auto;
                padding: var(--fb-spacing-lg) var(--fb-spacing-xl);
                display: flex;
                flex-direction: column;
                gap: var(--fb-spacing-md);
                scroll-behavior: smooth;
            }

            .message-list::-webkit-scrollbar {
                width: 8px;
            }
            .message-list::-webkit-scrollbar-track {
                background: transparent;
            }
            .message-list::-webkit-scrollbar-thumb {
                background-color: var(--md-sys-color-outline-variant);
                border-radius: 4px;
            }

            /* Message Bubbles */
            .message {
                max-width: 70%;
                padding: var(--fb-spacing-md) var(--fb-spacing-lg);
                border-radius: var(--md-sys-shape-corner-large);
                font: var(--md-sys-typescale-body-large);
                position: relative;
                animation: fadeIn 0.3s ease-out;
            }

            @keyframes fadeIn {
                from { opacity: 0; transform: translateY(10px); }
                to { opacity: 1; transform: translateY(0); }
            }

            .message.user {
                align-self: flex-end;
                background-color: var(--md-sys-color-primary-container);
                color: var(--md-sys-color-on-primary-container);
                border-bottom-right-radius: var(--md-sys-shape-corner-extra-small);
                box-shadow: var(--md-sys-elevation-1);
            }

            .message.assistant {
                align-self: flex-start;
                background-color: var(--md-sys-color-surface-container-high);
                color: var(--md-sys-color-on-surface);
                border-bottom-left-radius: var(--md-sys-shape-corner-extra-small);
                box-shadow: var(--md-sys-elevation-1);
            }

            .message.system {
                align-self: center;
                background: transparent;
                color: var(--md-sys-color-outline);
                font: var(--md-sys-typescale-label-medium);
                text-align: center;
                max-width: 90%;
                padding: var(--fb-spacing-xs);
                border: 1px dashed var(--md-sys-color-outline-variant);
                margin: var(--fb-spacing-sm) 0;
            }

            .processing-indicator {
                align-self: flex-start;
                padding: var(--fb-spacing-sm) var(--fb-spacing-md);
                color: var(--md-sys-color-secondary);
                font: var(--md-sys-typescale-body-medium);
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
                background-color: var(--md-sys-color-secondary);
                animation: pulse 1.5s infinite;
            }

            @keyframes pulse {
                0% { transform: scale(0.8); opacity: 0.5; }
                50% { transform: scale(1.2); opacity: 1; }
                100% { transform: scale(0.8); opacity: 0.5; }
            }

            /* Hero Dropzone Styling */
            .hero-dropzone {
                width: 100%;
                max-width: 600px;
                margin-top: var(--fb-spacing-xl);
                animation: slideUp 0.6s ease-out;
                --dropzone-min-height: 200px;
            }

            /* Compact Dropzone Styling */
            .dropzone-wrapper {
                flex-shrink: 0;
                padding: var(--fb-spacing-md) var(--fb-spacing-xl);
                background-color: var(--md-sys-color-surface);
                border-top: 1px solid var(--md-sys-color-outline-variant);
            }

            /* Greeting Hero */
            .greeting {
                flex: 1;
                display: flex;
                flex-direction: column;
                align-items: center;
                justify-content: center;
                padding: var(--fb-spacing-2xl);
                text-align: center;
                color: var(--md-sys-color-on-background);
                animation: slideUp 0.5s ease-out;
                overflow-y: auto; /* Allow scrolling if hero content is tall */
            }

            @keyframes slideUp {
                from { opacity: 0; transform: translateY(20px); }
                to { opacity: 1; transform: translateY(0); }
            }

            .greeting-icon {
                font: var(--md-sys-typescale-headline-medium);
                margin-bottom: var(--fb-spacing-md);
                background: var(--fb-dawn-gradient);
                -webkit-background-clip: text;
                -webkit-text-fill-color: transparent;
                display: inline-block;
            }

            .greeting strong {
                font: var(--md-sys-typescale-title-large);
                margin-bottom: var(--fb-spacing-sm);
                display: block;
            }

            .greeting p {
                font: var(--md-sys-typescale-body-medium);
                color: var(--md-sys-color-on-surface-variant);
                max-width: 480px;
                line-height: 1.6;
                margin-bottom: var(--fb-spacing-lg);
            }

            fb-input-bar {
                flex-shrink: 0;
            }
            
            /* ... (rest of styles preserved) ... */

            /* Extraction card styling */
            .extraction-card {
                background: var(--md-sys-color-surface-container);
                border: 1px solid var(--md-sys-color-outline-variant);
                border-radius: var(--md-sys-shape-corner-medium);
                padding: var(--fb-spacing-md);
                margin: var(--fb-spacing-xs) 0;
                width: 100%;
                box-sizing: border-box;
            }

            .extraction-header {
                font: var(--md-sys-typescale-title-small);
                color: var(--md-sys-color-primary);
                margin-bottom: var(--fb-spacing-sm);
                display: flex;
                align-items: center;
                gap: var(--fb-spacing-sm);
            }

            .extraction-header svg {
                width: 18px;
                height: 18px;
                fill: currentColor;
            }

            .extraction-fields {
                display: grid;
                grid-template-columns: repeat(auto-fit, minmax(140px, 1fr));
                gap: var(--fb-spacing-sm);
            }

            .extraction-field {
                display: flex;
                flex-direction: column;
            }

            .extraction-field-label {
                color: var(--md-sys-color-on-surface-variant);
                font: var(--md-sys-typescale-label-small);
                text-transform: uppercase;
                letter-spacing: 0.5px;
            }

            .extraction-field-value {
                color: var(--md-sys-color-on-surface);
                font: var(--md-sys-typescale-body-medium);
            }

            /* Procurement warning styling */
            .procurement-warning {
                background-color: var(--md-sys-color-error-container);
                color: var(--md-sys-color-on-error-container);
                border: none;
                border-radius: var(--md-sys-shape-corner-small);
                padding: var(--fb-spacing-md);
                margin-top: var(--fb-spacing-sm);
            }

            .procurement-warning-header {
                font: var(--md-sys-typescale-label-large);
                color: var(--md-sys-color-on-error-container);
                margin-bottom: var(--fb-spacing-xs);
                display: flex;
                align-items: center;
                gap: var(--fb-spacing-sm);
            }

            .procurement-warning-header svg {
                width: 18px;
                height: 18px;
                fill: currentColor;
            }

            .lead-item {
                font: var(--md-sys-typescale-body-small);
                color: var(--md-sys-color-on-error-container);
                opacity: 0.9;
            }

            /* Create button styling */
            .create-section {
                padding: var(--fb-spacing-md) var(--fb-spacing-xl);
                border-top: 1px solid var(--md-sys-color-outline-variant);
                background: var(--md-sys-color-surface-container-low);
                display: flex;
                flex-direction: column;
                gap: var(--fb-spacing-sm);
            }

            .create-summary {
                font: var(--md-sys-typescale-body-small);
                color: var(--md-sys-color-on-surface-variant);
                text-align: center;
            }

            .btn-create {
                display: flex;
                align-items: center;
                justify-content: center;
                gap: var(--fb-spacing-sm);
                width: 100%;
                padding: 12px 24px;
                background-color: var(--md-sys-color-primary);
                color: var(--md-sys-color-on-primary);
                font: var(--md-sys-typescale-label-large);
                border: none;
                border-radius: var(--md-sys-shape-corner-full);
                cursor: pointer;
                transition: all var(--fb-transition-base);
                box-shadow: var(--md-sys-elevation-2);
            }

            .btn-create:hover:not(:disabled) {
                background-color: var(--fb-primary-hover); /* Fallback or calculate lighter */
                box-shadow: var(--md-sys-elevation-3);
                transform: translateY(-1px);
            }

            .btn-create:disabled {
                background-color: var(--md-sys-color-surface-variant);
                color: var(--md-sys-color-on-surface-variant);
                box-shadow: none;
                cursor: not-allowed;
            }

            .btn-create svg {
                width: 20px;
                height: 20px;
                fill: currentColor;
            }

            .error-message {
                color: var(--md-sys-color-error);
                font: var(--md-sys-typescale-body-small);
                text-align: center;
            }
        `
    ];

    @state() private _showGreeting = true;
    @state() private _sessionId = crypto.randomUUID();
    @state() private _isSubmitting = false;
    @state() private _errorMessage = '';

    override connectedCallback(): void {
        super.connectedCallback();
        if (onboardingMessages.value.length === 0) {
            addMessage({
                id: `sys-${String(Date.now())}`,
                role: 'assistant',
                content: "Welcome! Drop your building plans or permit set here, and I'll extract everything I can to get your schedule started. You can also describe your project if you don't have documents handy.",
                timestamp: new Date()
            });
        }
    }

    private _getDropzone(): FBOnboardingDropzone | null {
        return this.shadowRoot?.querySelector<FBOnboardingDropzone>('fb-onboarding-dropzone') ?? null;
    }

    private _addSystemMessage(content: string): void {
        addMessage({
            id: `sys-${String(Date.now())}`,
            role: 'system',
            content,
            timestamp: new Date()
        });
    }

    // Fix 3: Auto-scroll message list when new messages arrive
    override updated(): void {
        const list = this.shadowRoot?.querySelector('.message-list');
        if (list) {
            list.scrollTop = list.scrollHeight;
        }
    }

    private async _handleSend(e: CustomEvent<{ content: string }>): Promise<void> {
        const content = e.detail.content.trim();
        // Fix 4: Guard against rapid sends while processing
        if (!content || isProcessing.value) return;

        this._showGreeting = false;
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

            if (response.extracted_values) {
                applyAIExtraction(
                    response.extracted_values,
                    response.confidence_scores ?? {}
                );
            }

            addMessage({
                id: `msg-${String(Date.now())}-assistant`,
                role: 'assistant',
                content: response.reply,
                timestamp: new Date()
            });

        } catch (err) {
            console.error('[FBOnboardingChat] Send failed:', err);
            this._addSystemMessage('Sorry, something went wrong. Please try again.');
        } finally {
            isProcessing.value = false;
        }
    }

    // Step 77: Magic Upload Trigger - real multipart upload to /agent/onboard
    private async _handleFileDrop(e: CustomEvent<{ files: File[] }>): Promise<void> {
        const file = e.detail.files[0];
        // Fix 4: Guard against uploads while already processing
        if (!file || isProcessing.value) return;

        this._showGreeting = false;
        isProcessing.value = true;

        const dropzone = this._getDropzone();
        dropzone?.setUploading(file.name);

        this._addSystemMessage(`Analyzing ${file.name}...`);

        try {
            // Build multipart form data
            const formData = new FormData();
            formData.append('file', file);
            formData.append('session_id', this._sessionId);
            formData.append('message', '');
            formData.append('current_state', JSON.stringify(onboardingValues.value));

            // Progress: uploading phase
            dropzone?.setProgress(30);

            // POST multipart to /agent/onboard
            // Phase 12: Use Clerk's cached token instead of legacy localStorage
            const token = clerkService.getToken();
            const headers: Record<string, string> = {};
            if (token) {
                headers['Authorization'] = `Bearer ${token}`;
            }

            const rawResponse = await fetch('/api/v1/agent/onboard', {
                method: 'POST',
                headers,
                body: formData
            });

            // Progress: analyzing phase
            dropzone?.setProgress(70);

            // Fix 6: Handle 401 explicitly (raw fetch bypasses centralized http.ts 401 handler)
            if (rawResponse.status === 401) {
                window.dispatchEvent(new CustomEvent('fb-unauthorized'));
                throw new Error('Session expired. Please log in again.');
            }

            if (!rawResponse.ok) {
                const errorText = await rawResponse.text();
                throw new Error(errorText || `Upload failed (${String(rawResponse.status)})`);
            }

            const response = await rawResponse.json() as OnboardProcessResponse;

            // Progress: done
            dropzone?.setProgress(100);

            // Apply extractions to form
            if (response.extracted_values) {
                applyAIExtraction(
                    response.extracted_values,
                    response.confidence_scores ?? {}
                );
            }

            // Store long-lead items
            if (response.long_lead_items && response.long_lead_items.length > 0) {
                setExtractedProcurement(response.long_lead_items);
            }

            // Mark document as uploaded
            markDocumentUploaded();

            // Add assistant response
            addMessage({
                id: `msg-${String(Date.now())}-assistant`,
                role: 'assistant',
                content: response.reply,
                timestamp: new Date()
            });

            dropzone?.setComplete();

            // If the uploaded file is a PDF, set the URL for the split-screen viewer
            if (file.type === 'application/pdf') {
                const pdfObjectUrl = URL.createObjectURL(file);
                setUploadedPdfUrl(pdfObjectUrl);
            }

            // Track ready state
            if (response.ready_to_create) {
                setReadyToCreate(true);
            }

        } catch (err) {
            console.error('[FBOnboardingChat] File upload failed:', err);
            const message = err instanceof Error ? err.message : 'Upload failed';
            dropzone?.setError(message);
            this._addSystemMessage('Sorry, I had trouble reading that file. Please try again.');
        } finally {
            isProcessing.value = false;
        }
    }

    private async _handleSubmit(): Promise<void> {
        if (this._isSubmitting || isProcessing.value || !isReadyToCreate.value) return;

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
            this.emit('project-created', { projectId: response.id });
        } catch (err) {
            console.error('[FBOnboardingChat] Project creation failed:', err);
            this._errorMessage = 'Failed to create project. Please try again.';
        } finally {
            this._isSubmitting = false;
        }
    }

    private _renderProcurementWarnings(): TemplateResult | typeof nothing {
        const items = extractedProcurement.value;
        if (items.length === 0) return nothing;

        return html`
            <div class="procurement-warning">
                <div class="procurement-warning-header">
                    <svg viewBox="0 0 24 24" fill="currentColor">
                        <path d="M1 21h22L12 2 1 21zm12-3h-2v-2h2v2zm0-4h-2v-4h2v4z"/>
                    </svg>
                    Long-lead items detected
                </div>
                ${items.map(item => html`
                    <div class="lead-item">
                        ${item.name}${item.brand ? ` (${item.brand})` : ''} -
                        <span class="lead-item-weeks">~${String(item.estimated_lead_weeks)} weeks</span>
                    </div>
                `)}
            </div>
        `;
    }

    private _renderCreateSection(): TemplateResult | typeof nothing {
        if (!isReadyToCreate.value) return nothing;

        const values = onboardingValues.value;
        const procurement = extractedProcurement.value;
        const maxLeadWeeks = procurement.length > 0
            ? Math.max(...procurement.map(p => p.estimated_lead_weeks))
            : 0;

        return html`
            <div class="create-section">
                <div class="create-summary">
                    ${values.square_footage ? `${String(values.square_footage)} sq ft` : ''}
                    ${values.foundation_type ? ` | ${values.foundation_type} foundation` : ''}
                    ${maxLeadWeeks > 0 ? ` | ${String(maxLeadWeeks)}-week longest lead time` : ''}
                </div>
                <button
                    class="btn-create"
                    @click=${(): void => { void this._handleSubmit(); }}
                    ?disabled=${this._isSubmitting || isProcessing.value}
                >
                    ${this._isSubmitting ? html`
                        Creating...
                    ` : html`
                        <svg viewBox="0 0 24 24" fill="currentColor">
                            <path d="M19 13h-6v6h-2v-6H5v-2h6V5h2v6h6v2z"/>
                        </svg>
                        Create Project
                    `}
                </button>
                ${this._errorMessage ? html`
                    <div class="error-message">${this._errorMessage}</div>
                ` : nothing}
            </div>
        `;
    }

    private _renderMessage(msg: OnboardingMessage): TemplateResult {
        return html`<div class="message ${msg.role}">${msg.content}</div>`;
    }

    private _renderMessages(): TemplateResult {
        const messages = onboardingMessages.value;
        return html`
            <div class="message-list">
                ${messages.map(msg => this._renderMessage(msg))}
                ${this._renderProcurementWarnings()}
                ${isProcessing.value ? html`
                    <div class="processing-indicator">Thinking...</div>
                ` : nothing}
            </div>
        `;
    }

    override render(): TemplateResult {
        const ready = isReadyToCreate.value;

        return html`
            <div class="chat-container">
                ${this._showGreeting ? html`
                    <div class="greeting">
                        <div class="greeting-icon">Welcome to FutureBuild</div>
                        <strong>I'm here to help set up your new project.</strong>
                        <p>Drop your building plans or permit set below, and I'll extract everything I need to generate your initial schedule. You can also describe your project if you don't have documents handy.</p>
                        
                        <div class="hero-dropzone">
                            <fb-onboarding-dropzone
                                @file-dropped=${(e: CustomEvent<{ files: File[] }>): void => { void this._handleFileDrop(e); }}
                                ?disabled=${isProcessing.value}
                            ></fb-onboarding-dropzone>
                        </div>
                    </div>
                ` : html`
                    ${this._renderMessages()}
                    
                    <div class="dropzone-wrapper">
                        <fb-onboarding-dropzone
                            @file-dropped=${(e: CustomEvent<{ files: File[] }>): void => { void this._handleFileDrop(e); }}
                            ?disabled=${isProcessing.value}
                        ></fb-onboarding-dropzone>
                    </div>
                `}
            </div>

            ${ready ? this._renderCreateSection() : html`
                <fb-input-bar
                    @send=${(e: CustomEvent<{ content: string }>): void => { void this._handleSend(e); }}
                    ?disabled=${isProcessing.value}
                ></fb-input-bar>
            `}
        `;
    }
}

declare global {
    interface HTMLElementTagNameMap {
        'fb-onboarding-chat': FBOnboardingChat;
    }
}
