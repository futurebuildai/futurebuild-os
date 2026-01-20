/**
 * FBPanelRight - Right Panel (Artifacts)
 * See FRONTEND_SCOPE.md Section 3.3
 * Step 58.5: Fortress Hardening - Added pre-render validation
 */
import { effect } from '@preact/signals-core';
import { html, css, TemplateResult } from 'lit';
import { customElement, state } from 'lit/decorators.js';
import { FBElement } from '../base/FBElement';
import { store } from '../../store/store';
import type { ArtifactRef } from '../../store/types';
import type { ArtifactData } from '../../types/artifacts';
import { ArtifactType } from '../../types/enums';
import { normalizeArtifactType, getArtifactIcon } from '../../utils/artifact-helpers';
import { validateArtifactData } from '../../utils/artifact-validation';

// Import error boundary component
import '../base/fb-error-boundary';

@customElement('fb-panel-right')
export class FBPanelRight extends FBElement {
    static override styles = [
        FBElement.styles,
        css`
            :host {
                display: flex;
                flex-direction: column;
                height: 100%;
                background: var(--fb-bg-panel);
                border-left: 1px solid var(--fb-border);
                overflow: hidden;
            }

            .header {
                display: flex;
                align-items: center;
                justify-content: space-between;
                padding: var(--fb-spacing-md);
                border-bottom: 1px solid var(--fb-border-light);
            }

            .header-title {
                font-size: var(--fb-text-sm);
                font-weight: 600;
                color: var(--fb-text-primary);
            }

            .collapse-btn {
                padding: var(--fb-spacing-xs);
                border: none;
                background: transparent;
                color: var(--fb-text-muted);
                cursor: pointer;
                border-radius: var(--fb-radius-sm);
            }

            .collapse-btn:hover {
                background: var(--fb-bg-tertiary);
                color: var(--fb-text-primary);
            }

            .collapse-btn svg {
                width: 16px;
                height: 16px;
                stroke: currentColor;
                fill: none;
                stroke-width: 2;
            }

            .artifacts-list {
                flex: 1;
                overflow-y: auto;
                padding: var(--fb-spacing-md);
                display: flex;
                flex-direction: column;
                gap: var(--fb-spacing-md);
            }

            .artifact-card {
                background: var(--fb-bg-card);
                border: 1px solid var(--fb-border);
                border-radius: var(--fb-radius-md);
                overflow: hidden;
            }

            .artifact-header {
                display: flex;
                align-items: center;
                justify-content: space-between;
                padding: var(--fb-spacing-sm) var(--fb-spacing-md);
                background: var(--fb-bg-tertiary);
                border-bottom: 1px solid var(--fb-border-light);
            }

            .artifact-title {
                font-size: var(--fb-text-sm);
                font-weight: 500;
                color: var(--fb-text-primary);
            }

            .artifact-type {
                font-size: var(--fb-text-xs);
                color: var(--fb-text-muted);
                text-transform: uppercase;
            }

            .expand-btn {
                padding: var(--fb-spacing-xs);
                border: none;
                background: transparent;
                color: var(--fb-text-muted);
                cursor: pointer;
                border-radius: var(--fb-radius-sm);
                font-size: 14px;
                line-height: 1;
                transition: background 0.15s, color 0.15s;
            }

            .expand-btn:hover {
                background: var(--fb-bg-tertiary);
                color: var(--fb-text-primary);
            }

            .artifact-content {
                padding: var(--fb-spacing-md);
                min-height: 100px;
                display: flex;
                align-items: center;
                justify-content: center;
                color: var(--fb-text-muted);
                font-size: var(--fb-text-sm);
            }

            .artifact-actions {
                display: flex;
                gap: var(--fb-spacing-sm);
                padding: var(--fb-spacing-sm) var(--fb-spacing-md);
                border-top: 1px solid var(--fb-border-light);
            }

            .action-btn {
                flex: 1;
                padding: var(--fb-spacing-xs) var(--fb-spacing-sm);
                border: none;
                border-radius: var(--fb-radius-sm);
                font-size: var(--fb-text-xs);
                font-weight: 500;
                cursor: pointer;
                transition: opacity 0.15s;
            }

            .action-btn:hover {
                opacity: 0.9;
            }

            .action-btn.approve {
                background: var(--fb-success, #22c55e);
                color: white;
            }

            .action-btn.edit {
                background: var(--fb-bg-tertiary);
                color: var(--fb-text-primary);
                border: 1px solid var(--fb-border);
            }

            .action-btn.deny {
                background: var(--fb-error, #ef4444);
                color: white;
            }

            .empty-state {
                flex: 1;
                display: flex;
                flex-direction: column;
                align-items: center;
                justify-content: center;
                color: var(--fb-text-muted);
                text-align: center;
                padding: var(--fb-spacing-xl);
            }

            .empty-icon {
                font-size: 32px;
                margin-bottom: var(--fb-spacing-sm);
            }

            .scope-tabs {
                display: flex;
                border-bottom: 1px solid var(--fb-border-light);
            }

            .scope-tab {
                flex: 1;
                padding: var(--fb-spacing-sm);
                border: none;
                background: transparent;
                color: var(--fb-text-muted);
                font-size: var(--fb-text-xs);
                font-weight: 500;
                cursor: pointer;
                border-bottom: 2px solid transparent;
            }

            .scope-tab:hover {
                color: var(--fb-text-secondary);
            }

            .scope-tab.active {
                color: var(--fb-primary);
                border-bottom-color: var(--fb-primary);
            }
        `,
    ];

    @state() private _activeScope: 'project' | 'thread' = 'thread';
    @state() private _artifacts: ArtifactRef[] = [];
    @state() private _artifactData: Record<string, ArtifactData> = {};

