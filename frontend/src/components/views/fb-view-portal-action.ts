/**
 * FBViewPortalAction - One-Time Action View from SMS Link
 * See LAUNCH_PLAN.md P2: Field Portal (Mobile)
 *
 * Single-purpose action view accessed via one-time link:
 * - Displays task context (name, project, what's needed)
 * - Shows appropriate action UI based on link type:
 *   - Status update: Status toggle buttons
 *   - Photo upload: Camera/upload interface
 *   - View: Read-only task details
 * - No login required - token-based auth from URL
 */
import { html, css, TemplateResult, nothing } from 'lit';
import { customElement, property, state } from 'lit/decorators.js';
import { FBViewElement } from '../base/FBViewElement';
import { notify } from '../../store/notifications';

import '../portal/fb-portal-shell';
import '../portal/fb-status-toggle';
import '../portal/fb-photo-upload';

import type { TaskStatusValue } from '../portal/fb-status-toggle';

interface ActionContext {
    actionType: string;
    contact: { id: string; name: string };
    project: { id: string; name: string; address: string };
    task: { id: string; wbs_code: string; name: string; status: string; start_date?: string; end_date?: string };
}

/**
 * Portal action view component.
 * @element fb-view-portal-action
 */
@customElement('fb-view-portal-action')
export class FBViewPortalAction extends FBViewElement {
    static override styles = [
        FBViewElement.styles,
        css`
            :host {
                display: block;
            }

            .loading {
                display: flex;
                flex-direction: column;
                align-items: center;
                justify-content: center;
                min-height: 300px;
                gap: 16px;
            }

            .spinner {
                width: 40px;
                height: 40px;
                border: 3px solid var(--fb-border, rgba(255,255,255,0.05));
                border-top-color: var(--fb-primary, #00FFA3);
                border-radius: 50%;
                animation: spin 1s linear infinite;
            }

            @keyframes spin {
                to { transform: rotate(360deg); }
            }

            .loading-text {
                color: var(--fb-text-secondary, #8B8D98);
                font-size: 14px;
            }

            .error-container {
                display: flex;
                flex-direction: column;
                align-items: center;
                justify-content: center;
                min-height: 300px;
                gap: 16px;
                text-align: center;
            }

            .error-icon {
                width: 64px;
                height: 64px;
                color: var(--fb-error, #F43F5E);
            }

            .error-title {
                color: var(--fb-text-primary, #fff);
                font-size: 20px;
                font-weight: 600;
                margin: 0;
            }

            .error-message {
                color: var(--fb-text-secondary, #8B8D98);
                font-size: 14px;
                margin: 0;
                max-width: 300px;
            }

            .content {
                display: flex;
                flex-direction: column;
                gap: 24px;
            }

            .task-card {
                background: var(--fb-bg-card, #161821);
                border: 1px solid var(--fb-border, rgba(255,255,255,0.05));
                border-radius: 12px;
                padding: 20px;
            }

            .task-header {
                display: flex;
                align-items: flex-start;
                justify-content: space-between;
                gap: 16px;
                margin-bottom: 16px;
            }

            .task-name {
                color: var(--fb-text-primary, #fff);
                font-size: 20px;
                font-weight: 600;
                margin: 0;
            }

            .task-wbs {
                color: var(--fb-primary, #00FFA3);
                font-size: 12px;
                font-weight: 500;
                padding: 4px 8px;
                background: var(--fb-primary-alpha, rgba(0, 255, 163, 0.1));
                border-radius: 4px;
            }

            .task-meta {
                display: flex;
                flex-direction: column;
                gap: 8px;
            }

            .meta-row {
                display: flex;
                align-items: center;
                gap: 8px;
                color: var(--fb-text-secondary, #8B8D98);
                font-size: 14px;
            }

            .meta-row svg {
                width: 18px;
                height: 18px;
                flex-shrink: 0;
            }

            .section-title {
                color: var(--fb-text-primary, #fff);
                font-size: 16px;
                font-weight: 600;
                margin: 0 0 12px 0;
            }

            .submit-btn {
                width: 100%;
                padding: 16px 24px;
                font-size: 16px;
                font-weight: 600;
                background: var(--fb-primary, #00FFA3);
                color: white;
                border: none;
                border-radius: 12px;
                cursor: pointer;
                transition: background 0.2s ease;
            }

            .submit-btn:hover:not([disabled]) {
                background: var(--fb-primary-hover, #5a6fd6);
            }

            .submit-btn[disabled] {
                opacity: 0.5;
                cursor: not-allowed;
            }

            .success-container {
                display: flex;
                flex-direction: column;
                align-items: center;
                justify-content: center;
                min-height: 300px;
                gap: 16px;
                text-align: center;
            }

            .success-icon {
                width: 64px;
                height: 64px;
                color: var(--fb-success, #00FFA3);
            }

            .success-title {
                color: var(--fb-text-primary, #fff);
                font-size: 20px;
                font-weight: 600;
                margin: 0;
            }

            .success-message {
                color: var(--fb-text-secondary, #8B8D98);
                font-size: 14px;
                margin: 0;
            }

            .upsell {
                background: var(--fb-bg-card, #161821);
                border: 1px solid var(--fb-border, rgba(255,255,255,0.05));
                border-radius: 12px;
                padding: 20px;
                text-align: center;
            }

            .upsell-title {
                color: var(--fb-text-primary, #fff);
                font-size: 14px;
                font-weight: 500;
                margin: 0 0 8px 0;
            }

            .upsell-link {
                color: var(--fb-primary, #00FFA3);
                text-decoration: none;
                font-size: 14px;
            }

            .upsell-link:hover {
                text-decoration: underline;
            }
        `,
    ];

