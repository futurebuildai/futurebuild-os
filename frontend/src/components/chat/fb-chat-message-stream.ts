/**
 * fb-chat-message-stream — Renders a streaming assistant message.
 *
 * Accumulates text chunks as they arrive via SSE and renders them
 * with a typing indicator. Transitions to a static message when done.
 */
import { html, css, nothing } from 'lit';
import { customElement, property, state } from 'lit/decorators.js';
import { FBElement } from '../base/FBElement';
import type { ChatStreamChunk } from '@/services/api';

@customElement('fb-chat-message-stream')
export class FBChatMessageStream extends FBElement {
    static override styles = [
        FBElement.styles,
        css`
            :host {
                display: block;
            }

            .message {
                padding: 12px 16px;
                border-radius: 12px;
                background: var(--surface-2, #1a2332);
                color: var(--text-primary, #e2e8f0);
                line-height: 1.6;
                white-space: pre-wrap;
                word-break: break-word;
            }

            .cursor {
                display: inline-block;
                width: 2px;
                height: 1em;
                background: var(--accent, #2d5a3d);
                animation: blink 0.8s step-end infinite;
                vertical-align: text-bottom;
                margin-left: 1px;
            }

            @keyframes blink {
                50% { opacity: 0; }
            }

            .tool-indicator {
                display: flex;
                align-items: center;
                gap: 8px;
                padding: 8px 12px;
                margin: 4px 0;
                border-radius: 8px;
                background: var(--surface-3, #243447);
                color: var(--text-secondary, #94a3b8);
                font-size: 0.85rem;
            }

            .tool-indicator .spinner {
                width: 14px;
                height: 14px;
                border: 2px solid var(--text-secondary, #94a3b8);
                border-top-color: var(--accent, #2d5a3d);
                border-radius: 50%;
                animation: spin 0.6s linear infinite;
            }

            @keyframes spin {
                to { transform: rotate(360deg); }
            }
        `,
    ];

    /** Whether the stream is still active. */
    @property({ type: Boolean }) streaming = true;

    @state() private text = '';
    @state() private activeToolName: string | null = null;
    private _disposed = false;

    override disconnectedCallback(): void {
        super.disconnectedCallback();
        this._disposed = true;
    }

    /** Append a stream chunk to the rendered content. */
    appendChunk(chunk: ChatStreamChunk): void {
        // Guard: ignore chunks after stream completed or component disposed
        if (this._disposed || !this.streaming) return;

        if (chunk.text) {
            this.text += chunk.text;
        }
        if (chunk.tool_use) {
            this.activeToolName = chunk.tool_use.name;
        }
        if (chunk.tool_result) {
            this.activeToolName = null;
        }
        if (chunk.done) {
            this.streaming = false;
            this.activeToolName = null;
        }
    }

    /** Get the full accumulated text. */
    get fullText(): string {
        return this.text;
    }

    override render() {
        return html`
            <div class="message" role="log" aria-live="polite">
                ${this.text}${this.streaming ? html`<span class="cursor"></span>` : nothing}
            </div>
            ${this.activeToolName
                ? html`
                    <div class="tool-indicator">
                        <div class="spinner"></div>
                        <span>Using ${this.activeToolName}...</span>
                    </div>
                `
                : nothing}
        `;
    }
}

declare global {
    interface HTMLElementTagNameMap {
        'fb-chat-message-stream': FBChatMessageStream;
    }
}
