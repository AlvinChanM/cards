import './ActionBar.css'

interface ActionBarProps {
  canPlay: boolean
  canPass: boolean
  isMyTurn: boolean
  onPlay: () => void
  onPass: () => void
  onHint: () => void
  hintAvailable: boolean
  hintCount: number
}

export function ActionBar({
  canPlay,
  canPass,
  isMyTurn,
  onPlay,
  onPass,
  onHint,
  hintAvailable,
  hintCount,
}: ActionBarProps) {
  return (
    <div className="action-bar">
      <div className="action-bar-status">{isMyTurn ? '轮到你出牌' : '等待对方...'}</div>
      <div className="action-bar-buttons">
        <button type="button" disabled={!isMyTurn} onClick={onHint} className="btn btn-hint">
          推荐打法{hintAvailable && hintCount > 1 ? ` (${hintCount})` : ''}
        </button>
        <button type="button" disabled={!canPass} onClick={onPass} className="btn btn-secondary">
          Pass
        </button>
        <button type="button" disabled={!canPlay} onClick={onPlay} className="btn btn-primary">
          出牌
        </button>
      </div>
    </div>
  )
}
