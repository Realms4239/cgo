// coalesce — latest-wins, fenêtre 20 ms. Le store React ne re-rend pas
// à chaque frame du flux : il prend la dernière valeur vue dans la fenêtre.
export function createCoalescer<A extends unknown[]>(fn: (...args: A) => void, windowMs = 20) {
  let pending: A | null = null
  let timer: ReturnType<typeof setTimeout> | null = null

  const flush = () => {
    timer = null
    if (pending === null) return
    const args = pending
    pending = null
    fn(...args)
  }
  const schedule = (...args: A) => {
    pending = args
    if (timer === null) timer = setTimeout(flush, windowMs)
  }
  const cancel = () => {
    if (timer !== null) clearTimeout(timer)
    timer = null
    pending = null
  }
  const pendingArgs = () => pending
  return { schedule, cancel, flush, pendingArgs, windowMs }
}
