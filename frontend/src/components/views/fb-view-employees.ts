import { html, css, TemplateResult, nothing } from 'lit';
import { customElement, state } from 'lit/decorators.js';
import { FBViewElement } from '../base/FBViewElement';
import { api } from '../../services/api';
import type { Employee, Certification } from '../../types/employee';
import '../shared/fb-modal';

type ViewState = 'loading' | 'ready' | 'error';
type StatusFilter = 'all' | 'active' | 'on_leave' | 'terminated';

const STATUS_FILTERS: { label: string; value: StatusFilter }[] = [
  { label: 'All', value: 'all' },
  { label: 'Active', value: 'active' },
  { label: 'On Leave', value: 'on_leave' },
  { label: 'Terminated', value: 'terminated' },
];

@customElement('fb-view-employees')
export class FBViewEmployees extends FBViewElement {
  static override styles = [
    FBViewElement.styles,
    css`
      :host { display: block; padding: var(--fb-spacing-lg, 24px); overflow-y: auto; }
      .header { display: flex; justify-content: space-between; align-items: center; margin-bottom: var(--fb-spacing-lg, 24px); }
      .header h1 { margin: 0; font-size: var(--fb-text-2xl, 24px); color: var(--fb-text-primary, #F0F0F5); font-weight: 600; }
      .btn-add {
        padding: 8px 18px; border-radius: var(--fb-radius-md, 8px); border: none;
        background: var(--fb-primary, #00FFA3); color: #0A0B10; font-size: var(--fb-text-sm, 13px);
        font-weight: 600; cursor: pointer; transition: opacity 0.15s;
      }
      .btn-add:hover { opacity: 0.85; }
      .filters { display: flex; gap: var(--fb-spacing-sm, 8px); margin-bottom: var(--fb-spacing-lg, 24px); flex-wrap: wrap; }
      .filter-chip {
        padding: 6px 14px; border-radius: 16px; border: 1px solid rgba(255,255,255,0.1);
        background: transparent; color: var(--fb-text-secondary, #8B8D98);
        font-size: 13px; cursor: pointer; transition: all 0.15s;
      }
      .filter-chip:hover { border-color: rgba(255,255,255,0.2); }
      .filter-chip.active {
        background: rgba(0, 255, 163, 0.1); color: var(--fb-primary, #00FFA3);
        border-color: rgba(0, 255, 163, 0.3);
      }
      .cert-warning {
        padding: var(--fb-spacing-sm, 8px) var(--fb-spacing-md, 16px); border-radius: var(--fb-radius-md, 8px);
        background: rgba(245,158,11,0.1); border: 1px solid rgba(245,158,11,0.3);
        color: #f59e0b; font-size: var(--fb-text-sm, 13px); margin-bottom: var(--fb-spacing-lg, 24px);
      }
      .card {
        background: rgba(22, 24, 33, 0.6); backdrop-filter: blur(24px);
        border: 1px solid rgba(255,255,255,0.05); border-radius: var(--fb-radius-lg, 12px);
        padding: var(--fb-spacing-md, 16px); margin-bottom: var(--fb-spacing-sm, 8px);
        display: grid; grid-template-columns: 2fr 1fr 1fr 1fr 1fr; align-items: center; gap: var(--fb-spacing-md, 16px);
      }
      .card-header {
        display: grid; grid-template-columns: 2fr 1fr 1fr 1fr 1fr; gap: var(--fb-spacing-md, 16px);
        padding: 0 var(--fb-spacing-md, 16px) var(--fb-spacing-sm, 8px);
        color: var(--fb-text-secondary, #8B8D98); font-size: var(--fb-text-xs, 11px);
        text-transform: uppercase; letter-spacing: 0.5px; font-weight: 600;
      }
      .emp-name { color: var(--fb-text-primary, #F0F0F5); font-size: var(--fb-text-md, 14px); font-weight: 500; }
      .emp-email { color: var(--fb-text-secondary, #8B8D98); font-size: var(--fb-text-xs, 11px); margin-top: 2px; }
      .emp-detail { color: var(--fb-text-secondary, #8B8D98); font-size: var(--fb-text-sm, 13px); }
      .status-badge { display: inline-block; padding: 3px 8px; border-radius: 4px; font-size: 11px; font-weight: 600; text-transform: uppercase; }
      .status-badge.active { background: rgba(0,255,163,0.12); color: #00FFA3; }
      .status-badge.on_leave { background: rgba(245,158,11,0.12); color: #f59e0b; }
      .status-badge.terminated { background: rgba(139,141,152,0.12); color: #8B8D98; }
      .empty-state, .loading-state, .error-state {
        display: flex; flex-direction: column; align-items: center; justify-content: center;
        padding: var(--fb-spacing-xl, 32px); color: var(--fb-text-secondary, #8B8D98);
        font-size: var(--fb-text-md, 14px); min-height: 200px;
      }
      .error-state { color: #ef4444; }
      .error-state button {
        margin-top: var(--fb-spacing-md, 16px); padding: 8px 16px; border-radius: var(--fb-radius-md, 8px);
        border: 1px solid rgba(255,255,255,0.1); background: transparent;
        color: var(--fb-text-primary, #F0F0F5); cursor: pointer; font-size: var(--fb-text-sm, 13px);
      }
      .form-overlay {
        background: rgba(22, 24, 33, 0.6); backdrop-filter: blur(24px);
        border: 1px solid rgba(255,255,255,0.05); border-radius: var(--fb-radius-lg, 12px);
        padding: var(--fb-spacing-lg, 24px); margin-bottom: var(--fb-spacing-lg, 24px);
      }
      .form-overlay h2 { margin: 0 0 var(--fb-spacing-md, 16px); font-size: var(--fb-text-xl, 18px); color: var(--fb-text-primary, #F0F0F5); }
      .form-row { display: grid; grid-template-columns: 1fr 1fr; gap: var(--fb-spacing-md, 16px); margin-bottom: var(--fb-spacing-md, 16px); }
      .form-field label {
        display: block; margin-bottom: var(--fb-spacing-xs, 4px); font-size: var(--fb-text-xs, 11px);
        color: var(--fb-text-secondary, #8B8D98); text-transform: uppercase; letter-spacing: 0.5px;
      }
      .form-field input, .form-field select {
        width: 100%; box-sizing: border-box; padding: 8px 12px; border-radius: var(--fb-radius-md, 8px);
        border: 1px solid rgba(255,255,255,0.1); background: rgba(10,11,16,0.6);
        color: var(--fb-text-primary, #F0F0F5); font-size: var(--fb-text-sm, 13px);
      }
      .form-field input:focus, .form-field select:focus { outline: none; border-color: rgba(0,255,163,0.4); }
      .form-actions { display: flex; gap: var(--fb-spacing-sm, 8px); justify-content: flex-end; margin-top: var(--fb-spacing-md, 16px); }
      .btn-cancel {
        padding: 8px 16px; border-radius: var(--fb-radius-md, 8px);
        border: 1px solid rgba(255,255,255,0.1); background: transparent;
        color: var(--fb-text-secondary, #8B8D98); font-size: var(--fb-text-sm, 13px); cursor: pointer;
      }
      .btn-submit {
        padding: 8px 18px; border-radius: var(--fb-radius-md, 8px); border: none;
        background: var(--fb-primary, #00FFA3); color: #0A0B10;
        font-size: var(--fb-text-sm, 13px); font-weight: 600; cursor: pointer;
      }
      .btn-submit:hover { opacity: 0.85; }
      
      .skeleton { background: rgba(255, 255, 255, 0.05); border-radius: 4px; position: relative; overflow: hidden; }
      .skeleton::after {
        content: ''; position: absolute; top: 0; left: -100%; width: 100%; height: 100%;
        background: linear-gradient(90deg, transparent, rgba(255, 255, 255, 0.08), transparent);
        animation: fb-skeleton-shimmer 1.5s infinite;
      }
      @keyframes fb-skeleton-shimmer { 100% { left: 100%; } }
    `,
  ];

