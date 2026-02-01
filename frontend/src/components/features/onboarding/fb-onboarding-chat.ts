/**
 * FBOnboardingChat - Left Panel Chat Interface for Project Onboarding
 * See STEP_74_SPLIT_SCREEN_WIZARD.md Task 2
 *
 * Conversational interface that:
 * - Displays initial greeting from "The Interrogator" agent
 * - Reuses existing fb-message-list and fb-input-bar components
 * - Includes drag-and-drop zone for blueprints/documents
 * - Emits extracted data events for real-time form population
 */
import { html, css, TemplateResult } from 'lit';
import { customElement, state } from 'lit/decorators.js';
import { SignalWatcher } from '@lit-labs/preact-signals';
import { FBElement } from '../../base/FBElement';
import { api } from '../../../services/api';
import type { ChatMessage } from '../../../store/types';
import type { CreateProjectRequest } from '../../../services/api';
import {
    onboardingMessages,
    onboardingValues,
    isProcessing,
    addMessage,
    applyAIExtraction,
    type OnboardingMessage
} from '../../../store/onboarding-store';

import '../../chat/fb-message-list';
import '../../chat/fb-input-bar';
import './fb-onboarding-dropzone';

/**
 * Formats an ISO timestamp to a display time string.
 */
function formatDisplayTime(isoString: string): string {
    try {
        return new Date(isoString).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
    } catch {
        return '';
    }
}

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

            fb-message-list {
                flex: 1;
                min-height: 0;
            }

            .dropzone-wrapper {
                flex-shrink: 0;
                padding: var(--fb-spacing-md);
                border-top: 1px solid var(--fb-border-light);
            }

            .greeting {
                padding: var(--fb-spacing-lg);
                background: var(--fb-bg-tertiary);
                border-radius: var(--fb-radius-lg);
                margin: var(--fb-spacing-lg);
                color: var(--fb-text-primary);
                font-size: var(--fb-text-sm);
                line-height: 1.6;
            }

            .greeting-icon {
                font-size: 24px;
                margin-bottom: var(--fb-spacing-sm);
            }
        `
    ];

    @state() private _showGreeting = true;
    @state() private _sessionId = crypto.randomUUID();

    override connectedCallback(): void {
        super.connectedCallback();
        // Add initial greeting to store
        if (onboardingMessages.value.length === 0) {
            addMessage({
                id: `sys-${String(Date.now())}`,
                role: 'system',
                content: "Hi! I'm here to help set up your new project. You can describe it to me, or drag a blueprint/document here to get started.",
                timestamp: new Date()
            });
        }
    }

    private _addSystemMessage(content: string): void {
        addMessage({
            id: `sys-${String(Date.now())}`,
            role: 'system',
            content,
            timestamp: new Date()
        });
    }

    private async _handleSend(e: CustomEvent<{ content: string }>): Promise<void> {
        const content = e.detail.content.trim();
        if (!content) return;

        this._showGreeting = false;
        isProcessing.value = true;

        // Add user message to store
        addMessage({
            id: `msg-${String(Date.now())}-user`,
            role: 'user',
            content,
            timestamp: new Date()
        });

        try {
            // Call Interrogator Agent API (Step 75)
            const response = await api.onboard.process({
                session_id: this._sessionId,
                message: content,
                current_state: onboardingValues.value
            });

            // Apply AI extractions to store (triggers form update)
            if (response.extracted_values) {
                applyAIExtraction(
                    response.extracted_values,
                    response.confidence_scores || {}
                );
            }

            // Add assistant message
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

    private async _handleFileDrop(e: CustomEvent<{ files: File[] }>): Promise<void> {
        this._showGreeting = false;
        isProcessing.value = true;

        const files = e.detail.files;
        const fileName = files[0]?.name ?? 'file';

        this._addSystemMessage(`Analyzing ${fileName}...`);

        try {
            // TODO Step 77: Implement magic upload trigger with real API
            // For now, simulate extraction
            await new Promise(resolve => setTimeout(resolve, 2000));

            this._addSystemMessage(
                `I found a ${fileName}. Let me extract the project details for you.`
            );

            // Simulate extraction (Step 77 will connect to real backend)
            applyAIExtraction(
                {
                    name: 'Extracted Project Name',
                    square_footage: 2500,
                    bedrooms: 4,
                    bathrooms: 3,
                },
                {
                    name: 0.85,
                    square_footage: 0.92,
                    bedrooms: 0.88,
                    bathrooms: 0.75, // Low confidence - will show yellow border
                }
            );
        } catch (err) {
            console.error('[FBOnboardingChat] File drop failed:', err);
            this._addSystemMessage('Sorry, I had trouble reading that file. Please try again.');
        } finally {
            isProcessing.value = false;
        }
    }

    override render(): TemplateResult {
        return html`
            <div class="chat-container">
                ${this._showGreeting ? html`
                    <div class="greeting">
                        <div class="greeting-icon">👋</div>
                        <strong>Welcome to FutureBuild!</strong>
                        <p>I'm The Interrogator, your AI project assistant. I'll help you set up your construction project by asking a few questions or analyzing your documents.</p>
                    </div>
                ` : html`
                    <fb-message-list .messages=${this._messages}></fb-message-list>
                `}

                <div class="dropzone-wrapper">
                    <fb-onboarding-dropzone
                        @file-dropped=${this._handleFileDrop}
                        ?disabled=${this._isProcessing}
                    ></fb-onboarding-dropzone>
                </div>
            </div>

            <fb-input-bar
                placeholder="Describe your project..."
                .disabled=${this._isProcessing}
                @send=${this._handleSend}
            ></fb-input-bar>
        `;
    }
}

declare global {
    interface HTMLElementTagNameMap {
        'fb-onboarding-chat': FBOnboardingChat;
    }
}
