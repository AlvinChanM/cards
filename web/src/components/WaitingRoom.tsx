import './WaitingRoom.css'

interface WaitingRoomProps {
  playersConnected: number
}

export function WaitingRoom({ playersConnected }: WaitingRoomProps) {
  return (
    <div className="waiting-room">
      <h2>等待对手加入...</h2>
      <p>
        当前已连接 {playersConnected} / 2 位玩家
      </p>
    </div>
  )
}
