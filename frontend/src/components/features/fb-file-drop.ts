/**
 * FBFileDrop - Global drag-and-drop overlay.
 * See FRONTEND_SCOPE.md Section 8.2
 * Step 56: Drag-and-Drop Ingestion
 *
 * Displays a full-screen glassmorphism overlay when files are being
 * dragged over the application. Visibility is controlled by store.isDragging$.
 *
 * The overlay uses CSS transitions for smooth fade-in/out and includes
 * a pulsing animation to guide the user to drop the file.
 *
 * @element fb-file-drop
 */
import { html, css, TemplateResult } from 'lit';
import { customElement, state } from 'lit/decorators.js';
import { effect } from '@preact/signals-core';
import { FBElement } from '../base/FBElement';
import { store } from '../../store/store';

@customElement('fb-file-drop')
export class FBFileDrop extends FBElement {
    static override styles = [
        FBElement.styles,
        css`
            :host {
                position: fixed;
                inset: 0;
                display: flex;
                align-items: center;
                justify-content: center;
                background: rgba(102, 126, 234, 0.1);
                backdrop-filter: blur(4px);
                z-index: var(--fb-z-overlay, 1000);
                opacity: 0;
                visibility: hidden;
                transition: opacity 0.15s ease, visibility 0.15s ease;
                pointer-events: none;
            }

            :host([active]) {
                opacity: 1;
                visibility: visible;
            }

            .drop-zone {
                display: flex;
                flex-direction: column;
                align-items: center;
                justify-content: center;
                gap: var(--fb-spacing-lg);
                padding: var(--fb-spacing-2xl) calc(var(--fb-spacing-2xl) * 2);
                border: 3px dashed var(--fb-primary);
                border-radius: var(--fb-radius-xl);
                background: rgba(255, 255, 255, 0.05);
                text-align: center;
                animation: pulse 1.5s ease-in-out infinite;
            }

            @keyframes pulse {
                0%, 100% {
                    transform: scale(1);
                    opacity: 1;
                }
                50% {
                    transform: scale(1.02);
                    opacity: 0.9;
                }
            }

            .icon {
                width: 64px;
                height: 64px;
                color: var(--fb-primary);
            }

            .title {
                font-size: var(--fb-text-xl);
                font-weight: 600;
                color: var(--fb-text-primary);
                margin: 0;
            }

            .subtitle {
                font-size: var(--fb-text-sm);
                color: var(--fb-text-secondary);
                margin: 0;
            }
        `,
    ];

    @state() private _isActive = false;

    private _disposeEffect?: () => void;

    override connectedCallback(): void {
        super.connectedCallback();
        this._disposeEffect = effect(() => {
            this._isActive = store.isDragging$.value;
            if (this._isActive) {
                this.setAttribute('active', '');
            } else {
                this.removeAttribute('active');
            }
        });
    }

    override disconnectedCallback(): void {
        this._disposeEffect?.();
        super.disconnectedCallback();
    }

    override render(): TemplateResult {
        return html`
            <div
                class="drop-zone"
                role="dialog"
                aria-label="Drop files here to upload"
                aria-hidden="${String(!this._isActive)}"
            >
                <svg class="icon" viewBox="0 0 24 24" aria-hidden="true">
                    <path
                        fill="currentColor"
                        d="M19.35 10.04C18.67 6.59 15.64 4 12 4 9.11 4 6.6 5.64 5.35 8.04 2.34 8.36 0 10.91 0 14c0 3.31 2.69 6 6 6h13c2.76 0 5-2.24 5-5 0-2.64-2.05-4.78-4.65-4.96zM14 13v4h-4v-4H7l5-5 5 5h-3z"
                    />
                </svg>
                <p class="title">Drop files to upload</p>
                <p class="subtitle">PDF, JPEG, PNG, GIF, or WebP</p>
            </div>
        `;
    }
}

declare global {
    interface HTMLElementTagNameMap {
        'fb-file-drop': FBFileDrop;
    }
}
