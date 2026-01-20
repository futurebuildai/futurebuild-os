/**
 * FBArtifactModal - Full-view scrollable artifact popout
 * Step 59.5: UX Enhancements
 *
 * Displays an artifact in a centered modal overlay with:
 * - 80vh/80vw max dimensions
 * - Scrollable content area
 * - Close via X button, Escape key, or backdrop click
 */
import { html, css, TemplateResult, nothing } from 'lit';
import { customElement, state } from 'lit/decorators.js';
import { effect } from '@preact/signals-core';
import { FBElement } from '../base/FBElement';
import { store } from '../../store/store';
import type { ArtifactPayload } from '../../services/realtime/types';
import { normalizeArtifactType, getArtifactIcon } from '../../utils/artifact-helpers';
import { validateArtifactData } from '../../utils/artifact-validation';
import '../base/fb-error-boundary';

@customElement('fb-artifact-modal')
export class FBArtifactModal extends FBElement {
    static override styles = [
        FBElement.styles,
        css`
            :host {
                position: fixed;
                top: 0;
                left: 0;
                width: 100vw;
                height: 100vh;
                z-index: var(--fb-z-modal, 1000);
                display: flex;
                align-items: center;
                justify-content: center;
                animation: fadeIn 0.2s ease;
            }

            @keyframes fadeIn {
                from { opacity: 0; }
                to { opacity: 1; }
            }

            .backdrop {
                position: absolute;
                top: 0;
                left: 0;
                width: 100%;
                height: 100%;
                background: rgba(0, 0, 0, 0.7);
                backdrop-filter: blur(4px);
            }

            .modal {
                position: relative;
                width: 80vw;
                max-width: 1200px;
                height: 80vh;
                max-height: 900px;
                background: var(--fb-bg-card);
                border: 1px solid var(--fb-border);
                border-radius: var(--fb-radius-lg);
                display: flex;
                flex-direction: column;
                overflow: hidden;
                box-shadow: 0 20px 60px rgba(0, 0, 0, 0.5);
                animation: slideUp 0.25s ease;
            }

            @keyframes slideUp {
                from {
                    opacity: 0;
                    transform: translateY(20px);
                }
                to {
                    opacity: 1;
                    transform: translateY(0);
                }
            }

            .modal-header {
                display: flex;
                align-items: center;
                justify-content: space-between;
                padding: var(--fb-spacing-md) var(--fb-spacing-lg);
                border-bottom: 1px solid var(--fb-border);
                background: var(--fb-bg-tertiary);
            }

            .modal-title {
                display: flex;
                align-items: center;
                gap: var(--fb-spacing-sm);
                font-size: var(--fb-text-lg);
                font-weight: 600;
                color: var(--fb-text-primary);
            }

            .modal-type {
                font-size: var(--fb-text-xs);
                color: var(--fb-text-muted);
                text-transform: uppercase;
                padding: var(--fb-spacing-xs) var(--fb-spacing-sm);
                background: var(--fb-bg-secondary);
                border-radius: var(--fb-radius-sm);
            }

            .close-btn {
                display: flex;
                align-items: center;
                justify-content: center;
                width: 36px;
                height: 36px;
                border: none;
                background: transparent;
                color: var(--fb-text-muted);
                cursor: pointer;
                border-radius: var(--fb-radius-sm);
                font-size: 20px;
                transition: background 0.15s ease, color 0.15s ease;
            }

            .close-btn:hover {
                background: var(--fb-bg-secondary);
                color: var(--fb-text-primary);
            }

            .close-btn:focus-visible {
                outline: 2px solid var(--fb-primary);
                outline-offset: 2px;
            }

            .modal-content {
                flex: 1;
                overflow-y: auto;
                padding: var(--fb-spacing-lg);
            }

            .modal-content fb-artifact-gantt,
            .modal-content fb-artifact-budget,
            .modal-content fb-artifact-invoice {
                width: 100%;
                min-height: 400px;
            }

            .error-message {
                display: flex;
                flex-direction: column;
                align-items: center;
                justify-content: center;
                height: 200px;
                color: var(--fb-text-muted);
            }

            .error-icon {
                font-size: 32px;
                margin-bottom: var(--fb-spacing-md);
            }
        `,
    ];

    @state() private _artifact: ArtifactPayload | null = null;

    private _disposeEffect: (() => void) | null = null;

    override connectedCallback(): void {
        super.connectedCallback();

        this._disposeEffect = effect(() => {
            this._artifact = store.popoutArtifact$.value;
        });

        // Listen for Escape key
        document.addEventListener('keydown', this._handleKeyDown);
    }

    override disconnectedCallback(): void {
        if (this._disposeEffect) this._disposeEffect();
        document.removeEventListener('keydown', this._handleKeyDown);
        super.disconnectedCallback();
    }

    private _handleKeyDown = (e: KeyboardEvent): void => {
        if (e.key === 'Escape' && this._artifact) {
            this._close();
        }
    };

    private _close(): void {
        store.actions.setPopoutArtifact(null);
    }

    private _handleBackdropClick = (e: MouseEvent): void => {
        if ((e.target as HTMLElement).classList.contains('backdrop')) {
            this._close();
        }
    };

    private _renderArtifactContent(): TemplateResult {
        if (!this._artifact) {
            return html`<div class="error-message">No artifact selected</div>`;
        }

        const validation = validateArtifactData(this._artifact.type, this._artifact.data);
        if (!validation.valid) {
            return html`
                <fb-error-boundary
                    .hasError=${true}
                    .errorMessage=${validation.error ?? 'Invalid artifact data'}
                ></fb-error-boundary>
            `;
        }

        const normalizedType = normalizeArtifactType(this._artifact.type);
        const data = this._artifact.data;

        switch (normalizedType) {
            case 'gantt':
                return html`<fb-artifact-gantt .data=${data}></fb-artifact-gantt>`;
            case 'budget':
                return html`<fb-artifact-budget .data=${data}></fb-artifact-budget>`;
            case 'invoice':
                return html`<fb-artifact-invoice .data=${data}></fb-artifact-invoice>`;
            default:
                return html`
                    <div class="error-message">
                        <span class="error-icon">📄</span>
                        <span>Preview not available for "${this._artifact.type}"</span>
                    </div>
                `;
        }
    }

    override render(): TemplateResult {
        if (!this._artifact) {
            return html`${nothing}`;
        }

        const normalizedType = normalizeArtifactType(this._artifact.type);
        const icon = getArtifactIcon(normalizedType);

        return html`
            <div class="backdrop" @click=${this._handleBackdropClick} role="presentation"></div>
            <div
                class="modal"
                role="dialog"
                aria-modal="true"
                aria-labelledby="modal-title"
            >
                <header class="modal-header">
                    <div class="modal-title">
                        <span>${icon}</span>
                        <span id="modal-title">${this._artifact.title}</span>
                        <span class="modal-type">${normalizedType}</span>
                    </div>
                    <button
                        class="close-btn"
                        @click=${this._close.bind(this)}
                        aria-label="Close modal"
                    >
                        ✕
                    </button>
                </header>
                <div class="modal-content">
                    ${this._renderArtifactContent()}
                </div>
            </div>
        `;
    }
}

declare global {
    interface HTMLElementTagNameMap {
        'fb-artifact-modal': FBArtifactModal;
    }
}