  @state() private _employees: Employee[] = [];
  @state() private _certWarnings: Certification[] = [];
  @state() private _viewState: ViewState = 'loading';
  @state() private _statusFilter: StatusFilter = 'all';
  @state() private _showAddForm = false;

  override async onViewActive(): Promise<void> {
    await this._loadEmployees();
    await this._loadCertWarnings();
  }

  private async _loadEmployees(): Promise<void> {
    try {
      this._viewState = 'loading';
      const filter = this._statusFilter === 'all' ? undefined : this._statusFilter;
      this._employees = await api.employees.list(filter);
      this._viewState = 'ready';
    } catch {
      this._viewState = 'error';
    }
  }

  private async _loadCertWarnings(): Promise<void> {
    try {
      this._certWarnings = await api.employees.getExpiringCertifications(30);
    } catch {
      // Non-critical — cert warnings are supplementary
    }
  }

  private async _handleFilterChange(filter: StatusFilter): Promise<void> {
    this._statusFilter = filter;
    await this._loadEmployees();
  }

  private async _handleSubmit(e: SubmitEvent): Promise<void> {
    e.preventDefault();
    const form = e.target as HTMLFormElement;
    const data = new FormData(form);
    try {
      const newEmp: Partial<Employee> = {
        first_name: data.get('first_name') as string,
        last_name: data.get('last_name') as string,
        email: data.get('email') as string,
        classification: data.get('classification') as string,
        pay_rate_cents: Math.round(parseFloat(data.get('pay_rate') as string || '0') * 100),
        pay_type: data.get('pay_type') as 'hourly' | 'salary',
      };
      const phone = data.get('phone') as string;
      if (phone) newEmp.phone = phone;
      await api.employees.create(newEmp);
      this._showAddForm = false;
      await this._loadEmployees();
    } catch {
      // TODO: surface form-level error
    }
  }

