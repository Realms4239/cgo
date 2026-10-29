// Réglages de campagne partagés entre Configurer et Observer — la matrice de
// progression lit exactement ce que le formulaire envoie.
import { create } from 'zustand'
import type { ChannelModel } from '../lib/canal'
import { loadSettings } from '../lib/settings'

interface BenchState {
  profiles: string[]
  reps: number
  flows: number
  channelModel: ChannelModel
  deadlineMs: number
  setProfiles: (p: string[]) => void
  setReps: (n: number) => void
  setFlows: (n: number) => void
  setChannelModel: (m: ChannelModel) => void
  setDeadlineMs: (n: number) => void
}

export const useBenchStore = create<BenchState>(set => ({
  profiles: ['P2'],
  reps: 3,
  flows: 4,
  channelModel: 'clean',
  deadlineMs: loadSettings().deadlineMs,
  setProfiles: profiles => set({ profiles }),
  setReps: reps => set({ reps }),
  setFlows: flows => set({ flows }),
  setChannelModel: channelModel => set({ channelModel }),
  setDeadlineMs: deadlineMs => set({ deadlineMs }),
}))
