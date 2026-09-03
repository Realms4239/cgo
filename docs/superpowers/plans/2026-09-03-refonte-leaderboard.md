# Refonte leaderboard Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Refonte benchmark-grade (DeepSWE sombre) des 5 vues, contrats API existants inchangés, tout ajout backend additif.

**Architecture:** Backend d'abord (3 ajouts additifs testés en Go), puis les 4 vues React avec les tokens CSS existants, puis vérification complète (typecheck/vitest/build/Edge). Aucune nouvelle dépendance.

**Tech Stack:** Go 1.26 stdlib, React 19 + Vite + echarts (déjà là), vitest, Playwright + Edge système.

## Global Constraints

- Copie UI en français uniquement.
- Zéro nouvelle dépendance npm/Go ; bundle gz ≤ 320 KB (`node scripts/check-bundle.mjs` doit passer, budget actuel 315,8 KB).
- Sélecteurs e2e intacts : `[data-panel]`, `.data-table`, `[data-testid]`, `[data-metric]`, `.mono.muted`, `input[name="import-id"]`, `.banner`, `#v-live`, `#v-campagne`.
- `book/`, `thesis-v1.docx`, `shots-thesis/`, figures figR* : INTOUCHÉS.
- Commits `fix(frontend):` / `feat(frontend):` / `feat(api):` en français ; `gofmt -l` vide ; `docs/api.md` mise à jour pour chaque ajout.
- Chaque tâche TDD : test qui échoue d'abord (preuve rouge via `git stash push` du fichier implémenté quand pertinent).

---

### Task 1: Backend — `Group.Smallp95IQR`

**Files:**
- Modify: `pkg/results/results.go` (struct Group + Scan)
- Test: `pkg/results/results_test.go` (étendre TestScanReadsByName)

**Interfaces:**
- Consumes: `metrics.Summarize([]float64)` → `Summary{Median, IQRLow, IQRHigh}` (existe).
- Produces: `Group.Smallp95IQR [2]float64` (`json:"small_p95_iqr"`) consommé par Task 4.

- [ ] **Step 1: Write the failing test**

Dans `pkg/results/results_test.go`, ajouter à TestScanReadsByName après l'assertion cake :
```go
if cake.Smallp95IQR != [2]float64{209.2, 209.2} {
    t.Fatalf("cake Smallp95IQR = %v, want [209.2 209.2]", cake.Smallp95IQR)
}
```
(une seule ligne valid → IQR dégénéré mais présent ; le point est la présence du champ.)

- [ ] **Step 2: Run test to verify it fails**

Run: `go test -count=1 -run TestScanReadsByName ./pkg/results/`
Expected: FAIL (`undefined: Smallp95IQR` à la compilation = rouge valide).

- [ ] **Step 3: Write minimal implementation**

Dans `pkg/results/results.go`, struct Group après `Smallp95Median` :
```go
Smallp95Median float64   `json:"small_p95_median"`
Smallp95IQR    [2]float64 `json:"small_p95_iqr"`
```
Dans Scan, après `ss := metrics.Summarize(b.smalls)` :
```go
Smallp95Median: ss.Median, Smallp95IQR: [2]float64{ss.IQRLow, ss.IQRHigh},
```
(remplace `Smallp95Median: ss.Median` existant dans le littéral Group).

- [ ] **Step 4: Run test to verify it passes**

