import { useEffect, useState } from 'react';
import { api } from '../../api/client';

export function SettingsPage() {
  const [profile, setProfile] = useState<any>(null);
  const [approvalConfig, setApprovalConfig] = useState<any>(null);
  const [ipList, setIpList] = useState('');
  const [msg, setMsg] = useState('');

  useEffect(() => {
    api.getProfile().then(setProfile);
    api.getApprovalConfig().then(d => setApprovalConfig(d.config));
    api.getIPWhitelist().then(d => setIpList((d.ips || []).join('\n')));
  }, []);

  const [threshold, setThreshold] = useState('');
  const [singleLimit, setSingleLimit] = useState('');
  const [dailyLimit, setDailyLimit] = useState('');
  const [autoRelease, setAutoRelease] = useState(false);

  useEffect(() => {
    if (approvalConfig) {
      setThreshold(approvalConfig.approval_threshold || '0');
      setSingleLimit(approvalConfig.single_tx_limit || '0');
      setDailyLimit(approvalConfig.daily_limit || '0');
      setAutoRelease(approvalConfig.auto_release || false);
    }
  }, [approvalConfig]);

  const saveApproval = async () => {
    try {
      await api.setApprovalConfig({
        approval_threshold: threshold,
        single_tx_limit: singleLimit,
        daily_limit: dailyLimit,
        auto_release: autoRelease,
        enabled: true,
      });
      setMsg('Approval config saved');
    } catch (e: any) { setMsg(e.message); }
  };

  const saveIPs = async () => {
    const ips = ipList.split('\n').map(s => s.trim()).filter(Boolean);
    try {
      await api.setIPWhitelist(ips);
      setMsg('IP whitelist updated');
    } catch (e: any) { setMsg(e.message); }
  };

  return (
    <div className="space-y-8 max-w-2xl">
      <h1 className="text-2xl font-bold text-gray-900">Settings</h1>

      {msg && (
        <div className="bg-blue-50 text-blue-700 px-4 py-2 rounded text-sm">{msg}</div>
      )}

      {/* Profile */}
      {profile && (
        <section className="bg-white rounded-lg shadow p-6 space-y-3">
          <h2 className="text-lg font-semibold">Merchant Profile</h2>
          <div className="grid grid-cols-2 gap-2 text-sm">
            <span className="text-gray-500">Name:</span><span>{profile.name}</span>
            <span className="text-gray-500">Email:</span><span>{profile.email}</span>
            <span className="text-gray-500">Webhook:</span><span className="break-all">{profile.webhook_url || 'Not set'}</span>
          </div>
        </section>
      )}

      {/* Approval Config */}
      <section className="bg-white rounded-lg shadow p-6 space-y-4">
        <h2 className="text-lg font-semibold">Withdrawal Approval Rules</h2>
        <div className="grid grid-cols-1 gap-3">
          <label className="text-sm">
            <span className="text-gray-600">Approval Threshold (amounts above this require approval)</span>
            <input value={threshold} onChange={e => setThreshold(e.target.value)}
              className="w-full mt-1 border rounded px-3 py-2 text-sm" />
          </label>
          <label className="text-sm">
            <span className="text-gray-600">Single Transaction Limit (0 = no limit)</span>
            <input value={singleLimit} onChange={e => setSingleLimit(e.target.value)}
              className="w-full mt-1 border rounded px-3 py-2 text-sm" />
          </label>
          <label className="text-sm">
            <span className="text-gray-600">Daily Limit (0 = no limit)</span>
            <input value={dailyLimit} onChange={e => setDailyLimit(e.target.value)}
              className="w-full mt-1 border rounded px-3 py-2 text-sm" />
          </label>
          <label className="flex items-center gap-2 text-sm">
            <input type="checkbox" checked={autoRelease} onChange={e => setAutoRelease(e.target.checked)} />
            Auto-release payouts below threshold
          </label>
        </div>
        <button onClick={saveApproval} className="bg-blue-600 text-white px-4 py-2 rounded text-sm hover:bg-blue-700">
          Save Approval Config
        </button>
      </section>

      {/* IP Whitelist */}
      <section className="bg-white rounded-lg shadow p-6 space-y-4">
        <h2 className="text-lg font-semibold">IP Whitelist</h2>
        <p className="text-sm text-gray-500">One IP or CIDR per line. Leave empty to allow all.</p>
        <textarea
          value={ipList}
          onChange={e => setIpList(e.target.value)}
          rows={4}
          className="w-full border rounded px-3 py-2 text-sm font-mono"
          placeholder="192.168.1.0/24&#10;10.0.0.1"
        />
        <button onClick={saveIPs} className="bg-blue-600 text-white px-4 py-2 rounded text-sm hover:bg-blue-700">
          Save IP Whitelist
        </button>
      </section>
    </div>
  );
}
