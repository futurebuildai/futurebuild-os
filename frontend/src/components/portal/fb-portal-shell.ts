/**
 * FBPortalShell - Simplified Mobile Shell for Portal
 * See LAUNCH_PLAN.md P2: Field Portal (Mobile)
 *
 * A minimal shell for portal views (no 3-panel layout).
 * - Header with project name (for permanent accounts)
 * - Minimal header for one-time action views
 * - Main content area
 */
import { html, css, TemplateResult, nothing } from 'lit';
import { customElement, property, state } from 'lit/decorators.js';
import { FBElement } from '../base/FBElement';
import './fb-portal-voice-input';
import './fb-portal-task-list';

/**
 * Portal shell component.
 * @element fb-portal-shell
 */
@customElement('fb-portal-shell')
export class FBPortalShell extends FBElement {
    static override styles = [
        FBElement.styles,
        css`
            :host {
                display: flex;
                flex-direction: column;
                min-height: 100vh;
                width: 100%;
                background: var(--fb-bg-primary, #000);
                color: var(--fb-text-primary, #fff);
                font-family: var(--fb-font-family, -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif);
                font-size: 16px;
                --fb-portal-accent: var(--fb-accent, #00FFA3);
            }

            .header {
                display: flex;
                align-items: center;
                justify-content: space-between;
                padding: 16px 20px;
                background: var(--fb-bg-secondary, #0a0a0a);
                border-bottom: 1px solid var(--fb-border, rgba(255,255,255,0.05));
            }

            .header--minimal {
                justify-content: center;
            }

            .logo {
                display: flex;
                align-items: center;
                gap: 8px;
                font-size: 18px;
                font-weight: 600;
                color: var(--fb-primary, #00FFA3);
            }

            .logo svg {
                width: 28px;
                height: 28px;
            }

            .project-name {
                font-size: 14px;
                color: var(--fb-text-secondary, #8B8D98);
            }

            .main {
                flex: 1;
                display: flex;
                flex-direction: column;
                padding: 20px;
                max-width: 600px;
                margin: 0 auto;
                width: 100%;
            }

            .footer {
                padding: 16px 20px;
                text-align: center;
                background: var(--fb-bg-secondary, #0a0a0a);
                border-top: 1px solid var(--fb-border, rgba(255,255,255,0.05));
            }

            .footer-text {
                color: var(--fb-text-muted, #4A4B55);
                font-size: 12px;
                margin: 0;
            }

            .footer-link {
                color: var(--fb-primary, #00FFA3);
                text-decoration: none;
            }

            .footer-link:hover {
                text-decoration: underline;
            }

            .offline-indicator {
                display: flex;
                align-items: center;
                gap: 6px;
                padding: 4px 10px;
                border-radius: 4px;
                background: rgba(245, 158, 11, 0.1);
                color: #f59e0b;
                font-size: 12px;
                font-weight: 600;
            }

            .offline-indicator .dot {
                width: 8px;
                height: 8px;
                border-radius: 50%;
                background: #f59e0b;
            }

            .online-indicator {
                display: flex;
                align-items: center;
                gap: 6px;
                padding: 4px 10px;
                border-radius: 4px;
                font-size: 12px;
                color: var(--fb-primary, #00FFA3);
            }

            .online-indicator .dot {
                width: 8px;
                height: 8px;
                border-radius: 50%;
                background: var(--fb-primary, #00FFA3);
            }

            .voice-section {
                display: flex;
                flex-direction: column;
                align-items: center;
                padding: 16px 0;
                border-bottom: 1px solid var(--fb-border, rgba(255,255,255,0.05));
                margin-bottom: 16px;
            }
        `,
    ];

    @property({ type: String, attribute: 'project-name' }) projectName = '';
    @property({ type: Boolean }) minimal = false;
    @property({ type: Boolean, attribute: 'show-voice' }) showVoice = false;
    @state() private _isOnline = typeof navigator !== 'undefined' ? navigator.onLine : true;

    override connectedCallback() {
        super.connectedCallback();
        window.addEventListener('online', this._handleOnline);
        window.addEventListener('offline', this._handleOffline);
    }

    override disconnectedCallback() {
        super.disconnectedCallback();
        window.removeEventListener('online', this._handleOnline);
        window.removeEventListener('offline', this._handleOffline);
    }

    private _handleOnline = () => { this._isOnline = true; };
    private _handleOffline = () => { this._isOnline = false; };

    override render(): TemplateResult {
        return html`
            <header class="header ${this.minimal ? 'header--minimal' : ''}">
                <div class="logo">
                    <svg viewBox="0 0 24 24" fill="currentColor">
                        <path d="M19 3H5c-1.1 0-2 .9-2 2v14c0 1.1.9 2 2 2h14c1.1 0 2-.9 2-2V5c0-1.1-.9-2-2-2zm-7 14l-5-5 1.41-1.41L12 14.17l4.59-4.58L18 11l-6 6z"/>
                    </svg>
                    FutureBuild
                </div>
                ${!this.minimal && this.projectName
                    ? html`<span class="project-name">${this.projectName}</span>`
                    : nothing}
                ${this._isOnline
                    ? html`<span class="online-indicator"><span class="dot"></span>Online</span>`
                    : html`<span class="offline-indicator"><span class="dot"></span>Offline</span>`}
            </header>

            <main class="main">
                ${this.showVoice ? html`
                    <div class="voice-section">
                        <fb-portal-voice-input></fb-portal-voice-input>
                    </div>
                ` : nothing}
                <slot></slot>
            </main>

            <footer class="footer">
                <p class="footer-text">
                    Powered by <a class="footer-link" href="https://futurebuild.ai" target="_blank">FutureBuild</a>
                </p>
            </footer>
        `;
    }
}

declare global {
    interface HTMLElementTagNameMap {
        'fb-portal-shell': FBPortalShell;
    }
}
