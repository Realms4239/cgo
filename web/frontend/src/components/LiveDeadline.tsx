// LiveDeadline — le pourcentage d'échéances du flux vivant. Une absence se dit
// par un mot et « — », jamais par un zéro ; un zéro mesuré garde son format.
import { numFr } from '../lib/format'
import type { LiveFrame } from '../lib/types'

export type DeadlineFrame = Pick<LiveFrame, 'deadline_ok_pct' | 'deadline_measured'>

export const DEADLINE_ABSENT = '— absent'

/** Un drapeau absent ou faux rend la valeur non mesurée, même si un 0 voyage. */
export function deadlinePctText(frame: DeadlineFrame | null | undefined): string {
  const pct = frame?.deadline_ok_pct
  if (frame?.deadline_measured !== true || pct == null || !Number.isFinite(pct)) return DEADLINE_ABSENT
  return numFr(pct, 1)
}

export default function LiveDeadline({ frame }: { frame?: DeadlineFrame | null }) {
  const text = deadlinePctText(frame)
  const absent = text === DEADLINE_ABSENT
  return (
    <span className="live-deadline mono" data-absent={absent ? 'true' : undefined}>
      <span className="live-deadline-label">échéances</span>{' '}
      <b className="live-deadline-value">{absent ? text : `${text} %`}</b>
    </span>
  )
}
