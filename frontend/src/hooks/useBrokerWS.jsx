import { useState, useEffect, useRef } from 'preact/hooks';

export function useBrokerWS() {
  const [ws, setWs] = useState(null);
  const [connected, setConnected] = useState(false);
  const handlersRef = useRef(new Map());
  const wsRef = useRef(null);

  useEffect(() => {
    const token = localStorage.getItem('jwt') || 'anonymous';
    const wsUrl = `ws://localhost:8082/v1/ws`;
    
    const ws = new WebSocket(wsUrl);
    wsRef.current = ws;

    ws.onopen = () => {
      setConnected(true);
      console.log('WebSocket connected');
    };

    ws.onmessage = (event) => {
      try {
        const message = JSON.parse(event.data);
        const channel = message.channel;
        
        if (channel && handlersRef.current.has(channel)) {
          const handler = handlersRef.current.get(channel);
          handler(message);
        }
      } catch (err) {
        console.error('Failed to parse WebSocket message:', err);
      }
    };

    ws.onerror = (error) => {
      console.error('WebSocket error:', error);
      setConnected(false);
    };

    ws.onclose = () => {
      setConnected(false);
      console.log('WebSocket disconnected');
      // Attempt to reconnect after 3 seconds
      setTimeout(() => {
        if (wsRef.current?.readyState === WebSocket.CLOSED) {
          setWs(null);
        }
      }, 3000);
    };

    setWs({
      subscribe: (channel) => {
        ws.send(JSON.stringify({
          type: 'subscribe',
          channel: channel,
        }));
      },
      unsubscribe: (channel) => {
        ws.send(JSON.stringify({
          type: 'unsubscribe',
          channel: channel,
        }));
        handlersRef.current.delete(channel);
      },
      on: (channel, handler) => {
        handlersRef.current.set(channel, handler);
      },
      send: (message) => {
        if (ws.readyState === WebSocket.OPEN) {
          ws.send(JSON.stringify(message));
        }
      },
      connected,
    });

    return () => {
      if (wsRef.current) {
        wsRef.current.close();
      }
    };
  }, []);

  return ws;
}

