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
                border: 1px dashed var(--fb-error, #ef4444);
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
                color: var(--fb-error, #ef4444);
                margin-bottom: var(--fb-spacing-xs);
            }

            .error-message {
                font-size: var(--fb-text-xs);
                color: var(--fb-text-muted);
                max-width: 300px;
                word-break: break-word;
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

    override render(): TemplateResult {
        if (this.hasError) {
            return html`
                <div class="error-fallback" role="alert" aria-live="polite">
                    <span class="error-icon">⚠️</span>
                    <span class="error-title">Data Error</span>
                    <span class="error-message">${this.errorMessage}</span>
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