    @property({ type: String }) token = '';

    @state() private _loading = true;
    @state() private _error: string | null = null;
    @state() private _context: ActionContext | null = null;
    @state() private _submitting = false;
    @state() private _submitted = false;
    @state() private _selectedStatus: TaskStatusValue | null = null;

    override connectedCallback(): void {
        super.connectedCallback();
        void this._loadContext();
    }

    private async _loadContext(): Promise<void> {
        if (!this.token) {
            this._error = 'Invalid link';
            this._loading = false;
            return;
        }

        const controller = new AbortController();
        const timeoutId = setTimeout(() => { controller.abort(); }, 15000);

        try {
            const response = await fetch(`/api/v1/portal/action/${this.token}`, {
                signal: controller.signal,
            });
            clearTimeout(timeoutId);

            if (!response.ok) {
                const data = await response.json() as { error?: string };
                throw new Error(data.error ?? 'Invalid or expired link');
            }

            this._context = await response.json() as ActionContext;
            // Set initial status from task
            this._selectedStatus = this._mapStatus(this._context.task.status);
        } catch (err) {
            clearTimeout(timeoutId);
            if (err instanceof DOMException && err.name === 'AbortError') {
                this._error = 'Request timed out. Please try again.';
            } else {
                this._error = err instanceof Error ? err.message : 'Failed to load task';
            }
        } finally {
            this._loading = false;
        }
    }

    private _mapStatus(status: string): TaskStatusValue {
        switch (status.toLowerCase()) {
            case 'in_progress':
                return 'in_progress';
            case 'completed':
                return 'completed';
            default:
                return 'pending';
        }
    }

    private _handleStatusChange(e: CustomEvent<{ status: TaskStatusValue }>): void {
        this._selectedStatus = e.detail.status;
    }

