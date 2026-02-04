/**
 * FBAdminShell - Platform Admin Layout Shell
 * 2-column grid (240px sidebar + 1fr content) with route dispatch for /admin/*.
 */
import { html, css, type TemplateResult } from 'lit';
import { customElement, state } from 'lit/decorators.js';
import { FBElement } from '../base/FBElement';

import './fb-admin-sidebar';
import './fb-admin-dashboard';
import '../views/fb-view-admin-invites';
import '../shadow/shadow-layout';
import '../feedback/fb-toast-container';

type AdminRoute = 'dashboard' | 'invitations' | 'shadow';

@customElement('fb-admin-shell')
export class FBAdminShell extends FBElement {
    static override styles = [
        FBElement.styles,
        css`
            :host {
                display: grid;
                grid-template-columns: 240px 1fr;
                height: 100vh;
                width: 100vw;
                overflow: hidden;
                background: var(--fb-bg-primary, #000);
                color: var(--fb-text-primary, #fff);
            }

            .content {
                overflow: hidden;
                display: flex;
                flex-direction: column;
            }

            @media (max-width: 768px) {
                :host {
                    grid-template-columns: 1fr;
                }

                fb-admin-sidebar {
                    display: none;
                }
            }
        `,
    ];

    @state() private _route: AdminRoute = 'dashboard';

    private _handlePopState = (): void => {
        this._resolveRoute();
    };

    override connectedCallback(): void {
        super.connectedCallback();
        this._resolveRoute();
        window.addEventListener('popstate', this._handlePopState);
    }

    override disconnectedCallback(): void {
        window.removeEventListener('popstate', this._handlePopState);
        super.disconnectedCallback();
    }

    private _resolveRoute(): void {
        const path = window.location.pathname;
        if (path === '/admin/invitations') {
            this._route = 'invitations';
        } else if (path === '/admin/shadow') {
            this._route = 'shadow';
        } else {
            this._route = 'dashboard';
        }
    }

    private _renderContent(): TemplateResult {
        switch (this._route) {
            case 'invitations':
                return html`<fb-view-admin-invites></fb-view-admin-invites>`;
            case 'shadow':
                return html`<shadow-layout></shadow-layout>`;
            default:
                return html`<fb-admin-dashboard></fb-admin-dashboard>`;
        }
    }

    override render(): TemplateResult {
        return html`
            <fb-admin-sidebar></fb-admin-sidebar>
            <div class="content">
                ${this._renderContent()}
            </div>
            <fb-toast-container></fb-toast-container>
        `;
    }
}

declare global {
    interface HTMLElementTagNameMap {
        'fb-admin-shell': FBAdminShell;
    }
}
