import { html, css, TemplateResult, nothing } from 'lit';
import { customElement, state } from 'lit/decorators.js';
import { FBViewElement } from '../base/FBViewElement';
import { api } from '../../services/api';
import type { FleetAsset, MaintenanceLog } from '../../types/fleet';

@customElement('fb-view-fleet')
export class FBViewFleet extends FBViewElement {
  @state() private _assets: FleetAsset[] = [];
  @state() private _upcomingMaintenance: MaintenanceLog[] = [];
  @state() private _viewState: 'loading' | 'ready' | 'error' = 'loading';
  @state() private _statusFilter: string = 'all';
  @state() private _showAddForm = false;

  static override styles = [
    FBViewElement.styles,
    css`
      :host { display: block; padding: var(--fb-spacing-lg); background: var(--fb-bg-primary); min-height: 100%; }

      .header { display: flex; justify-content: space-between; align-items: center; margin-bottom: var(--fb-spacing-lg); }
      .header h1 { margin: 0; font-size: 24px; color: var(--fb-text-primary); font-weight: 600; }

      .btn-add { padding: 8px 18px; border-radius: 8px; border: none; background: var(--fb-primary); color: #0a0b0f;
        font-size: 13px; font-weight: 600; cursor: pointer; transition: opacity 0.15s; }
      .btn-add:hover { opacity: 0.85; }

      .filters { display: flex; gap: var(--fb-spacing-sm); margin-bottom: var(--fb-spacing-lg); flex-wrap: wrap; }
      .filter-chip { padding: 6px 14px; border-radius: 16px; border: 1px solid rgba(255,255,255,0.1);
        background: transparent; color: var(--fb-text-secondary); font-size: 13px; cursor: pointer; transition: all 0.15s; }
      .filter-chip:hover { border-color: rgba(255,255,255,0.2); }
      .filter-chip.active { background: rgba(0,255,163,0.1); color: var(--fb-primary); border-color: rgba(0,255,163,0.3); }

      .maintenance-alert { padding: var(--fb-spacing-md); border-radius: 10px; background: rgba(245,158,11,0.08);
        border: 1px solid rgba(245,158,11,0.2); color: #f59e0b; font-size: 13px; margin-bottom: var(--fb-spacing-lg); }

      .asset-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(280px, 1fr)); gap: 16px; }

      .asset-card { background: rgba(22,24,33,0.6); backdrop-filter: blur(24px); border: 1px solid rgba(255,255,255,0.05);
        border-radius: 12px; padding: var(--fb-spacing-md); transition: border-color 0.15s; }
      .asset-card:hover { border-color: rgba(255,255,255,0.12); }

      .card-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: var(--fb-spacing-sm); }
      .asset-number { font-size: 15px; font-weight: 600; color: var(--fb-text-primary); }

      .status-badge { padding: 3px 10px; border-radius: 10px; font-size: 11px; font-weight: 500; text-transform: capitalize; }
      .status-badge.available { background: rgba(0,255,163,0.12); color: #00FFA3; }
      .status-badge.in_use { background: rgba(59,130,246,0.12); color: #3B82F6; }
      .status-badge.maintenance { background: rgba(245,158,11,0.12); color: #f59e0b; }
      .status-badge.retired { background: rgba(139,141,152,0.12); color: #8B8D98; }

      .card-body { margin-bottom: var(--fb-spacing-sm); }
      .card-body .make-model { font-size: 14px; color: var(--fb-text-primary); margin-bottom: 2px; }
      .card-body .detail { font-size: 12px; color: var(--fb-text-secondary); }

      .card-footer { display: flex; justify-content: space-between; align-items: center;
        padding-top: var(--fb-spacing-sm); border-top: 1px solid rgba(255,255,255,0.05); }
      .card-footer .value { font-size: 14px; font-weight: 600; color: var(--fb-primary); }
      .card-footer .location { font-size: 12px; color: var(--fb-text-secondary); }

      .empty-state, .loading-state, .error-state { display: flex; flex-direction: column; align-items: center;
        justify-content: center; padding: var(--fb-spacing-xl) 0; color: var(--fb-text-secondary); gap: var(--fb-spacing-sm); }
      .error-state .retry { padding: 6px 16px; border-radius: 8px; border: 1px solid rgba(255,255,255,0.1);
        background: transparent; color: var(--fb-text-primary); cursor: pointer; font-size: 13px; }

      .add-form { background: rgba(22,24,33,0.6); backdrop-filter: blur(24px); border: 1px solid rgba(255,255,255,0.05);
        border-radius: 12px; padding: var(--fb-spacing-lg); margin-bottom: var(--fb-spacing-lg); }
      .add-form h2 { margin: 0 0 var(--fb-spacing-md); font-size: 16px; color: var(--fb-text-primary); }
      .form-grid { display: grid; grid-template-columns: 1fr 1fr; gap: var(--fb-spacing-sm); }
      .form-field { display: flex; flex-direction: column; gap: 4px; }
      .form-field label { font-size: 12px; color: var(--fb-text-secondary); }
      .form-field input, .form-field select { padding: 8px 10px; border-radius: 8px; border: 1px solid rgba(255,255,255,0.1);
        background: rgba(10,11,15,0.5); color: var(--fb-text-primary); font-size: 13px; outline: none; }
      .form-field input:focus, .form-field select:focus { border-color: rgba(0,255,163,0.4); }
      .form-actions { display: flex; gap: var(--fb-spacing-sm); margin-top: var(--fb-spacing-md); justify-content: flex-end; }
      .btn-cancel { padding: 8px 16px; border-radius: 8px; border: 1px solid rgba(255,255,255,0.1);
        background: transparent; color: var(--fb-text-secondary); cursor: pointer; font-size: 13px; }
      .btn-submit { padding: 8px 18px; border-radius: 8px; border: none; background: var(--fb-primary);
        color: #0a0b0f; font-weight: 600; cursor: pointer; font-size: 13px; }
    `,
  ];

