const BASE = '/api/v1';

async function request<T>(method: string, path: string, body?: unknown): Promise<T> {
  const headers: Record<string, string> = { 'Content-Type': 'application/json' };
  const apiKey = localStorage.getItem('api_key');
  if (apiKey) headers['X-API-Key'] = apiKey;

  const controller = new AbortController();
  const timeoutId = setTimeout(() => controller.abort(), 15000);

  try {
    const res = await fetch(`${BASE}${path}`, {
      method,
      headers,
      body: body ? JSON.stringify(body) : undefined,
      signal: controller.signal,
    });
    if (!res.ok) {
      const err = await res.json().catch(() => ({ error: 'request failed' }));
      throw new Error(err.error || 'request failed');
    }
    return res.json();
  } finally {
    clearTimeout(timeoutId);
  }
}

export const api = {
  // Payments
  createPayment: (data: any) => request('POST', '/payments/create', data),
  getPayment: (id: string) => request<any>('GET', `/payments/${id}`),
  listPayments: (limit = 20, offset = 0) =>
    request<any>('GET', `/payments?limit=${limit}&offset=${offset}`),

  // Payouts
  createPayout: (data: any) => request('POST', '/payouts/create', data),
  getPayout: (id: string) => request<any>('GET', `/payouts/${id}`),
  listPayouts: (limit = 20, offset = 0) =>
    request<any>('GET', `/payouts?limit=${limit}&offset=${offset}`),
  approvePayout: (id: string, data: any) =>
    request('POST', `/payouts/${id}/approve`, data),
  rejectPayout: (id: string, data: any) =>
    request('POST', `/payouts/${id}/reject`, data),

  // Refunds
  createRefund: (data: any) => request('POST', '/refunds/create', data),

  // Batch
  createBatchPayout: (data: any) => request('POST', '/payouts/batch', data),
  listBatchPayouts: () => request<any>('GET', '/payouts/batches'),

  // Balances
  getBalances: () => request<any>('GET', '/balances'),

  // Config
  getApprovalConfig: () => request<any>('GET', '/approval/config'),
  setApprovalConfig: (data: any) => request('POST', '/approval/config', data),

  // Merchant
  getProfile: () => request<any>('GET', '/merchants/profile'),
  register: (data: any) => request<any>('POST', '/merchants/register', data),

  // Currencies (public)
  getCurrencies: () => fetch(`${BASE}/currencies`).then(r => r.json()),

  // Sweep
  getSweepTasks: () => request<any>('GET', '/sweep/tasks'),
  setCollectionAddress: (data: any) =>
    request('POST', '/sweep/collection-address', data),
  getCollectionAddresses: () => request<any>('GET', '/sweep/collection-address'),

  // Settings
  getIPWhitelist: () => request<any>('GET', '/security/ip-whitelist'),
  setIPWhitelist: (ips: string[]) =>
    request('POST', '/security/ip-whitelist', { ips }),
};
