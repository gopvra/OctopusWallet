import { useEffect, useState } from 'react';
import { api } from '../../api/client';
import { StatusBadge } from '../../components/StatusBadge';

export function PaymentsPage() {
  const [payments, setPayments] = useState<any[]>([]);
  const [offset, setOffset] = useState(0);
  const limit = 20;

  useEffect(() => {
    api.listPayments(limit, offset).then(d => setPayments(d.payments || []));
  }, [offset]);

  return (
    <div className="space-y-4">
      <h1 className="text-2xl font-bold text-gray-900">Payments</h1>
      <div className="bg-white rounded-lg shadow overflow-x-auto">
        <table className="w-full text-sm">
          <thead className="bg-gray-50">
            <tr>
              <th className="px-4 py-3 text-left text-gray-500">ID</th>
              <th className="px-4 py-3 text-left text-gray-500">Chain</th>
              <th className="px-4 py-3 text-left text-gray-500">Address</th>
              <th className="px-4 py-3 text-left text-gray-500">Expected</th>
              <th className="px-4 py-3 text-left text-gray-500">Received</th>
              <th className="px-4 py-3 text-left text-gray-500">Status</th>
              <th className="px-4 py-3 text-left text-gray-500">Date</th>
            </tr>
          </thead>
          <tbody className="divide-y">
            {payments.map((p: any) => (
              <tr key={p.id} className="hover:bg-gray-50">
                <td className="px-4 py-3 font-mono text-xs">{p.id.slice(0, 8)}...</td>
                <td className="px-4 py-3">{p.chain}</td>
                <td className="px-4 py-3 font-mono text-xs">{p.address?.slice(0, 12)}...</td>
                <td className="px-4 py-3">{p.amount_expected}</td>
                <td className="px-4 py-3">{p.amount_received}</td>
                <td className="px-4 py-3"><StatusBadge status={p.status} /></td>
                <td className="px-4 py-3 text-gray-500">{new Date(p.created_at).toLocaleString()}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
      <div className="flex justify-between">
        <button
          disabled={offset === 0}
          onClick={() => setOffset(Math.max(0, offset - limit))}
          className="px-4 py-2 bg-gray-200 rounded disabled:opacity-50"
        >Previous</button>
        <button
          disabled={payments.length < limit}
          onClick={() => setOffset(offset + limit)}
          className="px-4 py-2 bg-gray-200 rounded disabled:opacity-50"
        >Next</button>
      </div>
    </div>
  );
}
