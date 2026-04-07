import { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { api } from '../../api/client';
import { StatusBadge } from '../../components/StatusBadge';

export function DashboardPage() {
  const [balances, setBalances] = useState<any[]>([]);
  const [payments, setPayments] = useState<any[]>([]);
  const [payouts, setPayouts] = useState<any[]>([]);

  useEffect(() => {
    api.getBalances().then(d => setBalances(d.balances || []));
    api.listPayments(10, 0).then(d => setPayments(d.payments || []));
    api.listPayouts(10, 0).then(d => setPayouts(d.payouts || []));
  }, []);

  return (
    <div className="space-y-8">
      <h1 className="text-2xl font-bold text-gray-900">Dashboard</h1>

      {/* Balance Cards */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        {balances.length === 0 && (
          <div className="col-span-full text-gray-400 text-sm">No balances yet</div>
        )}
        {balances.map((b: any) => (
          <div key={b.id} className="bg-white rounded-lg shadow p-4">
            <p className="text-sm text-gray-500">{b.chain} {b.token || 'Native'}</p>
            <p className="text-xl font-bold text-gray-900 mt-1">{b.available}</p>
            <p className="text-xs text-gray-400">Pending: {b.pending}</p>
          </div>
        ))}
      </div>

      {/* Recent Payments */}
      <div>
        <div className="flex justify-between items-center mb-3">
          <h2 className="text-lg font-semibold text-gray-800">Recent Payments</h2>
          <Link to="/dashboard/payments" className="text-sm text-blue-600 hover:underline">View all</Link>
        </div>
        <div className="bg-white rounded-lg shadow overflow-hidden">
          <table className="w-full text-sm">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-4 py-3 text-left text-gray-500">ID</th>
                <th className="px-4 py-3 text-left text-gray-500">Chain</th>
                <th className="px-4 py-3 text-left text-gray-500">Amount</th>
                <th className="px-4 py-3 text-left text-gray-500">Status</th>
                <th className="px-4 py-3 text-left text-gray-500">Date</th>
              </tr>
            </thead>
            <tbody className="divide-y">
              {payments.map((p: any) => (
                <tr key={p.id} className="hover:bg-gray-50">
                  <td className="px-4 py-3 font-mono text-xs">{p.id.slice(0, 8)}...</td>
                  <td className="px-4 py-3">{p.chain}</td>
                  <td className="px-4 py-3 font-medium">{p.amount_expected}</td>
                  <td className="px-4 py-3"><StatusBadge status={p.status} /></td>
                  <td className="px-4 py-3 text-gray-500">{new Date(p.created_at).toLocaleDateString()}</td>
                </tr>
              ))}
              {payments.length === 0 && (
                <tr><td colSpan={5} className="px-4 py-6 text-center text-gray-400">No payments yet</td></tr>
              )}
            </tbody>
          </table>
        </div>
      </div>

      {/* Recent Payouts */}
      <div>
        <div className="flex justify-between items-center mb-3">
          <h2 className="text-lg font-semibold text-gray-800">Recent Payouts</h2>
          <Link to="/dashboard/payouts" className="text-sm text-blue-600 hover:underline">View all</Link>
        </div>
        <div className="bg-white rounded-lg shadow overflow-hidden">
          <table className="w-full text-sm">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-4 py-3 text-left text-gray-500">ID</th>
                <th className="px-4 py-3 text-left text-gray-500">Chain</th>
                <th className="px-4 py-3 text-left text-gray-500">To</th>
                <th className="px-4 py-3 text-left text-gray-500">Amount</th>
                <th className="px-4 py-3 text-left text-gray-500">Status</th>
              </tr>
            </thead>
            <tbody className="divide-y">
              {payouts.map((p: any) => (
                <tr key={p.id} className="hover:bg-gray-50">
                  <td className="px-4 py-3 font-mono text-xs">{p.id.slice(0, 8)}...</td>
                  <td className="px-4 py-3">{p.chain}</td>
                  <td className="px-4 py-3 font-mono text-xs">{p.to_address?.slice(0, 10)}...</td>
                  <td className="px-4 py-3 font-medium">{p.amount}</td>
                  <td className="px-4 py-3"><StatusBadge status={p.approval_status || p.status} /></td>
                </tr>
              ))}
              {payouts.length === 0 && (
                <tr><td colSpan={5} className="px-4 py-6 text-center text-gray-400">No payouts yet</td></tr>
              )}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
}
