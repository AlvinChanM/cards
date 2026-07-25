import './OpponentHandBadge.css'

interface OpponentHandBadgeProps {
  count: number
  name: string
  isOpponentTurn?: boolean
  lowCardAlert?: boolean
}

export function OpponentHandBadge({ count, name, isOpponentTurn, lowCardAlert }: OpponentHandBadgeProps) {
  return (
    <div className={`opponent-badge ${isOpponentTurn ? 'opponent-badge-active' : ''} ${lowCardAlert ? 'opponent-badge-alert' : ''}`}>
      <div className="opponent-name">{name || '对手'}</div>
      <div className="opponent-cards">
        {Array.from({ length: Math.min(count, 12) }).map((_, i) => (
          <div key={i} className="opponent-card-back" />
        ))}
      </div>
      <div className="opponent-count">剩 {count} 张</div>
    </div>
  )
}
