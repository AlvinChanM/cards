import { useRef } from 'react'
import { GameProvider, useGame } from './state/GameContext'
import { useWebSocket } from './hooks/useWebSocket'
import { WaitingRoom } from './components/WaitingRoom'
import { GameTable } from './components/GameTable'
import type { Envelope } from './types/protocol'
import './App.css'

function wsURL(): string {
  const proto = window.location.protocol === 'https:' ? 'wss' : 'ws'
  return `${proto}://${window.location.host}/ws`
}

function GameScreen() {
  const { state } = useGame()

  if (state.seat === 'unassigned' || !state.sync) {
    return <WaitingRoom playersConnected={state.playersConnected} />
  }
  return <GameTable />
}

// Bridges the WebSocket connection into GameProvider: GameProvider owns
// the reducer and needs a `send` function up front, while the socket's
// onMessage callback needs to reach that same provider's dispatch. A
// ref indirection lets both sides exist without a circular dependency.
function ConnectedGame() {
  const dispatchRef = useRef<((envelope: Envelope) => void) | null>(null)
  const { send } = useWebSocket({
    url: wsURL(),
    onMessage: (envelope) => dispatchRef.current?.(envelope),
  })

  return (
    <GameProvider send={send}>
      <DispatchRefSetter dispatchRef={dispatchRef} />
      <GameScreen />
    </GameProvider>
  )
}

function DispatchRefSetter({
  dispatchRef,
}: {
  dispatchRef: React.MutableRefObject<((envelope: Envelope) => void) | null>
}) {
  const { dispatchEnvelope } = useGame()
  dispatchRef.current = dispatchEnvelope
  return null
}

function App() {
  return <ConnectedGame />
}

export default App
