import { useCallback, useEffect, useRef, useState } from 'react'
import type { Envelope } from '../types/protocol'

export type ConnectionStatus = 'connecting' | 'open' | 'closed'

interface UseWebSocketOptions {
  url: string
  onMessage: (envelope: Envelope) => void
}

export function useWebSocket({ url, onMessage }: UseWebSocketOptions) {
  const [status, setStatus] = useState<ConnectionStatus>('connecting')
  const wsRef = useRef<WebSocket | null>(null)
  const onMessageRef = useRef(onMessage)
  onMessageRef.current = onMessage

  useEffect(() => {
    const ws = new WebSocket(url)
    wsRef.current = ws

    ws.onopen = () => setStatus('open')
    ws.onclose = () => setStatus('closed')
    ws.onerror = () => setStatus('closed')
    ws.onmessage = (ev) => {
      try {
        const envelope = JSON.parse(ev.data) as Envelope
        onMessageRef.current(envelope)
      } catch (err) {
        console.error('failed to parse WS message', err)
      }
    }

    return () => {
      ws.close()
    }
  }, [url])

  const send = useCallback((type: string, data: unknown) => {
    const ws = wsRef.current
    if (!ws || ws.readyState !== WebSocket.OPEN) {
      console.warn('WS not open, dropping message', type)
      return
    }
    ws.send(JSON.stringify({ type, data }))
  }, [])

  return { status, send }
}
