import { html, css, TemplateResult } from 'lit';
import { customElement, property, state } from 'lit/decorators.js';
import { FBElement } from '../base/FBElement';
import { store } from '../../store/store';

/**
 * FBInputBar - User input component for chat messages.
 * Features auto-resizing textarea, keyboard shortcuts, and accessibility.
 *
 * Step 73: Upload button wired to native file picker → store.handleFileDrop().
 * Drag-and-drop listeners live on fb-app-shell (global), not here.
 *
 * @element fb-input-bar
 * @fires send - When user submits a message
 */
@customElement('fb-input-bar')
export class FBInputBar extends FBElement {
    /** Maximum height for auto-resizing textarea as percentage of viewport */
    private static readonly MAX_TEXTAREA_HEIGHT_VH = 35;

    /** When true, disables input and send button (e.g., during AI processing). */
    @property({ type: Boolean }) disabled = false;

    static override styles = [
        FBElement.styles,
        css`
            :host {
                display: block;
                padding: var(--fb-spacing-md) var(--fb-spacing-lg);
                background: var(--md-sys-color-surface);
                border-top: 1px solid var(--md-sys-color-outline-variant);
            }

            .container {
                display: flex;
                align-items: flex-end;
                gap: var(--fb-spacing-sm);
                max-width: 100%;
                position: relative;
                background: var(--md-sys-color-surface-container-high);
                border-radius: var(--md-sys-shape-corner-extra-large);
                padding: 4px;
                box-shadow: var(--md-sys-elevation-1);
                transition: box-shadow var(--fb-transition-base);
            }

            .container:focus-within {
                box-shadow: var(--md-sys-elevation-2);
                background: var(--md-sys-color-surface-container-highest);
            }

            .input-wrapper {
                flex: 1;
                display: flex;
                flex-direction: column;
                position: relative;
            }

            textarea {
                width: 100%;
                min-height: 44px;
                max-height: 35vh;
                padding: 12px 16px;
                border: none;
                background: transparent;
                color: var(--md-sys-color-on-surface);
                font-family: var(--md-ref-typeface-plain);
                font-size: var(--md-sys-typescale-body-large);
                line-height: 1.5;
                resize: none;
                outline: none;
                box-sizing: border-box;
                overflow-y: auto;
            }

            textarea::placeholder {
                color: var(--md-sys-color-on-surface-variant);
            }

            .input-hint {
                position: absolute;
                width: 1px;
                height: 1px;
                padding: 0;
                margin: -1px;
                overflow: hidden;
                clip: rect(0, 0, 0, 0);
                white-space: nowrap;
                border: 0;
            }

            .icon-btn {
                width: 40px;
                height: 40px;
                display: flex;
                align-items: center;
                justify-content: center;
                border: none;
                background: transparent;
                color: var(--md-sys-color-on-surface-variant);
                cursor: pointer;
                border-radius: 50%;
                transition: all var(--fb-transition-base);
                flex-shrink: 0;
            }

            .icon-btn:hover:not(:disabled) {
                background-color: var(--md-sys-color-surface-variant);
                color: var(--md-sys-color-on-surface);
            }

            .icon-btn:focus-visible {
                outline: 2px solid var(--md-sys-color-primary);
                outline-offset: 2px;
            }

            .send-btn {
                background-color: var(--md-sys-color-primary);
                color: var(--md-sys-color-on-primary);
                margin-left: 4px;
            }

            .send-btn:hover:not(:disabled) {
                background-color: var(--md-sys-color-primary); /* Darken slightly in implementation if needed, but primary is standard */
                box-shadow: var(--md-sys-elevation-2);
                transform: scale(1.05);
            }

            .send-btn:disabled {
                background-color: var(--md-sys-color-surface-variant);
                color: var(--md-sys-color-on-surface-variant);
                cursor: not-allowed;
                box-shadow: none;
                transform: none;
            }

            svg {
                width: 24px;
                height: 24px;
                fill: currentColor;
            }
        `
    ];

    @state() private _value = '';

    private _handleInput(e: Event): void {
        const target = e.target as HTMLTextAreaElement;
        this._value = target.value;
        this._autoResize(target);
    }

