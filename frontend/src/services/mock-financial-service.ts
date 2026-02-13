
export interface FinancialSummary {
    project_id?: string;
    budget_total: number;
    spend_total: number;
    variance: number;
    last_updated: string;
    categories: {
        name: string;
        budget: number;
        spend: number;
        status: 'on_track' | 'at_risk' | 'over_budget';
    }[];
}

export const mockFinancialService = {
    async getSummary(projectId?: string): Promise<FinancialSummary> {
        await new Promise(resolve => setTimeout(resolve, 500));

        if (projectId === 'p1') { // Willow Creek
            return {
                project_id: 'p1',
                budget_total: 1250000,
                spend_total: 450000,
                variance: -12000, // slightly over
                last_updated: new Date().toISOString(),
                categories: [
                    { name: 'Site Work', budget: 150000, spend: 148000, status: 'on_track' },
                    { name: 'Concrete', budget: 200000, spend: 210000, status: 'at_risk' },
                    { name: 'Framing', budget: 300000, spend: 50000, status: 'on_track' },
                    { name: 'MEP', budget: 400000, spend: 42000, status: 'on_track' }
                ]
            };
        }

        // Global / Default
        return {
            budget_total: 5500000,
            spend_total: 2100000,
            variance: 45000, // under budget overall
            last_updated: new Date().toISOString(),
            categories: [
                { name: 'Labor', budget: 2000000, spend: 800000, status: 'on_track' },
                { name: 'Materials', budget: 3000000, spend: 1100000, status: 'on_track' },
                { name: 'Subcontractors', budget: 500000, spend: 200000, status: 'on_track' }
            ]
        };
    }
};
