// Mirrors internal/ws/protocol.go. Keep these two files in sync by hand.

export interface Envelope<T = unknown> {
  type: string
  data: T
}

export interface CardDTO {
  suit: 'S' | 'H' | 'C' | 'D'
  rank: '3' | '4' | '5' | '6' | '7' | '8' | '9' | '10' | 'J' | 'Q' | 'K' | 'A' | '2'
}

// --- Client -> Server ---

export interface JoinRequest {
  name: string
}

export interface PlayCardsRequest {
  cards: CardDTO[]
}

export type PassRequest = Record<string, never>
export type RequestStateRequest = Record<string, never>
export type RequestHintRequest = Record<string, never>

// --- Server -> Client ---

export interface JoinedMessage {
  seat: '0' | '1' | 'spectator'
  playerNames: [string, string]
}

export interface LastPlayDTO {
  playerIdx: number
  cards: CardDTO[]
  category: string
}

export interface StateSyncMessage {
  phase: 'waiting' | 'playing' | 'round_over' | 'unknown'
  yourSeat: number
  yourHand?: CardDTO[]
  opponentCardCount: number
  currentTurn: number
  leadPlayer: number
  lastPlay: LastPlayDTO | null
  roundsWon: [number, number]
  playerNames: [string, string]
}

export interface PlayResultMessage {
  playerIdx: number
  cards: CardDTO[]
  category: string
}

export interface PassResultMessage {
  playerIdx: number
}

export interface RoundOverMessage {
  winner: number
  roundsWon: [number, number]
}

export interface WaitingMessage {
  playersConnected: number
}

export interface ErrorMessage {
  code: string
  message: string
}

export interface HintMessage {
  plays: CardDTO[][]
}

export const MessageType = {
  Join: 'join',
  PlayCards: 'play_cards',
  Pass: 'pass',
  RequestState: 'request_state',
  RequestHint: 'request_hint',

  Joined: 'joined',
  StateSync: 'state_sync',
  PlayResult: 'play_result',
  PassResult: 'pass_result',
  RoundOver: 'round_over',
  Waiting: 'waiting_for_opponent',
  Error: 'error',
  Hint: 'hint',
} as const

// Rank ordering for client-side sorting/display, mirrors game.Rank.
export const RANK_ORDER: CardDTO['rank'][] = [
  '3', '4', '5', '6', '7', '8', '9', '10', 'J', 'Q', 'K', 'A', '2',
]

export function rankValue(rank: CardDTO['rank']): number {
  return RANK_ORDER.indexOf(rank)
}

export function sortCards(cards: CardDTO[]): CardDTO[] {
  return [...cards].sort((a, b) => rankValue(a.rank) - rankValue(b.rank))
}

export function cardKey(c: CardDTO): string {
  return `${c.suit}${c.rank}`
}
