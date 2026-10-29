type Props = { id: string; size?: number }

const common = {
  fill: 'none',
  stroke: 'currentColor',
  strokeWidth: 1.6,
  strokeLinecap: 'round' as const,
  strokeLinejoin: 'round' as const,
  'aria-hidden': true,
}

export function NavIcon({ id, size = 18 }: Props) {
  const p = { width: size, height: size, viewBox: '0 0 24 24', ...common }
  switch (id) {
    case 'synthese':
      return <svg {...p}><rect x="3.5" y="3.5" width="7" height="7" rx="1.2" /><rect x="13.5" y="3.5" width="7" height="7" rx="1.2" /><rect x="3.5" y="13.5" width="7" height="7" rx="1.2" /><rect x="13.5" y="13.5" width="7" height="7" rx="1.2" /></svg>
    case 'campagne':
      return <svg {...p}><path d="M4 7h9M19 7h1M4 17h3M13 17h7" /><circle cx="16" cy="7" r="2.2" /><circle cx="10" cy="17" r="2.2" /></svg>
    case 'live':
      return <svg {...p}><path d="M3 12h4l2-6 4 12 2-6h6" /></svg>
    case 'resultats':
      return <svg {...p}><path d="M4 20V10M10 20V4M16 20v-7M22 20H2" /></svg>
    case 'integrite':
      return <svg {...p}><path d="M12 3l7 3v5c0 5-3.5 8-7 10-3.5-2-7-5-7-10V6z" /><path d="M9 12l2 2 4-4" /></svg>
    default:
      return <svg {...p}><circle cx="12" cy="12" r="8" /></svg>
  }
}
