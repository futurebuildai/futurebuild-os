/**
 * FBViewProjects - Projects List View
 * See PRODUCTION_PLAN.md Step 51.4
 *
 * Displays all projects for the current user/org.
 */
import { html, css, TemplateResult } from 'lit';
import { customElement } from 'lit/decorators.js';
import { FBViewElement } from '../base/FBViewElement';

@customElement('fb-view-projects')
export class FBViewProjects extends FBViewElement {
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
            }
        `,
    ];

    override onViewActive(): void {
        console.log('[FBViewProjects] View activated - would fetch projects');
    }

    override render(): TemplateResult {
        return html`
            <h1>Projects</h1>
            <div class="placeholder">
                Project cards will be rendered here. Connect to store.projects$ for real data.
            </div>
        `;
    }
}

declare global {
    interface HTMLElementTagNameMap {
        'fb-view-projects': FBViewProjects;
    }
}
