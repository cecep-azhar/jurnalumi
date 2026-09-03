# Role & Menu Permission JurnalUmi

## Roles

1.  **Super Admin (Platform Owner)**
2.  **Family Owner (Kepala Keluarga / Husband / Primary Admin)**
3.  **Family Co-Owner (Pasangan / Spouse / Wife)**
4.  **Family Member (Anak / Dependents)**
5.  **Auditor / Financial Planner (Viewer)**

## Menu & Permissions Mapping

| Menu / Module | Super Admin | Family Owner | Family Co-Owner | Family Member | Auditor |
| :--- | :---: | :---: | :---: | :---: | :---: |
| **Platform Analytics & Settings** |
| SaaS Billing & Tenant Management | V (Full) | X | X | X | X |
| Global Variables (Gold/API API) | V (Full) | X | X | X | X |
| **Family Settings** |
| Manage Family Members | X | V (Full) | X | X | X |
| Setup Budget (Fixed, Sinking, Emergency) | X | V (Full) | X | X | X |
| **Daily Ledger (Income & Expense)** |
| Input/Edit Income | X | V (Full) | V (Full) | X | V (Read) |
| Input/Edit Expense | X | V (Full) | V (Full) | V (Limited*) | V (Read) |
| View Shared Wallets | X | V (Full) | V (Full) | X | V (Read) |
| **Net Worth & Assets** |
| View Net Worth Dashboard | X | V (Full) | V (Full) | X | V (Read) |
| Input/Edit Liquid Assets | X | V (Full) | V (Full) | X | V (Read) |
| Input/Edit Commodity (Gold/Silver) | X | V (Full) | V (Full) | X | V (Read) |
| Input/Edit Investments/Property | X | V (Full) | V (Full) | X | V (Read) |
| **Debt & Receivables** |
| View Debt/Receivables | X | V (Full) | V (Full) | X | V (Read) |
| Input/Edit Debt/Receivables | X | V (Full) | X** | X | V (Read) |
| **Protection & Goals** |
| Emergency Fund Tracker | X | V (Full) | V (Full) | X | V (Read) |
| Sinking Funds | X | V (Full) | V (Full) | X | V (Read) |

*\* Limited: Hanya bisa input/edit expense pada pos yang ditugaskan (uang saku pribadi).*
*\*\* Berdasarkan PRD, Family Owner memiliki full access utang/piutang. Co-Owner bisa jadi butuh akses, tapi default PRD menyebutkan Owner yang "Pengaturan Utang/Piutang & Aset Utama". Asumsi Co-Owner read-only atau input-only jika tidak dispesifikasikan eksplisit.*
