import { create } from 'zustand';

// The four LIEN panels (Partie III.IV) — nothing else exists.
export const PANELS = [
  { id: 'campagne', label: 'Campagne', key: '1', milestone: 1 },
  { id: 'live', label: 'Temps réel', key: '2', milestone: 1 },
  { id: 'resultats', label: 'Résultats', key: '3', milestone: 2 },
  { id: 'integrite', label: 'Intégrité', key: '4', milestone: 2 },
] as const;

export type PanelId = (typeof PANELS)[number]['id'];

interface UIState {
  panel: PanelId;
  setPanel: (p: PanelId) => void;
}

export const useUIStore = create<UIState>((set) => ({
  panel: 'campagne',
  setPanel: (panel) => set({ panel }),
}));
