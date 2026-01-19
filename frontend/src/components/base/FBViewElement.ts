/**
 * FBViewElement - Base Class for View Components
 * See FRONTEND_SCOPE.md Section 6.1, PRODUCTION_PLAN.md Step 51.4
 *
 * All FutureBuild view components extend this class to inherit:
 * - Strict viewport containment (height/width 100%)
 * - Internal scrolling (overflow-y: auto)
 * - Lifecycle hook for view activation
 *
 * L7 Constraint: Views manage their own scroll. The window NEVER scrolls.
 */
import { css, CSSResultGroup, PropertyValues } from 'lit';
import { property } from 'lit/decorators.js';
import { FBElement } from './FBElement';

/**
 * Abstract base class for all FutureBuild view components.
 *
 * @example
 * ```typescript
 * import { FBViewElement } from '@/components/base/FBViewElement';
 * import { html } from 'lit';
 * import { customElement } from 'lit/decorators.js';
 *
 * @customElement('fb-view-dashboard')
 * export class FBViewDashboard extends FBViewElement {
 *   override render() {
 *     return html`<h1>Dashboard</h1>`;
 *   }
 *
 *   override onViewActive(): void {
 *     console.log('Dashboard is now visible');
 *     // Fetch data, start animations, etc.
 *   }
 * }
 * ```
 */
export abstract class FBViewElement extends FBElement {
    /**
     * View-specific styles enforcing viewport containment.
     * Child views should merge these with their own styles.
     */
    static override styles: CSSResultGroup = [
        FBElement.styles,
        css`
            :host {
                display: block;
                width: 100%;
                height: 100%;
                overflow-y: auto;
                overflow-x: hidden;
                box-sizing: border-box;
                position: relative;
            }

            /* Hide scrollbar for Chrome, Safari, Opera */
            :host::-webkit-scrollbar {
                width: 8px;
            }

            :host::-webkit-scrollbar-track {
                background: transparent;
            }

            :host::-webkit-scrollbar-thumb {
                background: var(--fb-border);
                border-radius: var(--fb-radius-sm);
            }

            :host::-webkit-scrollbar-thumb:hover {
                background: var(--fb-text-muted);
            }
        `,
    ];

    /**
     * Whether this view is currently the active (visible) view.
     * Managed by the router in FBAppShell.
     */
    @property({ type: Boolean, reflect: true })
    active = false;

    /**
     * Lifecycle hook called when the view becomes active.
     * Override in subclasses to perform initialization when the view appears.
     *
     * Common use cases:
     * - Fetching data from API
     * - Starting animations or timers
     * - Focusing an input element
     */
    protected onViewActive(): void {
        // Override in subclasses
    }

    /**
     * Watch for `active` property changes to trigger lifecycle hook.
     */
    protected override updated(changedProperties: PropertyValues): void {
        super.updated(changedProperties);

        if (changedProperties.has('active') && this.active) {
            this.onViewActive();
        }
    }
}
