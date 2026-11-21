import type { ComponentProps } from 'react'
import { renderToStaticMarkup } from 'react-dom/server'
import { describe, expect, it } from 'vitest'
import { StateSlot } from './StateSlot'

type Props = ComponentProps<typeof StateSlot>

function host(props: Props): HTMLElement {
  const el = document.createElement('div')
  el.innerHTML = renderToStaticMarkup(<StateSlot {...props} />)
  return el
}

describe('StateSlot — absence honnête', () => {
  it('absent dit « — » + mot et réserve la même hauteur que la mesure', () => {
    const absent = host({ label: 'RTT repos', word: 'non mesuré', empty: true })
    const measured = host({ label: 'RTT repos', word: 'non mesuré', value: 28, unit: 'ms' })

    expect(absent.textContent).toContain('—')
    expect(absent.textContent).toContain('non mesuré')
    expect(absent.querySelector('[data-state="absent"]')).not.toBeNull()

    const a = absent.querySelector<HTMLElement>('.state-slot-value')
    const m = measured.querySelector<HTMLElement>('.state-slot-value')
    expect(a?.style.minHeight, 'hauteur réservée en permanence').toBeTruthy()
    expect(a?.style.minHeight).toBe(m?.style.minHeight)
  })

  it('une mesure à zéro affiche 0, jamais un tiret', () => {
    const slot = host({ label: 'Pertes', word: 'non mesuré', value: 0, unit: '%', empty: false })
    expect(slot.textContent).toContain('0')
    expect(slot.textContent).not.toContain('—')
    expect(slot.querySelector('[data-state="measured"]')).not.toBeNull()
  })

  it('une valeur nulle est une absence, pas un zéro fabriqué', () => {
    const slot = host({ label: 'Pertes', word: 'non mesuré', value: null })
    expect(slot.textContent).toContain('—')
    expect(slot.textContent).toContain('non mesuré')
    expect(slot.textContent).not.toContain('0')
  })

  it('le mot et la forme portent l’état, jamais la couleur seule', () => {
    const slot = host({ label: 'RTT', word: 'non mesuré', empty: true })
    expect(slot.textContent).toContain('non mesuré')
    expect(slot.textContent).toContain('—')
  })
})
