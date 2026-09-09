with open("web/views/debts.templ", "r") as f:
    content = f.read()

content = content.replace(
    'templ DebtManagement(tenant models.Tenant, user models.User, debts []models.Debt, totalDebt float64, totalReceivable float64) {',
    'templ DebtManagement(tenant models.Tenant, user models.User, debts []models.Debt, wallets []models.Wallet, totalDebt float64, totalReceivable float64) {'
)

content = content.replace(
    '@Layout("Utang, Piutang & Calculator Snowball", tenant, user, debtContent(debts, totalDebt, totalReceivable))',
    '@Layout("Utang, Piutang & Calculator Snowball", tenant, user, debtContent(debts, wallets, totalDebt, totalReceivable))'
)

content = content.replace(
    'templ debtContent(debts []models.Debt, totalDebt float64, totalReceivable float64) {',
    'templ debtContent(debts []models.Debt, wallets []models.Wallet, totalDebt float64, totalReceivable float64) {'
)

old_form = '''							<td class="py-4 text-right">
								<input type="hidden" name="csrf_token" value={ ctx.Value("csrf").(string) } />
<form action="/debts/pay" method="POST" class="inline">
				<input type="hidden" name="csrf_token" value={ ctx.Value("csrf").(string) } />
t			<input type="hidden" name="csrf_token" value={ ctx.Value("csrf").(string) } />
									<input type="hidden" name="debt_id" value={ d.ID.String() }/>
									<button type="submit" class="bg-emerald-600 hover:bg-emerald-700 text-white text-xs font-bold px-3 py-1.5 rounded-lg transition">Cicil / Lunas</button>
								</form>
							</td>'''

new_form = '''							<td class="py-4 text-right" x-data="{ payModal: false }">
								<button @click="payModal = true" class="bg-emerald-600 hover:bg-emerald-700 text-white text-xs font-bold px-3 py-1.5 rounded-lg transition">Cicil / Lunas</button>
								
								<div x-cloak x-show="payModal" class="fixed inset-0 z-50 flex items-center justify-center p-4 text-left" x-transition.opacity>
									<div class="fixed inset-0 bg-gray-900 bg-opacity-50 backdrop-blur-sm" @click="payModal = false"></div>
									<div class="bg-white rounded-2xl shadow-2xl max-w-md w-full p-6 relative z-10" x-transition.scale>
										<div class="flex justify-between items-center mb-6">
											<h3 class="text-lg font-bold text-gray-900">Bayar Utang/Piutang</h3>
											<button @click="payModal = false" class="text-gray-400 hover:text-gray-600">✖</button>
										</div>
										<form action="/debts/pay" method="POST" class="space-y-4">
											<input type="hidden" name="csrf_token" value={ ctx.Value("csrf").(string) } />
											<input type="hidden" name="debt_id" value={ d.ID.String() }/>
											<div>
												<label class="block text-sm font-bold text-gray-700 mb-1">Nominal Pembayaran (Rp)</label>
												<input type="number" name="amount" step="1000" class="w-full border-2 border-gray-200 rounded-xl px-4 py-2 outline-none" required>
											</div>
											<div>
												<label class="block text-sm font-bold text-gray-700 mb-1">Dari/Ke Dompet</label>
												<select name="wallet_id" class="w-full border-2 border-gray-200 rounded-xl px-4 py-2 outline-none" required>
													for _, w := range wallets {
														<option value={ w.ID.String() }>{ w.Name } - { FormatRupiah(w.Balance) }</option>
													}
												</select>
											</div>
											<button type="submit" class="w-full bg-emerald-600 hover:bg-emerald-700 text-white font-bold py-3 rounded-xl transition mt-4">Simpan Pembayaran</button>
										</form>
									</div>
								</div>
							</td>'''

content = content.replace(old_form, new_form)

with open("web/views/debts.templ", "w") as f:
    f.write(content)
