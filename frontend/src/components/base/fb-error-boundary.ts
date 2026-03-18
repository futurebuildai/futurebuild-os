/**
 * FBErrorBoundary - Safe Render Fallback Component
 * See FRONTEND_SCOPE.md Section 4.1
 * Step 58.5: Fortress Hardening - Flag 1 (White Screen of Death)
 *
 * Since Lit lacks React's componentDidCatch, this component provides a
 * fallback UI that can be triggered by parent components when data validation fails.
 * Parents set `hasError` and `errorMessage` properties instead of relying on exception catching.
 */
import { html, css, TemplateResult } from 'lit';
import { customElement, property } from 'lit/decorators.js';
import { FBElement } from './FBElement';
import { store } from '../../store/store';

@customElement('fb-error-boundary')
export class FBErrorBoundary extends FBElement {
    static override styles = [
        FBElement.styles,
        css`
            :host {
                display: block;
                width: 100%;
            }

            .error-fallback {
                display: flex;
                flex-direction: column;
                align-items: center;
                justify-content: center;
                padding: var(--fb-spacing-lg);
                background: var(--fb-bg-tertiary);
                border: 1px dashed var(--fb-error, #F43F5E);
                border-radius: var(--fb-radius-md);
                color: var(--fb-text-secondary);
                text-align: center;
                min-height: 100px;
            }

            .error-icon {
                font-size: 32px;
                margin-bottom: var(--fb-spacing-sm);
            }

            .error-title {
                font-size: var(--fb-text-sm);
                font-weight: 600;
                color: var(--fb-error, #F43F5E);
                margin-bottom: var(--fb-spacing-xs);
            }

            .error-message {
                font-size: var(--fb-text-xs);
                color: var(--fb-text-muted);
                max-width: 300px;
                word-break: break-word;
                margin-bottom: var(--fb-spacing-md);
            }

            .error-actions button {
                background: var(--fb-bg-secondary);
                border: 1px solid var(--fb-border);
                color: var(--fb-text-primary);
                padding: var(--fb-spacing-xs) var(--fb-spacing-md);
                border-radius: var(--fb-radius-sm);
                cursor: pointer;
                font-size: var(--fb-text-xs);
            }
            .error-actions button:hover {
                background: var(--fb-bg-tertiary);
            }
        `,
    ];

    /**
     * Whether an error state is active.
     * Set by parent component when data validation fails.
     */
    @property({ type: Boolean, attribute: 'has-error' })
    hasError = false;

    /**
     * Error message to display in fallback UI.
     */
    @property({ type: String, attribute: 'error-message' })
    errorMessage = 'An error occurred';

    /**
     * Expected category to dynamically select icon, title, and actions.
     */
    @property({ type: String, attribute: 'error-category' })
    errorCategory: 'network' | 'auth' | 'data' | 'ai' | 'unknown' = 'unknown';

    private _handleRetry() {
        this.dispatchEvent(new CustomEvent('fb-retry', { bubbles: true, composed: true }));
    }

    private _handleLogin() {
        store.actions.logout();
        window.location.href = '/login';
    }

    override render(): TemplateResult {
        if (this.hasError) {
            let icon = '⚠️';
            let title = 'Data Error';
            let action = html`<button @click=${this._handleRetry}>Retry</button>`;

            if (this.errorCategory === 'network') {
                icon = '📡';
                title = 'Network Error';
                action = html`<button @click=${this._handleRetry}>Retry Connection</button>`;
            } else if (this.errorCategory === 'auth') {
                icon = '🔒';
                title = 'Authentication Error';
                action = html`<button @click=${this._handleLogin}>Log In</button>`;
            } else if (this.errorCategory === 'ai') {
                icon = '🤖';
                title = 'AI Service Unavailable';
                action = html`<button @click=${this._handleRetry}>Retry AI Request</button>`;
            }

            return html`
                <div class="error-fallback" role="alert" aria-live="polite">
                    <span class="error-icon">${icon}</span>
                    <span class="error-title">${title}</span>
                    <span class="error-message">${this.errorMessage}</span>
                    <div class="error-actions">${action}</div>
                </div>
            `;
        }

        // Pass through children when no error
        return html`<slot></slot>`;
    }
}

declare global {
    interface HTMLElementTagNameMap {
        'fb-error-boundary': FBErrorBoundary;
    }
}
