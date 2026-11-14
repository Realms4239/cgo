// @vitest-environment happy-dom
import { renderToStaticMarkup } from 'react-dom/server'
import { describe, expect, it } from 'vitest'
import PhaseStepper from './PhaseStepper'

describe('PhaseStepper', () => {
  it('does not mark an unmeasured phase complete', () => {
    const host = document.createElement('div')
    host.innerHTML = renderToStaticMarkup(<PhaseStepper phase="charge" />)

    const repos = host.querySelector('[aria-label="Repos"]')?.closest('li')
    const charge = host.querySelector('[aria-label="Charge"]')?.closest('li')

    expect(repos?.getAttribute('data-state')).toBe('done')
    expect(charge?.getAttribute('aria-current')).toBe('step')
  })

  it('renders a failed phase without marking it complete', () => {
    const host = document.createElement('div')
    host.innerHTML = renderToStaticMarkup(<PhaseStepper phase="vidange" error="lecture impossible" />)

    expect(host.querySelector('[role="status"]')?.textContent).toBe('lecture impossible')
    expect(host.querySelector('[aria-label="Vidange"]')?.closest('li')?.getAttribute('aria-current')).toBe('step')
  })
})
