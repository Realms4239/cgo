import { create } from 'zustand';
import type { LiveFrame } from '../lib/types';
import type { BenchStage } from '../lib/bench';
import { preloadMode } from '../lib/preload';

// Deux modes seulement : une bascule, clavier 1 et 2. Aucun identifiant
// historique ne survit — la palette, le rail et les raccourcis lisent MODES.
export const MODES = [
  { id: 'cockpit', label: 'Synthèse', key: '1' },
  { id: 'labo', label: 'Labo', key: '2' },
] as const;

export type ModeId = (typeof MODES)[number]['id'];

// La destination survit au rechargement : le reload rouvre le même des deux
// modes, sans tiers caché ni identifiant historique.
const MODE_KEY = 'meteolink-mode'
function initialMode(): ModeId {
  try {
    return typeof localStorage !== 'undefined' && localStorage.getItem(MODE_KEY) === 'labo' ? 'labo' : 'cockpit'
  } catch { return 'cockpit' }
}

type ToastKind = '' | 'ok' | 'err' | 'blue'
export interface ToastItem { id: number; msg: string; cls: ToastKind }
let toastSeq = 1

// Un minuteur par notification : le survol les suspend, la sortie les réarme.
const toastTimers = new Map<number, ReturnType<typeof setTimeout>>()
function armToast(id: number, ms: number) {
  const prev = toastTimers.get(id)
  if (prev) clearTimeout(prev)
  toastTimers.set(id, setTimeout(() => {
    toastTimers.delete(id)
    useUIStore.getState().dropToast(id)
  }, ms))
}

export type FlashType = 'success' | 'danger' | 'info'
export interface FlashItem { type: FlashType; msg: string }

// Le tiroir de preuve : deux lectures du même rapport scellé.
export type ProofTab = 'commands' | 'proof'

interface UIState {
  mode: ModeId; setMode: (m: ModeId) => void
  benchStage: BenchStage; setBenchStage: (s: BenchStage) => void
  live: LiveFrame | null; setLive: (f: LiveFrame) => void
  connected: boolean; setConnected: (v: boolean) => void
  sseStatus: string; setSseStatus: (s: string) => void
  toasts: ToastItem[]; pushToast: (msg: string, cls?: ToastKind) => void; dropToast: (id: number) => void
  pauseToasts: () => void; resumeToasts: () => void
  flash: FlashItem | null; setFlash: (f: FlashItem | null) => void
  replayRunning: boolean; replayRunId: string | null; setReplay: (running: boolean, id: string | null) => void
  proof: { tab: ProofTab; auditId?: string } | null
  openProof: (tab: ProofTab, auditId?: string) => void
  closeProof: () => void
}

export const useUIStore = create<UIState>((set) => ({
  mode: initialMode(), setMode: (mode) => {
    preloadMode(mode)
    try { if (typeof localStorage !== 'undefined') localStorage.setItem(MODE_KEY, mode) } catch { /* mode privé */ }
    set({ mode })
  },
  benchStage: 'config', setBenchStage: (benchStage) => set({ benchStage }),
  live: null, setLive: (live) => set({ live }),
  connected: false, setConnected: (connected) => set({ connected }),
  sseStatus: 'déconnecté', setSseStatus: (sseStatus) => set({ sseStatus }),
  toasts: [],
  pushToast: (msg, cls = '') => {
    const id = toastSeq++
    set((s) => ({ toasts: [...s.toasts, { id, msg, cls }] }))
    armToast(id, 3200)
  },
  dropToast: (id) => {
    const timer = toastTimers.get(id)
    if (timer) clearTimeout(timer)
    toastTimers.delete(id)
    set((s) => ({ toasts: s.toasts.filter((t) => t.id !== id) }))
  },
  pauseToasts: () => {
    toastTimers.forEach((timer) => clearTimeout(timer))
    toastTimers.clear()
  },
  // Reprise courte : le message a déjà été lu pendant le survol.
  resumeToasts: () => {
    for (const t of useUIStore.getState().toasts) armToast(t.id, 1200)
  },
  flash: null, setFlash: (flash) => set({ flash }),
  replayRunning: false, replayRunId: null, setReplay: (replayRunning, replayRunId) => set({ replayRunning, replayRunId }),
  // Un dossier d'archive peut être demandé par son identifiant : le tiroir
  // charge alors CE rapport, pas celui qui se trouve être en mémoire.
  proof: null,
  openProof: (tab, auditId) => set({ proof: { tab, auditId } }),
  closeProof: () => set({ proof: null }),
}));
