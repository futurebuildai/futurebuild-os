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
            }

            .chat-container {
                flex: 1;
                display: flex;
                flex-direction: column;
                min-height: 0;
            }

            .message-list {
                flex: 1;
                overflow-y: auto;
                padding: var(--fb-spacing-md, 16px);
                display: flex;
                flex-direction: column;
                gap: var(--fb-spacing-sm, 8px);
            }

            .message {
                max-width: 85%;
                padding: var(--fb-spacing-sm, 8px) var(--fb-spacing-md, 16px);
                border-radius: var(--fb-radius-md, 8px);
                font-size: var(--fb-text-sm, 14px);
                line-height: 1.5;
                word-wrap: break-word;
            }

            .message.user {
                align-self: flex-end;
                background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
                color: white;
                border-bottom-right-radius: 2px;
            }

            .message.assistant {
                align-self: flex-start;
                background: var(--fb-bg-tertiary, #f5f5f5);
                color: var(--fb-text-primary, #1a1a1a);
                border-bottom-left-radius: 2px;
            }

            .message.system {
                align-self: center;
                background: rgba(102, 126, 234, 0.08);
                color: var(--fb-text-muted, #666);
                font-size: var(--fb-text-xs, 12px);
                text-align: center;
                max-width: 90%;
            }

            .processing-indicator {
                align-self: flex-start;
                padding: var(--fb-spacing-sm, 8px) var(--fb-spacing-md, 16px);
                color: var(--fb-text-muted, #666);
                font-size: var(--fb-text-sm, 14px);
                font-style: italic;
            }

            .dropzone-wrapper {
                flex-shrink: 0;
                padding: var(--fb-spacing-md, 16px);
                border-top: 1px solid var(--fb-border-light, #eee);
            }

            .greeting {
                padding: var(--fb-spacing-lg, 24px);
                background: var(--fb-bg-tertiary, #f5f5f5);
                border-radius: var(--fb-radius-lg, 12px);
                margin: var(--fb-spacing-lg, 24px);
                color: var(--fb-text-primary, #1a1a1a);
                font-size: var(--fb-text-sm, 14px);
                line-height: 1.6;
            }

            .greeting-icon {
                font-size: 24px;
                margin-bottom: var(--fb-spacing-sm, 8px);
            }

            fb-input-bar {
                flex-shrink: 0;
            }

            /* Extraction card styling */
            .extraction-card {
                background: var(--fb-bg-card, white);
                border: 1px solid var(--fb-border);
                border-radius: var(--fb-radius-md, 8px);
                padding: var(--fb-spacing-md, 16px);
                margin: var(--fb-spacing-sm, 8px) 0;
                max-width: 90%;
            }

            .extraction-header {
                font-weight: 600;
                font-size: var(--fb-text-sm, 14px);
                color: var(--fb-text-primary);
                margin-bottom: var(--fb-spacing-sm, 8px);
                display: flex;
                align-items: center;
                gap: var(--fb-spacing-xs, 4px);
            }

            .extraction-header svg {
                width: 16px;
                height: 16px;
                color: #667eea;
            }

            .extraction-fields {
                display: grid;
                grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
                gap: var(--fb-spacing-xs, 4px);
                font-size: var(--fb-text-sm, 14px);
            }

            .extraction-field {
                display: flex;
                flex-direction: column;
                padding: var(--fb-spacing-xs, 4px) 0;
            }

            .extraction-field-label {
                color: var(--fb-text-muted);
                font-size: 11px;
                text-transform: uppercase;
                letter-spacing: 0.5px;
            }

            .extraction-field-value {
                color: var(--fb-text-primary);
                font-weight: 500;
            }

            /* Procurement warning styling */
            .procurement-warning {
                background: rgba(251, 191, 36, 0.1);
                border: 1px solid #fbbf24;
                border-radius: var(--fb-radius-sm, 4px);
                padding: var(--fb-spacing-sm, 8px);
                margin-top: var(--fb-spacing-sm, 8px);
            }

            .procurement-warning-header {
                font-weight: 600;
                font-size: var(--fb-text-sm, 14px);
                color: #78350f;
                margin-bottom: var(--fb-spacing-xs, 4px);
                display: flex;
                align-items: center;
                gap: var(--fb-spacing-xs, 4px);
            }

            .procurement-warning-header svg {
                width: 16px;
                height: 16px;
            }

            .lead-item {
                font-size: var(--fb-text-sm, 14px);
                color: #78350f;
                padding: 2px 0;
            }

            .lead-item-weeks {
                font-weight: 600;
            }

            /* Create button styling */
            .create-section {
                padding: var(--fb-spacing-md, 16px);
                border-top: 1px solid var(--fb-border);
                background: var(--fb-bg-secondary);
            }

            .create-summary {
                font-size: var(--fb-text-sm, 14px);
                color: var(--fb-text-muted);
                margin-bottom: var(--fb-spacing-sm, 8px);
            }

            .btn-create {
                display: flex;
                align-items: center;
                justify-content: center;
                gap: var(--fb-spacing-xs, 4px);
                width: 100%;
                padding: var(--fb-spacing-md, 16px);
                background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
                color: white;
                font-size: var(--fb-text-md, 16px);
                font-weight: 600;
                border: none;
                border-radius: var(--fb-radius-md, 8px);
                cursor: pointer;
                transition: all 0.2s;
            }

            .btn-create:hover:not(:disabled) {
                transform: translateY(-1px);
                box-shadow: 0 4px 12px rgba(102, 126, 234, 0.4);
            }

            .btn-create:disabled {
                opacity: 0.5;
                cursor: not-allowed;
            }

            .btn-create svg {
                width: 20px;
                height: 20px;
            }

            .error-message {
                color: #dc2626;
                font-size: var(--fb-text-sm, 14px);
                margin-top: var(--fb-spacing-xs, 4px);
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
                    </div>
                ` : this._renderMessages()}

                <div class="dropzone-wrapper">
                    <fb-onboarding-dropzone
                        @file-dropped=${(e: CustomEvent<{ files: File[] }>): void => { void this._handleFileDrop(e); }}
                        ?disabled=${isProcessing.value}
                    ></fb-onboarding-dropzone>
                </div>
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
