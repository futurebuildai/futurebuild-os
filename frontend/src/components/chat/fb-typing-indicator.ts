/**
 * FBTypingIndicator - Agent Typing State Visual Feedback
 * See FRONTEND_SCOPE.md Section 8.4 (Streaming Response Handler)
 *
 * Displays an animated "processing" indicator when the AI agent is generating a response.
 * Uses a sleek pulsing dots animation to match the FutureBuild design language.
 */

import { html, css, TemplateResult } from 'lit';
import { customElement } from 'lit/decorators.js';
import { FBElement } from '../base/FBElement';

@customElement('fb-typing-indicator')
export class FBTypingIndicator extends FBElement {
    static override styles = [
        FBElement.styles,
        css`
            :host {
                display: flex;
                align-items: center;
                gap: var(--fb-spacing-sm);
                padding: var(--fb-spacing-sm) var(--fb-spacing-md);
                max-width: max-content;
            }

            .indicator-container {
                display: flex;
                align-items: center;
                gap: var(--fb-spacing-md);
                padding: var(--fb-spacing-sm) var(--fb-spacing-md);
                background: var(--fb-bg-tertiary);
                border-radius: var(--fb-radius-lg);
                border: 1px solid var(--fb-border-light);
            }

            .avatar {
                width: 28px;
                height: 28px;
                border-radius: var(--fb-radius-full);
                background: linear-gradient(135deg, var(--fb-primary), var(--fb-secondary));
                display: flex;
                align-items: center;
                justify-content: center;
                font-size: var(--fb-text-xs);
                color: white;
                font-weight: 600;
                flex-shrink: 0;
            }

            .content {
                display: flex;
                align-items: center;
                gap: var(--fb-spacing-sm);
            }

            .label {
                font-size: var(--fb-text-sm);
                color: var(--fb-text-secondary);
                font-weight: 500;
            }

            .dots {
                display: flex;
                align-items: center;
                gap: 4px;
            }

            .dot {
                width: 6px;
                height: 6px;
                border-radius: 50%;
                background: var(--fb-primary);
                animation: pulse 1.4s ease-in-out infinite;
            }

            .dot:nth-child(1) {
                animation-delay: 0s;
            }

            .dot:nth-child(2) {
                animation-delay: 0.2s;
            }

            .dot:nth-child(3) {
                animation-delay: 0.4s;
            }

            @keyframes pulse {
                0%,
                80%,
                100% {
                    opacity: 0.3;
                    transform: scale(0.8);
                }
                40% {
                    opacity: 1;
                    transform: scale(1);
                }
            }

            /* Subtle shimmer effect */
            .indicator-container::before {
                content: '';
                position: absolute;
                inset: 0;
                background: linear-gradient(
                    90deg,
                    transparent 0%,
                    rgba(255, 255, 255, 0.05) 50%,
                    transparent 100%
                );
                animation: shimmer 2s ease-in-out infinite;
            }

            .indicator-container {
                position: relative;
                overflow: hidden;
            }

            @keyframes shimmer {
                0% {
                    transform: translateX(-100%);
                }
                100% {
                    transform: translateX(100%);
                }
            }
        `,
    ];

    override render(): TemplateResult {
        return html`
            <div class="indicator-container" role="status" aria-label="FutureBuild AI is processing">
                <div class="avatar" aria-hidden="true">FB</div>
                <div class="content">
                    <span class="label">Processing</span>
                    <div class="dots" aria-hidden="true">
                        <span class="dot"></span>
                        <span class="dot"></span>
                        <span class="dot"></span>
                    </div>
                </div>
            </div>
        `;
    }
}

declare global {
    interface HTMLElementTagNameMap {
        'fb-typing-indicator': FBTypingIndicator;
    }
}
