# Task Tracker JurnalUmi

## Phase 1: Foundation & DB (W1)
- [x] Setup boilerplate (Go 1.22, Echo, Templ, Tailwind, Alpine.js, Air).
- [x] Setup Docker/Podman `docker-compose.yml` untuk PostgreSQL 16.
- [x] Implement DB connection & migration scripts (Schema dari PRD).

## Phase 2: Auth & Multi-Tenancy (W2)
- [x] Auth sistem (Login, Register, JWT/Session).
- [x] Middleware Multi-Tenant (Inject `tenant_id` ke context).
- [x] Middleware RBAC (Role checking).
- [x] CRUD Family Members.

## Phase 3: Core Ledger (W3)
- [x] CRUD Wallets (Cash, Bank, E-Wallet).
- [x] CRUD Categories (Income/Expense hierarchy).
- [x] CRUD Transactions (Income, Expense, Transfer).
- [ ] Budget capping check (Alpine.js UI warning).

## Phase 4: Asset Engine (W4)
- [x] CRUD Liquid Assets & Investment.
- [x] CRUD Commodity Assets (Gold, Dinar, Perak).
- [ ] API Provider integration (Harga Emas live).
- [ ] Net Worth Dashboard (Assets - Debts).

## Phase 5: Debt & Protection (W5)
- [ ] CRUD Debts & Receivables.
- [ ] Debt Snowball/Avalanche calculator.
- [ ] Sinking Funds & Emergency Fund tracker.

## Phase 6: PWA & Notifications (W6)
- [ ] Setup `manifest.json` & `sw.js` (Offline support).
- [ ] SMTP Client setup.
- [ ] Cron/Background worker untuk Debt Reminder & Budget Alert.

## Phase 7: SaaS & Export (W7-W8)
- [x] Billing Module (Mayar.id webhook / Voucher key).
- [x] PDF Export engine untuk Monthly Report.
- [x] Production Build (Dockerfile, Caddy Reverse Proxy).