    private _handleKeyDown(e: KeyboardEvent): void {
        if (e.key === 'Enter' && !e.shiftKey) {
            e.preventDefault();
            this._send();
        }
    }

    private _autoResize(el: HTMLTextAreaElement): void {
        el.style.height = 'auto';
        // Calculate max height as percentage of viewport
        const maxHeight = window.innerHeight * (FBInputBar.MAX_TEXTAREA_HEIGHT_VH / 100);
        const newHeight = Math.min(el.scrollHeight, maxHeight);
        el.style.height = `${String(newHeight)}px`;
    }

    private _send(): void {
        if (!this._value.trim() || this.disabled) return;

        const content = this._value;
        this._value = '';

        // Clear textarea value directly and reset height
        const textarea = this.shadowRoot?.querySelector('textarea');
        if (textarea) {
            textarea.value = '';
            textarea.style.height = 'auto';
        }

        this.emit('send', { content });
    }

    /** Step 73: Open native file picker, feed selection into store.handleFileDrop */
    private _openFilePicker(): void {
        const input = this.shadowRoot?.querySelector<HTMLInputElement>('#file-input');
        input?.click();
    }

    private _handleFileSelect(e: Event): void {
        const input = e.target as HTMLInputElement;
        if (input.files && input.files.length > 0) {
            store.actions.handleFileDrop(input.files);
        }
        // Reset so the same file can be re-selected
        input.value = '';
    }

    override render(): TemplateResult {
        return html`
            <input
                id="file-input"
                type="file"
                accept=".pdf,.jpg,.jpeg,.png,.gif,.webp"
                multiple
                hidden
                @change=${this._handleFileSelect.bind(this)}
            />
            <div class="container">
                <button
                    class="icon-btn"
                    title="Upload File"
                    aria-label="Upload file attachment"
                    @click=${this._openFilePicker.bind(this)}
                >
                    <svg viewBox="0 0 24 24" aria-hidden="true">
                        <path d="M16.5 6v11.5c0 2.21-1.79 4-4 4s-4-1.79-4-4V5a2.5 2.5 0 0 1 5 0v10.5c0 .55-.45 1-1 1s-1-.45-1-1V6H10v9.5a2.5 2.5 0 0 0 5 0V5c0-2.21-1.79-4-4-4S7 2.79 7 5v12.5c0 3.04 2.46 5.5 5.5 5.5s5.5-2.46 5.5-5.5V6h-1.5z"/>
                    </svg>
                </button>
                
                <div class="input-wrapper">
                    <span id="input-hint" class="input-hint">
                        Press Enter to send, Shift+Enter for new line
                    </span>
                    <textarea
                        placeholder="Type a message to the agent..."
                        aria-label="Message input"
                        aria-describedby="input-hint"
                        .value=${this._value}
                        ?disabled=${this.disabled}
                        @input=${this._handleInput.bind(this)}
                        @keydown=${this._handleKeyDown.bind(this)}
                        rows="1"
                    ></textarea>
                </div>

                <button 
                    class="icon-btn" 
                    title="Voice Input"
                    aria-label="Voice input (coming soon)"
                    disabled
                >
                    <svg viewBox="0 0 24 24" aria-hidden="true">
                        <path d="M12 14c1.66 0 3-1.34 3-3V5c0-1.66-1.34-3-3-3S9 3.34 9 5v6c0 1.66 1.34 3 3 3z"/>
                        <path d="M17 11c0 2.76-2.24 5-5 5s-5-2.24-5-5H5c0 3.53 2.61 6.43 6 6.92V21h2v-3.08c3.39-.49 6-3.39 6-6.92h-2z"/>
                    </svg>
                </button>

                <button 
                    class="icon-btn send-btn" 
                    ?disabled=${!this._value.trim() || this.disabled}
                    @click=${this._send.bind(this)}
                    title="Send Message"
                    aria-label="Send message"
                >
                    <svg viewBox="0 0 24 24" aria-hidden="true">
                        <path d="M2.01 21L23 12 2.01 3 2 10l15 2-15 2z"/>
                    </svg>
                </button>
            </div>
        `;
    }
}
