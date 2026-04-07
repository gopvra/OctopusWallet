import { useState } from 'react';
import { api } from '../../api/client';
import { StatusBadge } from '../../components/StatusBadge';

export function RefundsPage() {
  const [paymentId, setPaymentId] = useState('');
  const [refunds, setRefunds] = useState<any[]>([]);
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);

  const [toAddress, setToAddress] = useState('');
  const [amount, setAmount] = useState('');
  const [reason, setReason] = useState('');
  const [msg, setMsg] = useState('');

  const searchRefunds = async () => {
    if (!paymentId) return;
    setLoading(true);
    setError('');
    try {
      const data = await api.listRefunds(paymentId);
      setRefunds(data.refunds || []);
    } catch (e: any) {
      setError(e.message);
    } finally {
      setLoading(false);
    }
  };

  const createRefund = async () => {
    if (!paymentId || !toAddress || !amount) return;
    setMsg('');
    try {
      await api.createRefund({ payment_id: paymentId, to_address: toAddress, amount, reason });
      setMsg('Refund created');
      searchRefunds();
    } catch (e: any) {
      setMsg('Error: ' + e.message);
    }
  };

  return (
    <div className="space-y-6 max-w-3xl">
      <h1 className="text-2xl font-bold text-gray-900">Refunds</h1>

      {/* Create Refund */}
      <div className="bg-white rounded-lg shadow p-6 space-y-3">
        <h2 className="font-semibold text-gray-800">Create Refund</h2>
        <input placeholder="Payment ID" value={paymentId}
          onChange={e => setPaymentId(e.target.value)}
          className="w-full border rounded px-3 py-2 text-sm" />
        <input placeholder="Refund to address" value={toAddress}
          onChange={e => setToAddress(e.target.value)}
          className="w-full border rounded px-3 py-2 text-sm" />
        <div className="flex gap-3">
          <input placeholder="Amount" value={amount}
            onChange={e => setAmount(e.target.value)}
            className="flex-1 border rounded px-3 py-2 text-sm" />
          <input placeholder="Reason (optional)" value={reason}
            onChange={e => setReason(e.target.value)}
            className="flex-1 border rounded px-3 py-2 text-sm" />
        </div>
        <div className="flex gap-3">
          <button onClick={createRefund}
            className="px-4 py-2 bg-blue-600 text-white rounded text-sm hover:bg-blue-700">Create Refund</button>
          <button onClick={searchRefunds}
            className="px-4 py-2 bg-gray-200 rounded text-sm hover:bg-gray-300">Search Refunds</button>
        </div>
        {msg && <p className="text-sm text-blue-600">{msg}</p>}
        {error && <p className="text-sm text-red-500">{error}</p>}
      </div>

      {/* Refund List */}
      {loading && <p className="text-gray-500">Loading...</p>}
      {refunds.length > 0 && (
        <div className="bg-white rounded-lg shadow overflow-hidden">
          <table className="w-full text-sm">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-4 py-3 text-left text-gray-500">ID</th>
                <th className="px-4 py-3 text-left text-gray-500">To</th>
                <th className="px-4 py-3 text-left text-gray-500">Amount</th>
                <th className="px-4 py-3 text-left text-gray-500">Status</th>
              </tr>
            </thead>
            <tbody className="divide-y">
              {refunds.map((r: any) => (
                <tr key={r.id}>
                  <td className="px-4 py-3 font-mono text-xs">{r.id.slice(0, 8)}...</td>
                  <td className="px-4 py-3 font-mono text-xs">{r.to_address?.slice(0, 12)}...</td>
                  <td className="px-4 py-3">{r.amount}</td>
                  <td className="px-4 py-3"><StatusBadge status={r.status} /></td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
}
