import { useEffect, useState } from 'react';
import { api } from '../../api/client';
import { StatusBadge } from '../../components/StatusBadge';

export function PayoutsPage() {
  const [payouts, setPayouts] = useState<any[]>([]);
  const [offset, setOffset] = useState(0);
  const limit = 20;

  const load = () => api.listPayouts(limit, offset).then(d => setPayouts(d.payouts || []));
  useEffect(() => { load(); }, [offset]);

  const handleApprove = async (id: string) => {
    const note = prompt('Approval note (optional):') || '';
    await api.approvePayout(id, { approver_id: 'dashboard_user', note });
    load();
  };

  const handleReject = async (id: string) => {
    const note = prompt('Rejection reason:');
    if (!note) return;
    await api.rejectPayout(id, { approver_id: 'dashboard_user', note });
    load();
  };

  return (
    <div className="space-y-4">
      <h1 className="text-2xl font-bold text-gray-900">Payouts</h1>
      <div className="bg-white rounded-lg shadow overflow-x-auto">
        <table className="w-full text-sm">
          <thead className="bg-gray-50">
            <tr>
              <th className="px-4 py-3 text-left text-gray-500">ID</th>
              <th className="px-4 py-3 text-left text-gray-500">Chain</th>
              <th className="px-4 py-3 text-left text-gray-500">To</th>
              <th className="px-4 py-3 text-left text-gray-500">Amount</th>
              <th className="px-4 py-3 text-left text-gray-500">Approval</th>
              <th className="px-4 py-3 text-left text-gray-500">Status</th>
              <th className="px-4 py-3 text-left text-gray-500">Actions</th>
            </tr>
          </thead>
          <tbody className="divide-y">
            {payouts.map((p: any) => (
              <tr key={p.id} className="hover:bg-gray-50">
                <td className="px-4 py-3 font-mono text-xs">{p.id.slice(0, 8)}...</td>
                <td className="px-4 py-3">{p.chain}</td>
                <td className="px-4 py-3 font-mono text-xs">{p.to_address?.slice(0, 12)}...</td>
                <td className="px-4 py-3 font-medium">{p.amount}</td>
                <td className="px-4 py-3">
                  {p.approval_status && <StatusBadge status={p.approval_status} />}
                </td>
                <td className="px-4 py-3"><StatusBadge status={p.status} /></td>
                <td className="px-4 py-3">
                  {p.approval_status === 'pending_approval' && (
                    <div className="flex gap-2">
                      <button
                        onClick={() => handleApprove(p.id)}
                        className="px-2 py-1 bg-green-600 text-white text-xs rounded hover:bg-green-700"
                      >Approve</button>
                      <button
                        onClick={() => handleReject(p.id)}
                        className="px-2 py-1 bg-red-600 text-white text-xs rounded hover:bg-red-700"
                      >Reject</button>
                    </div>
                  )}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
      <div className="flex justify-between">
        <button disabled={offset === 0} onClick={() => setOffset(Math.max(0, offset - limit))} className="px-4 py-2 bg-gray-200 rounded disabled:opacity-50">Previous</button>
        <button disabled={payouts.length < limit} onClick={() => setOffset(offset + limit)} className="px-4 py-2 bg-gray-200 rounded disabled:opacity-50">Next</button>
      </div>
    </div>
  );
}