  private _renderHeader(): TemplateResult {
    return html`
      <div class="header">
        <h1>Employees</h1>
        <button class="btn-add" @click=${() => { this._showAddForm = !this._showAddForm; }}>
          ${this._showAddForm ? 'Cancel' : '+ Add Employee'}
        </button>
      </div>
    `;
  }

  private _renderFilters(): TemplateResult {
    return html`
      <div class="filters">
        ${STATUS_FILTERS.map(
          (f) => html`
            <button
              class="filter-chip ${this._statusFilter === f.value ? 'active' : ''}"
              @click=${() => this._handleFilterChange(f.value)}
            >${f.label}</button>
          `,
        )}
      </div>
    `;
  }

  private _renderCertWarning(): TemplateResult | typeof nothing {
    if (this._certWarnings.length === 0) return nothing;
    return html`
      <div class="cert-warning">
        ⚠ ${this._certWarnings.length} certification(s) expiring within 30 days
      </div>
    `;
  }

  private _renderAddForm(): TemplateResult | typeof nothing {
    return html`
      <fb-modal .open=${this._showAddForm} heading="New Employee"
        @fb-modal-close=${() => { this._showAddForm = false; }}>
        <form @submit=${this._handleSubmit}>
          <div class="form-row">
            <div class="form-field">
              <label>First Name</label>
              <input name="first_name" required />
            </div>
            <div class="form-field">
              <label>Last Name</label>
              <input name="last_name" required />
            </div>
          </div>
          <div class="form-row">
            <div class="form-field">
              <label>Email</label>
              <input name="email" type="email" required />
            </div>
            <div class="form-field">
              <label>Phone</label>
              <input name="phone" type="tel" />
            </div>
          </div>
          <div class="form-row">
            <div class="form-field">
              <label>Classification</label>
              <input name="classification" required placeholder="e.g. Electrician, Foreman" />
            </div>
            <div class="form-field">
              <label>Pay Rate ($)</label>
              <input name="pay_rate" type="number" step="0.01" min="0" required />
            </div>
          </div>
          <div class="form-row">
            <div class="form-field">
              <label>Pay Type</label>
              <select name="pay_type" required>
                <option value="hourly">Hourly</option>
                <option value="salary">Salary</option>
              </select>
            </div>
            <div></div>
          </div>
          <div class="form-actions">
            <button type="button" class="btn-cancel" @click=${() => { this._showAddForm = false; }}>Cancel</button>
            <button type="submit" class="btn-submit">Create Employee</button>
          </div>
        </form>
      </fb-modal>
    `;
  }

  private _renderEmployeeList(): TemplateResult | typeof nothing {
    if (this._viewState === 'loading') {
      return html`
        <div class="card-header">
          <span>Name</span><span>Employee #</span><span>Status</span><span>Pay Rate</span><span>Classification</span>
        </div>
        ${[1, 2, 3, 4, 5].map(() => html`
          <div class="card">
            <div>
              <div class="skeleton" style="width: 120px; height: 16px; margin-bottom: 6px;"></div>
              <div class="skeleton" style="width: 160px; height: 12px;"></div>
            </div>
            <div class="skeleton" style="width: 80px; height: 14px;"></div>
            <div class="skeleton" style="width: 60px; height: 20px; border-radius: 10px;"></div>
            <div class="skeleton" style="width: 70px; height: 14px;"></div>
            <div class="skeleton" style="width: 90px; height: 14px;"></div>
          </div>
        `)}
      `;
    }
    if (this._viewState === 'error') {
      return html`
        <div class="error-state">
          Failed to load employees.
          <button @click=${() => this._loadEmployees()}>Retry</button>
        </div>
      `;
    }
    if (this._employees.length === 0) {
      return html`<div class="empty-state">No employees found.</div>`;
    }
    return html`
      <div class="card-header">
        <span>Name</span><span>Employee #</span><span>Status</span><span>Pay Rate</span><span>Classification</span>
      </div>
      ${this._employees.map((emp) => this._renderEmployeeCard(emp))}
    `;
  }

  private _renderEmployeeCard(emp: Employee): TemplateResult {
    const rate = `$${((emp.pay_rate_cents ?? 0) / 100).toFixed(2)}`;
    const payLabel = emp.pay_type === 'hourly' ? '/hr' : '/yr';
    return html`
      <div class="card">
        <div>
          <div class="emp-name">${emp.first_name} ${emp.last_name}</div>
          <div class="emp-email">${emp.email}</div>
        </div>
        <div class="emp-detail">${emp.employee_number ?? '—'}</div>
        <div><span class="status-badge ${emp.status}">${emp.status.replace('_', ' ')}</span></div>
        <div class="emp-detail">${rate}${payLabel}</div>
        <div class="emp-detail">${emp.classification}</div>
      </div>
    `;
  }

  protected override render(): TemplateResult {
    return html`
      ${this._renderHeader()}
      ${this._renderCertWarning()}
      ${this._renderAddForm()}
      ${this._renderFilters()}
      ${this._renderEmployeeList()}
    `;
  }
}

declare global {
  interface HTMLElementTagNameMap {
    'fb-view-employees': FBViewEmployees;
  }
}
