import { describe, expect, it } from 'vitest'
import {
  archiveStatus,
  archiveVerdictWord,
  computeWindow,
  formatArchiveDate,
  type ArchiveRow,
} from './archive'

describe('archive window', () => {
  it('renders only visible rows plus overscan', () => {
    expect(computeWindow({ total: 179, scrollTop: 900, rowHeight: 40, viewportHeight: 400, overscan: 3 }))
      .toEqual({ start: 19, end: 36 })
  })

  it('never returns an invalid slice at the end', () => {
    const w = computeWindow({ total: 179, scrollTop: 10000, rowHeight: 40, viewportHeight: 400, overscan: 3 })
    expect(w.end).toBe(179)
    expect(w.start).toBeGreaterThanOrEqual(0)
    expect(w.start).toBeLessThan(w.end)
  })

  it('clamps an empty archive to an empty window', () => {
    expect(computeWindow({ total: 0, scrollTop: 0, rowHeight: 40, viewportHeight: 400, overscan: 3 }))
      .toEqual({ start: 0, end: 0 })
  })

  it('keeps a coherent window when the archive is smaller than the viewport', () => {
    const w = computeWindow({ total: 5, scrollTop: 0, rowHeight: 40, viewportHeight: 400, overscan: 3 })
    expect(w).toEqual({ start: 0, end: 5 })
  })
})

describe('archive sealed values', () => {
  const sealed = (verdict_level: string): ArchiveRow => ({ has_report: 'true', verdict_level })

  it('maps sealed verdicts to the three words, never to a colour guess', () => {
    expect(archiveStatus(sealed('green'))).toBe('Valide')
    expect(archiveStatus(sealed('amber'))).toBe('Dégradé')
    expect(archiveStatus(sealed('red'))).toBe('Dégradé')
    expect(archiveStatus(sealed('neutral'))).toBe('Observation seule')
  })

  it('shows nothing sealed when the dossier is absent', () => {
    expect(archiveStatus({ has_report: 'false' })).toBe('—')
    expect(archiveStatus({})).toBe('—')
    expect(archiveStatus({ has_report: 'true' })).toBe('—')
  })

  it('keeps the sealed verdict word, else says Non noté', () => {
    expect(archiveVerdictWord({ has_report: 'true', verdict_word: 'Lien sain' })).toBe('Lien sain')
    expect(archiveVerdictWord({ has_report: 'true' })).toBe('Non noté')
    expect(archiveVerdictWord({ has_report: 'false' })).toBe('Non noté')
  })
})

describe('formatArchiveDate', () => {
  it('renders YYYY-MM-DD HH:mm from an ISO timestamp', () => {
    expect(formatArchiveDate('2026-10-29T08:15:00Z')).toMatch(/^\d{4}-\d{2}-\d{2} \d{2}:\d{2}$/)
    expect(formatArchiveDate('nimportequoi')).toBe('—')
    expect(formatArchiveDate('')).toBe('—')
  })
})
