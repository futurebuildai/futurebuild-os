/**
 * FBViewTeam - Team Management View
 * See PHASE_12_PRD.md Step 80: Organization Manager
 *
 * Mounts the Clerk OrganizationProfile component for managing
 * organization members and invitations.
 */
import { html, css, TemplateResult } from 'lit';
import { customElement, query } from 'lit/decorators.js';
import { FBViewElement } from '../base/FBViewElement';
import { clerkService } from '../../services/clerk';

@customElement('fb-view-team')
export class FBViewTeam extends FBViewElement {
    static override styles = [
        FBViewElement.styles,
        css`
            :host {
                display: flex;
                flex-direction: column;
                background: var(--fb-bg-primary);
                padding: var(--fb-spacing-xl);
            }

            .header {
                margin-bottom: var(--fb-spacing-xl);
            }

            .title {
                font-size: var(--fb-text-2xl);
                font-weight: 600;
                color: var(--fb-text-primary);
            }

            .subtitle {
                font-size: var(--fb-text-sm);
                color: var(--fb-text-secondary);
                margin-top: var(--fb-spacing-xs);
            }

            .settings-card {
                background: var(--fb-bg-card);
                border: 1px solid var(--fb-border);
                border-radius: var(--fb-radius-lg);
                padding: var(--fb-spacing-xl);
            }
        `,
    ];

    @query('#clerk-org-profile') private _orgProfileContainer: HTMLDivElement | undefined;

    protected override firstUpdated(): void {
        if (this._orgProfileContainer) {
            clerkService.mountOrganizationProfile(this._orgProfileContainer);
        }
    }

    override disconnectedCallback(): void {
        if (this._orgProfileContainer) {
            clerkService.unmountOrganizationProfile(this._orgProfileContainer);
        }
        super.disconnectedCallback();
    }

    override render(): TemplateResult {
        return html`
            <div class="header">
                <div class="title">Team</div>
                <div class="subtitle">Manage your organization members and invitations</div>
            </div>

            <div class="settings-card">
                <div id="clerk-org-profile"></div>
            </div>
        `;
    }
}

declare global {
    interface HTMLElementTagNameMap {
        'fb-view-team': FBViewTeam;
    }
}
