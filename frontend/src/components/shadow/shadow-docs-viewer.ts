/**
 * Shadow Docs Viewer Component
 * Renders markdown documentation with XSS protection via DOMPurify.
 * SECURITY: Uses DOMPurify to sanitize HTML before rendering.
 * See SHADOW_VIEWER_specs.md Section 5.1 and 6.2
 */

import { html, css, type TemplateResult, type PropertyValues } from 'lit';
import { customElement, property, state } from 'lit/decorators.js';
import { unsafeHTML } from 'lit/directives/unsafe-html.js';
import { FBElement } from '../base/FBElement';
import { futureShadeService } from '../../futureshade/services/api';
import DOMPurify from 'dompurify';
import { marked } from 'marked';

@customElement('shadow-docs-viewer')
export class ShadowDocsViewer extends FBElement {
    static override styles = [
        FBElement.styles,
        css`
            :host {
                display: block;
                height: 100%;
                overflow-y: auto;
                padding: 24px;
                background: #0a0a12;
            }

            .markdown-content {
                line-height: 1.7;
                color: #F0F0F5;
            }

            .markdown-content h1 {
                font-size: 28px;
                font-weight: 600;
                color: #ffffff;
                margin-top: 0;
                margin-bottom: 16px;
                padding-bottom: 8px;
                border-bottom: 1px solid #161821;
            }

            .markdown-content h2 {
                font-size: 22px;
                font-weight: 600;
                color: #ffffff;
                margin-top: 32px;
                margin-bottom: 12px;
            }

            .markdown-content h3 {
                font-size: 18px;
                font-weight: 600;
                color: #ffffff;
                margin-top: 24px;
                margin-bottom: 8px;
            }

            .markdown-content h4,
            .markdown-content h5,
            .markdown-content h6 {
                font-size: 16px;
                font-weight: 600;
                color: #e5e7eb;
                margin-top: 20px;
                margin-bottom: 8px;
            }

            .markdown-content p {
                margin: 0 0 16px 0;
            }

            .markdown-content code {
                background: #161821;
                padding: 2px 6px;
                border-radius: 4px;
                font-size: 0.9em;
                font-family: 'Fira Code', 'Consolas', monospace;
            }

            .markdown-content pre {
                background: #161821;
                padding: 16px;
                border-radius: 8px;
                overflow-x: auto;
                margin: 16px 0;
            }

            .markdown-content pre code {
                background: transparent;
                padding: 0;
            }

            .markdown-content a {
                color: #33FFB8;
                text-decoration: none;
            }

            .markdown-content a:hover {
                text-decoration: underline;
            }

            .markdown-content blockquote {
                border-left: 4px solid #00CC82;
                margin: 16px 0;
                padding: 8px 16px;
                background: #161821;
                color: #9ca3af;
            }

            .markdown-content ul,
            .markdown-content ol {
                margin: 0 0 16px 0;
                padding-left: 24px;
            }

            .markdown-content li {
                margin-bottom: 4px;
            }

            .markdown-content table {
                width: 100%;
                border-collapse: collapse;
                margin: 16px 0;
            }

            .markdown-content th,
            .markdown-content td {
                border: 1px solid rgba(255,255,255,0.05);
                padding: 8px 12px;
                text-align: left;
            }

            .markdown-content th {
                background: #161821;
                font-weight: 600;
            }

            .markdown-content hr {
                border: none;
                border-top: 1px solid rgba(255,255,255,0.05);
                margin: 24px 0;
            }

            .empty-state {
                display: flex;
                align-items: center;
                justify-content: center;
                height: 100%;
                color: #6b7280;
                font-size: 14px;
            }

            .loading {
                display: flex;
                align-items: center;
                justify-content: center;
                height: 100%;
                color: #6b7280;
            }

            .error {
                padding: 16px;
                background: #7f1d1d;
                border-radius: 8px;
                color: #fca5a5;
                font-size: 14px;
            }

            .path-header {
                font-size: 12px;
                color: #6b7280;
                margin-bottom: 16px;
                padding-bottom: 8px;
                border-bottom: 1px solid #161821;
            }
        `,
    ];

    @property({ type: String }) path: string | null = null;

    @state() private _content: string = '';
    @state() private _loading = false;
    @state() private _error: string | null = null;

    override updated(changedProps: PropertyValues): void {
        if (changedProps.has('path') && this.path) {
            this._loadContent();
        }
    }

    private async _loadContent(): Promise<void> {
        if (!this.path) return;

        this._loading = true;
        this._error = null;

        try {
            const response = await futureShadeService.getDocContent(this.path);
            // Parse markdown to HTML
            const rawHtml = await marked(response.content);
            // SECURITY: Sanitize HTML to prevent XSS (SHADOW_VIEWER_specs.md Section 6.2)
            this._content = DOMPurify.sanitize(rawHtml);
        } catch (e) {
            this._error = e instanceof Error ? e.message : 'Failed to load document';
            this._content = '';
        } finally {
            this._loading = false;
        }
    }

    override render(): TemplateResult {
        if (this._loading) {
            return html`<div class="loading">Loading document...</div>`;
        }

        if (this._error) {
            return html`<div class="error">${this._error}</div>`;
        }

        if (!this.path) {
            return html`<div class="empty-state">Select a document from the tree</div>`;
        }

        return html`
            <div class="path-header">${this.path}</div>
            <div class="markdown-content">${unsafeHTML(this._content)}</div>
        `;
    }
}

declare global {
    interface HTMLElementTagNameMap {
        'shadow-docs-viewer': ShadowDocsViewer;
    }
}
