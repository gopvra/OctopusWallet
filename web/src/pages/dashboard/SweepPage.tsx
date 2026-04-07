import { useEffect, useState } from 'react';
import { api } from '../../api/client';
import { StatusBadge } from '../../components/StatusBadge';

export function SweepPage() {
  const [tasks, setTasks] = useState<any[]>([]);
  const [addresses, setAddresses] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);
  const [msg, setMsg] = useState('');

  // Config form
  const [chain, setChain] = useState('ethereum');
  const [address, setAddress] = useState('');
  const [threshold, setThreshold] = useState('0');

  const load = async () => {
    setLoading(true);
    try {
      const [t, a] = await Promise.all([
        api.getSweepTasks().catch(() => ({ sweep_tasks: [] })),
        api.getCollectionAddresses().catch(() => ({ collection_addresses: [] })),
      ]);
      setTasks(t.sweep_tasks || []);
      setAddresses(a.collection_addresses || []);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => { load(); }, []);

  const saveCollection = async () => {
    try {
      await api.setCollectionAddress({ chain, address, sweep_threshold: threshold });
      setMsg('Collection address saved');
      load();
    } catch (e: any) { setMsg('Error: ' + e.message); }
  };

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold text-gray-900">Auto-Sweep</h1>
      {msg && <div className="bg-blue-50 text-blue-700 px-4 py-2 rounded text-sm">{msg}</div>}

      {/* Config */}
      <div className="bg-white rounded-lg shadow p-6 space-y-3 max-w-xl">
        <h2 className="font-semibold text-gray-800">Collection Address</h2>
        <select value={chain} onChange={e => setChain(e.target.value)}
          className="w-full border rounded px-3 py-2 text-sm">
          <option value="ethereum">Ethereum</option>
          <option value="bsc">BSC</option>
          <option value="polygon">Polygon</option>
          <option value="solana">Solana</option>
          <option value="tron">TRON</option>
          <option value="bitcoin">Bitcoin</option>
        </select>
        <input placeholder="Collection address" value={address}
          onChange={e => setAddress(e.target.value)}
          className="w-full border rounded px-3 py-2 text-sm" />
        <input placeholder="Sweep threshold (min amount)" value={threshold}
          onChange={e => setThreshold(e.target.value)}
          className="w-full border rounded px-3 py-2 text-sm" />
        <button onClick={saveCollection}
          className="px-4 py-2 bg-blue-600 text-white rounded text-sm hover:bg-blue-700">Save</button>
      </div>

      {/* Current Addresses */}
      {addresses.length > 0 && (
        <div className="bg-white rounded-lg shadow p-6">
          <h2 className="font-semibold text-gray-800 mb-3">Configured Addresses</h2>
          <div className="space-y-2 text-sm">
            {addresses.map((a: any) => (
              <div key={a.id} className="flex justify-between border-b pb-2">
                <span className="font-medium">{a.chain}</span>
                <span className="font-mono text-xs">{a.address}</span>
                <span className="text-gray-500">threshold: {a.sweep_threshold}</span>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Tasks */}
      <div className="bg-white rounded-lg shadow overflow-hidden">
        <h2 className="font-semibold text-gray-800 p-4">Sweep Tasks</h2>
        {loading ? <p className="p-4 text-gray-500">Loading...</p> : (
          <table className="w-full text-sm">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-4 py-3 text-left text-gray-500">ID</th>
                <th className="px-4 py-3 text-left text-gray-500">Chain</th>
                <th className="px-4 py-3 text-left text-gray-500">From</th>
                <th className="px-4 py-3 text-left text-gray-500">Amount</th>
                <th className="px-4 py-3 text-left text-gray-500">Status</th>
              </tr>
            </thead>
            <tbody className="divide-y">
              {tasks.map((t: any) => (
                <tr key={t.id}>
                  <td className="px-4 py-3 font-mono text-xs">{t.id.slice(0, 8)}...</td>
                  <td className="px-4 py-3">{t.chain}</td>
                  <td className="px-4 py-3 font-mono text-xs">{t.from_address?.slice(0, 12)}...</td>
                  <td className="px-4 py-3">{t.amount}</td>
                  <td className="px-4 py-3"><StatusBadge status={t.status} /></td>
                </tr>
              ))}
              {tasks.length === 0 && (
                <tr><td colSpan={5} className="px-4 py-6 text-center text-gray-400">No sweep tasks yet</td></tr>
              )}
            </tbody>
          </table>
        )}
      </div>
    </div>
  );
}
