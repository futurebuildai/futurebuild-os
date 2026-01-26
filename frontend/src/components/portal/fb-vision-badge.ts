/**
 * FBVisionBadge - AI Verification Status Badge
 * See LAUNCH_PLAN.md P2: Field Portal (Mobile)
 *
 * Shows AI verification status for photos:
 * - pending: Awaiting verification
 * - verified: AI confirmed task completion
 * - flagged: AI detected potential issues
 * - failed: Verification could not complete
 */
import { html, css, TemplateResult } from 'lit';
import { customElement, property } from 'lit/decorators.js';
import { FBElement } from '../base/FBElement';

export type VisionStatus = 'pending' | 'verified' | 'flagged' | 'failed';

/**
 * Vision verification badge component.
 * @element fb-vision-badge
 */
@customElement('fb-vision-badge')
export class FBVisionBadge extends FBElement {
    static override styles = [
        FBElement.styles,
        css`
            :host {
                display: inline-flex;
            }

            .badge {
                display: inline-flex;
                align-items: center;
                gap: 6px;
                padding: 4px 10px;
                font-size: 12px;
                font-weight: 500;
                border-radius: 20px;
            }

            .badge svg {
                width: 14px;
                height: 14px;
            }

            .badge--pending {
                background: var(--fb-text-muted-alpha, rgba(102, 102, 102, 0.2));
                color: var(--fb-text-muted, #666);
            }

            .badge--verified {
                background: var(--fb-success-alpha, rgba(46, 125, 50, 0.2));
                color: var(--fb-success, #2e7d32);
            }

            .badge--flagged {
                background: var(--fb-warning-alpha, rgba(249, 168, 37, 0.2));
                color: var(--fb-warning, #f9a825);
            }

            .badge--failed {
                background: var(--fb-error-alpha, rgba(198, 40, 40, 0.2));
                color: var(--fb-error, #c62828);
            }

            .pulse {
                animation: pulse 2s infinite;
            }

            @keyframes pulse {
                0%, 100% { opacity: 1; }
                50% { opacity: 0.5; }
            }
        `,
    ];

    @property({ type: String }) status: VisionStatus = 'pending';

    private _renderIcon(): TemplateResult {
        switch (this.status) {
            case 'pending':
                return html`
                    <svg viewBox="0 0 24 24" fill="currentColor" class="pulse">
                        <path d="M12 4V1L8 5l4 4V6c3.31 0 6 2.69 6 6 0 1.01-.25 1.97-.7 2.8l1.46 1.46C19.54 15.03 20 13.57 20 12c0-4.42-3.58-8-8-8zm0 14c-3.31 0-6-2.69-6-6 0-1.01.25-1.97.7-2.8L5.24 7.74C4.46 8.97 4 10.43 4 12c0 4.42 3.58 8 8 8v3l4-4-4-4v3z"/>
                    </svg>
                `;
            case 'verified':
                return html`
                    <svg viewBox="0 0 24 24" fill="currentColor">
                        <path d="M9 16.17L4.83 12l-1.42 1.41L9 19 21 7l-1.41-1.41L9 16.17z"/>
                    </svg>
                `;
            case 'flagged':
                return html`
                    <svg viewBox="0 0 24 24" fill="currentColor">
                        <path d="M1 21h22L12 2 1 21zm12-3h-2v-2h2v2zm0-4h-2v-4h2v4z"/>
                    </svg>
                `;
            case 'failed':
                return html`
                    <svg viewBox="0 0 24 24" fill="currentColor">
                        <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm1 15h-2v-2h2v2zm0-4h-2V7h2v6z"/>
                    </svg>
                `;
        }
    }

    private _getLabel(): string {
        switch (this.status) {
            case 'pending':
                return 'Verifying...';
            case 'verified':
                return 'AI Verified';
            case 'flagged':
                return 'Needs Review';
            case 'failed':
                return 'Check Failed';
        }
    }

    override render(): TemplateResult {
        return html`
            <span class="badge badge--${this.status}">
                ${this._renderIcon()}
                ${this._getLabel()}
            </span>
        `;
    }
}

declare global {
    interface HTMLElementTagNameMap {
        'fb-vision-badge': FBVisionBadge;
    }
}
