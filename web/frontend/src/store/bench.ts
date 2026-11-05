// Réglages de campagne partagés entre Configurer et Observer — la matrice de
// progression lit exactement ce que le formulaire envoie.
import { create } from 'zustand'
import type { ChannelModel } from '../lib/canal'
import { loadSettings } from '../lib/settings'

const RESEARCH_KEY = 'meteolink-bench-research'

function readResearch(): boolean {
  try {
    return typeof localStorage !== 'undefined' && localStorage.getItem(RESEARCH_KEY) === 'true'
  } catch {
    return false
  }
}

function saveResearch(value: boolean) {
  try {
    localStorage.setItem(RESEARCH_KEY, String(value))
  } catch {
    // mode privé : le choix vaut pour la session
  }
}

interface BenchState {
  profiles: string[]
  reps: number
  flows: number
  channelModel: ChannelModel
  deadlineMs: number
  research: boolean
  setProfiles: (p: string[]) => void
  setReps: (n: number) => void
  setFlows: (n: number) => void
  setChannelModel: (m: ChannelModel) => void
  setDeadlineMs: (n: number) => void
  setResearch: (v: boolean) => void
}

export const useBenchStore = create<BenchState>(set => ({
  profiles: ['P2'],
  reps: 3,
  flows: 4,
  channelModel: 'clean',
  deadlineMs: loadSettings().deadlineMs,
  research: readResearch(),
  setProfiles: profiles => set({ profiles }),
  setReps: reps => set({ reps }),
  setFlows: flows => set({ flows }),
  setChannelModel: channelModel => set({ channelModel }),
  setDeadlineMs: deadlineMs => set({ deadlineMs }),
  setResearch: research => {
    saveResearch(research)
    set({ research })
  },
}))
