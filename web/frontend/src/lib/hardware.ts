export function hardwareRecommendation(bestQdisc: string, profile: string): string {
  if (bestQdisc === 'fq_codel' || bestQdisc === 'cake') return `Si MikroTik: Queue Tree PCQ/CAKE RouterOS v7+ pour ${profile} — ${bestQdisc} prouvé en lab`
  return `Si ISP/mini-PC gateway: transparent bridge CAKE en amont du CPE pour ${profile} — pilote isolé d'abord`
}
