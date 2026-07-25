import { useEffect } from 'react'
import type { CardDTO } from '../types/protocol'
import { CardComponent } from './CardComponent'
import './PlayCardFlyLayer.css'

// Category strings mirror internal/ws/convert.go's categoryToStr.
const STAMP_LABEL: Record<string, string> = {
  straight: '顺子',
  consecutive_pairs: '连对',
  airplane: '飞机',
  airplane_plus_singles: '飞机',
  airplane_plus_pairs: '飞机',
}

export interface FlyingPlay {
  id: number
  cards: CardDTO[]
  category: string
  direction: 'up' | 'down' // 'up' = played by you (flies from bottom), 'down' = opponent (flies from top)
}

interface PlayCardFlyLayerProps {
  play: FlyingPlay
  onDone: () => void
}

// Duration must match the longest keyframe animation below (fly-in card).
const FLY_DURATION_MS = 550
const STAMP_DURATION_MS = 700
const BOMB_DURATION_MS = 900

export function PlayCardFlyLayer({ play, onDone }: PlayCardFlyLayerProps) {
  const isBomb = play.category === 'bomb'
  const stampLabel = STAMP_LABEL[play.category]

  useEffect(() => {
    const duration = isBomb ? BOMB_DURATION_MS : stampLabel ? STAMP_DURATION_MS : FLY_DURATION_MS
    const t = setTimeout(onDone, duration)
    return () => clearTimeout(t)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [play.id])

  return (
    <div className={`play-fly-layer play-fly-${play.direction}`}>
      <div className="play-fly-cards">
        {play.cards.map((c, i) => (
          <div key={`${c.suit}${c.rank}-${i}`} className="play-fly-card-slot" style={{ animationDelay: `${i * 25}ms` }}>
            <CardComponent card={c} />
          </div>
        ))}
      </div>

      {stampLabel && (
        <div className="play-stamp">
          <span>{stampLabel}</span>
        </div>
      )}

      {isBomb && (
        <div className="bomb-fx">
          <div className="bomb-flash" />
          <div className="bomb-text">炸弹！</div>
        </div>
      )}
    </div>
  )
}
