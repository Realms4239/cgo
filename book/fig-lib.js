// Shared SVG helpers for thesis figures. Deterministic output, academic styling.
const PAL = {
  ink: '#1f2430',
  sub: '#5b6472',
  grid: '#e4e7ec',
  axis: '#9aa1ac',
  blue: '#2563eb',
  teal: '#0d9488',
  amber: '#d97706',
  red: '#dc2626',
  purple: '#7c3aed',
  slate: '#64748b',
  blueBg: '#eff4fe',
  tealBg: '#e6f7f5',
  amberBg: '#fdf3e3',
  redBg: '#fdeaea',
  purpleBg: '#f3eefe',
  grayBg: '#f4f6f8',
};
const FONT = "'Segoe UI', 'Helvetica Neue', Arial, sans-serif";

function esc(s) {
  return String(s).replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');
}

class Svg {
  constructor(w, h) {
    this.w = w; this.h = h;
    this.parts = [];
    this.defs = [];
  }
  raw(s) { this.parts.push(s); return this; }
  def(s) { this.defs.push(s); return this; }
  rect(x, y, w, h, o = {}) {
    const rx = o.rx != null ? o.rx : 0;
    this.parts.push(`<rect x="${x}" y="${y}" width="${w}" height="${h}" rx="${rx}" fill="${o.fill || 'none'}"${o.stroke ? ` stroke="${o.stroke}" stroke-width="${o.sw || 1.5}"` : ''}${o.dash ? ` stroke-dasharray="${o.dash}"` : ''}${o.op ? ` opacity="${o.op}"` : ''}/>`);
    return this;
  }
  line(x1, y1, x2, y2, o = {}) {
    this.parts.push(`<line x1="${x1}" y1="${y1}" x2="${x2}" y2="${y2}" stroke="${o.stroke || PAL.ink}" stroke-width="${o.sw || 1.5}"${o.dash ? ` stroke-dasharray="${o.dash}"` : ''}${o.marker ? ` marker-end="url(#${o.marker})"` : ''}${o.cap ? ` stroke-linecap="${o.cap}"` : ''}/>`);
    return this;
  }
  path(d, o = {}) {
    this.parts.push(`<path d="${d}" fill="${o.fill || 'none'}"${o.stroke ? ` stroke="${o.stroke}" stroke-width="${o.sw || 2}"` : ''}${o.dash ? ` stroke-dasharray="${o.dash}"` : ''}${o.op ? ` opacity="${o.op}"` : ''}${o.marker ? ` marker-end="url(#${o.marker})"` : ''} stroke-linejoin="round" stroke-linecap="round"/>`);
    return this;
  }
  circle(cx, cy, r, o = {}) {
    this.parts.push(`<circle cx="${cx}" cy="${cy}" r="${r}" fill="${o.fill || PAL.blue}"${o.stroke ? ` stroke="${o.stroke}" stroke-width="${o.sw || 1.5}"` : ''}${o.op ? ` opacity="${o.op}"` : ''}/>`);
    return this;
  }
  text(x, y, str, o = {}) {
    const anchor = o.anchor || 'start';
    const size = o.size || 13;
    const fill = o.fill || PAL.ink;
    const weight = o.weight || 'normal';
    const style = o.italic ? ' font-style="italic"' : '';
    const rot = o.rotate ? ` transform="rotate(${o.rotate} ${x} ${y})"` : '';
    const spacing = o.spacing ? ` letter-spacing="${o.spacing}"` : '';
    const family = o.mono ? "Consolas, 'Courier New', monospace" : FONT;
    this.parts.push(`<text x="${x}" y="${y}" font-family="${family}" font-size="${size}" fill="${fill}" font-weight="${weight}" text-anchor="${anchor}"${style}${rot}${spacing}>${esc(str)}</text>`);
    return this;
  }
  // multi-line centered text inside a box
  textLines(cx, cy, lines, o = {}) {
    const size = o.size || 12.5;
    const lh = o.lh || size * 1.25;
    const y0 = cy - ((lines.length - 1) * lh) / 2;
    lines.forEach((ln, i) => {
      this.text(cx, y0 + i * lh + size * 0.35, ln, { ...o, anchor: 'middle', size });
    });
    return this;
  }
  box(x, y, w, h, lines, o = {}) {
    this.rect(x, y, w, h, { rx: o.rx != null ? o.rx : 8, fill: o.fill || PAL.blueBg, stroke: o.stroke || PAL.blue, sw: o.sw || 1.6, dash: o.dash });
    if (lines && lines.length) {
      this.textLines(x + w / 2, y + h / 2 + (o.dy || 0), lines, { size: o.size || 12.5, fill: o.textFill || PAL.ink, weight: o.weight || '600', mono: o.mono, lh: o.lh });
    }
    return this;
  }
  arrowDef(id, color) {
    this.def(`<marker id="${id}" viewBox="0 0 10 10" refX="9" refY="5" markerWidth="7" markerHeight="7" orient="auto-start-reverse"><path d="M0 0 L10 5 L0 10 z" fill="${color}"/></marker>`);
    return this;
  }
  toString(title) {
    return `<?xml version="1.0" encoding="UTF-8"?>
<svg xmlns="http://www.w3.org/2000/svg" width="${this.w}" height="${this.h}" viewBox="0 0 ${this.w} ${this.h}">
<defs>${this.defs.join('')}</defs>
<rect width="${this.w}" height="${this.h}" fill="#ffffff"/>
${this.parts.join('\n')}
</svg>`;
  }
}

