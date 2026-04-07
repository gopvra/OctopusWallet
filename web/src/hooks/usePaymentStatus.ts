import { useEffect, useState, useRef, useCallback } from 'react';

interface PaymentUpdate {
  payment_id: string;
  status: string;
  tx_hash?: string;
  confirmations: number;
  amount_received?: string;
}

const MAX_RETRIES = 5;

export function usePaymentStatus(paymentId: string) {
  const [status, setStatus] = useState<PaymentUpdate | null>(null);
  const [connected, setConnected] = useState(false);
  const wsRef = useRef<WebSocket | null>(null);
  const retriesRef = useRef(0);

  const connect = useCallback(() => {
    if (!paymentId) return;

    let unmounted = false;
    let reconnectTimer: ReturnType<typeof setTimeout>;

    function connect() {
      if (unmounted) return;

      const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
      const ws = new WebSocket(`${protocol}//${window.location.host}/ws/payments/${paymentId}`);
      wsRef.current = ws;

      ws.onopen = () => {
        retriesRef.current = 0; // reset on successful connect
      };

      ws.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data) as PaymentUpdate;
          setStatus(data);
          // Stop reconnecting if payment is terminal
          if (data.status === 'completed' || data.status === 'expired') {
            retriesRef.current = MAX_RETRIES;
          }
        } catch { /* ignore malformed messages */ }
      };

      ws.onclose = () => {
        if (unmounted) return;
        if (retriesRef.current < MAX_RETRIES) {
          const delay = Math.min(1000 * Math.pow(2, retriesRef.current), 16000);
          retriesRef.current++;
          reconnectTimer = setTimeout(connect, delay);
        }
      };

      ws.onerror = () => {
        ws.close();
      };
    }

    connect();

  useEffect(() => {
    connect();
    return () => {
      unmounted = true;
      clearTimeout(reconnectTimer);
      wsRef.current?.close();
      wsRef.current = null;
    };
  }, [connect]);

  return { status, connected };
}
