import './PassBubble.css'

interface PassBubbleProps {
  side: 'you' | 'opponent'
}

export function PassBubble({ side }: PassBubbleProps) {
  return (
    <div className={`pass-bubble pass-bubble-${side}`}>
      <span>不要</span>
    </div>
  )
}
