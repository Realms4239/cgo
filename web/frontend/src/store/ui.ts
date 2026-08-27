import { create } from 'zustand';
import type { LiveFrame } from '../lib/types';

export const PANELS = [
  { id: 'campagne', label: 'Campagne', key: '1', milestone: 1 },
  { id: 'live', label: 'Temps réel', key: '2', milestone: 1 },
  { id: 'resultats', label: 'Résultats', key: '3', milestone: 2 },
  { id: 'integrite', label: 'Intégrité', key: '4', milestone: 2 },
] as const;

export type PanelId = (typeof PANELS)[number]['id'];

export type Density = 'airy' | 'dense'

type ToastKind = '' | 'ok' | 'err' | 'blue'
export interface ToastItem { id: number; msg: string; cls: ToastKind }
let toastSeq = 1

export type FlashType = 'success' | 'danger' | 'info'
export interface FlashItem { type: FlashType; msg: string }

interface UIState {
  panel: PanelId; setPanel: (p: PanelId) => void
  live: LiveFrame | null; setLive: (f: LiveFrame) => void
  connected: boolean; setConnected: (v: boolean) => void
  sseStatus: string; setSseStatus: (s: string) => void
  toasts: ToastItem[]; pushToast: (msg: string, cls?: ToastKind) => void; dropToast: (id: number) => void
  flash: FlashItem | null; setFlash: (f: FlashItem | null) => void
  replayRunning: boolean; replayRunId: string | null; setReplay: (running: boolean, id: string | null) => void
  railPinned: boolean; density: Density; setRailPinned: (v: boolean) => void; setDensity: (d: Density) => void
}

const getRailPinned = () => { try { return typeof localStorage !== 'undefined' && localStorage.getItem('railPinned') === '1' } catch { return false } }
const getDensity = (): Density => { try { const v = typeof localStorage !== 'undefined' ? localStorage.getItem('density') as Density : null; return v === 'dense' ? 'dense' : 'airy' } catch { return 'airy' } }

export const useUIStore = create<UIState>((set) => ({
  panel: 'campagne', setPanel: (panel) => set({ panel }),
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
  railPinned: getRailPinned(), density: getDensity(),
  setRailPinned: (railPinned) => { try { localStorage.setItem('railPinned', railPinned ? '1' : '0') } catch {} ; set({ railPinned }) },
  setDensity: (density) => { try { localStorage.setItem('density', density) } catch {} ; set({ density }) },
}));
