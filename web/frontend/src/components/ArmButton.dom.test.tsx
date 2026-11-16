// @vitest-environment happy-dom
import { describe, expect, it, vi } from 'vitest'
import { createRoot } from 'react-dom/client'
import { act } from 'react'
import { ArmButton } from './ArmButton'

function mount(node: React.ReactElement) {
  const host = document.createElement('div')
  document.body.appendChild(host)
  const root = createRoot(host)
  act(() => { root.render(node) })
  return { host, root }
}

describe('ArmButton — interactions réelles', () => {
  it('premier appui arme (libellé confirmer), second appui confirme une fois', () => {
    const onConfirm = vi.fn()
    const { host } = mount(<ArmButton label="DÉMARRER" confirmLabel="CONFIRMER ?" onConfirm={onConfirm} />)
    const btn = host.querySelector('button')!
    expect(btn.textContent).toBe('DÉMARRER')
    act(() => { (btn as HTMLButtonElement).click() })
    expect(host.querySelector('button')!.textContent).toBe('CONFIRMER ?')
    expect(onConfirm).not.toHaveBeenCalled()
    act(() => { (host.querySelector('button.btn-danger') as HTMLButtonElement).click() })
    expect(onConfirm).toHaveBeenCalledTimes(1)
    // après confirmation, composant revenu à l'état initial (DÉMARRER) — pas de re-confirmation fantôme
    expect(host.querySelector('button.btn-primary')).not.toBeNull()
  })

  it('a blur désarme — la confirmation ne part pas sans second appui', () => {
    const onConfirm = vi.fn()
    const { host } = mount(<ArmButton label="ARRÊTER" confirmLabel="CONFIRMER ?" onConfirm={onConfirm} />)
    const btn = host.querySelector('button')!
    act(() => { (btn as HTMLButtonElement).click() })
    expect(host.querySelector('button.btn-danger')).not.toBeNull()
    act(() => { host.querySelector('button.btn-danger')!.dispatchEvent(new FocusEvent('focusout', { bubbles: true })) })
    expect(host.querySelector('button.btn-danger')).toBeNull()
    expect(onConfirm).not.toHaveBeenCalled()
  })

  it('disabled refuse de s’armer', () => {
    const onConfirm = vi.fn()
    const { host } = mount(<ArmButton label="DÉMARRER" onConfirm={onConfirm} disabled />)
    act(() => { (host.querySelector('button') as HTMLButtonElement).click() })
    expect(host.querySelector('button.btn-danger')).toBeNull()
    expect(onConfirm).not.toHaveBeenCalled()
  })
})
