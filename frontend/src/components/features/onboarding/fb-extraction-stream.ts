/**
 * fb-extraction-stream — Real-time extraction display.
 * See FRONTEND_V2_SPEC.md §5.3
 *
 * Shows SSE-driven progress during document extraction:
 * ✓ Found address: 123 Main St
 * ✓ Created 47 tasks
 * ◎ Applying weather data...
 * ✓ Detected 3 long-lead items
 */
import { html, css, nothing } from 'lit';
import { customElement, property, state } from 'lit/decorators.js';
import { FBElement } from '../../base/FBElement';

export type ExtractionEventType = 'extraction' | 'scheduling' | 'weather' | 'procurement' | 'complete' | 'error';

export interface ExtractionEvent {
    type: ExtractionEventType;
    step: string;
    value?: string;
    count?: number;
    confidence?: number;
    items?: Array<{ name: string; lead_weeks: number; order_by: string }>;
    ready_to_create?: boolean;
    error?: string;
}

interface StreamItem {
    id: string;
    type: ExtractionEventType;
    label: string;
    detail?: string;
    status: 'pending' | 'active' | 'complete' | 'error';
    timestamp: number;
}

@customElement('fb-extraction-stream')
export class FBExtractionStream extends FBElement {
    static override styles = [
        FBElement.styles,
        css`
            :host {
                display: block;
                padding: 24px;
                background: var(--fb-surface-1, #161821);
                border-radius: 12px;
                border: 1px solid var(--fb-border, rgba(255,255,255,0.05));
                max-width: 480px;
            }

            .header {
                display: flex;
                align-items: center;
                gap: 12px;
                margin-bottom: 20px;
            }

            .header-icon {
                width: 40px;
                height: 40px;
                border-radius: 10px;
                background: linear-gradient(135deg, var(--fb-accent, #00FFA3), var(--fb-accent-light, #33FFB8));
                display: flex;
                align-items: center;
                justify-content: center;
            }

            .header-icon svg {
                width: 24px;
                height: 24px;
                color: #fff;
            }

            .header-text {
                flex: 1;
            }

            .header-title {
                font-size: 16px;
                font-weight: 600;
                color: var(--fb-text-primary, #F0F0F5);
            }

            .header-subtitle {
                font-size: 13px;
                color: var(--fb-text-secondary, #8B8D98);
                margin-top: 2px;
            }

            .stream-list {
                display: flex;
                flex-direction: column;
                gap: 12px;
            }

            .stream-item {
                display: flex;
                align-items: flex-start;
                gap: 12px;
                padding: 12px;
                background: var(--fb-surface-2, #1E2029);
                border-radius: 8px;
                transition: all 0.2s ease;
            }

            .stream-item.active {
                background: rgba(0, 255, 163, 0.1);
                border: 1px solid rgba(0, 255, 163, 0.3);
            }

            .stream-item.error {
                background: rgba(239, 68, 68, 0.1);
                border: 1px solid rgba(239, 68, 68, 0.3);
            }

            .status-icon {
                width: 24px;
                height: 24px;
                display: flex;
                align-items: center;
                justify-content: center;
                flex-shrink: 0;
            }

            .status-icon.pending {
                color: var(--fb-text-tertiary, #5A5B66);
            }

            .status-icon.active {
                color: var(--fb-accent, #00FFA3);
                animation: pulse 1s ease-in-out infinite;
            }

            .status-icon.complete {
                color: #00FFA3;
            }

            .status-icon.error {
                color: #F43F5E;
            }

            @keyframes pulse {
                0%, 100% { opacity: 1; }
                50% { opacity: 0.5; }
            }

            .item-content {
                flex: 1;
                min-width: 0;
            }

            .item-label {
                font-size: 14px;
                font-weight: 500;
                color: var(--fb-text-primary, #F0F0F5);
            }

            .item-detail {
                font-size: 13px;
                color: var(--fb-text-secondary, #8B8D98);
                margin-top: 4px;
                word-break: break-word;
            }

            .confidence-badge {
                display: inline-flex;
                align-items: center;
                padding: 2px 8px;
                border-radius: 4px;
                font-size: 11px;
                font-weight: 600;
                margin-left: 8px;
            }

            .confidence-badge.high {
                background: rgba(34, 197, 94, 0.15);
                color: #00FFA3;
            }

            .confidence-badge.medium {
                background: rgba(245, 158, 11, 0.15);
                color: #f59e0b;
            }

            .confidence-badge.low {
                background: rgba(239, 68, 68, 0.15);
                color: #F43F5E;
            }

            .empty-state {
                text-align: center;
                padding: 40px 20px;
                color: var(--fb-text-secondary, #8B8D98);
            }

            .empty-state svg {
                width: 48px;
                height: 48px;
                margin-bottom: 16px;
                opacity: 0.5;
            }

            .spinner {
                width: 20px;
                height: 20px;
                border: 2px solid var(--fb-border, rgba(255,255,255,0.05));
                border-top-color: var(--fb-accent, #00FFA3);
                border-radius: 50%;
                animation: spin 0.8s linear infinite;
            }

            @keyframes spin {
                to { transform: rotate(360deg); }
            }

            .complete-banner {
                display: flex;
                align-items: center;
                gap: 12px;
                padding: 16px;
                background: rgba(34, 197, 94, 0.1);
                border: 1px solid rgba(34, 197, 94, 0.3);
                border-radius: 8px;
                margin-top: 16px;
            }

            .complete-banner svg {
                width: 24px;
                height: 24px;
                color: #00FFA3;
            }

            .complete-text {
                font-size: 14px;
                font-weight: 500;
                color: #00FFA3;
            }
        `,
    ];

