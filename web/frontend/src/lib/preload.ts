// Préchauffe le chunk du banc au survol/focus de sa destination — le clic
// reste instantané sans charger ECharts pour un visiteur du tableau de bord.
export const preloadBench = () => import('../views/BenchView')
