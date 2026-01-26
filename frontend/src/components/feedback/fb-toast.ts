/**
 * FBToast - Individual Toast Notification Component
 * See LAUNCH_PLAN.md P2 (Notifications/Toast UI)
 *
 * Displays a single toast notification with:
 * - Icon based on type (success, error, warning, info)
 * - Close button
 * - Optional action button
 * - Slide-in animation
 */
import { html, css, TemplateResult, nothing } from 'lit';
import { customElement, property } from 'lit/decorators.js';
import { FBElement } from '../base/FBElement';
import type { NotificationType } from '../../store/notifications';

/**
 * Toast notification component.
 * @element fb-toast
 *
 * @fires fb-toast-dismiss - Fired when toast should be dismissed
 */
@customElement('fb-toast')
export class FBToast extends FBElement {
    static override styles = [
        FBElement.styles,
        css`
            :host {
                display: block;
                pointer-events: auto;
            }

            .toast {
                display: flex;
                align-items: flex-start;
                gap: 12px;
                padding: 12px 16px;
                background: var(--fb-bg-card, #111);
                border: 1px solid var(--fb-border, #333);
                border-radius: 8px;
                box-shadow: 0 4px 12px rgba(0, 0, 0, 0.4);
                min-width: 280px;
                max-width: 400px;
                animation: slideIn 0.3s ease;
            }

            @keyframes slideIn {
                from {
                    transform: translateX(100%);
                    opacity: 0;
                }
                to {
                    transform: translateX(0);
                    opacity: 1;
                }
            }

            :host([dismissing]) .toast {
                animation: slideOut 0.2s ease forwards;
            }

            @keyframes slideOut {
                from {
                    transform: translateX(0);
                    opacity: 1;
                }
                to {
                    transform: translateX(100%);
                    opacity: 0;
                }
            }

            /* Type-specific border colors */
            .toast--success {
                border-left: 4px solid var(--fb-success, #2e7d32);
            }

            .toast--error {
                border-left: 4px solid var(--fb-error, #c62828);
            }

            .toast--warning {
                border-left: 4px solid var(--fb-warning, #f9a825);
            }

            .toast--info {
                border-left: 4px solid var(--fb-primary, #667eea);
            }

            .icon {
                flex-shrink: 0;
                width: 20px;
                height: 20px;
                display: flex;
                align-items: center;
                justify-content: center;
            }

            .icon--success {
                color: var(--fb-success, #2e7d32);
            }

            .icon--error {
                color: var(--fb-error, #c62828);
            }

            .icon--warning {
                color: var(--fb-warning, #f9a825);
            }

            .icon--info {
                color: var(--fb-primary, #667eea);
            }

            .content {
                flex: 1;
                min-width: 0;
            }

            .message {
                color: var(--fb-text-primary, #fff);
                font-size: 14px;
                line-height: 1.5;
                margin: 0;
                word-wrap: break-word;
            }

            .actions {
                display: flex;
                gap: 8px;
                margin-top: 8px;
            }

            .action-btn {
                padding: 4px 12px;
                font-size: 13px;
                font-weight: 500;
                background: var(--fb-primary, #667eea);
                color: white;
                border: none;
                border-radius: 4px;
                cursor: pointer;
                transition: background 0.2s ease;
            }

            .action-btn:hover {
                background: var(--fb-primary-hover, #5a6fd6);
            }

            .close-btn {
                flex-shrink: 0;
                width: 24px;
                height: 24px;
                padding: 0;
                display: flex;
                align-items: center;
                justify-content: center;
                background: transparent;
                border: none;
                border-radius: 4px;
                color: var(--fb-text-secondary, #aaa);
                cursor: pointer;
                transition: background 0.2s ease, color 0.2s ease;
            }

            .close-btn:hover {
                background: var(--fb-bg-tertiary, #1a1a1a);
                color: var(--fb-text-primary, #fff);
            }

            .close-btn svg {
                width: 16px;
                height: 16px;
            }
        `,
    ];

    @property({ type: String }) type: NotificationType = 'info';
    @property({ type: String }) message = '';
    @property({ type: String }) actionLabel = '';
    @property({ type: String, attribute: 'notification-id' }) notificationId = '';

    private _actionCallback?: () => void;

    setActionCallback(callback: () => void): void {
        this._actionCallback = callback;
    }

    private _handleDismiss(): void {
        this.emit('fb-toast-dismiss', { id: this.notificationId });
    }

    private _handleAction(): void {
        if (this._actionCallback) {
            this._actionCallback();
        }
        this._handleDismiss();
    }

    private _renderIcon(): TemplateResult {
        switch (this.type) {
            case 'success':
                return html`
                    <svg viewBox="0 0 24 24" fill="currentColor">
                        <path d="M9 16.17L4.83 12l-1.42 1.41L9 19 21 7l-1.41-1.41L9 16.17z"/>
                    </svg>
                `;
            case 'error':
                return html`
                    <svg viewBox="0 0 24 24" fill="currentColor">
                        <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm1 15h-2v-2h2v2zm0-4h-2V7h2v6z"/>
                    </svg>
                `;
            case 'warning':
                return html`
                    <svg viewBox="0 0 24 24" fill="currentColor">
                        <path d="M1 21h22L12 2 1 21zm12-3h-2v-2h2v2zm0-4h-2v-4h2v4z"/>
                    </svg>
                `;
            case 'info':
            default:
                return html`
                    <svg viewBox="0 0 24 24" fill="currentColor">
                        <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm1 15h-2v-6h2v6zm0-8h-2V7h2v2z"/>
                    </svg>
                `;
        }
    }

    override render(): TemplateResult {
        return html`
            <div
                class="toast toast--${this.type}"
                role="status"
                aria-live="polite"
            >
                <div class="icon icon--${this.type}">
                    ${this._renderIcon()}
                </div>
                <div class="content">
                    <p class="message">${this.message}</p>
                    ${this.actionLabel
                        ? html`
                              <div class="actions">
                                  <button
                                      class="action-btn"
                                      @click=${this._handleAction}
                                  >
                                      ${this.actionLabel}
                                  </button>
                              </div>
                          `
                        : nothing}
                </div>
                <button
                    class="close-btn"
                    @click=${this._handleDismiss}
                    aria-label="Dismiss notification"
                >
                    <svg viewBox="0 0 24 24" fill="currentColor">
                        <path d="M19 6.41L17.59 5 12 10.59 6.41 5 5 6.41 10.59 12 5 17.59 6.41 19 12 13.41 17.59 19 19 17.59 13.41 12 19 6.41z"/>
                    </svg>
                </button>
            </div>
        `;
    }
}

declare global {
    interface HTMLElementTagNameMap {
        'fb-toast': FBToast;
    }
}