Run: `go test -count=1 ./pkg/results/`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add pkg/results/results.go pkg/results/results_test.go
git commit -m "feat(api): IQR small_p95 dans Group — l'incertitude s'affiche"
```

---

### Task 2: Backend — `breakdown` + `updated` dans `/api/integrity`

**Files:**
- Modify: `pkg/api/server.go` (handler `GET /api/integrity`)
- Test: `pkg/api/csvmapping_test.go` (nouveau TestIntegrityBreakdown)

**Interfaces:**
- Consumes: `results.Scan("data/runs", runID)` (existe, testé) ; `os.Stat` mtime.
- Produces: `{"breakdown": [{"run","rows","valid","quarantined"}], "updated": "<RFC3339>"}` consommé par Task 5.

- [ ] **Step 1: Write the failing test**

Dans `pkg/api/csvmapping_test.go` :
```go
// TestIntegrityBreakdown — détail par run + updated, triés du plus récent.
func TestIntegrityBreakdown(t *testing.T) {
	dir := chdirTemp(t)
	writeCSVRun(t, dir, "run-a", header18,
		"run-a,1,P2,cake,bbr,1,100,114,14,209.2,98.1,18.1,81,0,117288,12.10,0.0,valid",
		"run-a,2,P2,cake,bbr,2,100,115,15,210.0,97.0,18.0,82,0,117300,12.2,0.0,invalid",
	)
	writeCSVRun(t, dir, "run-b", header18,
		"run-b,1,P1,pfifo_fast,cubic,1,20,25,5,30.0,100,70.0,0,0,0,0,0,valid",
	)
	h := New(Deps{})
	srv := httptest.NewServer(h)
	defer srv.Close()
	resp, err := http.Get(srv.URL + "/api/integrity")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	s := string(body)
	for _, want := range []string{`"breakdown"`, `"run":"run-b"`, `"run":"run-a"`, `"rows":2`, `"quarantined":1`, `"updated":"`} {
		if !strings.Contains(s, want) {
			t.Fatalf("integrity missing %s in %s", want, s)
		}
	}
	if strings.Index(s, "run-b") > strings.Index(s, "run-a") {
		t.Fatalf("breakdown must list newest run first: %s", s)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test -count=1 -run TestIntegrityBreakdown ./pkg/api/`
Expected: FAIL (`breakdown` absent).

- [ ] **Step 3: Write minimal implementation**

Dans le handler `GET /api/integrity` de `pkg/api/server.go` : après la boucle
existante qui remplit `valid`/`quarantined`/`runIDs`, construire :
```go
type runBreak struct {
	Run         string `json:"run"`
	Rows        int    `json:"rows"`
	Valid       int    `json:"valid"`
	Quarantined int    `json:"quarantined"`
}
var breakdown []runBreak
for i := len(runIDs) - 1; i >= 0; i-- {
	gs, _ := results.Scan("data/runs", runIDs[i])
	b := runBreak{Run: runIDs[i]}
	for _, g := range gs {
		b.Rows += g.Count
		b.Quarantined += g.Quarantined
		b.Valid += g.Count - g.Quarantined
	}
	breakdown = append(breakdown, b)
}
```
`runIDs` est en ordre croissant (Glob trié ? NON — `filepath.Glob` trie déjà
par ordre lexical ; vérifier : si besoin `sort.Strings(runIDs)` après la
boucle. Le code actuel ne trie pas runIDs — ajouter `sort.Strings(runIDs)`
avant usage ; `sort` est déjà importé dans server.go).
`updated` : mtime du dernier csv :
```go
updated := ""
if files, _ := filepath.Glob(filepath.Join("data/runs", "*", "aqm_eval.csv")); len(files) > 0 {
	sort.Strings(files)
	if st, err := os.Stat(files[len(files)-1]); err == nil {
		updated = st.ModTime().UTC().Format(time.RFC3339)
	}
}
```
Ajouter `"breakdown": breakdown, "updated": updated` au `writeJSON` final
(remarque : si `breakdown` est nil, JSON rend `null` — initialiser
`breakdown := []runBreak{}` pour rendre `[]`).

- [ ] **Step 4: Run test to verify it passes**

Run: `go test -count=1 -run 'TestIntegrity|TestIntegrityBreakdown' ./pkg/api/`
Expected: PASS (l'ancien TestIntegrityProvenance doit rester vert).

- [ ] **Step 5: Commit**

```bash
git add pkg/api/server.go pkg/api/csvmapping_test.go
git commit -m "feat(api): breakdown par run + updated dans /api/integrity"
```

---

### Task 3: Backend — `GET /api/quarantine` + docs

**Files:**
- Modify: `pkg/api/server.go` (nouvelle route après `/api/run/rows`)
- Modify: `docs/api.md` (ligne Intégrité + champ breakdown/updated de la Task 2)
- Test: `pkg/api/csvmapping_test.go` (nouveau TestQuarantineRoute)

**Interfaces:**
- Consumes: `data/runs/<run>/quarantine.json` (écrit par `Writer.Freeze`,
  forme `[{"event_id","profile","qdisc","cc","gate_status"}]`).
- Produces: `{"quarantines": [{"run","event_id","profile","qdisc","cc","gate_status"}]}` ; `run` absent = tous runs.

- [ ] **Step 1: Write the failing test**

```go
// TestQuarantineRoute — quarantaine réelle depuis quarantine.json gelés.
func TestQuarantineRoute(t *testing.T) {
	dir := chdirTemp(t)
	for _, run := range []string{"run-a", "run-b"} {
		rd := filepath.Join(dir, "data", "runs", run)
		if err := os.MkdirAll(rd, 0755); err != nil {
			t.Fatal(err)
		}
	}
	os.WriteFile(filepath.Join(dir, "data", "runs", "run-a", "quarantine.json"),
		[]byte(`[{"event_id":3,"profile":"P2","qdisc":"cake","cc":"bbr","gate_status":"invalid"}]`), 0644)
	h := New(Deps{})
	srv := httptest.NewServer(h)
	defer srv.Close()
	get := func(q string) (int, string) {
		resp, err := http.Get(srv.URL + "/api/quarantine" + q)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		b, _ := io.ReadAll(resp.Body)
		return resp.StatusCode, string(b)
	}
	if code, s := get("?run=run-a"); code != 200 || !strings.Contains(s, `"event_id":3`) {
		t.Fatalf("run-a: %d %s", code, s)
	}
	if code, s := get(""); code != 200 || !strings.Contains(s, `"run":"run-a"`) {
		t.Fatalf("all runs: %d %s", code, s)
	}
	if code, _ := get("?run=../x"); code != 400 {
		t.Fatalf("traversal: status = %d, want 400", code)
	}
	if code, _ := get("?run=nope"); code != 404 {
		t.Fatalf("unknown run: status = %d, want 404", code)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test -count=1 -run TestQuarantineRoute ./pkg/api/`
Expected: FAIL (404 JSON du catch-all `/api/` — la route n'existe pas).

- [ ] **Step 3: Write minimal implementation**

Après le handler `GET /api/run/rows` dans `pkg/api/server.go` :
```go
// Quarantaine réelle — lignes quarantine.json gelées, tous runs si ?run= vide.
mux.HandleFunc("GET /api/quarantine", func(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("run")
	if strings.ContainsAny(id, `/\.`) {
		writeErr(w, r, "id de run invalide", http.StatusBadRequest)
		return
	}
	type qrow struct {
		Run     string `json:"run"`
		EventID int    `json:"event_id"`
		Profile string `json:"profile"`
		Qdisc   string `json:"qdisc"`
		CC      string `json:"cc"`
		Status  string `json:"gate_status"`
	}
	var out []qrow
	collect := func(run string) {
		b, err := os.ReadFile(filepath.Join("data", "runs", run, "quarantine.json"))
		if err != nil {
			return // aucun gel de quarantaine pour ce run : pas une erreur
		}
		var rows []qrow
		if err := json.Unmarshal(b, &rows); err != nil {
			return
		}
		for _, q := range rows {
			q.Run = run
			out = append(out, q)
		}
	}
	if id != "" {
		if st, err := os.Stat(filepath.Join("data", "runs", id)); err != nil || !st.IsDir() {
			writeErr(w, r, "run introuvable: "+id, http.StatusNotFound)
			return
		}
		collect(id)
	} else {
		runs, _ := filepath.Glob(filepath.Join("data", "runs", "*"))
		sort.Strings(runs)
		for i := len(runs) - 1; i >= 0; i-- {
			collect(filepath.Base(runs[i]))
		}
	}
	if out == nil {
		out = []qrow{}
	}
	writeJSON(w, map[string]any{"quarantines": out})
})
```
(`os`, `filepath`, `sort`, `json`, `strings` déjà importés dans server.go ;
la route exacte `/api/quarantine` ne collide pas avec le catch-all `/api/`
enregistré avant — même pattern que `/api/run/rows` qui fonctionne.)

Dans `docs/api.md`, table Intégrité, ajouter :
```md
| `/api/quarantine?run=` | GET | — | `{"quarantines": [{run, event_id, profile, qdisc, cc, gate_status}]}` — `run` vide = tous runs (récents d'abord), `[]` si aucune | `400` id invalide (anti-traversal) ; `404` run inconnu |
```
Et compléter la ligne `/api/integrity` : ajouter `breakdown: [{run, rows, valid, quarantined}] (récents d'abord), updated: <RFC3339 mtime dernier csv>`.

- [ ] **Step 4: Run test to verify it passes**

Run: `go test -count=1 ./pkg/api/`
Expected: PASS (tout le package, pas seulement le nouveau test).

- [ ] **Step 5: Commit**

```bash
git add pkg/api/server.go pkg/api/csvmapping_test.go docs/api.md
git commit -m "feat(api): GET /api/quarantine — quarantaine réelle depuis les gels"
```

---

### Task 4: Frontend — ResultatsView leaderboard

**Files:**
- Create: `web/frontend/src/lib/format.ts` (+ `web/frontend/src/lib/format.test.ts`)
- Create: `web/frontend/src/lib/gates.ts` (labels G0–G7 partagés)
- Modify: `web/frontend/src/views/ResultatsView.tsx`
- Modify: `web/frontend/src/views/CampagneView.tsx` (utiliser `GATE_LABELS`, supprimer `gateLabel` local)

**Interfaces:**
- Consumes: `/api/results[?run=]` (+`small_p95_iqr`, Task 1) ; `/api/replay/list` (existe) ; `/api/integrity` (`hash8`, existe) ; `/api/events` (`{events:[{ts,kind,msg}]`, existe).
- Produces: aucun (vue terminale). Sélecteurs préservés.

- [ ] **Step 1: Write the failing test** (`format.test.ts`)

```ts
import { fmtIQR } from './format'
import { describe, expect, it } from 'vitest'
describe('fmtIQR', () => {
  it('rend med [low–high] à 1 décimale', () => {
    expect(fmtIQR(209.25, [207.1, 212.4])).toBe('209.3 [207.1–212.4]')
  })
  it('sans IQR rend la médiane seule', () => {
    expect(fmtIQR(209.25, undefined)).toBe('209.3')
  })
})
```
Run: `npx vitest run src/lib/format.test.ts` → FAIL (module inexistant).

- [ ] **Step 2: Implement helpers**

`web/frontend/src/lib/format.ts` :
```ts
export function fmtIQR(med: number, iqr?: [number, number] | null): string {
  const m = med.toFixed(1)
  if (!iqr || iqr.length !== 2) return m
  return `${m} [${iqr[0].toFixed(1)}–${iqr[1].toFixed(1)}]`
}
```
`web/frontend/src/lib/gates.ts` :
```ts
export const GATE_LABELS = ['cible joignable','bulk démarré','sondes actives','latence plausible','débit cohérent','pas de doublon','baseline stable','CPU ok']
```

- [ ] **Step 3: Refonte ResultatsView** (éditions ciblées, garder tout le reste)

a) Type Group : ajouter `small_p95_iqr?: [number, number]`.
b) États : `const [runSel, setRunSel] = useState('')`, `const [runIds, setRunIds] = useState<string[]>([])`, `const [events, setEvents] = useState<{ts:string;kind:string;msg:string}[]>([])`, `const [showMethod, setShowMethod] = useState(false)`.
c) useEffect initial : ajouter `fetch('/api/replay/list')…setRunIds(j.runs||[])` et `fetch('/api/events')…setEvents((j.events||[]).slice(-20).reverse())`.
d) Fetch results : URL `/api/results${runSel ? `?run=${encodeURIComponent(runSel)}` : ''}`, dépendance `[runSel]` (ajouter runSel au tableau du premier useEffect ; garder les autres fetches inchangés).
e) Bandeau en tête (avant le constat) :
```tsx
<div className="card" style={{ display:'flex', alignItems:'center', gap:12, padding:'10px 14px', flexWrap:'wrap' }}>
  <span className="mono" style={{ fontSize:12, fontWeight:700 }}>Meteolink Leaderboard</span>
  <span className="mono muted" style={{ fontSize:11 }}>{safeGroups.length} groupes · hash {hash8}</span>
  <select value={runSel} onChange={e=>setRunSel(e.target.value)} style={{ marginLeft:'auto', background:'var(--surface-card)', color:'var(--text-body)', border:'1px solid var(--hairline)', padding:'6px 8px', fontFamily:'JetBrains Mono', fontSize:11 }}>
    <option value="">tous runs (gelés)</option>
    {runIds.map(id=><option key={id} value={id}>{id}</option>)}
  </select>
  <a className="btn btn-primary" href="/api/report/export?format=csv" download>Exporter CSV</a>
  <a className="btn" href="/api/report/export?format=md" download style={{ border:'1px solid var(--hairline)', padding:'7px 16px' }}>MD</a>
  <button className="btn" onClick={()=>setShowMethod(v=>!v)}>Méthode & limites</button>
</div>
{showMethod && (
  <div className="card"><div className="card-head">Méthode & limites</div>
    <p className="mono" style={{ fontSize:11, lineHeight:1.7, color:'#9aa3ad' }}>
      Médianes des lignes gelées non invalidées · portes G0–G7 ({GATE_LABELS.join(' · ')}) ·
      n = répétitions par cellule (n=1 : première limite — voir thèse § perspectives) ·
      provenance hash {hash8} depuis data/runs/*/aqm_eval.csv.
    </p>
  </div>
)}
{events.length > 0 && (
  <div className="card"><div className="card-head">Changelog — journal opérateur</div>
    <ul style={{ listStyle:'none', padding:0, margin:0 }}>
      {events.map((e,i)=><li key={i} className="mono" style={{ fontSize:11, padding:'3px 0', borderBottom:'1px solid var(--hairline-faint)' }}><span style={{ color:'#767b84' }}>{e.ts}</span> <span style={{ color:'#5ad3e3' }}>{e.kind}</span> {e.msg}</li>)}
    </ul>
  </div>
)}
```
f) Colonnes small/RTT : `{fmtIQR(g.small_p95_median, g.small_p95_iqr)}` dans la
cellule small (garder la barre + delta), `{fmtIQR(g.rtt_p95_median, g.rtt_p95_iqr)}`
dans la cellule RTT. Importer `fmtIQR` et `GATE_LABELS`.
g) CampagneView : remplacer `gateLabel(i)` par `GATE_LABELS[i]`, supprimer la
fonction locale `gateLabel`, importer depuis `../lib/gates`.