    /** Stream items to display */
    @state() private _items: StreamItem[] = [];

    /** Whether extraction is complete */
    @property({ type: Boolean }) complete = false;

    /** Whether there was an error */
    @property({ type: Boolean }) error = false;

    /** Add an extraction event to the stream */
    public addEvent(event: ExtractionEvent): void {
        const item = this._eventToItem(event);

        // Mark previous active items as complete
        this._items = this._items.map(i =>
            i.status === 'active' ? { ...i, status: 'complete' as const } : i
        );

        this._items = [...this._items, item];

        if (event.type === 'complete') {
            this.complete = true;
        }
        if (event.type === 'error') {
            this.error = true;
        }
    }

    /** Clear all items and reset state */
    public reset(): void {
        this._items = [];
        this.complete = false;
        this.error = false;
    }

    private _eventToItem(event: ExtractionEvent): StreamItem {
        const id = `item-${Date.now()}-${Math.random().toString(36).slice(2)}`;
        let label = '';
        let detail = '';

        switch (event.type) {
            case 'extraction':
                label = this._getExtractionLabel(event.step);
                detail = event.value ?? '';
                if (event.confidence !== undefined) {
                    const pct = Math.round(event.confidence * 100);
                    detail += ` (${pct}% confidence)`;
                }
                break;

            case 'scheduling':
                if (event.step === 'tasks_generated') {
                    label = `Created ${event.count} tasks`;
                } else if (event.step === 'dependencies_mapped') {
                    label = `Mapped ${event.count} dependencies`;
                } else {
                    label = event.step;
                }
                break;

            case 'weather':
                label = 'Applying weather data';
                detail = event.value ?? 'Checking forecast for outdoor work';
                break;

            case 'procurement':
                if (event.step === 'long_lead_detected' && event.items) {
                    label = `Detected ${event.items.length} long-lead items`;
                    detail = event.items.map(i => `${i.name} (${i.lead_weeks}wk lead)`).join(', ');
                } else {
                    label = 'Analyzing procurement';
                }
                break;

            case 'complete':
                label = 'Extraction complete';
                detail = 'Ready to create project';
                break;

            case 'error':
                label = 'Error occurred';
                detail = event.error ?? 'Unknown error';
                break;

            default:
                label = event.step;
        }

        return {
            id,
            type: event.type,
            label,
            detail,
            status: event.type === 'complete' ? 'complete' : event.type === 'error' ? 'error' : 'active',
            timestamp: Date.now(),
        };
    }

    private _getExtractionLabel(step: string): string {
        const labels: Record<string, string> = {
            address: 'Found address',
            scope: 'Identified scope',
            contract_value: 'Extracted contract value',
            start_date: 'Found start date',
            completion_date: 'Found completion date',
            contractor: 'Identified contractor',
            client: 'Identified client',
        };
        return labels[step] ?? `Extracted ${step}`;
    }

    // Reserved for future use when confidence badges are rendered
    // private _getConfidenceClass(confidence?: number): string {
    //     if (confidence === undefined) return '';
    //     if (confidence >= 0.9) return 'high';
    //     if (confidence >= 0.7) return 'medium';
    //     return 'low';
    // }

    private _renderStatusIcon(status: StreamItem['status']) {
        switch (status) {
            case 'pending':
                return html`<div class="status-icon pending">○</div>`;
            case 'active':
                return html`<div class="status-icon active"><div class="spinner"></div></div>`;
            case 'complete':
                return html`
                    <div class="status-icon complete">
                        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                            <polyline points="20 6 9 17 4 12"/>
                        </svg>
                    </div>
                `;
            case 'error':
                return html`
                    <div class="status-icon error">
                        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                            <circle cx="12" cy="12" r="10"/>
                            <line x1="15" y1="9" x2="9" y2="15"/>
                            <line x1="9" y1="9" x2="15" y2="15"/>
                        </svg>
                    </div>
                `;
        }
    }

    override render() {
        return html`
            <div class="header">
                <div class="header-icon">
                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                        <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/>
                        <polyline points="14 2 14 8 20 8"/>
                        <line x1="16" y1="13" x2="8" y2="13"/>
                        <line x1="16" y1="17" x2="8" y2="17"/>
                        <polyline points="10 9 9 9 8 9"/>
                    </svg>
                </div>
                <div class="header-text">
                    <div class="header-title">Analyzing Document</div>
                    <div class="header-subtitle">Extracting project details...</div>
                </div>
            </div>

            ${this._items.length === 0 ? html`
                <div class="empty-state">
                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                        <circle cx="12" cy="12" r="10"/>
                        <polyline points="12 6 12 12 16 14"/>
                    </svg>
                    <div>Waiting for document...</div>
                </div>
            ` : html`
                <div class="stream-list">
                    ${this._items.map(item => html`
                        <div class="stream-item ${item.status}">
                            ${this._renderStatusIcon(item.status)}
                            <div class="item-content">
                                <div class="item-label">${item.label}</div>
                                ${item.detail ? html`<div class="item-detail">${item.detail}</div>` : nothing}
                            </div>
                        </div>
                    `)}
                </div>

                ${this.complete ? html`
                    <div class="complete-banner">
                        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                            <path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"/>
                            <polyline points="22 4 12 14.01 9 11.01"/>
                        </svg>
                        <span class="complete-text">Ready to create your project</span>
                    </div>
                ` : nothing}
            `}
        `;
    }
}

declare global {
    interface HTMLElementTagNameMap {
        'fb-extraction-stream': FBExtractionStream;
    }
}
