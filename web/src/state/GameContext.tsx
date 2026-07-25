import { createContext, useContext, useReducer, type ReactNode } from 'react'
import type {
  CardDTO,
  Envelope,
  ErrorMessage,
  HintMessage,
  JoinedMessage,
  PassResultMessage,
  PlayResultMessage,
  RoundOverMessage,
  StateSyncMessage,
  WaitingMessage,
} from '../types/protocol'
import { MessageType } from '../types/protocol'

interface GameState {
  seat: 'unassigned' | '0' | '1' | 'spectator'
  playersConnected: number
  sync: StateSyncMessage | null
  lastPlayResult: PlayResultMessage | null
  lastPassResult: PassResultMessage | null
  lastRoundOver: RoundOverMessage | null
  lastError: ErrorMessage | null
  hintPlays: CardDTO[][]
  hintIndex: number
}

const initialState: GameState = {
  seat: 'unassigned',
  playersConnected: 0,
  sync: null,
  lastPlayResult: null,
  lastPassResult: null,
  lastRoundOver: null,
  lastError: null,
  hintPlays: [],
  hintIndex: -1,
}

type Action = { type: 'envelope'; envelope: Envelope } | { type: 'cycleHint' } | { type: 'clearHint' }

function reducer(state: GameState, action: Action): GameState {
  if (action.type === 'cycleHint') {
    if (state.hintPlays.length === 0) return state
    return { ...state, hintIndex: (state.hintIndex + 1) % state.hintPlays.length }
  }
  if (action.type === 'clearHint') {
    return { ...state, hintPlays: [], hintIndex: -1 }
  }

  const { envelope } = action

  switch (envelope.type) {
    case MessageType.Joined: {
      const data = envelope.data as JoinedMessage
      return { ...state, seat: data.seat }
    }
    case MessageType.Waiting: {
      const data = envelope.data as WaitingMessage
      return { ...state, playersConnected: data.playersConnected, sync: null }
    }
    case MessageType.StateSync: {
      const data = envelope.data as StateSyncMessage
      return { ...state, sync: data, lastRoundOver: null, hintPlays: [], hintIndex: -1 }
    }
    case MessageType.PlayResult: {
      const data = envelope.data as PlayResultMessage
      return { ...state, lastPlayResult: data, lastError: null, hintPlays: [], hintIndex: -1 }
    }
    case MessageType.PassResult: {
      const data = envelope.data as PassResultMessage
      return { ...state, lastPassResult: data, lastError: null, hintPlays: [], hintIndex: -1 }
    }
    case MessageType.RoundOver: {
      const data = envelope.data as RoundOverMessage
      return { ...state, lastRoundOver: data }
    }
    case MessageType.Error: {
      const data = envelope.data as ErrorMessage
      return { ...state, lastError: data }
    }
    case MessageType.Hint: {
      const data = envelope.data as HintMessage
      return { ...state, hintPlays: data.plays, hintIndex: data.plays.length > 0 ? 0 : -1 }
    }
    default:
      return state
  }
}

interface GameContextValue {
  state: GameState
  dispatchEnvelope: (envelope: Envelope) => void
  send: (type: string, data: unknown) => void
  cycleHint: () => void
  clearHint: () => void
}

const GameContext = createContext<GameContextValue | null>(null)

export function GameProvider({
  children,
  send,
}: {
  children: ReactNode
  send: (type: string, data: unknown) => void
}) {
  const [state, dispatch] = useReducer(reducer, initialState)
  const dispatchEnvelope = (envelope: Envelope) => dispatch({ type: 'envelope', envelope })
  const cycleHint = () => dispatch({ type: 'cycleHint' })
  const clearHint = () => dispatch({ type: 'clearHint' })

  return (
    <GameContext.Provider value={{ state, dispatchEnvelope, send, cycleHint, clearHint }}>
      {children}
    </GameContext.Provider>
  )
}

export function useGame() {
  const ctx = useContext(GameContext)
  if (!ctx) throw new Error('useGame must be used within a GameProvider')
  return ctx
}

export function isMyTurn(state: GameState): boolean {
  if (!state.sync) return false
  if (state.sync.yourSeat !== 0 && state.sync.yourSeat !== 1) return false
  return state.sync.currentTurn === state.sync.yourSeat
}

export function selectionKeys(cards: CardDTO[]): Set<string> {
  return new Set(cards.map((c) => `${c.suit}${c.rank}`))
}
