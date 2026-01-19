/**
 * FBViewChat - AI Chat Interface View
 * See PRODUCTION_PLAN.md Step 51.4
 *
 * The primary conversational interface for FutureBuild.
 */
import { html, css, TemplateResult } from 'lit';
import { customElement } from 'lit/decorators.js';
import { FBViewElement } from '../base/FBViewElement';

@customElement('fb-view-chat')
export class FBViewChat extends FBViewElement {
    static override styles = [
        FBViewElement.styles,
        css`
            :host {
                display: flex;
                flex-direction: column;
            }

            .chat-header {
                padding: var(--fb-spacing-lg);
                border-bottom: 1px solid var(--fb-border);
            }

            h1 {
                font-size: var(--fb-text-xl);
                font-weight: 600;
                color: var(--fb-text-primary);
                margin: 0;
            }

            .chat-body {
                flex: 1;
                display: flex;
                align-items: center;
                justify-content: center;
                padding: var(--fb-spacing-xl);
            }

            .placeholder {
                background: var(--fb-bg-card);
                border: 1px dashed var(--fb-border);
                border-radius: var(--fb-radius-lg);
                padding: var(--fb-spacing-2xl);
                text-align: center;
                color: var(--fb-text-muted);
                max-width: 500px;
            }
        `,
    ];

    override onViewActive(): void {
        console.log('[FBViewChat] View activated - would load chat history');
    }

    override render(): TemplateResult {
        return html`
            <div class="chat-header">
                <h1>AI Assistant</h1>
            </div>
            <div class="chat-body">
                <div class="placeholder">
                    Chat interface will be built in Step 52. Will include message list, input bar, and artifact rendering.
                </div>
            </div>
        `;
    }
}

declare global {
    interface HTMLElementTagNameMap {
        'fb-view-chat': FBViewChat;
    }
}
