import type { LastPlayDTO } from '../types/protocol'
import { CardComponent } from './CardComponent'
import './PlayedCardsArea.css'

interface PlayedCardsAreaProps {
  lastPlay: LastPlayDTO | null
  playerLabel: (idx: number) => string
}

export function PlayedCardsArea({ lastPlay, playerLabel }: PlayedCardsAreaProps) {
  if (!lastPlay) {
    return <div className="played-area played-area-empty">等待出牌...</div>
  }

  return (
    <div className="played-area">
      <div className="played-area-label">{playerLabel(lastPlay.playerIdx)} 出牌</div>
      <div className="played-cards">
        {lastPlay.cards.map((c, i) => (
          <div key={`${c.suit}${c.rank}-${i}`} className="played-card-slot">
            <CardComponent card={c} />
          </div>
        ))}
      </div>
    </div>
  )
}
