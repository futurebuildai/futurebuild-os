/**
 * Component Registry - Centralized Custom Element Registration
 * See FRONTEND_SCOPE.md Section 4.2
 *
 * Provides a clean API for registering web components without
 * repeating `customElements.define()` boilerplate in every file.
 */

/**
 * Type for component constructor that extends HTMLElement.
 * This is the standard signature for custom element classes.
 */
type ComponentConstructor = CustomElementConstructor;

/**
 * Map of tag names to component classes.
 *
 * @example
 * ```typescript
 * const components: ComponentMap = {
 *   'fb-button': FBButton,
 *   'fb-input': FBInput,
 * };
 * ```
 */
export type ComponentMap = Record<string, ComponentConstructor>;

/**
 * Registers multiple custom elements from a component map.
 *
 * Safely handles double-registration by checking `customElements.get()`
 * before defining. This allows the function to be called multiple times
 * without throwing duplicate registration errors.
 *
 * @param components - Map of tag names to component classes
 * @returns Number of newly registered components (excludes already-registered)
 *
 * @example
 * ```typescript
 * import { registerComponents } from '@/components/registry';
 * import { FBDemoButton } from '@/components/base/demo-button';
 *
 * const count = registerComponents({
 *   'fb-demo-button': FBDemoButton,
 * });
 * console.log(`Registered ${count} components`);
 * ```
 */
export function registerComponents(components: ComponentMap): number {
    let registered = 0;

    for (const [tagName, componentClass] of Object.entries(components)) {
        // Prevent double-registration errors
        if (!customElements.get(tagName)) {
            customElements.define(tagName, componentClass);
            registered++;
        }
    }

    return registered;
}
