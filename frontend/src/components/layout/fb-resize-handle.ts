/**
 * FBResizeHandle - Draggable resize handle for right panel
 * Step 59.5: UX Enhancements
 *
 * Provides drag-to-resize functionality for the artifact panel.
 * Min: 280px, Max: 600px
 */
import { html, css, TemplateResult } from 'lit';
import { customElement, state } from 'lit/decorators.js';
import { FBElement } from '../base/FBElement';
import { store } from '../../store/store';

@customElement('fb-resize-handle')
export class FBResizeHandle extends FBElement {
    static override styles = [
        FBElement.styles,
        css`
            :host {
                position: absolute;
                top: 0;
                right: var(--resize-handle-offset, 320px);
                width: 6px;
                height: 100%;
                cursor: col-resize;
                z-index: var(--fb-z-panel, 100);
                display: flex;
                align-items: center;
                justify-content: center;
            }

            .handle {
                width: 4px;
                height: 48px;
                background: var(--fb-border);
                border-radius: 2px;
                transition: background 0.15s ease, height 0.15s ease;
            }

            :host(:hover) .handle,
            .handle.dragging {
                background: var(--fb-primary);
                height: 80px;
            }
        `,
    ];

    @state() private _isDragging = false;
    private _startX = 0;
    private _startWidth = 0;

    private _handleMouseDown = (e: MouseEvent): void => {
        e.preventDefault();
        this._isDragging = true;
        this._startX = e.clientX;
        this._startWidth = store.rightPanelWidth$.value;

        document.addEventListener('mousemove', this._handleMouseMove);
        document.addEventListener('mouseup', this._handleMouseUp);
    };

    private _handleMouseMove = (e: MouseEvent): void => {
        if (!this._isDragging) return;

        // Moving left increases width, moving right decreases
        const deltaX = this._startX - e.clientX;
        const newWidth = this._startWidth + deltaX;

        store.actions.setRightPanelWidth(newWidth);
    };

    private _handleMouseUp = (): void => {
        this._isDragging = false;

        document.removeEventListener('mousemove', this._handleMouseMove);
        document.removeEventListener('mouseup', this._handleMouseUp);
    };

    // Touch support for mobile/tablet
    private _handleTouchStart = (e: TouchEvent): void => {
        const touch = e.touches[0];
        if (!touch) return;

        e.preventDefault();
        this._isDragging = true;
        this._startX = touch.clientX;
        this._startWidth = store.rightPanelWidth$.value;

        document.addEventListener('touchmove', this._handleTouchMove, { passive: false });
        document.addEventListener('touchend', this._handleTouchEnd);
    };

    private _handleTouchMove = (e: TouchEvent): void => {
        const touch = e.touches[0];
        if (!this._isDragging || !touch) return;

        e.preventDefault();
        const deltaX = this._startX - touch.clientX;
        const newWidth = this._startWidth + deltaX;

        store.actions.setRightPanelWidth(newWidth);
    };



    private _handleTouchEnd = (): void => {
        this._isDragging = false;

        document.removeEventListener('touchmove', this._handleTouchMove);
        document.removeEventListener('touchend', this._handleTouchEnd);
    };

    private _handleKeyDown = (e: KeyboardEvent): void => {
        if (e.key !== 'ArrowLeft' && e.key !== 'ArrowRight') return;

        e.preventDefault();
        const currentWidth = store.rightPanelWidth$.value;
        const step = 20; // 20px steps for keyboard resize

        // ArrowLeft increases width (expands panel), ArrowRight decreases (shrinks)
        const delta = e.key === 'ArrowLeft' ? step : -step;
        store.actions.setRightPanelWidth(currentWidth + delta);
    };

    override disconnectedCallback(): void {
        // Cleanup any lingering listeners
        document.removeEventListener('mousemove', this._handleMouseMove);
        document.removeEventListener('mouseup', this._handleMouseUp);
        document.removeEventListener('touchmove', this._handleTouchMove);
        document.removeEventListener('touchend', this._handleTouchEnd);
        super.disconnectedCallback();
    }

    override render(): TemplateResult {
        return html`
            <div
                class="handle ${this._isDragging ? 'dragging' : ''}"
                @mousedown=${this._handleMouseDown}
                @touchstart=${this._handleTouchStart}
                @keydown=${this._handleKeyDown}
                role="separator"
                aria-orientation="vertical"
                aria-label="Resize artifacts panel"
                tabindex="0"
            ></div>
        `;
    }
}

declare global {
    interface HTMLElementTagNameMap {
        'fb-resize-handle': FBResizeHandle;
    }
}
