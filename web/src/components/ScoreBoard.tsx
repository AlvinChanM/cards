import { useCountUp } from '../hooks/useCountUp'
import './ScoreBoard.css'

interface ScoreBoardProps {
  roundsWon: [number, number]
  playerNames: [string, string]
  yourSeat: number
}

export function ScoreBoard({ roundsWon, playerNames, yourSeat }: ScoreBoardProps) {
  const score0 = useCountUp(roundsWon[0])
  const score1 = useCountUp(roundsWon[1])

  const label = (idx: number) => {
    const name = playerNames[idx] || `玩家${idx}`
    return idx === yourSeat ? `${name}（你）` : name
  }

  return (
    <div className="score-board">
      <div className="score-entry">
        <span className="score-name">{label(0)}</span>
        <span className="score-value">{score0}</span>
      </div>
      <span className="score-sep">:</span>
      <div className="score-entry">
        <span className="score-value">{score1}</span>
        <span className="score-name">{label(1)}</span>
      </div>
    </div>
  )
}
