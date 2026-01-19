import { InvoiceArtifactData } from '../types/artifacts';

export const MOCK_INVOICE_DATA: InvoiceArtifactData = {
    vendor: 'ABC Lumber Co.',
    date: '2024-01-15T00:00:00Z',
    invoice_number: 'INV-2024-001',
    address: '123 Tree St, Woodville, OR',
    total_amount: 5600.00, // Sum: 2250 + 3200 + 150
    suggested_wbs_code: '5.1',
    confidence: 0.98,
    line_items: [
        { description: '2x4x8 Doug Fir Studs', quantity: 500, unit_price: 4.50, total: 2250.00 },
        { description: '3/4" Plywood Sheets (4x8)', quantity: 100, unit_price: 32.00, total: 3200.00 },
        { description: 'Delivery Fee', quantity: 1, unit_price: 150.00, total: 150.00 },
    ]
};