- [ ] **Step 4: Verify**

Run: `npm run typecheck && npx vitest run src/lib/format.test.ts`
Expected: PASS. Puis `npx vitest run` complet PASS.

- [ ] **Step 5: Commit**

```bash
git add web/frontend/src/lib/format.ts web/frontend/src/lib/format.test.ts web/frontend/src/lib/gates.ts web/frontend/src/views/ResultatsView.tsx web/frontend/src/views/CampagneView.tsx
git commit -m "feat(frontend): Resultats en leaderboard benchmark — versions, IQR, méthode, changelog"
```

---

### Task 5: Frontend — IntegriteView preuve

**Files:**
- Modify: `web/frontend/src/views/IntegriteView.tsx`

**Interfaces:**
- Consumes: `/api/integrity` (+`breakdown`, Task 2) ; `/api/quarantine` (Task 3) ; existants inchangés (`/api/replay/list`, `/api/results`, `/api/figures/regen`, `hardwareRecommendation`).

- [ ] **Step 1: Vérifier le contrat (pas de test vitest : vue branchée API)**

Exécuter le serveur local + `curl /api/integrity` et `/api/quarantine` pour
confirmer les formes (déjà couvert par Tasks 2–3). Cette tâche est visuelle :
la preuve de non-régression = specs Playwright existantes (Step 4).

