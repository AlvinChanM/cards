import type { CardDTO } from '../types/protocol'
import './CardComponent.css'

const SUIT_SYMBOL: Record<CardDTO['suit'], string> = {
  S: '♠',
  H: '♥',
  C: '♣',
  D: '♦',
}

const RED_SUITS = new Set(['H', 'D'])

interface CardComponentProps {
  card: CardDTO
  selected?: boolean
  faceDown?: boolean
  onClick?: () => void
  onPointerDown?: (e: React.PointerEvent) => void
  onPointerEnter?: (e: React.PointerEvent) => void
}

export function CardComponent({
  card,
  selected,
  faceDown,
  onClick,
  onPointerDown,
  onPointerEnter,
}: CardComponentProps) {
  if (faceDown) {
    return <div className="card card-back" />
  }

  const isRed = RED_SUITS.has(card.suit)
  return (
    <button
      type="button"
      className={`card ${isRed ? 'card-red' : 'card-black'} ${selected ? 'card-selected' : ''}`}
      onClick={onClick}
      onPointerDown={onPointerDown}
      onPointerEnter={onPointerEnter}
    >
      <span className="card-rank">{card.rank}</span>
      <span className="card-suit">{SUIT_SYMBOL[card.suit]}</span>
    </button>
  )
}
