/**
 * fb-empty-home — No-projects CTA leading to onboarding.
 * See FRONTEND_V2_SPEC.md §4, §2.3
 *
 * Displays a full-screen CTA when user has no projects.
 * "Your engine is ready. Start a project →"
 */
import { html, css } from 'lit';
import { customElement, property } from 'lit/decorators.js';
import { FBElement } from '../base/FBElement';

@customElement('fb-empty-home')
export class FBEmptyHome extends FBElement {
    static override styles = [
        FBElement.styles,
        css`
            :host {
                display: flex;
                flex-direction: column;
                align-items: center;
                justify-content: center;
                min-height: 60vh;
                text-align: center;
                padding: 40px 24px;
            }

            .icon {
                width: 80px;
                height: 80px;
                border-radius: 20px;
                background: linear-gradient(135deg, var(--fb-accent, #00FFA3), var(--fb-accent-light, #33FFB8));
                display: flex;
                align-items: center;
                justify-content: center;
                margin-bottom: 32px;
                box-shadow: 0 8px 32px rgba(0, 255, 163, 0.3);
            }

            .icon svg {
                width: 40px;
                height: 40px;
                color: #fff;
            }

            .title {
                font-size: 28px;
                font-weight: 700;
                color: var(--fb-text-primary, #F0F0F5);
                margin-bottom: 12px;
                line-height: 1.2;
            }

            .body {
                font-size: 16px;
                color: var(--fb-text-secondary, #8B8D98);
                line-height: 1.6;
                max-width: 360px;
                margin-bottom: 32px;
            }

            .cta {
                display: inline-flex;
                align-items: center;
                gap: 8px;
                padding: 14px 32px;
                border-radius: 10px;
                background: var(--fb-accent, #00FFA3);
                color: #fff;
                font-size: 16px;
                font-weight: 600;
                cursor: pointer;
                border: none;
                transition: all 0.2s ease;
                box-shadow: 0 4px 16px rgba(0, 255, 163, 0.3);
            }

            .cta:hover {
                transform: translateY(-1px);
                box-shadow: 0 6px 24px rgba(0, 255, 163, 0.4);
            }

            .cta:active {
                transform: translateY(0);
            }

            .cta:focus-visible {
                outline: 2px solid var(--fb-accent, #00FFA3);
                outline-offset: 3px;
            }

            .cta svg {
                width: 20px;
                height: 20px;
            }

            @media (max-width: 768px) {
                :host {
                    min-height: 50vh;
                    padding: 32px 16px;
                }

                .icon {
                    width: 64px;
                    height: 64px;
                    margin-bottom: 24px;
                }

                .icon svg {
                    width: 32px;
                    height: 32px;
                }

                .title {
                    font-size: 24px;
                }

                .body {
                    font-size: 15px;
                }

                .cta {
                    padding: 12px 28px;
                    font-size: 15px;
                }
            }
        `,
    ];

    /** Custom title (optional) */
    @property({ type: String }) title = 'Your engine is ready';

    /** Custom body text (optional) */
    @property({ type: String }) body =
        'Add your first project and I\'ll build your schedule, track your subs, and watch your deadlines.';

    /** CTA button text */
    @property({ type: String, attribute: 'cta-text' }) ctaText = 'Start a project';

    private _handleClick() {
        this.emit('fb-navigate', { view: 'onboard' });
    }

    override render() {
        return html`
            <div class="icon" aria-hidden="true">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                    <polygon points="13 2 3 14 12 14 11 22 21 10 12 10 13 2"/>
                </svg>
            </div>
            <h1 class="title">${this.title}</h1>
            <p class="body">${this.body}</p>
            <button class="cta" @click=${this._handleClick} aria-label="${this.ctaText}">
                ${this.ctaText}
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                    <line x1="5" y1="12" x2="19" y2="12"/>
                    <polyline points="12 5 19 12 12 19"/>
                </svg>
            </button>
        `;
    }
}

declare global {
    interface HTMLElementTagNameMap {
        'fb-empty-home': FBEmptyHome;
    }
}