- [ ] **Step 2: Refonte ciblée**

a) Type Integrity : ajouter `breakdown?: {run:string;rows:number;valid:number;quarantined:number}[]; updated?: string`.
b) État : `const [quar, setQuar] = useState<{run:string;event_id:number;profile:string;qdisc:string;cc:string;gate_status:string}[]>([])` ; charger dans `load()` : `fetch('/api/quarantine').then(r=>r.json()).then(j=>setQuar(j.quarantines||[])).catch(()=>{})`.
c) Bandeau preuve (remplace les 4 `div.kv`) :
```tsx
<div className="card" style={{ display:'flex', gap:16, alignItems:'baseline', flexWrap:'wrap', padding:'10px 14px' }}>
  <span className="mono" style={{ fontSize:12, fontWeight:700 }}>Preuve gelée</span>
  <span className="mono" data-testid="proof-runs" style={{ fontSize:11 }}>{data.runs} runs · {data.manifests} manifests</span>
  <span className="mono" style={{ fontSize:11, color:'var(--t-ok)' }}>{data.valid} valides</span>
  <span className="mono" style={{ fontSize:11, color:(data.quarantined||0)>0?'var(--t-danger)':'var(--text-muted)' }}>{data.quarantined} quarantaine</span>
  <span className="mono muted" style={{ fontSize:11, marginLeft:'auto' }}>maj {data.updated || '—'}</span>
</div>
```
d) Table des runs (remplace le `<ul>`) — triée par `breakdown` (déjà récent
d'abord côté serveur), repli sur `run_ids` :
```tsx
<div className="card"><div className="card-head">Runs archivés</div>
<table className="data-table" style={{ width:'100%', borderCollapse:'collapse', fontFamily:'var(--font-mono)', fontSize:12 }}>
<thead><tr style={{ color:'#c3c9d1', textAlign:'left', borderBottom:'1px solid var(--hairline)' }}>
<th style={{ padding:'6px 8px' }}>run</th><th>lignes</th><th>valides</th><th>quar.</th><th></th><th></th>
</tr></thead><tbody>
{(data.breakdown ?? (data.run_ids||[]).map(id=>({run:id,rows:0,valid:0,quarantined:0}))).map(b=>(
<tr key={b.run} style={{ borderBottom:'1px solid var(--hairline-faint)' }} onMouseEnter={e=>setPeek({rect:(e.currentTarget as HTMLElement).getBoundingClientRect(), run:b.run})} onMouseLeave={()=>setPeek(null)}>
<td style={{ padding:'6px 8px' }}>{b.run}</td><td>{b.rows}</td><td>{b.valid}</td><td>{b.quarantined}</td>
<td><a href={`/api/results?run=${b.run}`} target="_blank" rel="noreferrer" style={{ color:'var(--t-live)' }}>résultats</a></td>
<td><button className="btn btn-primary" onClick={()=>{ startReplay(b.run); setPanel('live') }} style={{ padding:'4px 10px', fontSize:11 }}>Rejouer</button></td>
</tr>))}
</tbody></table></div>
```
e) Table quarantaine (remplace le compteur seul) :
```tsx
<div className="card"><div className="card-head">Quarantaine — lignes invalidées par les portes</div>
{quar.length===0 ? <div style={{padding:'8px 0'}}><EmptyState kind="empty" hint="aucune mise en quarantaine (gate_status=valid)" /></div> :
<table className="data-table" style={{ width:'100%', borderCollapse:'collapse', fontFamily:'var(--font-mono)', fontSize:12 }}>
<thead><tr style={{ color:'#c3c9d1', textAlign:'left', borderBottom:'1px solid var(--hairline)' }}>
<th style={{ padding:'6px 8px' }}>run</th><th>événement</th><th>cellule</th><th>statut</th>
</tr></thead><tbody>
{quar.map((q,i)=><tr key={i} style={{ borderBottom:'1px solid var(--hairline-faint)' }}>
<td style={{ padding:'6px 8px' }}>{q.run}</td><td>#{q.event_id}</td><td>{q.profile}·{q.qdisc}·{q.cc}</td><td style={{ color:'#f4b400' }}>{q.gate_status}</td>
</tr>)}
</tbody></table>}
<div className="mono" style={{ fontSize:10, color:'#767b84', marginTop:8 }}>source: quarantine.json · gate_status != valid</div></div>
```
f) Figures : sous chaque `<img>`, ajouter `<div className="mono muted" style={{fontSize:10}}>hash {…}</div>` — réutiliser le hash8 du fetch integrity existant (`data.hash8` via `(data as any).hash8` comme le fait déjà la pastille provenance).
g) Le reste (recommandations, replay card, Provenance, peek) : inchangé.

