/**
 * FBNotificationList - Notification Stream Dropdown
 * See STEP_91_NOTIFICATION_UI.md Section 2.2
 *
 * Renders a scrollable list of system notifications as a popover dropdown.
 * Includes "Mark all as read" action and relative timestamps.
 *
 * @element fb-notification-list
 */
import { html, css, TemplateResult, nothing } from 'lit';
import { customElement, property, state } from 'lit/decorators.js';
import { effect } from '@preact/signals-core';
import { FBElement } from '../base/FBElement';
import {
    systemNotifications$,
    markAsRead,
    markAllAsRead,
    type SystemNotification,
    type SystemNotificationType,
} from '../../services/notification-service';

/**
 * Formats an ISO timestamp to a relative time string (e.g., "5m ago").
 */
function formatRelativeTime(isoString: string): string {
    const now = Date.now();
    const then = new Date(isoString).getTime();
    const diffMs = now - then;

    if (diffMs < 0) return 'now';

    const minutes = Math.floor(diffMs / 60_000);
    if (minutes < 1) return 'now';
    if (minutes < 60) return `${String(minutes)}m ago`;

    const hours = Math.floor(minutes / 60);
    if (hours < 24) return `${String(hours)}h ago`;

    const days = Math.floor(hours / 24);
    if (days < 7) return `${String(days)}d ago`;

    return new Date(isoString).toLocaleDateString();
}

@customElement('fb-notification-list')
export class FBNotificationList extends FBElement {
    static override styles = [
        FBElement.styles,
        css`
            :host {
                display: block;
                position: absolute;
                top: 100%;
                right: 0;
                margin-top: var(--fb-spacing-xs, 4px);
                width: 360px;
                max-height: 420px;
                background: var(--fb-bg-card, #111);
                border: 1px solid var(--fb-border, #333);
                border-radius: var(--fb-radius-lg, 12px);
                box-shadow: var(--fb-shadow-lg, 0 10px 15px rgba(0, 0, 0, 0.5));
                overflow: hidden;
                z-index: 1001;
            }

            /* Mobile: full-width dropdown */
            @media (max-width: 767px) {
                :host {
                    width: calc(100vw - var(--fb-spacing-lg, 24px));
                    right: calc(-1 * var(--fb-spacing-sm, 8px));
                }
            }

            .header {
                display: flex;
                align-items: center;
                justify-content: space-between;
                padding: var(--fb-spacing-sm, 8px) var(--fb-spacing-md, 16px);
                border-bottom: 1px solid var(--fb-border-light, #222);
            }

            .header-title {
                font-size: var(--fb-text-sm, 0.875rem);
                font-weight: 600;
                color: var(--fb-text-primary, #fff);
            }

            .mark-read-btn {
                border: none;
                background: transparent;
                color: var(--fb-primary, #667eea);
                font-size: var(--fb-text-xs, 0.75rem);
                font-weight: 500;
                font-family: var(--fb-font-family, 'Poppins', system-ui, sans-serif);
                cursor: pointer;
                padding: var(--fb-spacing-xs, 4px) var(--fb-spacing-sm, 8px);
                border-radius: var(--fb-radius-sm, 4px);
                transition: background var(--fb-transition-fast, 150ms ease);
            }

            .mark-read-btn:hover {
                background: rgba(102, 126, 234, 0.1);
            }

            .mark-read-btn:focus-visible {
                outline: 2px solid var(--fb-primary, #667eea);
                outline-offset: 2px;
            }

            .list {
                overflow-y: auto;
                max-height: 360px;
            }

            .item {
                display: flex;
                gap: var(--fb-spacing-sm, 8px);
                padding: var(--fb-spacing-sm, 8px) var(--fb-spacing-md, 16px);
                border-bottom: 1px solid var(--fb-border-light, #222);
                cursor: pointer;
                transition: background var(--fb-transition-fast, 150ms ease);
                border: none;
                background: transparent;
                width: 100%;
                text-align: left;
                font-family: var(--fb-font-family, 'Poppins', system-ui, sans-serif);
            }

            .item:hover {
                background: var(--fb-bg-tertiary, #1a1a1a);
            }

            .item:focus-visible {
                outline: 2px solid var(--fb-primary, #667eea);
                outline-offset: -2px;
            }

            .item:last-child {
                border-bottom: none;
            }

            .item.unread {
                background: rgba(102, 126, 234, 0.05);
            }

            .item.unread::before {
                content: '';
                position: absolute;
                left: 0;
                top: 0;
                bottom: 0;
                width: 3px;
                background: var(--fb-primary, #667eea);
            }

            .item {
                position: relative;
            }

            .icon-wrapper {
                flex-shrink: 0;
                width: 32px;
                height: 32px;
                border-radius: 50%;
                display: flex;
                align-items: center;
                justify-content: center;
            }

            .icon-wrapper svg {
                width: 16px;
                height: 16px;
                stroke: currentColor;
                fill: none;
                stroke-width: 2;
                stroke-linecap: round;
                stroke-linejoin: round;
            }

            .icon-wrapper.alert {
                background: rgba(198, 40, 40, 0.15);
                color: var(--fb-error, #c62828);
            }

            .icon-wrapper.success {
                background: rgba(46, 125, 50, 0.15);
                color: var(--fb-success, #2e7d32);
            }

            .icon-wrapper.system {
                background: rgba(21, 101, 192, 0.15);
                color: var(--fb-info, #1565c0);
            }

            .icon-wrapper.mention {
                background: rgba(102, 126, 234, 0.15);
                color: var(--fb-primary, #667eea);
            }

            .content {
                flex: 1;
                min-width: 0;
            }

            .title {
                font-size: var(--fb-text-sm, 0.875rem);
                font-weight: 500;
                color: var(--fb-text-primary, #fff);
                margin: 0;
                line-height: 1.3;
            }

            .item.unread .title {
                font-weight: 600;
            }

            .message {
                font-size: var(--fb-text-xs, 0.75rem);
                color: var(--fb-text-secondary, #aaa);
                margin: 2px 0 0 0;
                line-height: 1.4;
                display: -webkit-box;
                -webkit-line-clamp: 2;
                -webkit-box-orient: vertical;
                overflow: hidden;
            }

            .time {
                font-size: 10px;
                color: var(--fb-text-muted, #666);
                margin-top: 4px;
                flex-shrink: 0;
                align-self: flex-start;
                padding-top: 2px;
            }

            .empty {
                display: flex;
                align-items: center;
                justify-content: center;
                padding: var(--fb-spacing-xl, 32px);
                color: var(--fb-text-muted, #666);
                font-size: var(--fb-text-sm, 0.875rem);
            }
        `,
    ];