  override async onViewActive(): Promise<void> {
    await this._loadAssets();
    await this._loadMaintenance();
  }

  private async _loadAssets(): Promise<void> {
    try {
      this._viewState = 'loading';
      const filter = this._statusFilter === 'all' ? undefined : this._statusFilter;
      this._assets = await api.fleet.list(filter);
      this._viewState = 'ready';
    } catch {
      this._viewState = 'error';
    }
  }

  private async _loadMaintenance(): Promise<void> {
    try {
      this._upcomingMaintenance = await api.fleet.getUpcomingMaintenance(14);
    } catch {
      this._upcomingMaintenance = [];
    }
  }

  private async _onFilterChange(status: string): Promise<void> {
    this._statusFilter = status;
    await this._loadAssets();
  }

  private async _onAddSubmit(e: SubmitEvent): Promise<void> {
    e.preventDefault();
    const form = e.target as HTMLFormElement;
    const data = new FormData(form);
    const costStr = data.get('purchase_cost') as string | null;
    const yearStr = data.get('year') as string | null;

    try {
      const newAsset: Partial<FleetAsset> = {
        asset_number: (data.get('asset_number') as string) ?? '',
        asset_type: (data.get('asset_type') as string) ?? '',
        make: (data.get('make') as string) ?? '',
        model: (data.get('model') as string) ?? '',
        status: (data.get('status') as FleetAsset['status']) ?? 'available',
      };
      if (yearStr) newAsset.year = parseInt(yearStr, 10);
      const vin = data.get('vin') as string;
      if (vin) newAsset.vin = vin;
      const plate = data.get('license_plate') as string;
      if (plate) newAsset.license_plate = plate;
      if (costStr) newAsset.purchase_cost_cents = Math.round(parseFloat(costStr) * 100);
      await api.fleet.create(newAsset);
      this._showAddForm = false;
      await this._loadAssets();
    } catch {
      // keep form open on error
    }
  }

  private _formatCents(cents: number): string {
    return (cents / 100).toLocaleString('en-US', { style: 'currency', currency: 'USD' });
  }

