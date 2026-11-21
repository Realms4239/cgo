import { act, useState, type ReactNode } from 'react'
import { createRoot } from 'react-dom/client'
import { afterEach, beforeEach, describe, expect, it } from 'vitest'
import { Drawer } from './Drawer'

;(globalThis as unknown as { IS_REACT_ACT_ENVIRONMENT: boolean }).IS_REACT_ACT_ENVIRONMENT = true

function mount(ui: ReactNode) {
  const host = document.createElement('div')
  document.body.appendChild(host)
  const root = createRoot(host)
  act(() => { root.render(ui) })
  return {
    host,
    unmount() {
      act(() => root.unmount())
      host.remove()
    },
  }
}

function key(k: string, shiftKey = false) {
  act(() => {
    document.dispatchEvent(new KeyboardEvent('keydown', { key: k, shiftKey, bubbles: true }))
  })
}

function panels() {
  return Array.from(document.querySelectorAll<HTMLElement>('.ui-drawer-panel'))
}

function Stack() {
  const [bottom, setBottom] = useState(true)
  const [top, setTop] = useState(true)
  return (
    <>
      <Drawer open={bottom} onClose={() => setBottom(false)} title="Bas">
        <button type="button">Bas</button>
      </Drawer>
      <Drawer open={top} onClose={() => setTop(false)} title="Haut">
        <button type="button">Haut</button>
      </Drawer>
    </>
  )
}

beforeEach(() => {
  document.getElementById('shell')?.remove()
  document.querySelectorAll('.ui-drawer').forEach(n => n.remove())
})

afterEach(() => {
  document.getElementById('shell')?.remove()
  document.querySelectorAll('.ui-drawer').forEach(n => n.remove())
})

describe('Drawer — profondeur sans perte de contexte', () => {
  it('garde le focus dans le panneau (Tab et Maj-Tab bouclent)', () => {
    const m = mount(
      <Drawer open onClose={() => {}} title="Détails">
        <button type="button">A</button>
        <button type="button">B</button>
      </Drawer>,
    )
    const focusables = Array.from(panels()[0].querySelectorAll<HTMLElement>('button'))
    const last = focusables[focusables.length - 1]
    act(() => { last.focus() })
    expect(document.activeElement).toBe(last)

    key('Tab')
    expect(document.activeElement, 'Tab boucle sur le premier').toBe(focusables[0])
    key('Tab', true)
    expect(document.activeElement, 'Maj-Tab boucle sur le dernier').toBe(last)
    m.unmount()
  })

  it('Échap ne ferme que le tiroir du dessus, puis le suivant', () => {
    const m = mount(<Stack />)
    expect(panels().map(p => p.getAttribute('aria-label'))).toEqual(['Bas', 'Haut'])

    key('Escape')
    expect(panels().map(p => p.getAttribute('aria-label')), 'le tiroir du bas reste ouvert').toEqual(['Bas'])
    key('Escape')
    expect(panels(), 'les deux sont fermés').toHaveLength(0)
    m.unmount()
  })

  it('rend le focus au déclencheur à la fermeture', () => {
    const trigger = document.createElement('button')
    document.body.appendChild(trigger)
    act(() => { trigger.focus() })

    function One() {
      const [open, setOpen] = useState(true)
      return (
        <Drawer open={open} onClose={() => setOpen(false)} title="Seul">
          <button type="button">Dedans</button>
        </Drawer>
      )
    }
    const m = mount(<One />)
    act(() => { panels()[0].querySelector('button')!.focus() })
    expect(document.activeElement).not.toBe(trigger)

    key('Escape')
    expect(document.activeElement, 'focus rendu au déclencheur').toBe(trigger)
    m.unmount()
    trigger.remove()
  })

  it('rend #shell inerte jusqu’à la fermeture du dernier tiroir', () => {
    const shell = document.createElement('div')
    shell.id = 'shell'
    document.body.appendChild(shell)
    const m = mount(<Stack />)

    expect((shell as unknown as { inert: boolean }).inert).toBe(true)
    key('Escape')
    expect((shell as unknown as { inert: boolean }).inert, 'un tiroir reste ouvert').toBe(true)
    key('Escape')
    expect((shell as unknown as { inert: boolean }).inert, 'plus aucun tiroir').toBe(false)
    m.unmount()
    shell.remove()
  })

  it('ne couvre jamais la colonne décisionnelle au-delà de 1280px', () => {
    const m = mount(
      <Drawer open onClose={() => {}} title="Large" width="900px">
        <span>contenu</span>
      </Drawer>,
    )
    const panel = panels()[0]
    expect(panel.style.width).toBe('900px')
    expect(panel.style.maxWidth, 'largeur plafonnée hors colonne').toContain('100vw')
    expect(panel.style.maxWidth).toContain('var(--ui-spine-w')
    m.unmount()
  })
})
