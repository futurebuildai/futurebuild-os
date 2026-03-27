/**
 * DOM-rendering tests for fb-settings-brain.
 *
 * Phase 18 ERP Transition — Frontend Test Harness Validation
 * Validates the OS-to-Brain Bridge renders successfully with active A2A connections.
 */
import { fixture, html, expect, elementUpdated } from '@open-wc/testing';
import sinon from 'sinon';
import { api } from '../src/services/api';
import type { BrainConnectionResponse } from '../src/services/api';
import type { ActiveAgentConnection, A2AExecutionLog } from '../src/types/a2a';

// Import component to register the custom element
import '../src/components/settings/fb-settings-brain';

// Cast api.settings to mutable so sinon can stub readonly `as const` properties
const settingsApi = api.settings as Record<string, unknown>;

// ---------------------------------------------------------------------------
// Mock Data
// ---------------------------------------------------------------------------

const mockConn: BrainConnectionResponse = {
    id: 'conn-1',
    org_id: 'org-1',
    brain_url: 'https://brain.test',
    integration_key: 'fbk_test_abc123',
    status: 'connected',
    last_sync_at: '2024-06-15T12:00:00Z',
    platforms: [
        { name: 'QuickBooks', type: 'accounting', status: 'active' },
        { name: 'Procore', type: 'project_mgmt', status: 'inactive' },
    ],
    updated_at: '2024-06-15T12:00:00Z',
};

const mockAgents: ActiveAgentConnection[] = [
    {
        id: 'a1',
        org_id: 'org-1',
        agent_name: 'ProcurementSync',
        agent_type: 'cron',
        status: 'active',
        execution_count: 42,
        error_count: 1,
        last_execution_at: '2024-06-15T05:00:00Z',
        created_at: '2024-01-01T00:00:00Z',
        updated_at: '2024-06-15T05:00:00Z',
    },
    {
        id: 'a2',
        org_id: 'org-1',
        agent_name: 'DailyFocusSync',
        agent_type: 'cron',
        status: 'paused',
        execution_count: 10,
        error_count: 0,
        created_at: '2024-01-01T00:00:00Z',
        updated_at: '2024-06-15T05:00:00Z',
    },
];

const mockLogs: A2AExecutionLog[] = [
    {
        id: 'l1',
        org_id: 'org-1',
        source_system: 'FutureBuild',
        target_system: 'QuickBooks',
        action_type: 'invoice_sync',
        status: 'completed',
        duration_ms: 450,
        executed_at: '2024-06-15T12:00:00Z',
        created_at: '2024-06-15T12:00:00Z',
    },
    {
        id: 'l2',
        org_id: 'org-1',
        source_system: 'FutureBuild',
        target_system: 'Procore',
        action_type: 'schedule_push',
        status: 'failed',
        error_message: 'Connection timeout',
        duration_ms: 30000,
        executed_at: '2024-06-15T11:00:00Z',
        created_at: '2024-06-15T11:00:00Z',
    },
    {
        id: 'l3',
        org_id: 'org-1',
        source_system: 'Brain',
        target_system: 'FutureBuild',
        action_type: 'config_sync',
        status: 'pending',
        executed_at: '2024-06-15T10:00:00Z',
        created_at: '2024-06-15T10:00:00Z',
    },
];

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

/** Click a shadow DOM element matching `selector`. */
function clickShadow(el: HTMLElement, selector: string) {
    const target = el.shadowRoot!.querySelector<HTMLElement>(selector);
    if (!target) throw new Error(`clickShadow: no element for "${selector}"`);
    target.click();
}

/** Query all matching elements in shadow DOM. */
function queryAll(el: HTMLElement, selector: string) {
    return el.shadowRoot!.querySelectorAll(selector);
}

