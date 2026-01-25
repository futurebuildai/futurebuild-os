/**
 * FBProjectDialog - Modal Wrapper for Project Creation
 * See PROJECT_ONBOARDING_SPEC.md Step 62.5
 *
 * Provides a modal dialog container with backdrop, escape key handling,
 * and focus trapping for the project creation form.
 *
 * @element fb-project-dialog
 */
import { html, css, TemplateResult, nothing } from 'lit';
import { customElement, property } from 'lit/decorators.js';
import { FBElement } from '../../base/FBElement';

// Import form component
import './fb-project-form';

@customElement('fb-project-dialog')
export class FBProjectDialog extends FBElement {
    static override styles = [
        FBElement.styles,
        css`
            :host {
                display: none;
            }

            :host([open]) {
                display: block;
            }

            .backdrop {
                position: fixed;
                inset: 0;
                background: rgba(0, 0, 0, 0.7);
                backdrop-filter: blur(4px);
                z-index: var(--fb-z-modal, 1000);
                animation: fadeIn 0.15s ease;
            }

            @keyframes fadeIn {
                from { opacity: 0; }
                to { opacity: 1; }
            }

            .modal {
                position: fixed;
                top: 50%;
                left: 50%;
                transform: translate(-50%, -50%);
                width: 100%;
                max-width: 480px;
                max-height: 90vh;
                overflow-y: auto;
                background: var(--fb-bg-card);
                border: 1px solid var(--fb-border);
                border-radius: var(--fb-radius-xl, 16px);
                box-shadow: var(--fb-shadow-lg, 0 25px 50px -12px rgba(0, 0, 0, 0.25));
                z-index: calc(var(--fb-z-modal, 1000) + 1);
                animation: slideIn 0.2s ease;
            }

            @keyframes slideIn {
                from {
                    opacity: 0;
                    transform: translate(-50%, -48%);
                }
                to {
                    opacity: 1;
                    transform: translate(-50%, -50%);
                }
            }

            .header {
                display: flex;
                align-items: center;
                justify-content: space-between;
                padding: var(--fb-spacing-lg);
                border-bottom: 1px solid var(--fb-border-light);
            }

            .title {
                font-size: var(--fb-text-xl);
                font-weight: 600;
                color: var(--fb-text-primary);
                margin: 0;
            }

            .close-btn {
                display: flex;
                align-items: center;
                justify-content: center;
                width: 32px;
                height: 32px;
                padding: 0;
                border: none;
                background: transparent;
                color: var(--fb-text-muted);
                border-radius: var(--fb-radius-md);
                cursor: pointer;
                transition: background 0.15s ease, color 0.15s ease;
            }

            .close-btn:hover {
                background: var(--fb-bg-tertiary);
                color: var(--fb-text-primary);
            }

            .close-btn:focus-visible {
                outline: 2px solid var(--fb-primary);
                outline-offset: 2px;
            }

            .close-btn svg {
                width: 20px;
                height: 20px;
                stroke: currentColor;
                fill: none;
                stroke-width: 2;
            }

            .content {
                padding: var(--fb-spacing-lg);
            }

            @media (max-width: 520px) {
                .modal {
                    max-width: calc(100% - var(--fb-spacing-lg) * 2);
                }
            }
        `,
    ];

    @property({ type: Boolean, reflect: true })
    open = false;

    private _boundHandleKeyDown = this._handleKeyDown.bind(this);

    override connectedCallback(): void {
        super.connectedCallback();
        document.addEventListener('keydown', this._boundHandleKeyDown);
    }

    override disconnectedCallback(): void {
        document.removeEventListener('keydown', this._boundHandleKeyDown);
        super.disconnectedCallback();
    }

    private _handleKeyDown(e: KeyboardEvent): void {
        if (this.open && e.key === 'Escape') {
            e.preventDefault();
            this._close();
        }
    }

    private _handleBackdropClick(e: MouseEvent): void {
        // Only close if clicking the backdrop itself, not the modal
        if (e.target === e.currentTarget) {
            this._close();
        }
    }

    private _close(): void {
        this.emit('close');
    }

    private _handleProjectCreated(e: CustomEvent): void {
        // Re-emit the event from the form
        this.emit('project-created', e.detail);
    }

    private _handleCancel(): void {
        this._close();
    }

    override render(): TemplateResult {
        if (!this.open) {
            return html`${nothing}`;
        }

        return html`
            <div
                class="backdrop"
                @click=${this._handleBackdropClick.bind(this)}
                aria-hidden="true"
            ></div>
            <div
                class="modal"
                role="dialog"
                aria-modal="true"
                aria-labelledby="dialog-title"
            >
                <header class="header">
                    <h2 id="dialog-title" class="title">Create New Project</h2>
                    <button
                        class="close-btn"
                        @click=${this._close.bind(this)}
                        aria-label="Close dialog"
                    >
                        <svg viewBox="0 0 24 24" aria-hidden="true">
                            <path d="M18 6L6 18M6 6l12 12" />
                        </svg>
                    </button>
                </header>
                <div class="content">
                    <fb-project-form
                        @project-created=${this._handleProjectCreated.bind(this)}
                        @cancel=${this._handleCancel.bind(this)}
                    ></fb-project-form>
                </div>
            </div>
        `;
    }
}

declare global {
    interface HTMLElementTagNameMap {
        'fb-project-dialog': FBProjectDialog;
    }
}