- [ ] **Step 3: Verify**

Run: `npm run typecheck`
Expected: PASS.

- [ ] **Step 4: Commit**

```bash
git add web/frontend/src/views/IntegriteView.tsx
git commit -m "fix(frontend): Integrité en table de preuve — runs, quarantaine réelle, hash"
```

---

### Task 6: Frontend — Live resserré + Timeline honnête

**Files:**
- Modify: `web/frontend/src/views/LiveView.tsx`
- Modify: `web/frontend/src/views/CampagneView.tsx` (Timeline uniquement)

**Interfaces:**
- Consumes: `liveRing.phaseSince` de `../lib/live` (vérifier la forme exacte avant d'écrire — voir Step 1) ; `liveSnap.phase`, `liveSnap.phase_total_s` (existent).

- [ ] **Step 1: Inspecter (pas de code)**

Lire `web/frontend/src/lib/live.ts` : confirmer l'export `phaseSince` (objet
`{phase: timestamp_ms}` ?) et dans `LiveView.tsx` l'usage existant
(`live.phaseSince['charge']`). Lire le composant `Timeline` (props
`baselineStart, chargeStart, chargeEnd, recupEnd, currentPhase`).
Si `phaseSince` couvre baseline/charge/recup : brancher. Sinon : masquer la
Timeline (ne jamais simuler).