/** Query a single element in shadow DOM. */
function query(el: HTMLElement, selector: string) {
    return el.shadowRoot!.querySelector(selector);
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

describe('fb-settings-brain', () => {
    let getBrainStub: sinon.SinonStub;
    let getBrainAgentsStub: sinon.SinonStub;
    let getBrainLogsStub: sinon.SinonStub;

    beforeEach(() => {
        // Stubs must be set BEFORE fixture() so connectedCallback picks them up
        getBrainStub = sinon.stub(settingsApi, 'getBrain');
        getBrainAgentsStub = sinon.stub(settingsApi, 'getBrainAgents');
        getBrainLogsStub = sinon.stub(settingsApi, 'getBrainLogs');
    });

    afterEach(() => {
        sinon.restore();
    });

    it('renders loading state initially', async () => {
        // Stubs that never resolve — component stays in loading
        getBrainStub.returns(new Promise(() => {}));
        getBrainAgentsStub.returns(new Promise(() => {}));
        getBrainLogsStub.returns(new Promise(() => {}));

        const el = await fixture(html`<fb-settings-brain></fb-settings-brain>`);
        const shadow = el.shadowRoot!;

        // Should show loading text since _loading is true and _conn is null
        expect(shadow.textContent).to.include('Loading...');
    });

    it('renders integrations tab by default after data loads', async () => {
        getBrainStub.resolves(mockConn);
        getBrainAgentsStub.resolves(mockAgents);
        getBrainLogsStub.resolves(mockLogs);

        const el = await fixture(html`<fb-settings-brain></fb-settings-brain>`);
        // Wait for async loads + re-render
        await elementUpdated(el);
        // May need an extra tick for all 3 async calls to settle
        await new Promise(r => setTimeout(r, 0));
        await elementUpdated(el);

        const shadow = el.shadowRoot!;

        // The active tab should be "Integrations"
        const activeTab = shadow.querySelector('.tab-btn.active');
        expect(activeTab).to.exist;
        expect(activeTab!.textContent!.trim()).to.equal('Integrations');

        // Connection status card should render
        const cardTitle = shadow.querySelector('.card-title');
        expect(cardTitle).to.exist;
        expect(cardTitle!.textContent!.trim()).to.equal('Connection Status');
    });

    it('renders agent list on Agents tab', async () => {
        getBrainStub.resolves(mockConn);
        getBrainAgentsStub.resolves(mockAgents);
        getBrainLogsStub.resolves(mockLogs);

        const el = await fixture(html`<fb-settings-brain></fb-settings-brain>`);
        await elementUpdated(el);
        await new Promise(r => setTimeout(r, 0));
        await elementUpdated(el);

        // Click "Active Agents" tab
        clickShadow(el, '.tab-btn');
        await elementUpdated(el);

        const agents = queryAll(el, '.agent-item');
        expect(agents.length).to.equal(2);

        // Verify agent names
        const names = Array.from(queryAll(el, '.agent-name')).map(n => n.textContent!.trim());
        expect(names).to.include('ProcurementSync');
        expect(names).to.include('DailyFocusSync');
    });

    it('shows pause/resume toggle per agent', async () => {
        getBrainStub.resolves(mockConn);
        getBrainAgentsStub.resolves(mockAgents);
        getBrainLogsStub.resolves(mockLogs);

        const el = await fixture(html`<fb-settings-brain></fb-settings-brain>`);
        await elementUpdated(el);
        await new Promise(r => setTimeout(r, 0));
        await elementUpdated(el);

        // Switch to Agents tab
        clickShadow(el, '.tab-btn');
        await elementUpdated(el);

        const toggles = Array.from(queryAll(el, '.btn-toggle'));
        expect(toggles.length).to.equal(2);

        // First agent is active → "Pause", second is paused → "Resume"
        expect(toggles[0]!.textContent!.trim()).to.equal('Pause');
        expect(toggles[1]!.textContent!.trim()).to.equal('Resume');
    });

    it('renders execution logs table on Logs tab', async () => {
        getBrainStub.resolves(mockConn);
        getBrainAgentsStub.resolves(mockAgents);
        getBrainLogsStub.resolves(mockLogs);

        const el = await fixture(html`<fb-settings-brain></fb-settings-brain>`);
        await elementUpdated(el);
        await new Promise(r => setTimeout(r, 0));
        await elementUpdated(el);

        // Click "Execution Logs" tab (third .tab-btn)
        const tabs = queryAll(el, '.tab-btn');
        (tabs[2] as HTMLElement).click();
        await elementUpdated(el);

        // Verify table exists
        const table = query(el, '.logs-table');
        expect(table).to.exist;

        // Verify 3 data rows (thead has 1 row + tbody has 3)
        const dataRows = table!.querySelectorAll('tbody tr');
        expect(dataRows.length).to.equal(3);

        // Verify status badges
        expect(table!.querySelector('.badge-completed')).to.exist;
        expect(table!.querySelector('.badge-failed')).to.exist;
        expect(table!.querySelector('.badge-pending')).to.exist;
    });

    it('shows connection status indicator', async () => {
        getBrainStub.resolves(mockConn);
        getBrainAgentsStub.resolves(mockAgents);
        getBrainLogsStub.resolves(mockLogs);

        const el = await fixture(html`<fb-settings-brain></fb-settings-brain>`);
        await elementUpdated(el);
        await new Promise(r => setTimeout(r, 0));
        await elementUpdated(el);

        // Default tab is integrations — should show green connected dot
        const dot = query(el, '.status-dot.connected');
        expect(dot).to.exist;
    });

    it('shows empty state when no agents', async () => {
        getBrainStub.resolves(mockConn);
        getBrainAgentsStub.resolves([]);
        getBrainLogsStub.resolves(mockLogs);

        const el = await fixture(html`<fb-settings-brain></fb-settings-brain>`);
        await elementUpdated(el);
        await new Promise(r => setTimeout(r, 0));
        await elementUpdated(el);

        // Switch to Agents tab
        clickShadow(el, '.tab-btn');
        await elementUpdated(el);

        const empty = query(el, '.empty-state');
        expect(empty).to.exist;
        expect(empty!.textContent).to.include('No active agent connections');
    });

    it('shows empty state when no logs', async () => {
        getBrainStub.resolves(mockConn);
        getBrainAgentsStub.resolves(mockAgents);
        getBrainLogsStub.resolves([]);

        const el = await fixture(html`<fb-settings-brain></fb-settings-brain>`);
        await elementUpdated(el);
        await new Promise(r => setTimeout(r, 0));
        await elementUpdated(el);

        // Click "Execution Logs" tab (third .tab-btn)
        const tabs = queryAll(el, '.tab-btn');
        (tabs[2] as HTMLElement).click();
        await elementUpdated(el);

        const empty = query(el, '.empty-state');
        expect(empty).to.exist;
        expect(empty!.textContent).to.include('No execution logs');
    });
});
