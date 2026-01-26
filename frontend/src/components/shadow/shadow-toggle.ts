/**
 * Shadow Toggle Component
 * Toggle button to switch between Standard and Shadow mode.
 * See SHADOW_VIEWER_specs.md Section 5.1
 */

import { html, css, type TemplateResult } from 'lit';
import { customElement, state } from 'lit/decorators.js';
import { effect } from '@preact/signals-core';
import { FBElement } from '../base/FBElement';
import { store } from '../../store/store';

@customElement('shadow-toggle')
export class ShadowToggle extends FBElement {
    static override styles = [
        FBElement.styles,
        css`
            :host {
                display: block;
            }

            .toggle-btn {
                display: flex;
                align-items: center;
                gap: var(--fb-spacing-sm, 8px);
                padding: var(--fb-spacing-sm, 8px) var(--fb-spacing-md, 12px);
                background: transparent;
                border: 1px solid var(--fb-border, #e5e7eb);
                border-radius: var(--fb-radius-md, 6px);
                color: var(--fb-text-secondary, #6b7280);
                cursor: pointer;
                font-size: var(--fb-text-sm, 14px);
                font-family: inherit;
                transition: all 0.15s ease;
                width: 100%;
                justify-content: center;
            }

            .toggle-btn:hover {
                background: var(--fb-bg-tertiary, #f3f4f6);
                color: var(--fb-text-primary, #111827);
            }

            .toggle-btn:focus-visible {
                outline: 2px solid var(--fb-focus-ring, #3b82f6);
                outline-offset: 2px;
            }

            .toggle-btn.active {
                background: #1a1a2e;
                border-color: #4f46e5;
                color: #818cf8;
            }

            .icon {
                width: 16px;
                height: 16px;
                flex-shrink: 0;
            }
        `,
    ];

    @state() private _isEnabled = false;

    private _disposeEffect?: () => void;

    override connectedCallback(): void {
        super.connectedCallback();
        this._disposeEffect = effect(() => {
            this._isEnabled = store.shadowModeEnabled$.value;
        });
    }

    override disconnectedCallback(): void {
        this._disposeEffect?.();
        super.disconnectedCallback();
    }

    private _handleClick(): void {
        store.actions.toggleShadowMode();
    }

    override render(): TemplateResult {
        return html`
            <button
                class="toggle-btn ${this._isEnabled ? 'active' : ''}"
                @click=${this._handleClick}
                aria-pressed="${this._isEnabled}"
                aria-label="Toggle Shadow Mode"
            >
                <svg class="icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <circle cx="12" cy="12" r="5" />
                    <path d="M12 1v2M12 21v2M4.22 4.22l1.42 1.42M18.36 18.36l1.42 1.42M1 12h2M21 12h2M4.22 19.78l1.42-1.42M18.36 5.64l1.42-1.42" />
                </svg>
                ${this._isEnabled ? 'Exit Shadow Mode' : 'Shadow Mode'}
            </button>
        `;
    }
}

declare global {
    interface HTMLElementTagNameMap {
        'shadow-toggle': ShadowToggle;
    }
}
