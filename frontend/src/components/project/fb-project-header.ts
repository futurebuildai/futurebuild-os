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

    private _handleNav(tab: string) {
        switch (tab) {
            case 'schedule':
                this.emit('fb-navigate', { view: 'project-schedule', id: this.projectId });
                break;
            case 'budget':
                this.emit('fb-navigate', { view: 'project-budget', id: this.projectId });
                break;
            case 'chat':
                this.emit('fb-navigate', { view: 'project-chat', id: this.projectId });
                break;
            case 'contacts':
                this.emit('fb-navigate', { view: 'project-contacts', id: this.projectId });
                break;
            default:
                this.emit('fb-navigate', { view: 'project', id: this.projectId });
        }
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

                <div class="nav-buttons">
                    <button
                        class="nav-btn ${this.activeTab === 'schedule' ? 'active' : ''}"
                        @click=${() => this._handleNav('schedule')}
                    >
                        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                            <rect x="3" y="4" width="18" height="18" rx="2" ry="2"/>
                            <line x1="16" y1="2" x2="16" y2="6"/>
                            <line x1="8" y1="2" x2="8" y2="6"/>
                            <line x1="3" y1="10" x2="21" y2="10"/>
                        </svg>
                        Schedule
                    </button>
                    <button
                        class="nav-btn ${this.activeTab === 'budget' ? 'active' : ''}"
                        @click=${() => this._handleNav('budget')}
                    >
                        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                            <line x1="12" y1="1" x2="12" y2="23"/>
                            <path d="M17 5H9.5a3.5 3.5 0 0 0 0 7h5a3.5 3.5 0 0 1 0 7H6"/>
                        </svg>
                        Budget
                    </button>
                    <button
                        class="nav-btn ${this.activeTab === 'chat' ? 'active' : ''}"
                        @click=${() => this._handleNav('chat')}
                    >
                        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                            <path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"/>
                        </svg>
                        Chat
                    </button>
                    <button
                        class="nav-btn ${this.activeTab === 'contacts' ? 'active' : ''}"
                        @click=${() => this._handleNav('contacts')}
                    >
                        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                            <path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"/>
                            <circle cx="9" cy="7" r="4"/>
                            <path d="M23 21v-2a4 4 0 0 0-3-3.87"/>
                            <path d="M16 3.13a4 4 0 0 1 0 7.75"/>
                        </svg>
                        Contacts
                    </button>
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
