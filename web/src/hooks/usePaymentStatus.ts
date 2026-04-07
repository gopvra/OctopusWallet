import { useEffect, useState, useRef, useCallback } from 'react';

interface PaymentUpdate {
  payment_id: string;
  status: string;
  tx_hash?: string;
  confirmations: number;
  amount_received?: string;
}

export function usePaymentStatus(paymentId: string) {
  const [status, setStatus] = useState<PaymentUpdate | null>(null);
  const [connected, setConnected] = useState(false);
  const wsRef = useRef<WebSocket | null>(null);
  const retriesRef = useRef(0);

  const connect = useCallback(() => {
    if (!paymentId) return;

    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const ws = new WebSocket(`${protocol}//${window.location.host}/ws/payments/${paymentId}`);
    wsRef.current = ws;

    ws.onopen = () => {
      setConnected(true);
      retriesRef.current = 0;
    };

    ws.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data) as PaymentUpdate;
        setStatus(data);
      } catch { /* ignore malformed messages */ }
    };

    ws.onclose = () => {
      setConnected(false);
      // Reconnect with exponential backoff (max 30s)
      const delay = Math.min(1000 * Math.pow(2, retriesRef.current), 30000);
      retriesRef.current++;
      setTimeout(() => {
        if (wsRef.current === ws) connect();
      }, delay);
    };

    ws.onerror = () => ws.close();
  }, [paymentId]);

  useEffect(() => {
    connect();
    return () => {
      if (wsRef.current) {
        const ws = wsRef.current;
        wsRef.current = null;
        ws.close();
      }
    };
  }, [connect]);

  return { status, connected };
}
