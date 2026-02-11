import { Contact } from '../types/models';

export const mockContactsService = {
    async list(search?: string): Promise<Contact[]> {
        await new Promise(resolve => setTimeout(resolve, 400)); // Simulate latency

        const contacts: Contact[] = [
            {
                id: 'c1',
                name: 'Mike Johnson',
                company: 'MJ Electrical',
                role: 'Subcontractor' as any,
                phone: '555-0101',
                email: 'mike@mjelectric.com',
                contact_preference: 'SMS',
                portal_enabled: true,
                org_id: 'org_1',
                source: 'manual',
                created_at: new Date().toISOString(),
                updated_at: new Date().toISOString()
            },
            {
                id: 'c2',
                name: 'Sarah Connor',
                company: 'Skyline Concrete',
                role: 'Subcontractor' as any,
                phone: '555-0102',
                email: 'sarah@skyline.com',
                contact_preference: 'Email',
                portal_enabled: true,
                org_id: 'org_1',
                source: 'manual',
                created_at: new Date().toISOString(),
                updated_at: new Date().toISOString()
            },
            {
                id: 'c3',
                name: 'David Builder',
                company: 'Urban Living',
                role: 'Client' as any,
                phone: '555-0103',
                email: 'david@urbanliving.com',
                contact_preference: 'Email',
                portal_enabled: true,
                org_id: 'org_1',
                source: 'manual',
                created_at: new Date().toISOString(),
                updated_at: new Date().toISOString()
            },
            {
                id: 'c4',
                name: 'Jessica Lee',
                company: 'City Plumbing',
                role: 'Subcontractor' as any,
                phone: '555-0104',
                email: 'jessica@cityplumbing.com',
                contact_preference: 'SMS',
                portal_enabled: false,
                org_id: 'org_1',
                source: 'manual',
                created_at: new Date().toISOString(),
                updated_at: new Date().toISOString()
            },
            {
                id: 'c5',
                name: 'Tom Architect',
                company: 'Design Studio A',
                role: 'Architect' as any,
                phone: '555-0105',
                email: 'tom@designstudioa.com',
                contact_preference: 'Email',
                portal_enabled: true,
                org_id: 'org_1',
                source: 'manual',
                created_at: new Date().toISOString(),
                updated_at: new Date().toISOString()
            }
        ];

        if (search) {
            const lowerSearch = search.toLowerCase();
            return contacts.filter(c =>
                c.name.toLowerCase().includes(lowerSearch) ||
                (c.company && c.company.toLowerCase().includes(lowerSearch)) ||
                (c.email && c.email.toLowerCase().includes(lowerSearch))
            );
        }

        return contacts;
    }
};
