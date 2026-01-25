import { LitElement, html, css } from 'lit';
import { customElement, state } from 'lit/decorators.js';
import { futureShadeService } from '../services/api';
import type { HealthResponse } from '../types';

/**
 * FutureShade Status Indicator
 * Displays the current health status of the FutureShade service.
 * Admin-only component.
 */
@customElement('fb-shadow-status')
export class FBShadowStatus extends LitElement {
  static override styles = css`
    :host {
      display: inline-flex;
      align-items: center;
      gap: 0.5rem;
      font-size: 0.875rem;
    }

    .indicator {
      width: 8px;
      height: 8px;
      border-radius: 50%;
    }

    .indicator--active {
      background-color: var(--color-success, #22c55e);
    }

    .indicator--disabled {
      background-color: var(--color-muted, #6b7280);
    }

    .indicator--degraded {
      background-color: var(--color-warning, #f59e0b);
    }

    .indicator--error {
      background-color: var(--color-error, #ef4444);
    }

    .label {
      color: var(--color-text-secondary, #9ca3af);
    }
  `;

  @state()
  private health: HealthResponse | null = null;

  @state()
  private error: string | null = null;

  override connectedCallback(): void {
    super.connectedCallback();
    this.checkHealth();
  }

  private async checkHealth(): Promise<void> {
    try {
      this.health = await futureShadeService.health();
      this.error = null;
    } catch (e) {
      this.error = e instanceof Error ? e.message : 'Unknown error';
      this.health = null;
    }
  }

  private getIndicatorClass(): string {
    if (this.error) return 'indicator indicator--error';
    if (!this.health) return 'indicator indicator--disabled';
    return `indicator indicator--${this.health.status}`;
  }

  private getStatusLabel(): string {
    if (this.error) return 'Error';
    if (!this.health) return 'Loading...';
    return this.health.status.charAt(0).toUpperCase() + this.health.status.slice(1);
  }

  override render() {
    return html`
      <span class=${this.getIndicatorClass()}></span>
      <span class="label">Shadow: ${this.getStatusLabel()}</span>
    `;
  }
}

declare global {
  interface HTMLElementTagNameMap {
    'fb-shadow-status': FBShadowStatus;
  }
}