    private _disposeEffects: (() => void)[] = [];

    override connectedCallback(): void {
        super.connectedCallback();

        // Subscribe to active artifact from Realtime Service
        this._disposeEffects.push(
            effect(() => {
                const active = store.activeArtifact$.value;
                if (active) {
                    // Store data
                    this._artifactData = {
                        ...this._artifactData,
                        [active.id]: active.data
                    };

                    // 2. Ensure ref exists in list
                    const exists = this._artifacts.find(a => a.id === active.id);
                    if (!exists) {
                        const newRef: ArtifactRef = {
                            id: active.id,
                            type: normalizeArtifactType(active.type),
                            title: active.title,
                            scope: 'thread'
                        };
                        this._artifacts = [newRef, ...this._artifacts];
                    }
                }
            })
        );

    }

    override disconnectedCallback(): void {
        this._disposeEffects.forEach((d) => { d(); });
        this._disposeEffects = [];
        super.disconnectedCallback();
    }

    private _handleCollapse(): void {
        store.actions.toggleRightPanel();
    }

    private _setScope(scope: 'project' | 'thread'): void {
        this._activeScope = scope;
    }

    private _getFilteredArtifacts(): ArtifactRef[] {
        return this._artifacts.filter((a) => a.scope === this._activeScope);
    }

    private _handleExpand(artifact: ArtifactRef): void {
        const data = this._artifactData[artifact.id];
        if (!data) return;

        // Map normalized type back to ArtifactType enum for ArtifactPayload
        const typeMap: Record<string, ArtifactType> = {
            'gantt': ArtifactType.GanttView,
            'budget': ArtifactType.BudgetView,
            'invoice': ArtifactType.Invoice,
        };

        const enumType = typeMap[artifact.type] ?? ArtifactType.Invoice;

        store.actions.setPopoutArtifact({
            id: artifact.id,
            type: enumType,
            title: artifact.title,
            data,
        });
    }

    // _getArtifactIcon removed - use getArtifactIcon from artifact-helpers.ts

    override render(): TemplateResult {
        const filteredArtifacts = this._getFilteredArtifacts();

        return html`
            <div class="header">
                <span class="header-title">📊 Artifacts</span>
                            <button class="collapse-btn" @click=${this._handleCollapse.bind(this)} aria-label="Collapse panel">
                    <svg viewBox="0 0 24 24" aria-hidden="true"><path d="M9 18l6-6-6-6"/></svg>
                </button>
            </div>

            <!-- Scope Tabs -->
            <div class="scope-tabs" role="tablist" aria-label="Artifact scope">
                <button 
                    class="scope-tab ${this._activeScope === 'thread' ? 'active' : ''}"
                    role="tab"
                    aria-selected=${this._activeScope === 'thread'}
                    @click=${(): void => { this._setScope('thread'); }}
                >
                    Thread
                </button>
                <button 
                    class="scope-tab ${this._activeScope === 'project' ? 'active' : ''}"
                    role="tab"
                    aria-selected=${this._activeScope === 'project'}
                    @click=${(): void => { this._setScope('project'); }}
                >
                    Project
                </button>
            </div>

            <!-- Artifacts List -->
            <div class="artifacts-list" role="list" aria-label="${this._activeScope} artifacts">
                ${filteredArtifacts.length > 0 ? filteredArtifacts.map((artifact) => html`
                    <div class="artifact-card" role="listitem">
                        <div class="artifact-header">
                            <span class="artifact-title">${getArtifactIcon(artifact.type)} ${artifact.title}</span>
                            <span class="artifact-type">${artifact.type}</span>
                            <button
                                class="expand-btn"
                                @click=${(): void => { this._handleExpand(artifact); }}
                                aria-label="Expand artifact to full view"
                                title="Expand"
                            >⛶</button>
                        </div>
                        
                        <div class="artifact-content">
                            ${this._renderArtifactContent(artifact)}
                        </div>

                        <!-- TODO: Wire up action handlers in future sprint -->
                        <div class="artifact-actions">
                            <button class="action-btn approve" aria-label="Approve artifact">✓ Approve</button>
                            <button class="action-btn edit" aria-label="Edit artifact">✎ Edit</button>
                            <button class="action-btn deny" aria-label="Deny artifact">✗ Deny</button>
                        </div>
                    </div>
                `) : html`
                    <div class="empty-state">
                        <div class="empty-icon">📭</div>
                        <div>No ${this._activeScope} artifacts</div>
                    </div>
                `}
            </div>
        `;
    }

    private _renderArtifactContent(artifact: ArtifactRef): TemplateResult {
        const data = this._artifactData[artifact.id];

        // Step 58.5: Pre-render validation to prevent crash from corrupt data
        const validation = validateArtifactData(artifact.type, data);
        if (!validation.valid) {
            return html`
                <fb-error-boundary
                    .hasError=${true}
                    .errorMessage=${validation.error ?? 'Invalid data'}
                ></fb-error-boundary>
            `;
        }

        // Use shared normalization helper
        const normalizedType = normalizeArtifactType(artifact.type);

        switch (normalizedType) {
            case 'gantt':
                return html`<fb-artifact-gantt .data=${data as ArtifactData}></fb-artifact-gantt>`;
            case 'budget':
                return html`<fb-artifact-budget .data=${data as ArtifactData}></fb-artifact-budget>`;
            case 'invoice':
                return html`<fb-artifact-invoice .data=${data as ArtifactData}></fb-artifact-invoice>`;
            default:
                return html`<div class="placeholder">Preview not available for ${artifact.type}</div>`;
        }
    }

}

declare global {
    interface HTMLElementTagNameMap {
        'fb-panel-right': FBPanelRight;
    }
}
