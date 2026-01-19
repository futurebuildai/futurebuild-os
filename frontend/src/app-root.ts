/**
 * FutureBuild App Root Component
 * See FRONTEND_SCOPE.md Section 4.1
 *
 * This is the top-level Lit component that wraps the application shell.
 */
import { LitElement, html, css, TemplateResult } from 'lit';
import { customElement } from 'lit/decorators.js';

// Import app shell (triggers component registration)
import './components/layout/fb-app-shell';

@customElement('app-root')
export class AppRoot extends LitElement {
  static override styles = css`
        :host {
            display: block;
            width: 100%;
            height: 100%;
        }
    `;

  override render(): TemplateResult {
    return html`
            <fb-app-shell>
                <div style="padding: var(--fb-spacing-xl); color: var(--fb-text-primary);">
                    <h1 style="font-size: var(--fb-text-2xl); font-weight: 600; margin-bottom: var(--fb-spacing-md);">
                        FutureBuild Command Center
                    </h1>
                    <p style="color: var(--fb-text-secondary);">
                        Select a view from the navigation rail.
                    </p>
                </div>
            </fb-app-shell>
        `;
  }
}

// TypeScript declaration for HTML type checking
declare global {
  interface HTMLElementTagNameMap {
    'app-root': AppRoot;
  }
}

