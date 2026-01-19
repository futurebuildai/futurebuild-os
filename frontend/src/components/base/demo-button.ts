/**
 * FBDemoButton - Verification Component for FBElement Base Class
 * See PRODUCTION_PLAN.md Step 51.1 (Litmus Test)
 *
 * This temporary component verifies:
 * 1. CSS variable inheritance from global design tokens
 * 2. FBElement.emit() helper works correctly
 * 3. Shadow DOM encapsulation
 *
 * DELETE THIS FILE after Step 51.1 verification is complete.
 */
import { html, css, TemplateResult, CSSResultGroup } from 'lit';
import { FBElement } from './FBElement';

/**
 * Demo button component for testing FBElement infrastructure.
 *
 * @fires fb-click - Emitted when button is clicked, detail contains timestamp
 */
export class FBDemoButton extends FBElement {
    static override styles: CSSResultGroup = [
        FBElement.styles,
        css`
            :host {
                display: inline-block;
            }

            button {
                padding: var(--fb-spacing-md) var(--fb-spacing-xl);
                font-family: var(--fb-font-family);
                font-size: var(--fb-text-base);
                font-weight: 600;
                color: var(--fb-text-primary);
                background: var(--fb-dawn-gradient);
                border: none;
                border-radius: var(--fb-radius-md);
                cursor: pointer;
                transition: transform var(--fb-transition-fast),
                    box-shadow var(--fb-transition-fast);
            }

            button:hover {
                transform: translateY(-2px);
                box-shadow: var(--fb-shadow-md);
            }

            button:active {
                transform: translateY(0);
            }

            button:focus-visible {
                outline: 2px solid var(--fb-primary);
                outline-offset: 2px;
            }
        `,
    ];

    override render(): TemplateResult {
        return html`
            <button @click=${(): void => { this.handleClick(); }}>
                <slot>Demo Button</slot>
            </button>
        `;
    }

    /**
     * Handles click event and emits a custom event using FBElement.emit()
     */
    private handleClick(): void {
        // Test the emit() helper with typed payload
        this.emit('fb-click', {
            timestamp: Date.now(),
            source: 'demo-button',
        });

        // Log for visual verification during testing
        console.log('[fb-demo-button] Clicked! Event emitted.');
    }
}

// TypeScript declaration for HTML type checking
declare global {
    interface HTMLElementTagNameMap {
        'fb-demo-button': FBDemoButton;
    }
}