  private _renderFilters(): TemplateResult {
    const statuses = ['all', 'available', 'in_use', 'maintenance', 'retired'] as const;
    const labels: Record<string, string> = { all: 'All', available: 'Available', in_use: 'In Use', maintenance: 'Maintenance', retired: 'Retired' };
    return html`
      <div class="filters">
        ${statuses.map(s => html`
          <button class="filter-chip ${this._statusFilter === s ? 'active' : ''}"
            @click=${() => this._onFilterChange(s)}>${labels[s]}</button>
        `)}
      </div>
    `;
  }

  private _renderMaintenanceAlert(): TemplateResult | typeof nothing {
    if (this._upcomingMaintenance.length === 0) return nothing;
    return html`
      <div class="maintenance-alert">
        \uD83D\uDD27 ${this._upcomingMaintenance.length} maintenance item(s) due within 14 days
      </div>
    `;
  }

  private _renderAssetCard(asset: FleetAsset): TemplateResult {
    return html`
      <div class="asset-card">
        <div class="card-header">
          <span class="asset-number">${asset.asset_number}</span>
          <span class="status-badge ${asset.status}">${asset.status.replace('_', ' ')}</span>
        </div>
        <div class="card-body">
          <div class="make-model">${asset.make} ${asset.model}</div>
          <div class="detail">${asset.year != null ? asset.year : ''} &middot; ${asset.asset_type}</div>
        </div>
        <div class="card-footer">
          <span class="value">${asset.current_value_cents != null ? this._formatCents(asset.current_value_cents) : '--'}</span>
          <span class="location">${asset.location ?? ''}</span>
        </div>
      </div>
    `;
  }

  private _renderAddForm(): TemplateResult | typeof nothing {
    if (!this._showAddForm) return nothing;
    return html`
      <div class="add-form">
        <h2>Add New Asset</h2>
        <form @submit=${this._onAddSubmit}>
          <div class="form-grid">
            <div class="form-field"><label>Asset Number</label><input name="asset_number" required /></div>
            <div class="form-field"><label>Asset Type</label><input name="asset_type" required /></div>
            <div class="form-field"><label>Make</label><input name="make" required /></div>
            <div class="form-field"><label>Model</label><input name="model" required /></div>
            <div class="form-field"><label>Year</label><input name="year" type="number" /></div>
            <div class="form-field"><label>VIN</label><input name="vin" /></div>
            <div class="form-field"><label>License Plate</label><input name="license_plate" /></div>
            <div class="form-field"><label>Purchase Cost ($)</label><input name="purchase_cost" type="number" step="0.01" /></div>
            <div class="form-field">
              <label>Status</label>
              <select name="status">
                <option value="available">Available</option>
                <option value="in_use">In Use</option>
                <option value="maintenance">Maintenance</option>
                <option value="retired">Retired</option>
              </select>
            </div>
          </div>
          <div class="form-actions">
            <button type="button" class="btn-cancel" @click=${() => { this._showAddForm = false; }}>Cancel</button>
            <button type="submit" class="btn-submit">Create Asset</button>
          </div>
        </form>
      </div>
    `;
  }

  override render(): TemplateResult {
    return html`
      <div class="header">
        <h1>Fleet &amp; Equipment</h1>
        <button class="btn-add" @click=${() => { this._showAddForm = !this._showAddForm; }}>Add Asset</button>
      </div>

      ${this._renderAddForm()}
      ${this._renderMaintenanceAlert()}
      ${this._renderFilters()}

      ${this._viewState === 'loading' ? html`
        <div class="loading-state"><span>Loading fleet assets...</span></div>
      ` : this._viewState === 'error' ? html`
        <div class="error-state">
          <span>Failed to load fleet data</span>
          <button class="retry" @click=${() => this._loadAssets()}>Retry</button>
        </div>
      ` : this._assets.length === 0 ? html`
        <div class="empty-state"><span>No assets found</span></div>
      ` : html`
        <div class="asset-grid">
          ${this._assets.map(a => this._renderAssetCard(a))}
        </div>
      `}
    `;
  }
}

declare global {
  interface HTMLElementTagNameMap {
    'fb-view-fleet': FBViewFleet;
  }
}
