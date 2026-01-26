/**
 * FBTaskCard - Compact Task Card for Portal Dashboard
 * See LAUNCH_PLAN.md P2: Field Portal (Mobile)
 *
 * Displays task summary with:
 * - Task name, WBS code
 * - Status indicator
 * - Due date
 * - Tap to expand or change status
 */
import { html, css, TemplateResult, nothing } from 'lit';
import { customElement, property } from 'lit/decorators.js';
import { FBElement } from '../base/FBElement';

export interface TaskCardData {
    id: string;
    wbsCode: string;
    name: string;
    status: string;
    projectName: string;
    dueDate?: string;
    hasPhotos?: boolean;
}

/**
 * Compact task card for portal dashboard.
 * @element fb-task-card
 *
 * @fires fb-task-select - Fired when card is tapped
 */
@customElement('fb-task-card')
export class FBTaskCard extends FBElement {
    static override styles = [
        FBElement.styles,
        css`
            :host {
                display: block;
            }

            .card {
                display: flex;
                align-items: center;
                gap: 16px;
                padding: 16px;
                background: var(--fb-bg-card, #111);
                border: 1px solid var(--fb-border, #333);
                border-radius: 12px;
                cursor: pointer;
                transition: all 0.2s ease;
            }

            .card:hover {
                border-color: var(--fb-primary, #667eea);
            }

            .card:active {
                transform: scale(0.98);
            }

            .status-indicator {
                flex-shrink: 0;
                width: 12px;
                height: 12px;
                border-radius: 50%;
            }

            .status-indicator--pending {
                background: var(--fb-text-muted, #666);
            }

            .status-indicator--in_progress {
                background: var(--fb-primary, #667eea);
            }

            .status-indicator--completed {
                background: var(--fb-success, #2e7d32);
            }

            .status-indicator--blocked {
                background: var(--fb-error, #c62828);
            }

            .content {
                flex: 1;
                min-width: 0;
            }

            .header {
                display: flex;
                align-items: center;
                gap: 8px;
                margin-bottom: 4px;
            }

            .name {
                color: var(--fb-text-primary, #fff);
                font-size: 15px;
                font-weight: 500;
                margin: 0;
                white-space: nowrap;
                overflow: hidden;
                text-overflow: ellipsis;
            }

            .wbs {
                flex-shrink: 0;
                color: var(--fb-primary, #667eea);
                font-size: 11px;
                font-weight: 500;
                padding: 2px 6px;
                background: var(--fb-primary-alpha, rgba(102, 126, 234, 0.1));
                border-radius: 4px;
            }

            .meta {
                display: flex;
                align-items: center;
                gap: 12px;
                color: var(--fb-text-secondary, #aaa);
                font-size: 13px;
            }

            .meta-item {
                display: flex;
                align-items: center;
                gap: 4px;
            }

            .meta-item svg {
                width: 14px;
                height: 14px;
            }

            .chevron {
                flex-shrink: 0;
                width: 20px;
                height: 20px;
                color: var(--fb-text-muted, #666);
            }

            .photo-icon {
                color: var(--fb-success, #2e7d32);
            }
        `,
    ];

    @property({ type: Object }) task: TaskCardData | null = null;

    private _handleClick(): void {
        if (this.task) {
            this.emit('fb-task-select', { taskId: this.task.id });
        }
    }

    private _getStatusClass(status: string): string {
        switch (status.toLowerCase()) {
            case 'in_progress':
                return 'in_progress';
            case 'completed':
                return 'completed';
            case 'blocked':
            case 'delayed':
                return 'blocked';
            default:
                return 'pending';
        }
    }

    override render(): TemplateResult {
        if (!this.task) return html``;

        const { name, wbsCode, status, projectName, dueDate, hasPhotos } = this.task;
        const statusClass = this._getStatusClass(status);

        return html`
            <div class="card" @click=${this._handleClick.bind(this)}>
                <div class="status-indicator status-indicator--${statusClass}"></div>
                <div class="content">
                    <div class="header">
                        <h3 class="name">${name}</h3>
                        <span class="wbs">${wbsCode}</span>
                    </div>
                    <div class="meta">
                        <span class="meta-item">${projectName}</span>
                        ${dueDate
                            ? html`
                                  <span class="meta-item">
                                      <svg viewBox="0 0 24 24" fill="currentColor">
                                          <path d="M19 3h-1V1h-2v2H8V1H6v2H5c-1.11 0-1.99.9-1.99 2L3 19c0 1.1.89 2 2 2h14c1.1 0 2-.9 2-2V5c0-1.1-.9-2-2-2zm0 16H5V8h14v11z"/>
                                      </svg>
                                      ${dueDate}
                                  </span>
                              `
                            : nothing}
                        ${hasPhotos
                            ? html`
                                  <span class="meta-item photo-icon">
                                      <svg viewBox="0 0 24 24" fill="currentColor">
                                          <path d="M21 19V5c0-1.1-.9-2-2-2H5c-1.1 0-2 .9-2 2v14c0 1.1.9 2 2 2h14c1.1 0 2-.9 2-2zM8.5 13.5l2.5 3.01L14.5 12l4.5 6H5l3.5-4.5z"/>
                                      </svg>
                                  </span>
                              `
                            : nothing}
                    </div>
                </div>
                <svg class="chevron" viewBox="0 0 24 24" fill="currentColor">
                    <path d="M10 6L8.59 7.41 13.17 12l-4.58 4.59L10 18l6-6z"/>
                </svg>
            </div>
        `;
    }
}

declare global {
    interface HTMLElementTagNameMap {
        'fb-task-card': FBTaskCard;
    }
}
