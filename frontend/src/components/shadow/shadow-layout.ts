/**
 * Shadow Layout Component
 * Main layout for Shadow Mode - 3-pane design with dark theme.
 * See SHADOW_VIEWER_specs.md Section 5.1
 */

import { html, css, type TemplateResult } from 'lit';
import { customElement, state } from 'lit/decorators.js';
import { effect } from '@preact/signals-core';
import { FBElement } from '../base/FBElement';
import { store } from '../../store/store';
import './shadow-nav';
import './tribunal-log-feed';
import './tribunal-case-detail';
import './shadow-docs-viewer';

@customElement('shadow-layout')
export class ShadowLayout extends FBElement {
    static override styles = [
        FBElement.styles,
        css`
            :host {
                display: grid;
                grid-template-columns: 240px 1fr 400px;
                height: 100%;
                background: #0a0a12;
                color: #e0e0e0;
            }

            .panel {
                border-right: 1px solid #1a1a2e;
                overflow: hidden;
                display: flex;
                flex-direction: column;
            }

            .panel:last-child {
                border-right: none;
            }

            .center-panel {
                overflow: hidden;
            }

            .right-panel {
                background: #0d0d15;
            }

            /* Responsive adjustments */
            @media (max-width: 1024px) {
                :host {
                    grid-template-columns: 200px 1fr 320px;
                }
            }

            @media (max-width: 768px) {
                :host {
                    grid-template-columns: 1fr;
                }

                .panel:not(.center-panel) {
                    display: none;
                }
            }
        `,
    ];

    @state() private _activeView: 'log' | 'docs' = 'log';
    @state() private _selectedDecisionId: string | null = null;
    @state() private _selectedDocPath: string | null = null;

    private _disposeEffects: (() => void)[] = [];

    override connectedCallback(): void {
        super.connectedCallback();
        this._disposeEffects.push(
            effect(() => {
                this._activeView = store.shadowActiveView$.value;
            }),
            effect(() => {
                this._selectedDecisionId = store.selectedDecisionId$.value;
            }),
            effect(() => {
                this._selectedDocPath = store.selectedDocPath$.value;
            })
        );
    }

    override disconnectedCallback(): void {
        this._disposeEffects.forEach((d) => d());
        super.disconnectedCallback();
    }

    override render(): TemplateResult {
        return html`
            <div class="panel">
                <shadow-nav></shadow-nav>
            </div>

            <div class="panel center-panel">
                ${this._activeView === 'log'
                    ? html`<tribunal-log-feed></tribunal-log-feed>`
                    : html`<shadow-docs-viewer .path=${this._selectedDocPath}></shadow-docs-viewer>`}
            </div>

            <div class="panel right-panel">
                <tribunal-case-detail .decisionId=${this._selectedDecisionId}></tribunal-case-detail>
            </div>
        `;
    }
}

declare global {
    interface HTMLElementTagNameMap {
        'shadow-layout': ShadowLayout;
    }
}
