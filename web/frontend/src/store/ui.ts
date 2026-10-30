import { create } from 'zustand';
import type { LiveFrame } from '../lib/types';
import type { BenchStage } from '../lib/bench';
import { preloadMode } from '../lib/preload';

// Une seule taxonomie de navigation : quatre destinations, aucune imbriquée.
export const MODES = [
  { id: 'dashboard', label: 'Tableau de bord', key: '1' },
  { id: 'audit', label: 'Auditer', key: '2' },
  { id: 'bench', label: "Banc d'essai", key: '3' },
  { id: 'evidence', label: 'Preuves & Rapports', key: '4' },
] as const;

export type ModeId = (typeof MODES)[number]['id'];

type ToastKind = '' | 'ok' | 'err' | 'blue'
export interface ToastItem { id: number; msg: string; cls: ToastKind }
let toastSeq = 1

export type FlashType = 'success' | 'danger' | 'info'
export interface FlashItem { type: FlashType; msg: string }

interface UIState {
  mode: ModeId; setMode: (m: ModeId) => void
  benchStage: BenchStage; setBenchStage: (s: BenchStage) => void
  live: LiveFrame | null; setLive: (f: LiveFrame) => void
  connected: boolean; setConnected: (v: boolean) => void
  sseStatus: string; setSseStatus: (s: string) => void
  toasts: ToastItem[]; pushToast: (msg: string, cls?: ToastKind) => void; dropToast: (id: number) => void
  flash: FlashItem | null; setFlash: (f: FlashItem | null) => void
  replayRunning: boolean; replayRunId: string | null; setReplay: (running: boolean, id: string | null) => void
  railPinned: boolean; setRailPinned: (v: boolean) => void
}

const railPinned = () => {
  try { return localStorage.getItem('mode-rail-pinned') !== '0' } catch { return true }
}

export const useUIStore = create<UIState>((set) => ({
  // le chunk de la destination part dès que l'opérateur demande le mode —
  // clavier, palette ou clic, pas seulement au survol du rail.
  mode: 'dashboard', setMode: (mode) => { preloadMode(mode); set({ mode }) },
  benchStage: 'config', setBenchStage: (benchStage) => set({ benchStage }),
  live: null, setLive: (live) => set({ live }),
  connected: false, setConnected: (connected) => set({ connected }),
  sseStatus: 'déconnecté', setSseStatus: (sseStatus) => set({ sseStatus }),
  toasts: [],
  pushToast: (msg, cls = '') => set((s) => {
    const id = toastSeq++
    setTimeout(() => useUIStore.getState().dropToast(id), 3200)
    return { toasts: [...s.toasts, { id, msg, cls }] }
  }),
  dropToast: (id) => set((s) => ({ toasts: s.toasts.filter((t) => t.id !== id) })),
  flash: null, setFlash: (flash) => set({ flash }),
  replayRunning: false, replayRunId: null, setReplay: (replayRunning, replayRunId) => set({ replayRunning, replayRunId }),
  railPinned: railPinned(),
  setRailPinned: (v) => { try { localStorage.setItem('mode-rail-pinned', v ? '1' : '0') } catch {} ; set({ railPinned: v }) },
}));
