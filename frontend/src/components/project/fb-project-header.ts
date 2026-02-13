/**
 * fb-project-header — Project detail header bar.
 * See FRONTEND_V2_SPEC.md §2.3.C
 *
 * Shows: ← Back | Project Name | Status | Completion% | Completion Date
 * Quick-nav: [Schedule] [Budget] [Chat] [Contacts]
 */
import { html, css, nothing } from 'lit';
import { customElement, property } from 'lit/decorators.js';
import { FBElement } from '../base/FBElement';

export type ProjectStatus = 'planning' | 'active' | 'on_hold' | 'completed';

@customElement('fb-project-header')
export class FBProjectHeader extends FBElement {
    static override styles = [
        FBElement.styles,
        css`
            :host {
                display: block;
                background: var(--fb-surface-1, #1a1a2e);
                border-bottom: 1px solid var(--fb-border, #2a2a3e);
                padding: 12px 20px;
            }

            .header {
                display: flex;
                align-items: center;
                gap: 16px;
                flex-wrap: wrap;
            }

            .back-btn {
                display: flex;
                align-items: center;
                gap: 6px;
                padding: 6px 12px;
                border-radius: 6px;
                background: transparent;
                border: 1px solid var(--fb-border, #2a2a3e);
                color: var(--fb-text-secondary, #a0a0b0);
                font-size: 13px;
                font-weight: 500;
                cursor: pointer;
                transition: all 0.15s ease;
            }

            .back-btn:hover {
                background: var(--fb-surface-2, #252540);
                color: var(--fb-text-primary, #e0e0e0);
            }

            .back-btn svg {
                width: 16px;
                height: 16px;
            }

            .divider {
                width: 1px;
                height: 24px;
                background: var(--fb-border, #2a2a3e);
            }

            .project-info {
                display: flex;
                align-items: center;
                gap: 16px;
                flex: 1;
            }

            .project-name {
                font-size: 16px;
                font-weight: 600;
                color: var(--fb-text-primary, #e0e0e0);
                white-space: nowrap;
                overflow: hidden;
                text-overflow: ellipsis;
                max-width: 300px;
            }

            .status-badge {
                padding: 4px 10px;
                border-radius: 12px;
                font-size: 11px;
                font-weight: 600;
                text-transform: uppercase;
                letter-spacing: 0.5px;
            }

            .status-badge.planning {
                background: rgba(59, 130, 246, 0.15);
                color: #3b82f6;
            }

            .status-badge.active {
                background: rgba(34, 197, 94, 0.15);
                color: #22c55e;
            }

            .status-badge.on_hold {
                background: rgba(245, 158, 11, 0.15);
                color: #f59e0b;
            }

            .status-badge.completed {
                background: rgba(156, 163, 175, 0.15);
                color: #9ca3af;
            }

            .completion {
                display: flex;
                align-items: center;
                gap: 8px;
                font-size: 13px;
                color: var(--fb-text-secondary, #a0a0b0);
            }

            .completion-bar {
                width: 80px;
                height: 6px;
                background: var(--fb-surface-2, #252540);
                border-radius: 3px;
                overflow: hidden;
            }

            .completion-fill {
                height: 100%;
                background: var(--fb-accent, #6366f1);
                border-radius: 3px;
                transition: width 0.3s ease;
            }

            .completion-date {
                font-size: 13px;
                color: var(--fb-text-tertiary, #707080);
            }

            .nav-buttons {
                display: flex;
                align-items: center;
                gap: 8px;
                margin-left: auto;
            }

            .nav-btn {
                display: flex;
                align-items: center;
                gap: 6px;
                padding: 8px 14px;
                border-radius: 6px;
                background: transparent;
                border: none;
                color: var(--fb-text-secondary, #a0a0b0);
                font-size: 13px;
                font-weight: 500;
                cursor: pointer;
                transition: all 0.15s ease;
            }

            .nav-btn:hover {
                background: var(--fb-surface-2, #252540);
                color: var(--fb-text-primary, #e0e0e0);
            }

            .nav-btn.active {
                background: var(--fb-accent, #6366f1);
                color: #fff;
            }

            .nav-btn svg {
                width: 16px;
                height: 16px;
            }

            @media (max-width: 768px) {
                :host {
                    padding: 10px 12px;
                }

                .header {
                    gap: 12px;
                }

                .project-info {
                    flex-wrap: wrap;
                    gap: 8px;
                }

                .project-name {
                    max-width: 200px;
                    font-size: 14px;
                }

                .completion-bar {
                    width: 60px;
                }

                .nav-buttons {
                    width: 100%;
                    justify-content: flex-start;
                    overflow-x: auto;
                    scrollbar-width: none;
                    margin-top: 8px;
                }

                .nav-buttons::-webkit-scrollbar {
                    display: none;
                }

                .nav-btn {
                    padding: 6px 12px;
                    font-size: 12px;
                }
            }
        `,
    ];

    /** Project ID */
    @property({ type: String, attribute: 'project-id' }) projectId = '';

    /** Project name/address */
    @property({ type: String }) name = '';

    /** Project status */
    @property({ type: String }) status: ProjectStatus = 'active';

    /** Completion percentage (0-100) */
    @property({ type: Number }) completion = 0;

    /** Estimated completion date */
    @property({ type: String, attribute: 'completion-date' }) completionDate = '';

    /** Currently active nav tab */
    @property({ type: String, attribute: 'active-tab' }) activeTab: 'feed' | 'schedule' | 'budget' | 'chat' | 'contacts' = 'feed';

    private _handleBack() {
        this.emit('fb-navigate', { view: 'home' });
    }

    private _getStatusLabel(status: ProjectStatus): string {
        const labels: Record<ProjectStatus, string> = {
            planning: 'Planning',
            active: 'Active',
            on_hold: 'On Hold',
            completed: 'Completed',
        };
        return labels[status] || 'Active';
    }

    private _formatDate(dateStr: string): string {
        if (!dateStr) return '';
        const date = new Date(dateStr);
        return date.toLocaleDateString('en-US', {
            month: 'short',
            day: 'numeric',
        });
    }

    override render() {
        return html`
            <div class="header">
                <button class="back-btn" @click=${this._handleBack} aria-label="Back to feed">
                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                        <polyline points="15 18 9 12 15 6"/>
                    </svg>
                    Back
                </button>

                <div class="divider"></div>

                <div class="project-info">
                    <span class="project-name" title=${this.name}>${this.name}</span>
                    <span class="status-badge ${this.status}">${this._getStatusLabel(this.status)}</span>

                    <div class="completion">
                        <div class="completion-bar">
                            <div class="completion-fill" style="width: ${Math.min(100, Math.max(0, this.completion))}%"></div>
                        </div>
                        <span>${this.completion}%</span>
                    </div>

                    ${this.completionDate ? html`
                        <span class="completion-date">${this._formatDate(this.completionDate)} completion</span>
                    ` : nothing}
                </div>
            </div>
        `;
    }
}

declare global {
    interface HTMLElementTagNameMap {
        'fb-project-header': FBProjectHeader;
    }
}
