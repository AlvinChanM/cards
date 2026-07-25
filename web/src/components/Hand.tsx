import { useEffect, useRef, useState } from 'react'
import type { CardDTO } from '../types/protocol'
import { cardKey, sortCards } from '../types/protocol'
import { CardComponent } from './CardComponent'
import { playSelectSound, playDeselectSound } from '../sound'
import './Hand.css'

interface HandProps {
  cards: CardDTO[]
  onSelectionChange: (selected: CardDTO[]) => void
  forcedSelection?: CardDTO[] | null
  isMyTurn?: boolean
  lowCardAlert?: boolean
}

export function Hand({ cards, onSelectionChange, forcedSelection, isMyTurn, lowCardAlert }: HandProps) {
  const [selected, setSelected] = useState<Set<string>>(new Set())
  const sorted = sortCards(cards)

  // When a hint is active, forcedSelection drives what's highlighted --
  // sync it into local selection state so the user can still tweak it
  // (e.g. deselect one suggested card) before playing.
  useEffect(() => {
    if (forcedSelection) {
      setSelected(new Set(forcedSelection.map(cardKey)))
      onSelectionChange(forcedSelection)
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [forcedSelection])

  const isDragging = useRef(false)
  const dragTouched = useRef<Set<string>>(new Set())

  const applySelection = (next: Set<string>) => {
    setSelected(next)
    onSelectionChange(sorted.filter((c) => next.has(cardKey(c))))
  }

  const toggleCard = (card: CardDTO) => {
    const key = cardKey(card)
    const next = new Set(selected)
    if (next.has(key)) {
      next.delete(key)
      playDeselectSound()
    } else {
      next.add(key)
      playSelectSound()
    }
    applySelection(next)
  }

  const handlePointerDown = (card: CardDTO) => (e: React.PointerEvent) => {
    if (e.button !== 0) return // left button only
    isDragging.current = true
    dragTouched.current = new Set([cardKey(card)])
    toggleCard(card)
  }

  const handlePointerEnter = (card: CardDTO) => () => {
    if (!isDragging.current) return
    const key = cardKey(card)
    if (dragTouched.current.has(key)) return
    dragTouched.current.add(key)
    toggleCard(card)
  }

  const endDrag = () => {
    isDragging.current = false
    dragTouched.current = new Set()
  }

  return (
    <div
      className={`hand ${isMyTurn ? 'hand-active' : ''} ${lowCardAlert ? 'hand-alert' : ''}`}
      onPointerUp={endDrag}
      onPointerLeave={endDrag}
    >
      {sorted.map((card, i) => (
        <div key={`${cardKey(card)}-${i}`} className="hand-card-slot">
          <CardComponent
            card={card}
            selected={selected.has(cardKey(card))}
            onPointerDown={handlePointerDown(card)}
            onPointerEnter={handlePointerEnter(card)}
          />
        </div>
      ))}
    </div>
  )
}
