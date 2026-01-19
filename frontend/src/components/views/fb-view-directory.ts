/**
 * FBViewDirectory - Contacts/Team Directory View
 * See PRODUCTION_PLAN.md Step 51.4
 *
 * Displays project team members, subs, and vendors.
 */
import { html, css, TemplateResult } from 'lit';
import { customElement } from 'lit/decorators.js';
import { FBViewElement } from '../base/FBViewElement';

@customElement('fb-view-directory')
export class FBViewDirectory extends FBViewElement {
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
        console.log('[FBViewDirectory] View activated - would fetch contacts');
    }

    override render(): TemplateResult {
        return html`
            <h1>Directory</h1>
            <div class="placeholder">
                Contact cards and team roster will be rendered here.
            </div>
        `;
    }
}

declare global {
    interface HTMLElementTagNameMap {
        'fb-view-directory': FBViewDirectory;
    }
}