- [ ] **Step 2: Timeline honnête dans CampagneView**

Remplacer le bloc `useState(() => ({baselineStart: n-90000, …}))` simulé par :
```tsx
const phaseSince: Record<string, number> = (liveRing as any).phaseSince ?? {}
const hasPhases = phaseSince.baseline != null && phaseSince.charge != null
…
{hasData && hasPhases && <Timeline baselineStart={phaseSince.baseline} chargeStart={phaseSince.charge} chargeEnd={phaseSince.recup ?? Date.now()} recupEnd={(phaseSince.recup ?? Date.now()) + 15000} currentPhase={phase} />}
```
Adapter aux clés réelles trouvées au Step 1 (noms exacts des phases dans
`phaseSince`). Si `chargeEnd`/`recupEnd` indisponibles : passer `Date.now()`
pour la borne ouverte + commentaire `// borne ouverte : phase en cours`.

- [ ] **Step 3: Resserrer LiveView (éditions minimales)**

a) MetricCards : ne toucher ni aux `data-metric` ni aux valeurs — uniquement
le style du conteneur si besoin (ne rien casser des specs wall).
b) Journal : tronquer `msg` à 120 caractères (`e.msg.slice(0,120)`) pour la
densité ; garder kind coloré.
c) Ne PAS toucher : charts, bannière, contrôles shape/burst/watch, SSE.

