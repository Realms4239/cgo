import type { APIResponse } from '@playwright/test'

// Données de test déterministes : aucune preuve réelle, aucune source de prod.
export const VALID_REPORT = JSON.stringify({
  audit_id: 'archive-valid',
  site: 'Fort-Dauphin',
  link_type: 'fiber',
  timestamp: '2026-11-01T10:00:00Z',
  bloat_grade: 'A',
  remediation_state: 'available',
  measured: {
    direction: 'down',
    load_measured: true,
    rtt_idle_p95_ms: 42,
    rtt_loaded_p95_ms: 84,
    bloat_delta_ms: 42,
    throughput_mbps: 78,
    loss_pct: 0.1,
  },
  provenance: {
    source_file: 'audit_report.json',
    source_sha256: 'a'.repeat(64),
    telemetry_file: 'archive.csv',
    telemetry_sha256: 'b'.repeat(64),
    row_count: 1,
  },
  cap_sla: { measured: true, breach: false, ok_pct: 99.2 },
  remediation: { qdisc: 'cake', linux_tc: ['tc qdisc add dev eth0 root cake'] },
})

export const ABSENT_REPORT = JSON.stringify({
  audit_id: 'archive-absent',
  site: 'Toleara',
  link_type: 'vsat',
  timestamp: '2026-11-02T10:00:00Z',
  measured: { direction: 'down', load_measured: false },
})

export const CORRUPT_PAYLOAD = '{"audit_id": "archive-corrupt",'

const row = (n: number) => ({
  audit_id: `archive-${n}`,
  site: `site-${String(n).padStart(3, '0')}`,
  link_type: n % 3 === 0 ? 'fiber' : n % 3 === 1 ? '4g' : 'vsat',
  timestamp: `2026-10-${String(1 + (n % 28)).padStart(2, '0')}T10:00:00Z`,
  has_report: 'true',
  verdict_level: n % 4 === 0 ? 'neutral' : 'green',
  verdict_word: n % 4 === 0 ? 'Creux' : 'Tenue',
  rtt_loaded_p95_ms: String(60 + (n % 40)),
})

export const ARCHIVE_179_ROWS = JSON.stringify({
  audits: Array.from({ length: 179 }, (_, i) => row(i + 1)),
})

export const ARCHIVE_ONE_ROW = JSON.stringify({ audits: [row(1)] })

// Limiteur déclaré par principal (60 lectures/min rafale 20, 10 mutations/min
// rafale 5) : un 429 est transitoire, jamais une réussite, et il porte un
// Retry-After entier en secondes. On l'honore avec un plafond court pour qu'un
// test ne puisse pas attendre indéfiniment ; budget épuisé, on renvoie le
// dernier statut réel et l'assertion échoue sur ce statut, sans le masquer.
const MAX_RETRIES_429 = 4
const MAX_WAIT_429_MS = 8_000

export async function retryOn429(send: () => Promise<APIResponse>): Promise<APIResponse> {
  let res = await send()
  for (let attempt = 0; attempt < MAX_RETRIES_429 && res.status() === 429; attempt++) {
    const seconds = Number(res.headers()['retry-after'])
    const waitMs = Number.isFinite(seconds) && seconds > 0
      ? Math.min(seconds * 1000, MAX_WAIT_429_MS)
      : 1_000
    await new Promise((resolve) => setTimeout(resolve, waitMs))
    res = await send()
  }
  return res
}
