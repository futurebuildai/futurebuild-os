/**
 * FBViewSchedule - Gantt/Schedule View
 * See PRODUCTION_PLAN.md Step 51.4
 *
 * Displays the project timeline and Gantt chart.
 */
import { html, css, TemplateResult } from 'lit';
import { customElement } from 'lit/decorators.js';
import { FBViewElement } from '../base/FBViewElement';

@customElement('fb-view-schedule')
export class FBViewSchedule extends FBViewElement {
    static override styles = [
        FBViewElement.styles,
        css`
            :host {
                padding: var(--fb-spacing-xl);
            }

            h1 {
                font-size: var(--fb-text-2xl);
                font-weight: 600;
                color: var(--fb-text-primary);
                margin: 0 0 var(--fb-spacing-xl) 0;
            }

            .placeholder {
                background: var(--fb-bg-card);
                border: 1px dashed var(--fb-border);
                border-radius: var(--fb-radius-lg);
                padding: var(--fb-spacing-2xl);
                text-align: center;
                color: var(--fb-text-muted);
                min-height: 400px;
                display: flex;
                align-items: center;
                justify-content: center;
            }
        `,
    ];

    override onViewActive(): void {
        console.log('[FBViewSchedule] View activated - would fetch schedule');
    }

    override render(): TemplateResult {
        return html`
            <h1>Schedule</h1>
            <div class="placeholder">
                Gantt chart component will be built in a future step. Will use Canvas/SVG for performance.
            </div>
        `;
    }
}

declare global {
    interface HTMLElementTagNameMap {
        'fb-view-schedule': FBViewSchedule;
    }
}
