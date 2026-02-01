/**
 * FBStatusCard - Reusable Status Card for Agent Outputs
 * See STEP_71_FOCUS_CARD.md
 *
 * Displays high-priority tasks and status updates from Agent 1.
 * Supports 4 visual variants: critical, success, warning, info.
 */
import { html, css, TemplateResult, nothing } from 'lit';
import { customElement, property } from 'lit/decorators.js';
import { FBElement } from '../base/FBElement';

export type StatusCardType = 'critical' | 'info' | 'success' | 'warning';

@customElement('fb-status-card')
export class FBStatusCard extends FBElement {
    static override styles = [
        FBElement.styles,
        css`
            :host {
                display: block;
            }

            .card {
                display: flex;
                align-items: flex-start;
                gap: var(--fb-spacing-md);
                background: var(--fb-bg-card);
                border: 1px solid var(--fb-border);
                border-radius: var(--fb-radius-lg);
                padding: var(--fb-spacing-md) var(--fb-spacing-lg);
                transition: border-color var(--fb-transition-fast),
                    background var(--fb-transition-fast);
            }

            /* --- Type Variants --- */

            :host([type='critical']) .card {
                border-color: var(--fb-error);
                background: color-mix(
                    in srgb,
                    var(--fb-error) 8%,
                    var(--fb-bg-card)
                );
            }

            :host([type='success']) .card {
                border-color: var(--fb-success);
                background: color-mix(
                    in srgb,
                    var(--fb-success) 8%,
                    var(--fb-bg-card)
                );
            }

            :host([type='warning']) .card {
                border-color: var(--fb-warning);
                background: color-mix(
                    in srgb,
                    var(--fb-warning) 8%,
                    var(--fb-bg-card)
                );
            }

            :host([type='info']) .card {
                border-color: var(--fb-info, #1565c0);
                background: color-mix(
                    in srgb,
                    var(--fb-info, #1565c0) 8%,
                    var(--fb-bg-card)
                );
            }

            /* --- Icon --- */

            .icon {
                flex-shrink: 0;
                width: 24px;
                height: 24px;
                display: flex;
                align-items: center;
                justify-content: center;
                font-size: 20px;
            }

            :host([type='critical']) .icon {
                color: var(--fb-error);
            }

            :host([type='success']) .icon {
                color: var(--fb-success);
            }

            :host([type='warning']) .icon {
                color: var(--fb-warning);
            }

            :host([type='info']) .icon {
                color: var(--fb-info, #1565c0);
            }

            /* --- Content --- */

            .content {
                flex: 1;
                min-width: 0;
            }

            .title {
                font-size: var(--fb-text-base);
                font-weight: 500;
                color: var(--fb-text-primary);
                margin: 0;
                overflow: hidden;
                text-overflow: ellipsis;
                white-space: nowrap;
            }

            .subtitle {
                font-size: var(--fb-text-sm);
                color: var(--fb-text-secondary);
                margin: var(--fb-spacing-xs) 0 0 0;
                overflow: hidden;
                text-overflow: ellipsis;
                white-space: nowrap;
            }
        `,
    ];

    @property({ type: String, reflect: true }) type: StatusCardType = 'info';

    @property({ type: String }) override title = '';

    @property({ type: String }) subtitle = '';

    @property({ type: String }) icon = '';

    override render(): TemplateResult {
        return html`
            <div class="card">
                ${this.icon
                    ? html`<span class="icon material-symbols-outlined"
                          >${this.icon}</span
                      >`
                    : nothing}
                <div class="content">
                    <p class="title">${this.title}</p>
                    ${this.subtitle
                        ? html`<p class="subtitle">${this.subtitle}</p>`
                        : nothing}
                </div>
            </div>
        `;
    }
}

declare global {
    interface HTMLElementTagNameMap {
        'fb-status-card': FBStatusCard;
    }
}