    private async _handleSubmit(): Promise<void> {
        if (!this._context || this._submitting) return;

        this._submitting = true;

        const controller = new AbortController();
        const timeoutId = setTimeout(() => { controller.abort(); }, 15000);

        try {
            const body: Record<string, string> = {};

            if (this._context.actionType === 'status_update' && this._selectedStatus) {
                body.status = this._selectedStatus;
            }

            const response = await fetch(`/api/v1/portal/action/${this.token}`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(body),
                signal: controller.signal,
            });
            clearTimeout(timeoutId);

            if (!response.ok) {
                const data = await response.json() as { error?: string };
                throw new Error(data.error ?? 'Failed to submit');
            }

            this._submitted = true;
            notify.success('Task updated successfully!');
        } catch (err) {
            clearTimeout(timeoutId);
            if (err instanceof DOMException && err.name === 'AbortError') {
                notify.error('Request timed out. Please try again.');
            } else {
                const message = err instanceof Error ? err.message : 'Failed to submit';
                notify.error(message);
            }
        } finally {
            this._submitting = false;
        }
    }

    private _renderLoading(): TemplateResult {
        return html`
            <div class="loading">
                <div class="spinner"></div>
                <span class="loading-text">Loading task details...</span>
            </div>
        `;
    }

    private _renderError(): TemplateResult {
        return html`
            <div class="error-container">
                <svg class="error-icon" viewBox="0 0 24 24" fill="currentColor">
                    <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm1 15h-2v-2h2v2zm0-4h-2V7h2v6z"/>
                </svg>
                <h2 class="error-title">Link Unavailable</h2>
                <p class="error-message">${this._error}</p>
            </div>
        `;
    }

    private _renderSuccess(): TemplateResult {
        return html`
            <div class="success-container">
                <svg class="success-icon" viewBox="0 0 24 24" fill="currentColor">
                    <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm-2 15l-5-5 1.41-1.41L10 14.17l7.59-7.59L19 8l-9 9z"/>
                </svg>
                <h2 class="success-title">Task Updated!</h2>
                <p class="success-message">Your response has been recorded.</p>
            </div>

            <div class="upsell">
                <p class="upsell-title">Want easier access to all your tasks?</p>
                <a class="upsell-link" href="/portal/signup">Create a portal account</a>
            </div>
        `;
    }

    private _renderContent(): TemplateResult {
        if (!this._context) return html``;

        const { actionType, project, task } = this._context;

        return html`
            <div class="content">
                <div class="task-card">
                    <div class="task-header">
                        <h1 class="task-name">${task.name}</h1>
                        <span class="task-wbs">${task.wbs_code}</span>
                    </div>
                    <div class="task-meta">
                        <div class="meta-row">
                            <svg viewBox="0 0 24 24" fill="currentColor">
                                <path d="M12 2C8.13 2 5 5.13 5 9c0 5.25 7 13 7 13s7-7.75 7-13c0-3.87-3.13-7-7-7zm0 9.5c-1.38 0-2.5-1.12-2.5-2.5s1.12-2.5 2.5-2.5 2.5 1.12 2.5 2.5-1.12 2.5-2.5 2.5z"/>
                            </svg>
                            ${project.name} - ${project.address}
                        </div>
                        ${task.start_date || task.end_date
                            ? html`
                                  <div class="meta-row">
                                      <svg viewBox="0 0 24 24" fill="currentColor">
                                          <path d="M19 3h-1V1h-2v2H8V1H6v2H5c-1.11 0-1.99.9-1.99 2L3 19c0 1.1.89 2 2 2h14c1.1 0 2-.9 2-2V5c0-1.1-.9-2-2-2zm0 16H5V8h14v11zM9 10H7v2h2v-2zm4 0h-2v2h2v-2zm4 0h-2v2h2v-2z"/>
                                      </svg>
                                      ${task.start_date ?? ''} ${task.start_date && task.end_date ? ' - ' : ''} ${task.end_date ?? ''}
                                  </div>
                              `
                            : nothing}
                    </div>
                </div>

                ${actionType === 'status_update'
                    ? html`
                          <div>
                              <h2 class="section-title">Update Status</h2>
                              <fb-status-toggle
                                  value="${this._selectedStatus ?? 'pending'}"
                                  ?disabled=${this._submitting}
                                  @fb-status-change=${this._handleStatusChange.bind(this)}
                              ></fb-status-toggle>
                          </div>
                          <button
                              class="submit-btn"
                              ?disabled=${this._submitting || !this._selectedStatus}
                              @click=${this._handleSubmit.bind(this)}
                          >
                              ${this._submitting ? 'Submitting...' : 'Submit Update'}
                          </button>
                      `
                    : nothing}

                ${actionType === 'photo_upload'
                    ? html`
                          <div>
                              <h2 class="section-title">Upload Photo</h2>
                              <fb-photo-upload></fb-photo-upload>
                          </div>
                      `
                    : nothing}

                ${actionType === 'view'
                    ? html`
                          <div class="upsell">
                              <p class="upsell-title">This is a view-only link.</p>
                              <a class="upsell-link" href="/portal/signup">Create an account to make updates</a>
                          </div>
                      `
                    : nothing}
            </div>
        `;
    }

    override render(): TemplateResult {
        let content: TemplateResult;

        if (this._loading) {
            content = this._renderLoading();
        } else if (this._error) {
            content = this._renderError();
        } else if (this._submitted) {
            content = this._renderSuccess();
        } else {
            content = this._renderContent();
        }

        return html`
            <fb-portal-shell minimal>
                ${content}
            </fb-portal-shell>
        `;
    }
}

declare global {
    interface HTMLElementTagNameMap {
        'fb-view-portal-action': FBViewPortalAction;
    }
}
