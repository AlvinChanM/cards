import './PlayerAvatar.css'

interface PlayerAvatarProps {
  seatNumber: number
  playerName?: string
  size?: 'small' | 'medium' | 'large'
}

export function PlayerAvatar({ seatNumber, playerName, size = 'medium' }: PlayerAvatarProps) {
  const initials = (playerName || `玩家${seatNumber}`)
    .slice(0, 2)
    .toUpperCase()

  return (
    <div className={`player-avatar player-avatar-${size} player-avatar-seat-${seatNumber}`}>
      <div className="avatar-circle">{initials}</div>
    </div>
  )
}