- [ ] **Step 4: Verify**

Run: `npm run typecheck && npx vitest run`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add web/frontend/src/views/LiveView.tsx web/frontend/src/views/CampagneView.tsx
git commit -m "fix(frontend): Timeline sur phases réelles + journal densifié"
```

---

### Task 7: Frontend — table matrice en CampagneView

**Files:**
- Modify: `web/frontend/src/views/CampagneView.tsx`

**Interfaces:**
- Consumes: `live.event_id`, `live.running` (store, existent) ; `profiles`, `reps` (state local, existent) ; `ALL_QDISCS`, `ALL_CC` (constantes, existent).
- Ordre matrice (serveur `matrix.go`) : profils × [pfifo_fast, fq_codel, cake] × [cubic, bbr] × reps, event_id séquentiel dès 1.

- [ ] **Step 1: Implémenter la table matrice**

Après la carte « État — flux SSE », ajouter :
```tsx
{profiles.length > 0 && (
<div className="card"><div className="card-head">Matrice — état par cellule</div>
<table className="data-table" style={{ width:'100%', borderCollapse:'collapse', fontFamily:'var(--font-mono)', fontSize:11 }}>
<thead><tr style={{ color:'#c3c9d1', textAlign:'left', borderBottom:'1px solid var(--hairline)' }}>
<th style={{ padding:'6px 8px' }}>profil</th><th>file × CC</th><th>répétitions</th><th>état</th>
</tr></thead><tbody>
{profiles.flatMap((p, pi) => ALL_QDISCS.flatMap((q, qi) => ALL_CC.map((c, ci) => {
  const cellStart = ((pi * ALL_QDISCS.length + qi) * ALL_CC.length + ci) * reps + 1
  const cur = live?.event_id ?? 0
  const done = Math.max(0, Math.min(reps, cur - cellStart + (live?.running ? 0 : 0)))
```
ATTENTION — état correct sans `running` : `cur==0` (jamais démarré) → toutes
« en attente ». Sinon : `cellStart + reps - 1 < cur` → « terminée » ;
`cellStart <= cur && cur < cellStart + reps` → « en cours » ; sinon
« en attente ». `doneReps = clamp(cur - cellStart + 1, 0, reps)` quand
`cur >= cellStart`, en tenant compte que `cur` pointe la cellule en cours
(événements précédents gelés) :
```tsx
  const st = cur === 0 ? 'attente' : (cellStart + reps - 1 < cur ? 'terminée' : (cellStart <= cur ? 'en cours' : 'attente'))
  const doneReps = cur <= cellStart ? (st === 'en cours' ? 1 : 0) : Math.min(reps, cur - cellStart + 1)
```
Hmm — `cur` = event_id de la cellule EN COURS (push au début de RunEvent) :
cellules `cellStart+reps-1 < cur` terminées ; cellule contenant `cur` en cours
avec `doneReps = cur - cellStart` terminées + 1 en cours → afficher
`{doneReps}/{reps}` avec doneReps = `Math.max(0, Math.min(reps, cur - cellStart + (st==='en cours'?1:0)))`.
Simplifier : afficher `{Math.max(0, Math.min(reps, cur - cellStart + (st === 'en cours' ? 1 : 0)))}/{reps}`.
```tsx
  const n = Math.max(0, Math.min(reps, cur - cellStart + (st === 'en cours' ? 1 : 0)))
  const color = st === 'terminée' ? '#1fa348' : st === 'en cours' ? '#5ad3e3' : '#767b84'
  return (<tr key={`${p}/${q}/${c}`} style={{ borderBottom:'1px solid var(--hairline-faint)' }}>
    <td style={{ padding:'6px 8px' }}>{p}</td><td>{q} × {c}</td>
    <td style={{ fontVariantNumeric:'tabular-nums' }}>{n}/{reps}</td>
    <td style={{ color }}>{st}</td></tr>)
})))}
</tbody></table>
<div className="mono muted" style={{ fontSize:10, marginTop:6 }}>ordre serveur : profils × files × CC × répétitions · événement #{live?.event_id ?? '—'}/{live?.total_events ?? '—'}</div></div>
)}
```
Cas limite : reprise d'un run (`seen` saute des cellules, `event_id` avance
quand même — la table reste cohérente car `id` s'incrémente toujours).

- [ ] **Step 2: Verify**

Run: `npm run typecheck && npx vitest run`
Expected: PASS.

- [ ] **Step 3: Commit**

```bash
git add web/frontend/src/views/CampagneView.tsx
git commit -m "feat(frontend): table matrice — état par cellule depuis event_id"
```

---

### Task 8: Vérification finale + rebuild dist

**Files:** aucun (rebuild `web/frontend/dist/index.html` si le hash change).

- [ ] **Step 1: Backend complet**

Run: `go vet ./... && go build ./... && go test -count=1 ./...`
Expected: tout PASS, `gofmt -l cmd pkg internal` vide.

- [ ] **Step 2: Frontend complet**

Run: `npm run typecheck && npx vitest run && npm run build`
Expected: PASS + bundle sous budget (le script `check-bundle.mjs` l'affirme).

- [ ] **Step 3: e2e Edge sur données seedées**

  1. `go build -o <temp>/cgo.exe ./cmd/cgo` ; copier 3 runs avec manifests
     depuis `book/data/runs` vers `<temp>/data/runs` ; lancer
     `./cgo.exe --serve --addr 127.0.0.1:18099 --mode observe` (CWD = temp).
  2. Recréer `pw-edge.config.ts` (Edge système, baseURL 18099) + spec tour
     temporaire (5 vues + import + console) — NE PAS committer, supprimer après.
  3. Lancer le tour + `smoke, wall, chart-probe, compare-probe (AUDIT_BASE),
     archives` ; exiger 100 % PASS et zéro erreur console.
  4. Tuer le serveur, supprimer config + spec.

- [ ] **Step 4: Commits**

`git add web/frontend/dist/index.html` si modifié :
`git commit -m "chore(frontend): rebuild dist post-refonte"` (uniquement si le hash change).

- [ ] **Step 5: Revue diff**

`git diff --stat` : aucun fichier sous `book/`, `thesis-v1.docx`,
`shots-thesis/` ne doit apparaître.