// ---- chart scaffolding ----
function linScale(d0, d1, r0, r1) {
  const f = v => r0 + ((v - d0) / (d1 - d0)) * (r1 - r0);
  f.ticks = n => {
    const step = niceStep((d1 - d0) / n);
    const t = [];
    for (let v = Math.ceil(d0 / step) * step; v <= d1 + 1e-9; v += step) t.push(Math.round(v * 1e6) / 1e6);
    return t;
  };
  return f;
}
function niceStep(raw) {
  const mag = Math.pow(10, Math.floor(Math.log10(raw)));
  const norm = raw / mag;
  const n = norm < 1.5 ? 1 : norm < 3.5 ? 2 : norm < 7.5 ? 5 : 10;
  return n * mag;
}
// draws axes + grid, returns {sx, sy}
function chartFrame(svg, m, xDom, yDom, o = {}) {
  const { left, top, right, bottom } = m; // pixel bounds of plot area
  const sx = linScale(xDom[0], xDom[1], left, right);
  const sy = linScale(yDom[0], yDom[1], bottom, top);
  const xt = o.xTicks || sx.ticks(o.nx || 6);
  const yt = o.yTicks || sy.ticks(o.ny || 5);
  for (const t of yt) {
    svg.line(left, sy(t), right, sy(t), { stroke: PAL.grid, sw: 1 });
    svg.text(left - 8, sy(t) + 4, o.yFmt ? o.yFmt(t) : String(t), { anchor: 'end', size: 11.5, fill: PAL.sub });
  }
  for (const t of xt) {
    if (o.xGrid) svg.line(sx(t), top, sx(t), bottom, { stroke: PAL.grid, sw: 1 });
    svg.text(sx(t), bottom + 18, o.xFmt ? o.xFmt(t) : String(t), { anchor: 'middle', size: 11.5, fill: PAL.sub });
  }
  svg.line(left, bottom, right, bottom, { stroke: PAL.axis, sw: 1.4 });
  svg.line(left, top, left, bottom, { stroke: PAL.axis, sw: 1.4 });
  if (o.xLabel) svg.text((left + right) / 2, bottom + 40, o.xLabel, { anchor: 'middle', size: 12.5, fill: PAL.ink, weight: '600' });
  if (o.yLabel) svg.text(left - 46, (top + bottom) / 2, o.yLabel, { anchor: 'middle', size: 12.5, fill: PAL.ink, weight: '600', rotate: -90 });
  return { sx, sy };
}
function legend(svg, x, y, entries, o = {}) {
  let cx = x;
  for (const [label, color, kind] of entries) {
    if (kind === 'line') {
      svg.line(cx, y - 4, cx + 22, y - 4, { stroke: color, sw: 2.5 });
    } else if (kind === 'dash') {
      svg.line(cx, y - 4, cx + 22, y - 4, { stroke: color, sw: 2.5, dash: '6 4' });
    } else {
      svg.rect(cx, y - 12, 14, 14, { fill: color, rx: 3 });
    }
    const tw = label.length * 6.6;
    svg.text(cx + (kind === 'line' || kind === 'dash' ? 28 : 20), y, label, { size: 12, fill: PAL.ink });
    cx += (kind === 'line' || kind === 'dash' ? 32 : 24) + tw + 18;
  }
}

module.exports = { PAL, FONT, Svg, linScale, chartFrame, legend, esc };
