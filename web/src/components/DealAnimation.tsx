import { useEffect } from 'react'
import './DealAnimation.css'

const DEAL_COUNT = 8
const STAGGER_MS = 55
const FLIGHT_MS = 420
export const DEAL_DURATION_MS = (DEAL_COUNT - 1) * STAGGER_MS + FLIGHT_MS

interface DealAnimationProps {
  onDone: () => void
}

// Simplified deal animation: fires a fixed number of card-back sprites
// from a central "deck" toward the bottom (your hand) and top
// (opponent's hand), staggered to mimic a dealer's rhythm. Does not
// attempt to model the real 24-card alternating deal -- that would
// take several seconds and add little beyond this.
export function DealAnimation({ onDone }: DealAnimationProps) {
  useEffect(() => {
    const t = setTimeout(onDone, DEAL_DURATION_MS)
    return () => clearTimeout(t)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  return (
    <div className="deal-layer">
      <div className="deal-deck" />
      {Array.from({ length: DEAL_COUNT }).map((_, i) => (
        <div key={`you-${i}`} className="deal-card deal-card-down" style={{ animationDelay: `${i * STAGGER_MS}ms` }} />
      ))}
      {Array.from({ length: DEAL_COUNT }).map((_, i) => (
        <div key={`opp-${i}`} className="deal-card deal-card-up" style={{ animationDelay: `${i * STAGGER_MS}ms` }} />
      ))}
    </div>
  )
}
