/**
 * FBOnboardingChat - Left Panel Chat Interface for Project Onboarding
 * See STEP_74_SPLIT_SCREEN_WIZARD.md Task 2, STEP_76_REALTIME_FORM_FILLING.md,
 * STEP_77_MAGIC_UPLOAD_TRIGGER.md
 *
 * Conversational interface that:
 * - Displays initial greeting from "The Interrogator" agent
 * - Renders messages inline (separate from global chat store)
 * - Includes drag-and-drop zone for blueprints/documents
 * - Calls Interrogator Agent API for field extraction
 * - Step 77: Uploads files via multipart to /agent/onboard
 */
import { html, css, TemplateResult, nothing } from 'lit';
import { customElement, state } from 'lit/decorators.js';
import { SignalWatcher } from '@lit-labs/preact-signals';
import { FBElement } from '../../base/FBElement';
import { api, type OnboardProcessResponse } from '../../../services/api';
import {
    onboardingMessages,
    onboardingValues,
    isProcessing,
    addMessage,
    applyAIExtraction,
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
        `
    ];

    @state() private _showGreeting = true;
    @state() private _sessionId = crypto.randomUUID();

    override connectedCallback(): void {
        super.connectedCallback();
        if (onboardingMessages.value.length === 0) {
            addMessage({
                id: `sys-${String(Date.now())}`,
                role: 'system',
                content: "Hi! I'm here to help set up your new project. You can describe it to me, or drag a blueprint/document here to get started.",
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

    private _renderMessage(msg: OnboardingMessage): TemplateResult {
        return html`<div class="message ${msg.role}">${msg.content}</div>`;
    }

    private _renderMessages(): TemplateResult {
        const messages = onboardingMessages.value;
        return html`
            <div class="message-list">
                ${messages.map(msg => this._renderMessage(msg))}
                ${isProcessing.value ? html`
                    <div class="processing-indicator">Thinking...</div>
                ` : nothing}
            </div>
        `;
    }

    override render(): TemplateResult {
        return html`
            <div class="chat-container">
                ${this._showGreeting ? html`
                    <div class="greeting">
                        <div class="greeting-icon">Welcome to FutureBuild</div>
                        <strong>I'm The Interrogator, your AI project assistant.</strong>
                        <p>I'll help you set up your construction project by asking a few questions or analyzing your documents.</p>
                    </div>
                ` : this._renderMessages()}

                <div class="dropzone-wrapper">
                    <fb-onboarding-dropzone
                        @file-dropped=${(e: CustomEvent<{ files: File[] }>): void => { void this._handleFileDrop(e); }}
                        ?disabled=${isProcessing.value}
                    ></fb-onboarding-dropzone>
                </div>
            </div>

            <fb-input-bar
                @send=${(e: CustomEvent<{ content: string }>): void => { void this._handleSend(e); }}
                ?disabled=${isProcessing.value}
            ></fb-input-bar>
        `;
    }
}

declare global {
    interface HTMLElementTagNameMap {
        'fb-onboarding-chat': FBOnboardingChat;
    }
}
