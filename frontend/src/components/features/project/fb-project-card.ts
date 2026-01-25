/**
 * FBProjectCard - Presentational Project Card Component
 * See PROJECT_ONBOARDING_SPEC.md Step 62.5
 *
 * Displays a project summary card with name, address, and status badge.
 * Clicking the card dispatches 'project-selected' event with project ID.
 *
 * @element fb-project-card
 */
import { html, css, TemplateResult } from 'lit';
import { customElement, property } from 'lit/decorators.js';
import { FBElement } from '../../base/FBElement';
import type { ProjectSummary } from '../../../store/types';

@customElement('fb-project-card')
export class FBProjectCard extends FBElement {
    static override styles = [
        FBElement.styles,
        css`
            :host {
                display: block;
            }

            .card {
                background: var(--fb-bg-card);
                border: 1px solid var(--fb-border);
                border-radius: var(--fb-radius-lg);
                padding: var(--fb-spacing-lg);
                cursor: pointer;
                transition: transform 0.15s ease, box-shadow 0.15s ease;
            }

            .card:hover {
                transform: translateY(-2px);
                box-shadow: var(--fb-shadow-md, 0 4px 12px rgba(0, 0, 0, 0.15));
            }

            .card:focus-visible {
                outline: 2px solid var(--fb-primary);
                outline-offset: 2px;
            }

            .name {
                font-size: var(--fb-text-lg);
                font-weight: 600;
                color: var(--fb-text-primary);
                margin: 0 0 var(--fb-spacing-xs) 0;
            }

            .address {
                font-size: var(--fb-text-sm);
                color: var(--fb-text-secondary);
                margin: 0 0 var(--fb-spacing-md) 0;
            }

            .footer {
                display: flex;
                align-items: center;
                justify-content: space-between;
            }

            .status {
                display: inline-flex;
                align-items: center;
                padding: var(--fb-spacing-xs) var(--fb-spacing-sm);
                border-radius: var(--fb-radius-full, 9999px);
                font-size: var(--fb-text-xs);
                font-weight: 500;
                text-transform: capitalize;
            }

            .status--planning {
                background: rgba(102, 126, 234, 0.15);
                color: var(--fb-primary, #667eea);
            }

            .status--in_progress,
            .status--active {
                background: rgba(72, 187, 120, 0.15);
                color: #48bb78;
            }

            .status--completed {
                background: rgba(160, 174, 192, 0.15);
                color: #a0aec0;
            }

            .status--on_hold {
                background: rgba(237, 137, 54, 0.15);
                color: #ed8936;
            }

            .progress {
                font-size: var(--fb-text-xs);
                color: var(--fb-text-muted);
            }
        `,
    ];

    @property({ type: Object })
    project!: ProjectSummary;

    private _handleClick(): void {
        this.emit('project-selected', { id: this.project.id });
    }

    private _handleKeyDown(e: KeyboardEvent): void {
        if (e.key === 'Enter' || e.key === ' ') {
            e.preventDefault();
            this._handleClick();
        }
    }

    private _getStatusClass(): string {
        const status = this.project.status.toLowerCase().replace(/\s+/g, '_');
        return `status--${status}`;
    }

    override render(): TemplateResult {
        return html`
            <article
                class="card"
                role="button"
                tabindex="0"
                @click=${this._handleClick.bind(this)}
                @keydown=${this._handleKeyDown.bind(this)}
                aria-label="Project: ${this.project.name}"
            >
                <h3 class="name">${this.project.name}</h3>
                <p class="address">${this.project.address}</p>
                <div class="footer">
                    <span class="status ${this._getStatusClass()}">
                        ${this.project.status}
                    </span>
                    <span class="progress">
                        ${this.project.completionPercentage}% complete
                    </span>
                </div>
            </article>
        `;
    }
}

declare global {
    interface HTMLElementTagNameMap {
        'fb-project-card': FBProjectCard;
    }
}
