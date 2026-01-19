/**
 * FBElement - Base Component Class for FutureBuild
 * See FRONTEND_SCOPE.md Section 4.1
 *
 * All FutureBuild web components extend this class to inherit:
 * - Shared foundational styles (box-sizing, etc.)
 * - Standardized event dispatching via emit()
 * - Future: SignalWatcher integration (Step 51.2)
 */
import { LitElement, css, CSSResultGroup } from 'lit';

/**
 * Abstract base class for all FutureBuild web components.
 *
 * @example
 * ```typescript
 * import { FBElement } from '@/components/base/FBElement';
 * import { html, css } from 'lit';
 * import { customElement } from 'lit/decorators.js';
 *
 * @customElement('fb-my-component')
 * export class FBMyComponent extends FBElement {
 *   static override styles = [
 *     FBElement.styles,
 *     css`:host { display: block; }`
 *   ];
 *
 *   override render() {
 *     return html`<button @click=${this.handleClick}>Click me</button>`;
 *   }
 *
 *   private handleClick() {
 *     this.emit('fb-action', { action: 'clicked' });
 *   }
 * }
 * ```
 */
export abstract class FBElement extends LitElement {
    /**
     * Shared foundational styles injected into every component's Shadow DOM.
     * Child components should merge these with their own styles:
     *
     * ```typescript
     * static override styles = [FBElement.styles, css`...`];
     * ```
     */
    static override styles: CSSResultGroup = css`
        :host {
            box-sizing: border-box;
        }

        :host *,
        :host *::before,
        :host *::after {
            box-sizing: inherit;
        }

        /* Shared Skeleton Loading Styles */
        .skeleton {
            background: linear-gradient(90deg, #f0f0f0 25%, #e0e0e0 50%, #f0f0f0 75%);
            background-size: 200% 100%;
            animation: shimmer 1.5s infinite;
            border-radius: 4px;
        }
        .skeleton-text { height: 1em; width: 100%; margin-bottom: 0.5em; display: block; }
        .skeleton-box { height: 100px; width: 100%; display: block; }
        .skeleton-bar { height: 8px; width: 100%; display: block; }

        @keyframes shimmer {
            0% { background-position: 200% 0; }
            100% { background-position: -200% 0; }
        }
    `;

    /**
     * Dispatches a typed CustomEvent that bubbles through Shadow DOM boundaries.
     *
     * Use this instead of raw `dispatchEvent` to ensure consistent event behavior
     * across the application. All events dispatched via `emit()`:
     * - Bubble up the DOM tree
     * - Cross Shadow DOM boundaries (composed: true)
     * - Carry optional typed detail payload
     *
     * @param name - The event name (conventionally prefixed with 'fb-')
     * @param detail - Optional data payload attached to event.detail
     * @returns The dispatched CustomEvent for chaining or inspection
     *
     * @example
     * ```typescript
     * // Dispatch with payload
     * this.emit('fb-item-selected', { itemId: 123, label: 'Task A' });
     *
     * // Listen from parent
     * myElement.addEventListener('fb-item-selected', (e) => {
     *   console.log(e.detail.itemId); // 123
     * });
     * ```
     */
    protected emit<T = unknown>(name: string, detail?: T): CustomEvent<T> {
        const event = new CustomEvent<T>(name, {
            bubbles: true,
            composed: true,
            detail: detail as T,
        });
        this.dispatchEvent(event);
        return event;
    }
}
