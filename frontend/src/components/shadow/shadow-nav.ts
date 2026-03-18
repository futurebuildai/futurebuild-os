/**
 * Shadow Navigation Component
 * Navigation sidebar for Shadow Mode - switches between Tribunal Log and ShadowDocs.
 * See SHADOW_VIEWER_specs.md Section 5.1
 */

import { html, css, type TemplateResult } from 'lit';
import { customElement, state } from 'lit/decorators.js';
import { effect } from '@preact/signals-core';
import { FBElement } from '../base/FBElement';
import { store } from '../../store/store';
import { futureShadeService } from '../../futureshade/services/api';
import type { TreeNode } from '../../futureshade/types';

@customElement('shadow-nav')
export class ShadowNav extends FBElement {
    static override styles = [
        FBElement.styles,
        css`
            :host {
                display: flex;
                flex-direction: column;
                height: 100%;
                background: #0d0d15;
            }

            .tabs {
                display: flex;
                border-bottom: 1px solid #161821;
            }

            .tab {
                flex: 1;
                padding: 12px;
                background: transparent;
                border: none;
                border-bottom: 2px solid transparent;
                color: #9ca3af;
                cursor: pointer;
                font-size: 13px;
                font-family: inherit;
                transition: all 0.15s ease;
            }

            .tab:hover {
                background: #161821;
                color: #e5e7eb;
            }

            .tab.active {
                color: #33FFB8;
                border-bottom-color: #33FFB8;
            }

            .content {
                flex: 1;
                overflow-y: auto;
                padding: 8px;
            }

            .tree-item {
                display: flex;
                align-items: center;
                gap: 6px;
                padding: 6px 10px;
                border-radius: 4px;
                cursor: pointer;
                font-size: 13px;
                color: #9ca3af;
                transition: all 0.1s ease;
            }

            .tree-item:hover {
                background: #161821;
                color: #e5e7eb;
            }

            .tree-item.selected {
                background: #1a1a3e;
                color: #33FFB8;
            }

            .tree-children {
                margin-left: 16px;
            }

            .icon {
                font-size: 14px;
                flex-shrink: 0;
            }

            .empty-state {
                padding: 16px;
                text-align: center;
                color: #6b7280;
                font-size: 13px;
            }
        `,
    ];

    @state() private _activeView: 'log' | 'docs' = 'log';
    @state() private _tree: TreeNode[] = [];
    @state() private _selectedPath: string | null = null;
    @state() private _loading = false;

    private _disposeEffects: (() => void)[] = [];

    override connectedCallback(): void {
        super.connectedCallback();
        this._loadTree();
        this._disposeEffects.push(
            effect(() => {
                this._activeView = store.shadowActiveView$.value;
            }),
            effect(() => {
                this._selectedPath = store.selectedDocPath$.value;
            })
        );
    }

    override disconnectedCallback(): void {
        this._disposeEffects.forEach((d) => { d(); });
        super.disconnectedCallback();
    }

    private async _loadTree(): Promise<void> {
        this._loading = true;
        try {
            const response = await futureShadeService.getDocsTree();
            this._tree = response.roots;
        } catch (e) {
            console.error('Failed to load doc tree:', e);
        } finally {
            this._loading = false;
        }
    }

    private _handleTabClick(view: 'log' | 'docs'): void {
        store.actions.setShadowActiveView(view);
    }

    private _handleDocClick(node: TreeNode): void {
        if (node.type === 'file' && node.path) {
            store.actions.selectDoc(node.path);
        }
    }

    private _renderTree(nodes: TreeNode[]): TemplateResult {
        return html`
            ${nodes.map(
                (node) => html`
                    <div
                        class="tree-item ${this._selectedPath === node.path ? 'selected' : ''}"
                        @click=${() => { this._handleDocClick(node); }}
                    >
                        <span class="icon">${node.type === 'dir' ? '📁' : '📄'}</span>
                        ${node.name}
                    </div>
                    ${node.children && node.children.length > 0
                        ? html`<div class="tree-children">${this._renderTree(node.children)}</div>`
                        : ''}
                `
            )}
        `;
    }

    override render(): TemplateResult {
        return html`
            <div class="tabs">
                <button
                    class="tab ${this._activeView === 'log' ? 'active' : ''}"
                    @click=${() => { this._handleTabClick('log'); }}
                >
                    Tribunal Log
                </button>
                <button
                    class="tab ${this._activeView === 'docs' ? 'active' : ''}"
                    @click=${() => { this._handleTabClick('docs'); }}
                >
                    ShadowDocs
                </button>
            </div>

            <div class="content">
                ${this._activeView === 'docs'
                    ? this._loading
                        ? html`<div class="empty-state">Loading...</div>`
                        : this._tree.length > 0
                          ? this._renderTree(this._tree)
                          : html`<div class="empty-state">No documents found</div>`
                    : html`<div class="empty-state">Select a decision from the list</div>`}
            </div>
        `;
    }
}

declare global {
    interface HTMLElementTagNameMap {
        'shadow-nav': ShadowNav;
    }
}
