import { useEffect, useState } from 'react';
import { useParams } from 'react-router-dom';
import { QRCode } from '../../components/QRCode';
import { CountdownTimer } from '../../components/CountdownTimer';
import { CopyButton } from '../../components/CopyButton';
import { usePaymentStatus } from '../../hooks/usePaymentStatus';

interface Payment {
  id: string;
  chain: string;
  token: string;
  amount_expected: string;
  amount_received: string;
  address: string;
  status: string;
  tx_hash?: string;
  confirmations: number;
  currency: string;
  description: string;
  order_id: string;
  redirect_url: string;
  expires_at: string;
}

const chainPrefixes: Record<string, string> = {
  ethereum: 'ethereum:',
  bsc: '',
  polygon: '',
  solana: 'solana:',
  tron: '',
  bitcoin: 'bitcoin:',
};

const chainNames: Record<string, string> = {
  ethereum: 'Ethereum',
  bsc: 'BSC',
  polygon: 'Polygon',
  solana: 'Solana',
  tron: 'TRON',
  bitcoin: 'Bitcoin',
};

const explorerUrls: Record<string, string> = {
  ethereum: 'https://etherscan.io/tx/',
  bsc: 'https://bscscan.com/tx/',
  polygon: 'https://polygonscan.com/tx/',
  solana: 'https://solscan.io/tx/',
  tron: 'https://tronscan.org/#/transaction/',
  bitcoin: 'https://mempool.space/tx/',
};

export function CheckoutPage() {
  const { id } = useParams<{ id: string }>();
  const [payment, setPayment] = useState<Payment | null>(null);
  const [error, setError] = useState('');
  const { status: wsStatus } = usePaymentStatus(id || '');

  useEffect(() => {
    if (!id) return;
    fetch(`/api/v1/payments/${id}`)
      .then(r => r.ok ? r.json() : Promise.reject('not found'))
      .then(setPayment)
      .catch(() => setError('Payment not found'));
  }, [id]);

  // Update from WebSocket
  useEffect(() => {
    if (wsStatus && payment) {
      setPayment(p => p ? {
        ...p,
        status: wsStatus.status,
        tx_hash: wsStatus.tx_hash || p.tx_hash,
        confirmations: wsStatus.confirmations,
        amount_received: wsStatus.amount_received || p.amount_received,
      } : p);
    }
  }, [wsStatus]);

  // Redirect on completion
  useEffect(() => {
    if (payment?.status === 'completed' && payment.redirect_url) {
      const timer = setTimeout(() => {
        window.location.href = payment.redirect_url;
      }, 3000);
      return () => clearTimeout(timer);
    }
  }, [payment?.status, payment?.redirect_url]);

  if (error) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="bg-white p-8 rounded-xl shadow-lg text-center">
          <p className="text-red-500 text-lg">{error}</p>
        </div>
      </div>
    );
  }

  if (!payment) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="text-gray-500">Loading...</div>
      </div>
    );
  }

  const prefix = chainPrefixes[payment.chain] || '';
  const qrValue = `${prefix}${payment.address}?amount=${payment.amount_expected}`;
  const truncAddr = `${payment.address.slice(0, 10)}...${payment.address.slice(-8)}`;

  return (
    <div className="min-h-screen bg-gray-50 flex items-center justify-center p-4">
      <div className="bg-white rounded-2xl shadow-xl max-w-md w-full overflow-hidden">
        {/* Header */}
        <div className="bg-gray-900 text-white px-6 py-4 flex items-center justify-between">
          <h1 className="font-bold text-lg">OctopusWallet</h1>
          {payment.status === 'pending' && payment.expires_at && (
            <CountdownTimer expiresAt={payment.expires_at} />
          )}
        </div>

        {/* Body */}
        <div className="p-6 space-y-6">
          {/* Order info */}
          {(payment.order_id || payment.description) && (
            <div className="text-center">
              {payment.order_id && (
                <p className="text-sm text-gray-500">Order: {payment.order_id}</p>
              )}
              {payment.description && (
                <p className="text-sm text-gray-600">{payment.description}</p>
              )}
            </div>
          )}

          {/* Amount */}
          <div className="text-center">
            <p className="text-3xl font-bold text-gray-900">{payment.amount_expected}</p>
            <p className="text-sm text-gray-500 mt-1">
              {chainNames[payment.chain] || payment.chain}
              {payment.token && ` (${payment.token})`}
            </p>
          </div>

          {/* Status: Completed */}
          {payment.status === 'completed' && (
            <div className="text-center py-6">
              <div className="w-16 h-16 bg-green-100 rounded-full flex items-center justify-center mx-auto mb-3">
                <svg className="w-8 h-8 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                </svg>
              </div>
              <p className="text-green-600 font-semibold text-lg">Payment Complete</p>
              {payment.tx_hash && (
                <a
                  href={`${explorerUrls[payment.chain] || '#'}${payment.tx_hash}`}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="text-sm text-blue-500 hover:underline mt-2 block"
                >
                  View transaction
                </a>
              )}
              {payment.redirect_url && (
                <p className="text-sm text-gray-400 mt-2">Redirecting in 3 seconds...</p>
              )}
            </div>
          )}

          {/* Status: Expired */}
          {payment.status === 'expired' && (
            <div className="text-center py-6">
              <p className="text-red-500 font-semibold text-lg">Payment Expired</p>
              <p className="text-sm text-gray-500 mt-1">Please create a new payment.</p>
            </div>
          )}

          {/* Status: Confirming */}
          {payment.status === 'confirming' && (
            <div className="text-center py-4">
              <div className="animate-spin w-8 h-8 border-4 border-blue-500 border-t-transparent rounded-full mx-auto mb-3" />
              <p className="text-blue-600 font-semibold">Confirming...</p>
              <p className="text-sm text-gray-500">{payment.confirmations} confirmations</p>
              {payment.tx_hash && (
                <a
                  href={`${explorerUrls[payment.chain] || '#'}${payment.tx_hash}`}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="text-sm text-blue-500 hover:underline mt-2 block"
                >
                  View transaction
                </a>
              )}
            </div>
          )}

          {/* Status: Pending — show QR and address */}
          {payment.status === 'pending' && (
            <>
              <div className="flex justify-center">
                <QRCode value={qrValue} size={180} />
              </div>

              <div className="space-y-2">
                <p className="text-xs text-gray-500 text-center">Send to this address:</p>
                <div className="bg-gray-50 rounded-lg p-3 flex items-center justify-between gap-2">
                  <code className="text-sm text-gray-800 break-all flex-1">{payment.address}</code>
                  <CopyButton text={payment.address} />
                </div>
              </div>

              {/* Status steps */}
              <div className="space-y-2">
                <Step active label="Waiting for payment..." />
                <Step label="Confirming on blockchain" />
                <Step label="Complete" />
              </div>
            </>
          )}
        </div>

        {/* Footer */}
        <div className="border-t px-6 py-3 text-center">
          <p className="text-xs text-gray-400">Powered by OctopusWallet</p>
        </div>
      </div>
    </div>
  );
}

function Step({ label, active = false }: { label: string; active?: boolean }) {
  return (
    <div className="flex items-center gap-3">
      <div className={`w-3 h-3 rounded-full ${active ? 'bg-blue-500 animate-pulse' : 'bg-gray-300'}`} />
      <span className={`text-sm ${active ? 'text-gray-900 font-medium' : 'text-gray-400'}`}>
        {label}
      </span>
    </div>
  );
}
