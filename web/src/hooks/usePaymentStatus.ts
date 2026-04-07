import { useEffect, useState, useRef } from 'react';

interface PaymentUpdate {
  payment_id: string;
  status: string;
  tx_hash?: string;
  confirmations: number;
  amount_received?: string;
}

export function usePaymentStatus(paymentId: string) {
  const [status, setStatus] = useState<PaymentUpdate | null>(null);
  const wsRef = useRef<WebSocket | null>(null);

  useEffect(() => {
    if (!paymentId) return;

    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const ws = new WebSocket(`${protocol}//${window.location.host}/ws/payments/${paymentId}`);
    wsRef.current = ws;

    ws.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data) as PaymentUpdate;
        setStatus(data);
      } catch {}
    };

    ws.onerror = () => ws.close();

    return () => {
      ws.close();
      wsRef.current = null;
    };
  }, [paymentId]);

  return status;
}
