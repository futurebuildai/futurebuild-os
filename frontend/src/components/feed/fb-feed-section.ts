/**
 * fb-feed-section — Time horizon section header.
 * See FRONTEND_V2_SPEC.md §4
 *
 * Renders a decorative section divider with label:
 * [── TODAY ──────────────────────────────────────]
 */
import { html, css } from 'lit';
import { customElement, property } from 'lit/decorators.js';
import { FBElement } from '../base/FBElement';

export type FeedSectionHorizon = 'today' | 'this_week' | 'horizon';

const HORIZON_LABELS: Record<FeedSectionHorizon, string> = {
    today: 'TODAY',
    this_week: 'THIS WEEK',
    horizon: 'ON THE HORIZON',
};

@customElement('fb-feed-section')
export class FBFeedSection extends FBElement {
    static override styles = [
        FBElement.styles,
        css`
            :host {
                display: block;
                margin: 28px 0 16px;
            }

            :host(:first-of-type) {
                margin-top: 0;
            }

            .header {
                display: flex;
                align-items: center;
                gap: 12px;
            }

            .line {
                flex: 0 0 12px;
                height: 1px;
                background: var(--fb-border, #3a3a4a);
            }

            .line-end {
                flex: 1;
            }

            .label {
                font-size: 11px;
                font-weight: 600;
                text-transform: uppercase;
                letter-spacing: 1px;
                color: var(--fb-text-tertiary, #5A5B66);
                white-space: nowrap;
            }

            .content {
                display: flex;
                flex-direction: column;
                gap: 12px;
                margin-top: 12px;
            }

            /* Horizon-specific accents */
            :host([horizon="today"]) .label {
                color: var(--fb-accent, #00FFA3);
            }

            :host([horizon="today"]) .line {
                background: var(--fb-accent, #00FFA3);
                opacity: 0.4;
            }

            @media (max-width: 768px) {
                :host {
                    margin: 24px 0 12px;
                }

                .label {
                    font-size: 10px;
                }
            }
        `,
    ];

    /** Time horizon identifier */
    @property({ type: String, reflect: true }) horizon: FeedSectionHorizon = 'today';

    /** Optional custom label (overrides default) */
    @property({ type: String }) label?: string;

    /** Card count for aria */
    @property({ type: Number, attribute: 'card-count' }) cardCount = 0;

    override render() {
        const displayLabel = this.label ?? HORIZON_LABELS[this.horizon] ?? 'UPCOMING';

        return html`
            <section
                role="region"
                aria-label="${displayLabel} section"
                aria-description="${this.cardCount} card${this.cardCount !== 1 ? 's' : ''}"
            >
                <div class="header">
                    <div class="line"></div>
                    <span class="label">${displayLabel}</span>
                    <div class="line line-end"></div>
                </div>
                <div class="content" role="feed">
                    <slot></slot>
                </div>
            </section>
        `;
    }
}

declare global {
    interface HTMLElementTagNameMap {
        'fb-feed-section': FBFeedSection;
    }
}
