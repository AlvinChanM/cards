import './RoundOverModal.css'

interface RoundOverModalProps {
  winner: number
  yourSeat: number
  playerNames: [string, string]
  onDismiss: () => void
}

export function RoundOverModal({ winner, yourSeat, playerNames, onDismiss }: RoundOverModalProps) {
  const won = winner === yourSeat
  const winnerName = playerNames[winner] || `玩家${winner}`

  return (
    <div className="modal-backdrop" onClick={onDismiss}>
      <div
        className={`modal-content ${won ? 'modal-content-win' : 'modal-content-lose'}`}
        onClick={(e) => e.stopPropagation()}
      >
        <div className="modal-icon">{won ? '🏆' : '😢'}</div>
        <h2>{won ? '你赢了本局！' : `${winnerName} 赢了本局`}</h2>
        <p>下一局即将开始，赢家先出牌</p>
        <button type="button" className="btn btn-primary" onClick={onDismiss}>
          好的
        </button>
      </div>
    </div>
  )
}
