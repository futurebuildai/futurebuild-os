/**
 * fb-modal — Reusable modal dialog component.
 * Phase 20: UI/UX refinement — replaces inline form overlays.
 * GableLBM Industrial Dark glassmorphic styling.
 */
import { html, css, TemplateResult, nothing } from 'lit';
import { customElement, property } from 'lit/decorators.js';
import { FBElement } from '../base/FBElement';

@customElement('fb-modal')
export class FBModal extends FBElement {
  @property({ type: Boolean, reflect: true }) open = false;
  @property({ type: String }) heading = '';

  static override styles = [
    FBElement.styles,
    css`
      :host { display: contents; }

      .backdrop {
        position: fixed; inset: 0; z-index: 1000;
        background: rgba(0, 0, 0, 0.5);
        display: flex; align-items: center; justify-content: center;
        animation: fadeIn 0.15s ease;
      }

      .modal-card {
        background: rgba(22, 24, 33, 0.95);
        backdrop-filter: blur(24px);
        -webkit-backdrop-filter: blur(24px);
        border: 1px solid rgba(255, 255, 255, 0.08);
        border-radius: var(--fb-radius-lg, 12px);
        width: 90%;
        max-width: 560px;
        max-height: 85vh;
        overflow-y: auto;
        animation: slideUp 0.25s ease;
      }

      .modal-header {
        display: flex; justify-content: space-between; align-items: center;
        padding: var(--fb-spacing-lg, 24px);
        padding-bottom: 0;
      }

      .modal-header h2 {
        margin: 0; font-size: var(--fb-text-xl, 18px);
        color: var(--fb-text-primary, #F0F0F5); font-weight: 600;
      }

      .btn-close {
        background: transparent; border: none; cursor: pointer;
        color: var(--fb-text-secondary, #8B8D98); font-size: 18px;
        padding: 4px 8px; border-radius: var(--fb-radius-sm, 4px);
        transition: color 0.15s;
        line-height: 1;
      }
      .btn-close:hover { color: var(--fb-text-primary, #F0F0F5); }

      .modal-body {
        padding: var(--fb-spacing-lg, 24px);
      }

      @keyframes fadeIn {
        from { opacity: 0; }
        to { opacity: 1; }
      }
      @keyframes slideUp {
        from { opacity: 0; transform: translateY(20px); }
        to { opacity: 1; transform: translateY(0); }
      }
    `,
  ];

  override connectedCallback(): void {
    super.connectedCallback();
    this._onKeyDown = this._onKeyDown.bind(this);
    document.addEventListener('keydown', this._onKeyDown);
  }

  override disconnectedCallback(): void {
    super.disconnectedCallback();
    document.removeEventListener('keydown', this._onKeyDown);
  }

  private _onKeyDown(e: KeyboardEvent): void {
    if (e.key === 'Escape' && this.open) {
      this._close();
    }
  }

  private _onBackdropClick(e: MouseEvent): void {
    if ((e.target as HTMLElement).classList.contains('backdrop')) {
      this._close();
    }
  }

  private _close(): void {
    this.emit('fb-modal-close');
  }

  override render(): TemplateResult | typeof nothing {
    if (!this.open) return nothing;
    return html`
      <div class="backdrop" @click=${this._onBackdropClick} role="dialog" aria-modal="true">
        <div class="modal-card">
          <div class="modal-header">
            <h2>${this.heading}</h2>
            <button class="btn-close" @click=${this._close} aria-label="Close">&times;</button>
          </div>
          <div class="modal-body">
            <slot></slot>
          </div>
        </div>
      </div>
    `;
  }
}

declare global {
  interface HTMLElementTagNameMap {
    'fb-modal': FBModal;
  }
}
