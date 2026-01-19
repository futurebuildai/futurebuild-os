import { BudgetArtifactData } from '../types/artifacts';

export const MOCK_BUDGET_DATA: BudgetArtifactData = {
    totalBudget: 450000,
    totalSpent: 125000,
    categories: [
        { name: 'Materials', budget: 200000, spent: 65000 },
        { name: 'Labor', budget: 150000, spent: 45000 },
        { name: 'Permits & Fees', budget: 25000, spent: 12000 },
        { name: 'Contingency', budget: 75000, spent: 3000 }
    ]
};