    @property({ type: Boolean }) open = false;

    @state() private _notifications: SystemNotification[] = [];

    private _disposeEffects: (() => void)[] = [];

    override connectedCallback(): void {
        super.connectedCallback();
        this._disposeEffects.push(
            effect(() => {
                this._notifications = systemNotifications$.value;
            })
        );
    }

    override disconnectedCallback(): void {
        this._disposeEffects.forEach((d) => { d(); });
        this._disposeEffects = [];
        super.disconnectedCallback();
    }

    private _handleMarkAllRead(): void {
        markAllAsRead();
    }

    private _handleItemClick(notification: SystemNotification): void {
        if (!notification.isRead) {
            markAsRead(notification.id);
        }
        if (notification.link) {
            window.history.pushState({}, '', notification.link);
            window.dispatchEvent(new PopStateEvent('popstate'));
            // Close the dropdown
            this.emit('notification-close');
        }
    }

    private _renderIcon(type: SystemNotificationType): TemplateResult {
        switch (type) {
            case 'alert':
                return html`<svg viewBox="0 0 24 24"><path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z"/><line x1="12" y1="9" x2="12" y2="13"/><line x1="12" y1="17" x2="12.01" y2="17"/></svg>`;
            case 'success':
                return html`<svg viewBox="0 0 24 24"><path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"/><polyline points="22 4 12 14.01 9 11.01"/></svg>`;
            case 'mention':
                return html`<svg viewBox="0 0 24 24"><path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"/></svg>`;
            case 'system':
            default:
                return html`<svg viewBox="0 0 24 24"><circle cx="12" cy="12" r="10"/><line x1="12" y1="16" x2="12" y2="12"/><line x1="12" y1="8" x2="12.01" y2="8"/></svg>`;
        }
    }

    override render(): TemplateResult {
        if (!this.open) return html``;

        const hasUnread = this._notifications.some((n) => !n.isRead);

        return html`
            <div class="header">
                <span class="header-title">Notifications</span>
                ${hasUnread ? html`
                    <button
                        class="mark-read-btn"
                        @click=${this._handleMarkAllRead.bind(this)}
                        aria-label="Mark all notifications as read"
                    >
                        Mark all read
                    </button>
                ` : nothing}
            </div>
            <div class="list" role="list" aria-label="Notification list">
                ${this._notifications.length === 0 ? html`
                    <div class="empty">No notifications</div>
                ` : this._notifications.map((n) => html`
                    <button
                        class="item ${n.isRead ? '' : 'unread'}"
                        role="listitem"
                        @click=${(): void => { this._handleItemClick(n); }}
                        aria-label="${n.title}: ${n.message}"
                    >
                        <div class="icon-wrapper ${n.type}" aria-hidden="true">
                            ${this._renderIcon(n.type)}
                        </div>
                        <div class="content">
                            <p class="title">${n.title}</p>
                            <p class="message">${n.message}</p>
                        </div>
                        <span class="time" aria-label="Time: ${formatRelativeTime(n.createdAt)}">${formatRelativeTime(n.createdAt)}</span>
                    </button>
                `)}
            </div>
        `;
    }
}

declare global {
    interface HTMLElementTagNameMap {
        'fb-notification-list': FBNotificationList;
    }
}
