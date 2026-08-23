export interface LiveFrame {
  ts: number
  phase: string
  profile: string
  qdisc: string
  cc: string
  repetition: number
  event_id: number
  load_status: string
  rtt_p50_ms: number
  rtt_p95_ms: number
  small_p95_ms: number
  bulk_goodput_mbps: number
  drops: number
  gates: (boolean | null)[]
  running: boolean
}

export type GateIdx = 0|1|2|3|4|5|6|7
