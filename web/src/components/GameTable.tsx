import { useEffect, useRef, useState } from 'react'
import type { CardDTO } from '../types/protocol'
import { MessageType } from '../types/protocol'
import { useGame, isMyTurn } from '../state/GameContext'
import { Hand } from './Hand'
import { PlayedCardsArea } from './PlayedCardsArea'
import { OpponentHandBadge } from './OpponentHandBadge'
import { ActionBar } from './ActionBar'
import { ScoreBoard } from './ScoreBoard'
import { RoundOverModal } from './RoundOverModal'
import { PlayCardFlyLayer, type FlyingPlay } from './PlayCardFlyLayer'
import { PassBubble } from './PassBubble'
import { DealAnimation } from './DealAnimation'
import { playCardSound, passSound, roundOverSound, dealSound } from '../sound'
import './GameTable.css'

const LOW_CARD_THRESHOLD = 2

export function GameTable() {
  const { state, send, cycleHint, clearHint } = useGame()
  const [selected, setSelected] = useState<CardDTO[]>([])
  const [flyingPlay, setFlyingPlay] = useState<FlyingPlay | null>(null)
  const [passBubble, setPassBubble] = useState<{ id: number; side: 'you' | 'opponent' } | null>(null)
  const [tableShake, setTableShake] = useState(false)
  const [dealing, setDealing] = useState(false)

  const sync = state.sync

  const lastPlaySeen = useRef<unknown>(null)
  const lastPassSeen = useRef<unknown>(null)
  const lastRoundOverSeen = useRef<unknown>(null)
  const prevPhaseRef = useRef<string | null>(null)
  const flyIdRef = useRef(0)
  const passIdRef = useRef(0)

  useEffect(() => {
    if (state.lastPlayResult && state.lastPlayResult !== lastPlaySeen.current) {
      lastPlaySeen.current = state.lastPlayResult
      playCardSound()

      const result = state.lastPlayResult
      const direction = sync && result.playerIdx === sync.yourSeat ? 'up' : 'down'
      flyIdRef.current += 1
      setFlyingPlay({ id: flyIdRef.current, cards: result.cards, category: result.category, direction })

      if (result.category === 'bomb') {
        setTableShake(true)
        setTimeout(() => setTableShake(false), 500)
      }
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [state.lastPlayResult])

  useEffect(() => {
    if (state.lastPassResult && state.lastPassResult !== lastPassSeen.current) {
      lastPassSeen.current = state.lastPassResult
      passSound()

      const result = state.lastPassResult
      const side = sync && result.playerIdx === sync.yourSeat ? 'you' : 'opponent'
      passIdRef.current += 1
      setPassBubble({ id: passIdRef.current, side })
      setTimeout(() => setPassBubble(null), 1000)
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [state.lastPassResult])

  useEffect(() => {
    if (state.lastRoundOver && state.lastRoundOver !== lastRoundOverSeen.current) {
      lastRoundOverSeen.current = state.lastRoundOver
      roundOverSound()
    }
  }, [state.lastRoundOver])

  useEffect(() => {
    const phase = sync?.phase ?? null
    if (phase === 'playing' && prevPhaseRef.current !== 'playing') {
      setDealing(true)
      dealSound()
    }
    prevPhaseRef.current = phase
  }, [sync?.phase])

  if (!sync) return null

  const myTurn = isMyTurn(state)
  const opponentIdx = sync.yourSeat === 0 ? 1 : 0
  const yourLowCards = (sync.yourHand?.length ?? 0) > 0 && (sync.yourHand?.length ?? 0) <= LOW_CARD_THRESHOLD
  const opponentLowCards = sync.opponentCardCount > 0 && sync.opponentCardCount <= LOW_CARD_THRESHOLD

  const playerLabel = (idx: number) => {
    const name = sync.playerNames[idx] || `玩家${idx}`
    return idx === sync.yourSeat ? `你 (${name})` : name
  }

  const handlePlay = () => {
    if (selected.length === 0) return
    send(MessageType.PlayCards, { cards: selected })
    setSelected([])
    clearHint()
  }

  const handlePass = () => {
    send(MessageType.Pass, {})
    clearHint()
  }

  const handleHint = () => {
    if (state.hintPlays.length === 0) {
      send(MessageType.RequestHint, {})
    } else {
      cycleHint()
    }
  }

  const dismissRoundOver = () => {
    send(MessageType.RequestState, {})
  }

  return (
    <div className={`game-table ${tableShake ? 'game-table-shake' : ''}`}>
      <ScoreBoard roundsWon={sync.roundsWon} playerNames={sync.playerNames} yourSeat={sync.yourSeat} />

      <div className="game-table-opponent">
        <OpponentHandBadge
          count={sync.opponentCardCount}
          name={sync.playerNames[opponentIdx]}
          seatNumber={opponentIdx}
          isOpponentTurn={!myTurn && sync.phase === 'playing'}
          lowCardAlert={opponentLowCards}
        />
      </div>

      <div className="game-table-center">
        <PlayedCardsArea lastPlay={sync.lastPlay} playerLabel={playerLabel} />
      </div>

      {state.lastError && <div className="error-toast">{state.lastError.message}</div>}

      {flyingPlay && <PlayCardFlyLayer play={flyingPlay} onDone={() => setFlyingPlay(null)} />}
      {passBubble && <PassBubble key={passBubble.id} side={passBubble.side} />}
      {dealing && <DealAnimation onDone={() => setDealing(false)} />}

      <Hand
        cards={sync.yourHand ?? []}
        onSelectionChange={setSelected}
        forcedSelection={state.hintIndex >= 0 ? state.hintPlays[state.hintIndex] : null}
        isMyTurn={myTurn}
        lowCardAlert={yourLowCards}
      />

      <ActionBar
        canPlay={myTurn && selected.length > 0}
        canPass={myTurn && sync.lastPlay !== null}
        isMyTurn={myTurn}
        onPlay={handlePlay}
        onPass={handlePass}
        onHint={handleHint}
        hintAvailable={state.hintPlays.length > 0}
        hintCount={state.hintPlays.length}
      />

      {state.lastRoundOver && (
        <RoundOverModal
          winner={state.lastRoundOver.winner}
          yourSeat={sync.yourSeat}
          playerNames={sync.playerNames}
          onDismiss={dismissRoundOver}
        />
      )}
    </div>
  )
}

