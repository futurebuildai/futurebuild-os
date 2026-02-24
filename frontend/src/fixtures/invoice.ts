import { InvoiceArtifactData } from '../types/artifacts';

export const MOCK_INVOICE_DATA: InvoiceArtifactData = {
    vendor: 'ABC Lumber Co.',
    date: '2024-01-15T00:00:00Z',
    invoice_number: 'INV-2024-001',
    address: '123 Tree St, Woodville, OR',
    total_amount_cents: '560000', // 5600.00 in cents as string
    suggested_wbs_code: '5.1',
    confidence: 0.98,
    line_items: [
        { description: '2x4x8 Doug Fir Studs', quantity: 500, unit_price_cents: '450', total_cents: '225000' },
        { description: '3/4" Plywood Sheets (4x8)', quantity: 100, unit_price_cents: '3200', total_cents: '320000' },
        { description: 'Delivery Fee', quantity: 1, unit_price_cents: '15000', total_cents: '15000' },
    ],
    fieldConfidences: [
        { field: 'vendor', score: 0.98, source: 'vision_extraction' },
        { field: 'date', score: 0.72, source: 'vision_extraction', boundingBox: { page: 1, x: 0.7, y: 0.05, w: 0.2, h: 0.03 } },
        { field: 'invoice_number', score: 0.95, source: 'vision_extraction' },
        { field: 'line_items[0].description', score: 0.91, source: 'vision_extraction' },
        { field: 'line_items[0].unit_price_cents', score: 0.65, source: 'vision_extraction', boundingBox: { page: 1, x: 0.6, y: 0.35, w: 0.15, h: 0.025 } },
        { field: 'line_items[1].description', score: 0.88, source: 'vision_extraction' },
        { field: 'line_items[1].unit_price_cents', score: 0.92, source: 'vision_extraction' },
        { field: 'line_items[2].description', score: 0.55, source: 'vision_extraction', boundingBox: { page: 1, x: 0.1, y: 0.55, w: 0.4, h: 0.025 } },
        { field: 'line_items[2].unit_price_cents', score: 0.60, source: 'vision_extraction', boundingBox: { page: 1, x: 0.6, y: 0.55, w: 0.15, h: 0.025 } },
    ],
};
