import { html, css, TemplateResult, nothing } from 'lit';
import { customElement, property } from 'lit/decorators.js';
import { FBElement } from '../base/FBElement';
import { InvoiceFieldBoundingBox } from '../../types/artifacts';

/**
 * Provenance tooltip for AI-extracted invoice fields.
 * Shows the confidence score and optionally a bounding box preview
 * indicating where in the source PDF the value was extracted from.
 *
 * Sprint 3.1 — Task 3.1.3
 *
 * @fires fb-provenance-dismiss - When the user dismisses the tooltip
 */
@customElement('fb-field-provenance')
export class FBFieldProvenance extends FBElement {

    static override styles = [
        FBElement.styles,
        css`
            :host {
                display: block;
                position: absolute;
                z-index: 1000;
                pointer-events: auto;
            }

            .tooltip {
                background: #1f2937;
                color: #f9fafb;
                border-radius: 8px;
                padding: 10px 14px;
                font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;
                font-size: 12px;
                line-height: 1.5;
                min-width: 200px;
                max-width: 300px;
                box-shadow: 0 4px 16px rgba(0, 0, 0, 0.25), 0 0 0 1px rgba(255, 255, 255, 0.05);
                animation: fadeIn 0.15s ease-out;
            }

            .tooltip-arrow {
                position: absolute;
                bottom: -6px;
                left: 50%;
                transform: translateX(-50%);
                width: 12px;
                height: 6px;
                overflow: hidden;
            }

            .tooltip-arrow::after {
                content: '';
                position: absolute;
                width: 8px;
                height: 8px;
                background: #1f2937;
                transform: translateX(-50%) translateY(-50%) rotate(45deg);
                top: 0;
                left: 50%;
            }

            .confidence-header {
                display: flex;
                align-items: center;
                gap: 8px;
                margin-bottom: 6px;
            }

            .confidence-bar {
                flex: 1;
                height: 4px;
                background: rgba(255, 255, 255, 0.15);
                border-radius: 2px;
                overflow: hidden;
            }

            .confidence-fill {
                height: 100%;
                border-radius: 2px;
                transition: width 0.3s ease;
            }

            .confidence-fill--high { background: #34d399; }
            .confidence-fill--medium { background: #fbbf24; }
            .confidence-fill--low { background: #f87171; }

            .confidence-pct {
                font-weight: 600;
                font-size: 13px;
                font-variant-numeric: tabular-nums;
            }

            .source-label {
                color: #9ca3af;
                font-size: 11px;
                margin-top: 4px;
            }

            .bbox-preview {
                margin-top: 8px;
                background: #374151;
                border-radius: 4px;
                padding: 6px;
                position: relative;
                aspect-ratio: 8.5 / 11;
                max-height: 120px;
                overflow: hidden;
            }

            .bbox-page-label {
                position: absolute;
                top: 4px;
                left: 6px;
                font-size: 9px;
                color: #9ca3af;
                text-transform: uppercase;
                letter-spacing: 0.05em;
            }

            .bbox-highlight {
                position: absolute;
                border: 2px solid #f59e0b;
                background: rgba(245, 158, 11, 0.15);
                border-radius: 2px;
                animation: pulse-border 2s infinite;
            }

            @keyframes fadeIn {
                from { opacity: 0; transform: translateY(4px); }
                to { opacity: 1; transform: translateY(0); }
            }

            @keyframes pulse-border {
                0%, 100% { border-color: rgba(245, 158, 11, 1); }
                50% { border-color: rgba(245, 158, 11, 0.4); }
            }
        `
    ];

    /** Human-readable field label */
    @property()
    fieldName = '';

    /** Confidence score 0.0–1.0 */
    @property({ type: Number })
    confidence = 1.0;

    /** PDF bounding box (if available) */
    @property({ type: Object })
    boundingBox?: InvoiceFieldBoundingBox;

    /** Extraction source label */
    @property()
    source = 'vision_extraction';

    private _confidenceLevel(): 'high' | 'medium' | 'low' {
        if (this.confidence >= 0.85) return 'high';
        if (this.confidence >= 0.6) return 'medium';
        return 'low';
    }

    private _formatPercent(): string {
        return `${Math.round(this.confidence * 100)}%`;
    }

    private _sourceLabel(): string {
        if (this.source === 'manual') return 'Manually entered';
        return 'Extracted by AI';
    }

    private _renderBoundingBox(): TemplateResult | typeof nothing {
        if (!this.boundingBox) return nothing;
        const { page, x, y, w, h } = this.boundingBox;
        // Convert normalized coords to percentage positions
        const style = `left: ${x * 100}%; top: ${y * 100}%; width: ${w * 100}%; height: ${h * 100}%;`;
        return html`
            <div class="bbox-preview">
                <span class="bbox-page-label">Page ${page}</span>
                <div class="bbox-highlight" style=${style}></div>
            </div>
        `;
    }

    override render(): TemplateResult {
        const level = this._confidenceLevel();

        return html`
            <div class="tooltip" role="tooltip" aria-live="polite">
                <div class="confidence-header">
                    <div class="confidence-bar">
                        <div
                            class="confidence-fill confidence-fill--${level}"
                            style="width: ${this.confidence * 100}%"
                        ></div>
                    </div>
                    <span class="confidence-pct">${this._formatPercent()}</span>
                </div>
                <div class="source-label">${this._sourceLabel()}${this.fieldName ? ` — ${this.fieldName}` : ''}</div>
                ${this._renderBoundingBox()}
                <div class="tooltip-arrow"></div>
            </div>
        `;
    }
}
