/**
 * fb-top-bar — V2 Top navigation bar.
 * See FRONTEND_V2_SPEC.md §4.2
 *
 * Renders: [Logo] [All] [Proj1] [Proj2] [+New] ... [Bell] [Avatar]
 * Project pills filter the feed. "All" shows cross-project feed.
 */
import { html, css } from 'lit';
import { customElement, property } from 'lit/decorators.js';
import { FBElement } from '../base/FBElement';
import type { ProjectPill } from '../../types/feed';

@customElement('fb-top-bar')
export class FBTopBar extends FBElement {
    static override styles = [
        FBElement.styles,
        css`
            :host {
                display: block;
                height: 56px;
                background: var(--fb-surface-1, #1a1a2e);
                border-bottom: 1px solid var(--fb-border, #2a2a3e);
                padding: 0 16px;
                z-index: 10;
            }

            .bar {
                display: flex;
                align-items: center;
                height: 100%;
                gap: 12px;
            }

            .logo {
                font-size: 18px;
                font-weight: 700;
                color: var(--fb-text-primary, #e0e0e0);
                white-space: nowrap;
                cursor: pointer;
                letter-spacing: -0.3px;
            }

            .logo span {
                color: var(--fb-accent, #6366f1);
            }

            .pills {
                display: flex;
                align-items: center;
                gap: 6px;
                overflow-x: auto;
                flex: 1;
                scrollbar-width: none;
                padding: 0 8px;
            }

            .pills::-webkit-scrollbar {
                display: none;
            }

            .pill {
                display: inline-flex;
                align-items: center;
                padding: 6px 14px;
                border-radius: 20px;
                font-size: 13px;
                font-weight: 500;
                white-space: nowrap;
                cursor: pointer;
                border: 1px solid var(--fb-border, #2a2a3e);
                background: transparent;
                color: var(--fb-text-secondary, #a0a0b0);
                transition: all 0.15s ease;
            }

            .pill:hover {
                background: var(--fb-surface-2, #252540);
                color: var(--fb-text-primary, #e0e0e0);
            }

            .pill[data-active] {
                background: var(--fb-accent, #6366f1);
                color: #fff;
                border-color: var(--fb-accent, #6366f1);
            }

            .pill-new {
                border-style: dashed;
                color: var(--fb-text-tertiary, #707080);
            }

            .pill-new:hover {
                border-color: var(--fb-accent, #6366f1);
                color: var(--fb-accent, #6366f1);
            }

            .actions {
                display: flex;
                align-items: center;
                gap: 8px;
                margin-left: auto;
            }

            .avatar {
                width: 32px;
                height: 32px;
                border-radius: 50%;
                background: var(--fb-accent, #6366f1);
                color: #fff;
                display: flex;
                align-items: center;
                justify-content: center;
                font-size: 13px;
                font-weight: 600;
                cursor: pointer;
            }

            .avatar:hover {
                opacity: 0.85;
            }

            @media (max-width: 768px) {
                :host {
                    padding: 0 12px;
                }

                .logo {
                    font-size: 16px;
                }

                .pill {
                    font-size: 12px;
                    padding: 5px 10px;
                }
            }
        `,
    ];

    @property({ type: Array }) projects: ProjectPill[] = [];
    @property({ type: String, attribute: 'active-filter' }) activeFilter: string | null = null;
    @property({ type: String, attribute: 'user-name' }) userName = '';

    private _handlePillClick(projectId: string | null) {
        this.emit('fb-filter-change', { projectId });
    }

    private _handleNewProject() {
        this.emit('fb-navigate', { view: 'onboard' });
    }

    private _handleLogoClick() {
        this.emit('fb-navigate', { view: 'home' });
    }

    private _handleAvatarClick() {
        this.emit('fb-navigate', { view: 'settings-profile' });
    }

    private _getInitials(): string {
        if (!this.userName) return '?';
        const parts = this.userName.trim().split(/\s+/);
        if (parts.length >= 2) {
            return (parts[0]![0]! + parts[parts.length - 1]![0]!).toUpperCase();
        }
        return this.userName.substring(0, 2).toUpperCase();
    }

    override render() {
        return html`
            <div class="bar">
                <div class="logo" @click=${this._handleLogoClick}>
                    Future<span>Build</span>
                </div>

                <div class="pills">
                    <button
                        class="pill"
                        ?data-active=${this.activeFilter === null}
                        @click=${() => this._handlePillClick(null)}
                    >
                        All
                    </button>

                    ${this.projects.map(
                        (p) => html`
                            <button
                                class="pill"
                                ?data-active=${this.activeFilter === p.id}
                                @click=${() => this._handlePillClick(p.id)}
                                title=${p.address}
                            >
                                ${p.name}
                            </button>
                        `
                    )}

                    <button class="pill pill-new" @click=${this._handleNewProject}>
                        + New
                    </button>
                </div>

                <div class="actions">
                    <fb-notification-bell></fb-notification-bell>
                    <div class="avatar" @click=${this._handleAvatarClick} title="Settings">
                        ${this._getInitials()}
                    </div>
                </div>
            </div>
        `;
    }
}

declare global {
    interface HTMLElementTagNameMap {
        'fb-top-bar': FBTopBar;
    }
}
