/**
 * FBStatusToggle - Mobile-Friendly Status Toggle Component
 * See LAUNCH_PLAN.md P2: Field Portal (Mobile)
 *
 * Large touch targets for mobile usage.
 * States: Pending | In Progress | Completed
 */
import { html, css, TemplateResult } from 'lit';
import { customElement, property } from 'lit/decorators.js';
import { FBElement } from '../base/FBElement';

export type TaskStatusValue = 'pending' | 'in_progress' | 'completed';

/**
 * Mobile-friendly status toggle component.
 * @element fb-status-toggle
 *
 * @fires fb-status-change - Fired when status changes
 */
@customElement('fb-status-toggle')
export class FBStatusToggle extends FBElement {
    static override styles = [
        FBElement.styles,
        css`
            :host {
                display: block;
            }

            .toggle-group {
                display: flex;
                flex-direction: column;
                gap: 12px;
            }

            .toggle-btn {
                display: flex;
                align-items: center;
                gap: 12px;
                padding: 16px 20px;
                background: var(--fb-bg-card, #111);
                border: 2px solid var(--fb-border, #333);
                border-radius: 12px;
                cursor: pointer;
                transition: all 0.2s ease;
                min-height: 64px;
            }

            .toggle-btn:hover:not([disabled]) {
                border-color: var(--fb-primary, #667eea);
            }

            .toggle-btn[aria-pressed="true"] {
                border-color: var(--fb-primary, #667eea);
                background: var(--fb-primary-alpha, rgba(102, 126, 234, 0.1));
            }

            .toggle-btn[disabled] {
                opacity: 0.5;
                cursor: not-allowed;
            }

            .icon {
                flex-shrink: 0;
                width: 32px;
                height: 32px;
                display: flex;
                align-items: center;
                justify-content: center;
                border-radius: 50%;
            }

            .icon--pending {
                background: var(--fb-warning-alpha, rgba(249, 168, 37, 0.2));
                color: var(--fb-warning, #f9a825);
            }

            .icon--in_progress {
                background: var(--fb-primary-alpha, rgba(102, 126, 234, 0.2));
                color: var(--fb-primary, #667eea);
            }

            .icon--completed {
                background: var(--fb-success-alpha, rgba(46, 125, 50, 0.2));
                color: var(--fb-success, #2e7d32);
            }

            .icon svg {
                width: 20px;
                height: 20px;
            }

            .label {
                flex: 1;
                text-align: left;
            }

            .label-text {
                color: var(--fb-text-primary, #fff);
                font-size: 16px;
                font-weight: 500;
                margin: 0;
            }

            .label-desc {
                color: var(--fb-text-secondary, #aaa);
                font-size: 14px;
                margin: 4px 0 0 0;
            }

            .checkmark {
                flex-shrink: 0;
                width: 24px;
                height: 24px;
                color: var(--fb-primary, #667eea);
                opacity: 0;
                transition: opacity 0.2s ease;
            }

            .toggle-btn[aria-pressed="true"] .checkmark {
                opacity: 1;
            }
        `,
    ];

    @property({ type: String }) value: TaskStatusValue = 'pending';
    @property({ type: Boolean }) disabled = false;

    private _handleSelect(status: TaskStatusValue): void {
        if (this.disabled || this.value === status) return;
        this.value = status;
        this.emit('fb-status-change', { status });
    }

    private _renderIcon(status: TaskStatusValue): TemplateResult {
        switch (status) {
            case 'pending':
                return html`
                    <svg viewBox="0 0 24 24" fill="currentColor">
                        <circle cx="12" cy="12" r="10" fill="none" stroke="currentColor" stroke-width="2"/>
                    </svg>
                `;
            case 'in_progress':
                return html`
                    <svg viewBox="0 0 24 24" fill="currentColor">
                        <path d="M12 2C6.5 2 2 6.5 2 12s4.5 10 10 10 10-4.5 10-10S17.5 2 12 2zm0 18c-4.41 0-8-3.59-8-8s3.59-8 8-8 8 3.59 8 8-3.59 8-8 8z"/>
                        <path d="M12 6c-3.31 0-6 2.69-6 6s2.69 6 6 6 6-2.69 6-6-2.69-6-6-6zm0 10c-2.21 0-4-1.79-4-4s1.79-4 4-4v8z"/>
                    </svg>
                `;
            case 'completed':
                return html`
                    <svg viewBox="0 0 24 24" fill="currentColor">
                        <path d="M9 16.17L4.83 12l-1.42 1.41L9 19 21 7l-1.41-1.41L9 16.17z"/>
                    </svg>
                `;
        }
    }

    override render(): TemplateResult {
        const options: Array<{ status: TaskStatusValue; label: string; desc: string }> = [
            { status: 'pending', label: 'Not Started', desc: 'Task has not begun yet' },
            { status: 'in_progress', label: 'In Progress', desc: 'Work is underway' },
            { status: 'completed', label: 'Completed', desc: 'Task is finished' },
        ];

        return html`
            <div class="toggle-group" role="radiogroup" aria-label="Task Status">
                ${options.map(
                    (opt) => html`
                        <button
                            class="toggle-btn"
                            role="radio"
                            aria-pressed="${this.value === opt.status}"
                            ?disabled=${this.disabled}
                            @click=${() => { this._handleSelect(opt.status); }}
                        >
                            <div class="icon icon--${opt.status}">
                                ${this._renderIcon(opt.status)}
                            </div>
                            <div class="label">
                                <p class="label-text">${opt.label}</p>
                                <p class="label-desc">${opt.desc}</p>
                            </div>
                            <svg class="checkmark" viewBox="0 0 24 24" fill="currentColor">
                                <path d="M9 16.17L4.83 12l-1.42 1.41L9 19 21 7l-1.41-1.41L9 16.17z"/>
                            </svg>
                        </button>
                    `
                )}
            </div>
        `;
    }
}

declare global {
    interface HTMLElementTagNameMap {
        'fb-status-toggle': FBStatusToggle;
    }
}
