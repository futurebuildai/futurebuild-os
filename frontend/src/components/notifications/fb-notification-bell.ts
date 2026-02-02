/**
 * FBNotificationBell - Bell Icon with Unread Badge
 * See STEP_91_NOTIFICATION_UI.md Section 2.1
 *
 * Renders a bell icon in the app header. Shows a red badge with unread count.
 * Clicking toggles the notification dropdown list.
 *
 * @element fb-notification-bell
 * @fires notification-toggle - Dispatched when bell is clicked
 */
import { html, css, TemplateResult, nothing } from 'lit';
import { customElement, state } from 'lit/decorators.js';
import { effect } from '@preact/signals-core';
import { FBElement } from '../base/FBElement';
import { unreadCount$ } from '../../services/notification-service';

@customElement('fb-notification-bell')
export class FBNotificationBell extends FBElement {
    static override styles = [
        FBElement.styles,
        css`
            :host {
                display: inline-flex;
                position: relative;
            }

            .bell-btn {
                position: relative;
                display: flex;
                align-items: center;
                justify-content: center;
                width: 36px;
                height: 36px;
                border: none;
                background: transparent;
                color: var(--fb-text-secondary, #aaa);
                cursor: pointer;
                border-radius: var(--fb-radius-md, 8px);
                transition: background var(--fb-transition-fast, 150ms ease),
                            color var(--fb-transition-fast, 150ms ease);
                -webkit-tap-highlight-color: transparent;
            }

            .bell-btn:hover {
                background: var(--fb-bg-tertiary, #1a1a1a);
                color: var(--fb-text-primary, #fff);
            }

            .bell-btn:focus-visible {
                outline: 2px solid var(--fb-primary, #667eea);
                outline-offset: 2px;
            }

            .bell-btn svg {
                width: 20px;
                height: 20px;
                stroke: currentColor;
                fill: none;
                stroke-width: 2;
                stroke-linecap: round;
                stroke-linejoin: round;
            }

            .badge {
                position: absolute;
                top: 2px;
                right: 2px;
                display: flex;
                align-items: center;
                justify-content: center;
                min-width: 16px;
                height: 16px;
                padding: 0 4px;
                border-radius: 8px;
                background: var(--fb-error, #c62828);
                color: #fff;
                font-size: 10px;
                font-weight: 700;
                font-family: var(--fb-font-family, 'Poppins', system-ui, sans-serif);
                line-height: 1;
                pointer-events: none;
            }
        `,
    ];

    @state() private _unreadCount = 0;

    private _disposeEffects: (() => void)[] = [];

    override connectedCallback(): void {
        super.connectedCallback();
        this._disposeEffects.push(
            effect(() => {
                this._unreadCount = unreadCount$.value;
            })
        );
    }

    override disconnectedCallback(): void {
        this._disposeEffects.forEach((d) => { d(); });
        this._disposeEffects = [];
        super.disconnectedCallback();
    }

    private _handleClick(): void {
        this.emit('notification-toggle');
    }

    override render(): TemplateResult {
        return html`
            <button
                class="bell-btn"
                @click=${this._handleClick.bind(this)}
                aria-label="Notifications${this._unreadCount > 0 ? `, ${String(this._unreadCount)} unread` : ''}"
                aria-haspopup="true"
            >
                <svg viewBox="0 0 24 24" aria-hidden="true">
                    <path d="M18 8A6 6 0 0 0 6 8c0 7-3 9-3 9h18s-3-2-3-9"/>
                    <path d="M13.73 21a2 2 0 0 1-3.46 0"/>
                </svg>
                ${this._unreadCount > 0 ? html`
                    <span class="badge" aria-hidden="true">${this._unreadCount > 9 ? '9+' : String(this._unreadCount)}</span>
                ` : nothing}
            </button>
        `;
    }
}

declare global {
    interface HTMLElementTagNameMap {
        'fb-notification-bell': FBNotificationBell;
    }
}
