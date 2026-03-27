import { html, css, TemplateResult, nothing } from 'lit';
import { customElement, property, state } from 'lit/decorators.js';
import { FBElement } from '../base/FBElement';
import { InvoiceStatus } from '../../types/enums';
import { api } from '../../services/api';

/**
 * Shared action bar for artifact approval/rejection.
 * Renders Approve/Reject buttons for Draft artifacts, and
 * finality badges for Approved/Rejected states.
 *
 * See STEP_83_APPROVAL_ACTIONS.md Section 1.1
 *
 * @fires artifact-approved - When the artifact is approved
 * @fires artifact-rejected - When the artifact is rejected
 */
@customElement('fb-artifact-actions')
export class FBArtifactActions extends FBElement {
    static override styles = [
        FBElement.styles,
        css`
            :host {
                display: block;
                width: 100%;
            }

            .actions-bar {
                display: flex;
                justify-content: space-between;
                align-items: center;
                padding: 12px 0;
                border-top: 1px solid #e5e7eb;
                margin-top: 8px;
            }

            .btn {
                padding: 8px 20px;
                border: none;
                border-radius: 6px;
                font-size: 13px;
                font-weight: 600;
                cursor: pointer;
                font-family: inherit;
                transition: background 0.15s, opacity 0.15s, transform 0.1s;
                min-height: 44px;
                min-width: 44px;
            }

            .btn:active:not(:disabled) { transform: scale(0.98); }
            .btn:disabled { opacity: 0.5; cursor: not-allowed; }

            .btn--approve {
                background: #059669;
                color: white;
            }
            .btn--approve:hover:not(:disabled) { background: #047857; }

            .btn--reject {
                background: #dc2626;
                color: white;
            }
            .btn--reject:hover:not(:disabled) { background: #b91c1c; }

            .btn-group {
                display: flex;
                gap: 8px;
            }

            /* Finality badges */
            .finality-badge {
                display: inline-flex;
                align-items: center;
                gap: 6px;
                padding: 6px 14px;
                border-radius: 6px;
                font-size: 13px;
                font-weight: 600;
            }

            .finality-badge--approved {
                background: #d1fae5;
                color: #065f46;
            }

            .finality-badge--rejected {
                background: #fee2e2;
                color: #991b1b;
            }

            .finality-meta {
                font-size: 11px;
                color: #6b7280;
                font-weight: 400;
            }

            /* Confirmation modal */
            .modal-overlay {
                position: fixed;
                top: 0;
                left: 0;
                right: 0;
                bottom: 0;
                background: rgba(0,0,0,0.5);
                display: flex;
                align-items: center;
                justify-content: center;
                z-index: 1000;
            }

            .modal {
                background: white;
                border-radius: 8px;
                padding: 24px;
                max-width: 400px;
                width: 90%;
                box-shadow: 0 20px 60px rgba(0,0,0,0.3);
                font-family: system-ui, sans-serif;
            }

            .modal h3 {
                margin: 0 0 12px;
                font-size: 16px;
                color: #111;
            }

            .modal p {
                margin: 0 0 16px;
                font-size: 13px;
                color: #4b5563;
                line-height: 1.5;
            }

            .modal textarea {
                width: 100%;
                min-height: 80px;
                padding: 8px;
                border: 1px solid #d1d5db;
                border-radius: 4px;
                font-family: inherit;
                font-size: 13px;
                resize: vertical;
                margin-bottom: 12px;
                box-sizing: border-box;
            }

            .modal textarea:focus {
                outline: none;
                border-color: #2563eb;
                box-shadow: 0 0 0 2px rgba(37,99,235,0.15);
            }

            .modal-actions {
                display: flex;
                justify-content: flex-end;
                gap: 8px;
            }

            .btn--modal-cancel {
                background: #f3f4f6;
                color: #374151;
            }
            .btn--modal-cancel:hover:not(:disabled) { background: #e5e7eb; }

            .error-text {
                color: #dc2626;
                font-size: 11px;
                margin-top: 4px;
            }
        `
    ];

    /** The invoice/artifact ID for API calls */
    @property({ type: String })
    artifactId = '';

    /** Current status of the artifact */
    @property({ type: String })
    status: InvoiceStatus = InvoiceStatus.Pending;

    /** Whether the user has approve permission */
    @property({ type: Boolean, attribute: 'can-approve' })
    canApprove = false;

    /** Name/email of the approver (for display) */
    @property({ type: String, attribute: 'approved-by' })
    approvedBy = '';

    /** Name/email of the rejector (for display) */
    @property({ type: String, attribute: 'rejected-by' })
    rejectedBy = '';

    /** Rejection reason (for display) */
    @property({ type: String, attribute: 'rejection-reason' })
    rejectionReason = '';

    @state()
    private _showApproveModal = false;

    @state()
    private _showRejectModal = false;

    @state()
    private _rejectReason = '';

    @state()
    private _isProcessing = false;

    @state()
    private _error = '';

