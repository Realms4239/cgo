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
    case 'dashboard':
      return <svg {...p}><rect x="3.5" y="3.5" width="7" height="7" rx="1.2" /><rect x="13.5" y="3.5" width="7" height="7" rx="1.2" /><rect x="3.5" y="13.5" width="7" height="7" rx="1.2" /><rect x="13.5" y="13.5" width="7" height="7" rx="1.2" /></svg>
    case 'audit':
      return <svg {...p}><circle cx="10.5" cy="10.5" r="6.5" /><path d="M20 20l-4.6-4.6" /><path d="M7 10.5h1.8l1.1-2.6 1.9 5.2 1-2.6H15" /></svg>
    case 'bench':
      return <svg {...p}><path d="M4 6h10M18 6h2M4 12h3M11 12h9M4 18h7M15 18h5" /><circle cx="16" cy="6" r="2" /><circle cx="9" cy="12" r="2" /><circle cx="13" cy="18" r="2" /></svg>
    case 'evidence':
      return <svg {...p}><path d="M12 3l7 3v5c0 5-3.5 8-7 10-3.5-2-7-5-7-10V6z" /><path d="M9 12l2 2 4-4" /></svg>
    default:
      return <svg {...p}><circle cx="12" cy="12" r="8" /></svg>
  }
}
