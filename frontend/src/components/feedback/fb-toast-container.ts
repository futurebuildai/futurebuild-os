/**
 * FBToastContainer - Global Toast Notification Container
 * See LAUNCH_PLAN.md P2 (Notifications/Toast UI)
 *
 * Fixed position container that renders a stack of toast notifications.
 * - Subscribes to notifications$ signal
 * - Handles auto-dismiss timers
 * - Accessible: role="status" + aria-live="polite"
 */
import { html, css, TemplateResult } from 'lit';
import { customElement, state } from 'lit/decorators.js';
import { effect } from '@preact/signals-core';
import { FBElement } from '../base/FBElement';
import { notifications$, notify, type Notification } from '../../store/notifications';
import './fb-toast';

/**
 * Global toast container component.
 * @element fb-toast-container
 *
 * Mount this component once in the app shell:
 * ```html
 * <fb-toast-container></fb-toast-container>
 * ```
 */
@customElement('fb-toast-container')
export class FBToastContainer extends FBElement {
    static override styles = [
        FBElement.styles,
        css`
            :host {
                position: fixed;
                bottom: 24px;
                right: 24px;
                z-index: var(--fb-z-toast, 1000);
                display: flex;
                flex-direction: column-reverse;
                gap: 12px;
                pointer-events: none;
                max-height: calc(100vh - 48px);
                overflow: hidden;
            }

            /* Mobile: Center at bottom, full width with padding */
            @media (max-width: 768px) {
                :host {
                    left: 16px;
                    right: 16px;
                    bottom: 16px;
                }

                fb-toast {
                    width: 100%;
                }
            }
        `,
    ];

    @state() private _notifications: Notification[] = [];
    private _disposeEffect?: () => void;
    private _timers = new Map<string, number>();

    override connectedCallback(): void {
        super.connectedCallback();

        // Subscribe to notifications signal
        this._disposeEffect = effect(() => {
            const notifications = notifications$.value;
            this._notifications = notifications;

            // Set up auto-dismiss timers for new notifications
            for (const notification of notifications) {
                if (!this._timers.has(notification.id) && notification.duration > 0) {
                    const timerId = window.setTimeout(() => {
                        this._dismissWithAnimation(notification.id);
                    }, notification.duration);
                    this._timers.set(notification.id, timerId);
                }
            }

            // Clear timers for removed notifications
            for (const [id, timerId] of this._timers.entries()) {
                if (!notifications.find((n) => n.id === id)) {
                    window.clearTimeout(timerId);
                    this._timers.delete(id);
                }
            }
        });

        // Listen for dismiss events from toast components
        this.addEventListener('fb-toast-dismiss', this._handleDismiss as EventListener);
    }

    override disconnectedCallback(): void {
        this._disposeEffect?.();
        this.removeEventListener('fb-toast-dismiss', this._handleDismiss as EventListener);

        // Clear all timers
        for (const timerId of this._timers.values()) {
            window.clearTimeout(timerId);
        }
        this._timers.clear();

        super.disconnectedCallback();
    }

    private _handleDismiss = (e: CustomEvent<{ id: string }>): void => {
        e.stopPropagation();
        this._dismissWithAnimation(e.detail.id);
    };

    private _dismissWithAnimation(id: string): void {
        // Clear timer if exists
        const timerId = this._timers.get(id);
        if (timerId) {
            window.clearTimeout(timerId);
            this._timers.delete(id);
        }

        // Find toast element and add dismissing attribute
        const toastElement = this.shadowRoot?.querySelector(
            `fb-toast[notification-id="${id}"]`
        );
        if (toastElement) {
            toastElement.setAttribute('dismissing', '');
            // Wait for animation to complete
            setTimeout(() => {
                notify.dismiss(id);
            }, 200);
        } else {
            notify.dismiss(id);
        }
    }

    override render(): TemplateResult {
        return html`
            ${this._notifications.map((notification) => {
                const toastEl = document.createElement('fb-toast') as HTMLElement & {
                    type: string;
                    message: string;
                    actionLabel: string;
                    notificationId: string;
                    setActionCallback: (callback: () => void) => void;
                };
                return html`
                    <fb-toast
                        type="${notification.type}"
                        message="${notification.message}"
                        notification-id="${notification.id}"
                        .actionLabel=${notification.action?.label ?? ''}
                        @connected=${(e: Event) => {
                            const el = e.target as typeof toastEl;
                            if (notification.action?.callback) {
                                el.setActionCallback(notification.action.callback);
                            }
                        }}
                    ></fb-toast>
                `;
            })}
        `;
    }
}

declare global {
    interface HTMLElementTagNameMap {
        'fb-toast-container': FBToastContainer;
    }
}