    private async _handleApprove(): Promise<void> {
        if (!this.artifactId || this._isProcessing) return;

        this._isProcessing = true;
        this._error = '';

        try {
            const response = await api.invoices.approve(this.artifactId);
            this._showApproveModal = false;
            this.emit('artifact-approved', {
                artifactId: this.artifactId,
                status: response.status,
            });
        } catch (err: unknown) {
            this._error = err instanceof Error ? err.message : 'Approval failed';
        } finally {
            this._isProcessing = false;
        }
    }

    private async _handleReject(): Promise<void> {
        if (!this.artifactId || this._isProcessing) return;

        this._isProcessing = true;
        this._error = '';

        try {
            const reason = this._rejectReason;
            const response = await api.invoices.reject(this.artifactId, reason);
            this._showRejectModal = false;
            this._rejectReason = '';
            this.emit('artifact-rejected', {
                artifactId: this.artifactId,
                status: response.status,
                reason,
            });
        } catch (err: unknown) {
            this._error = err instanceof Error ? err.message : 'Rejection failed';
        } finally {
            this._isProcessing = false;
        }
    }

    private _renderApproveModal(): TemplateResult {
        return html`
            <div class="modal-overlay" @click=${(e: Event) => {
                if (e.target === e.currentTarget) this._showApproveModal = false;
            }}>
                <div class="modal">
                    <h3>Approve Invoice</h3>
                    <p>Are you sure? This will finalize the document. This action cannot be undone.</p>
                    ${this._error ? html`<div class="error-text">${this._error}</div>` : nothing}
                    <div class="modal-actions">
                        <button
                            class="btn btn--modal-cancel"
                            @click=${() => { this._showApproveModal = false; this._error = ''; }}
                            ?disabled=${this._isProcessing}
                        >Cancel</button>
                        <button
                            class="btn btn--approve"
                            @click=${this._handleApprove}
                            ?disabled=${this._isProcessing}
                        >${this._isProcessing ? 'Approving...' : 'Confirm Approve'}</button>
                    </div>
                </div>
            </div>
        `;
    }

    private _renderRejectModal(): TemplateResult {
        return html`
            <div class="modal-overlay" @click=${(e: Event) => {
                if (e.target === e.currentTarget) this._showRejectModal = false;
            }}>
                <div class="modal">
                    <h3>Reject Invoice</h3>
                    <p>Provide a reason for rejection (optional):</p>
                    <textarea
                        .value=${this._rejectReason}
                        @input=${(e: Event) => { this._rejectReason = (e.target as HTMLTextAreaElement).value; }}
                        placeholder="e.g., Labor costs exceed estimate..."
                        maxlength="2000"
                    ></textarea>
                    ${this._error ? html`<div class="error-text">${this._error}</div>` : nothing}
                    <div class="modal-actions">
                        <button
                            class="btn btn--modal-cancel"
                            @click=${() => { this._showRejectModal = false; this._error = ''; }}
                            ?disabled=${this._isProcessing}
                        >Cancel</button>
                        <button
                            class="btn btn--reject"
                            @click=${this._handleReject}
                            ?disabled=${this._isProcessing}
                        >${this._isProcessing ? 'Rejecting...' : 'Confirm Reject'}</button>
                    </div>
                </div>
            </div>
        `;
    }

    override render(): TemplateResult {
        // Draft: show action buttons
        if (this.status === InvoiceStatus.Draft && this.canApprove) {
            return html`
                <div class="actions-bar">
                    <div></div>
                    <div class="btn-group">
                        <button
                            class="btn btn--reject"
                            @click=${() => { this._showRejectModal = true; this._error = ''; }}
                        >Reject</button>
                        <button
                            class="btn btn--approve"
                            @click=${() => { this._showApproveModal = true; this._error = ''; }}
                        >Approve</button>
                    </div>
                </div>
                ${this._showApproveModal ? this._renderApproveModal() : nothing}
                ${this._showRejectModal ? this._renderRejectModal() : nothing}
            `;
        }

        // Approved: show finality badge
        if (this.status === InvoiceStatus.Approved) {
            return html`
                <div class="actions-bar">
                    <span class="finality-badge finality-badge--approved">
                        Approved
                        ${this.approvedBy ? html`<span class="finality-meta">by ${this.approvedBy}</span>` : nothing}
                    </span>
                </div>
            `;
        }

        // Rejected: show finality badge with reason
        if (this.status === InvoiceStatus.Rejected) {
            return html`
                <div class="actions-bar">
                    <span class="finality-badge finality-badge--rejected">
                        Rejected
                        ${this.rejectedBy ? html`<span class="finality-meta">by ${this.rejectedBy}</span>` : nothing}
                    </span>
                    ${this.rejectionReason
                        ? html`<span class="finality-meta">${this.rejectionReason}</span>`
                        : nothing
                    }
                </div>
            `;
        }

        // Other statuses: no actions shown
        return html``;
    }
}
