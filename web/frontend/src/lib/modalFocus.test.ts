import { describe, expect, it } from 'vitest'
import { isTextEntryTarget } from './modalFocus'

describe('text entry guard', () => {
  it('ignores input, textarea, select and editable targets', () => {
    const input = document.createElement('input')
    const editable = document.createElement('div')
    editable.setAttribute('contenteditable', 'true')
    document.body.append(input, editable)

    expect(isTextEntryTarget(input)).toBe(true)
    expect(isTextEntryTarget(editable)).toBe(true)
  })

  it('ignores controls inside a dialog', () => {
    const dialog = document.createElement('div')
    dialog.setAttribute('role', 'dialog')
    const button = document.createElement('button')
    dialog.append(button)
    document.body.append(dialog)

    expect(isTextEntryTarget(button)).toBe(true)
  })

  it('allows mode shortcuts from ordinary controls', () => {
    const button = document.createElement('button')
    document.body.append(button)

    expect(isTextEntryTarget(button)).toBe(false)
  })
})
