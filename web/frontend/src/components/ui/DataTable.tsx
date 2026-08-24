import { useMemo, useState } from 'react';
export interface Col<T> {
  key: keyof T & string;
  label: string;
  align?: 'l' | 'r';
  render?: (row: T) => string;
}
export function DataTable<T extends object>({ cols, rows, initialSort, rowKey }: {
  cols: Col<T>[];
  rows: T[];
  initialSort?: { key: string; dir: 1 | -1 };
  rowKey: (r: T, i: number) => string;
}) {
  const [sort, setSort] = useState<{ key: string; dir: 1 | -1 } | null>(initialSort ?? null);
  const sorted = useMemo(() => {
    if (!sort) return rows;
    const col = cols.find((c) => c.key === sort.key);
    const val = (r: T) => (col?.render ? col.render(r) : String(r[sort.key as keyof T]));
    return [...rows].sort((a, b) => {
      const va = val(a), vb = val(b);
      const na = parseFloat(va), nb = parseFloat(vb);
      const cmp = Number.isFinite(na) && Number.isFinite(nb) && /^\d/.test(va) && /^\d/.test(vb)
        ? na - nb : va.localeCompare(vb);
      return cmp * sort.dir;
    });
  }, [rows, sort, cols]);
  const flip = (key: string) =>
    setSort((s) => (s?.key === key ? { key, dir: (s.dir * -1) as 1 | -1 } : { key, dir: 1 }));
  return (
    <div className="c-tbl-wrap">
      <table className="c-tbl">
        <thead>
          <tr>
            {cols.map((c) => (
              <th key={c.key} className={c.align === 'r' ? 'r' : 'l'} aria-sort={sort?.key === c.key ? (sort.dir === 1 ? 'ascending' : 'descending') : 'none'}>
                <button type="button" className="c-tbl-sort" onClick={() => flip(c.key)}>{c.label}</button>
              </th>
            ))}
          </tr>
        </thead>
        <tbody>
          {sorted.map((r, i) => (
            <tr key={rowKey(r, i)}>
              {cols.map((c) => <td key={c.key} className={c.align === 'r' ? 'r' : 'l'}>{c.render ? c.render(r) : String(r[c.key])}</td>)}
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
