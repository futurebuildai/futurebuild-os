/**
 * FBVisionBadge - AI Verification Status Badge
 * See STEP_85_VISION_BADGES.md Section 1.1
 *
 * Badge Variants:
 * - verifying: Yellow with spinner animation (AI is processing)
 * - verified: Green with check icon (AI confirmed)
 * - flagged: Red with warning icon (AI detected issues)
 * - failed: Red with error icon (Verification could not complete)
 *
 * @fires fb-badge-click - When user clicks the badge (detail includes status + summary)
 */
import { html, css, TemplateResult, nothing } from 'lit';
import { customElement, property } from 'lit/decorators.js';
import { FBElement } from '../base/FBElement';

export type VisionBadgeStatus = 'verifying' | 'verified' | 'flagged' | 'failed';

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
                font-weight: 600;
                border-radius: 20px;
                cursor: default;
                user-select: none;
                white-space: nowrap;
            }

            .badge svg {
                width: 14px;
                height: 14px;
                flex-shrink: 0;
            }

            /* Verifying: Yellow/Warning — AI is processing */
            .badge--verifying {
                background: rgba(245, 158, 11, 0.15);
                color: #d97706;
            }

            /* Verified: Green/Success — AI confirmed */
            .badge--verified {
                background: rgba(5, 150, 105, 0.15);
                color: #059669;
                cursor: pointer;
            }
            .badge--verified:hover { background: rgba(5, 150, 105, 0.25); }

            /* Flagged: Red/Danger — AI detected issues */
            .badge--flagged {
                background: rgba(220, 38, 38, 0.15);
                color: #dc2626;
                cursor: pointer;
            }
            .badge--flagged:hover { background: rgba(220, 38, 38, 0.25); }

            /* Failed: Gray/Muted — Verification could not complete */
            .badge--failed {
                background: rgba(107, 114, 128, 0.15);
                color: #6b7280;
            }

            /* Spinner animation for verifying state */
            .spin {
                animation: spin 1s linear infinite;
            }

            @keyframes spin {
                to { transform: rotate(360deg); }
            }

            /* Overlay positioning (when used inside gallery) */
            :host([overlay]) {
                position: absolute;
                top: 6px;
                right: 6px;
                z-index: 2;
            }
        `,
    ];

    /** Current verification status */
    @property({ type: String }) status: VisionBadgeStatus = 'verifying';

    /** Optional summary text (shown on click for flagged/verified) */
    @property({ type: String }) summary = '';

    private _handleClick(): void {
        if (this.status === 'flagged' || this.status === 'verified') {
            this.emit('fb-badge-click', {
                status: this.status,
                summary: this.summary,
            });
        }
    }

    private _renderIcon(): TemplateResult {
        switch (this.status) {
            case 'verifying':
                return html`
                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" class="spin">
                        <path d="M12 2a10 10 0 0 1 10 10" stroke-linecap="round"/>
                    </svg>
                `;
            case 'verified':
                return html`
                    <svg viewBox="0 0 24 24" fill="currentColor">
                        <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm-2 15l-5-5 1.41-1.41L10 14.17l7.59-7.59L19 8l-9 9z"/>
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
            case 'verifying':
                return 'Verifying...';
            case 'verified':
                return 'Verified';
            case 'flagged':
                return 'Flagged';
            case 'failed':
                return 'Failed';
        }
    }

    override render(): TemplateResult {
        const isClickable = this.status === 'flagged' || this.status === 'verified';
        return html`
            <span
                class="badge badge--${this.status}"
                @click=${this._handleClick}
                title=${isClickable && this.summary ? this.summary : nothing}
            >
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
