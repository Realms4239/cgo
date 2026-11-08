import { describe, expect, it } from 'vitest'
import { selectAuditViewState } from './viewState'

const report = { audit_id: 'a1' } as any

describe('viewState', () => {
  it('a completed report cannot coexist with a running state', () => {
    const s = selectAuditViewState({
      stage: 'observing',
      live: { running: true, phase: 'charge', elapsed_sec: 11, total_sec: 60 } as any,
      report,
      failure: null,
    })
    expect(s.status).toBe('running')
    expect('report' in s).toBe(false)
  })

  it('a report becomes complete only when the stage closes and report exists', () => {
    const s = selectAuditViewState({ stage: 'done', live: null, report, failure: null })
    expect(s).toEqual({ status: 'complete', report })
  })

  it('a request failure never exposes a previous verdict as current', () => {
    const s = selectAuditViewState({
      stage: 'observing',
      live: null,
      report: null,
      failure: { code: 'HTTP_500', message: 'lecture impossible' },
    })
    expect(s).toEqual({
      status: 'error',
      code: 'HTTP_500',
      message: 'lecture impossible',
      retryAction: expect.any(Function),
    })
  })
})
