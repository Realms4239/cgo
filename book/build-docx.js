// build-docx.js : assemble thesis-v1.docx à partir des modules de contenu.
// Contrat de style : Times New Roman 12 pt, interligne 1,5, justifié, alinéa 12,5 mm,
// marges 25 mm, légendes de tableaux AU-DESSUS et de figures EN-DESSOUS (italique centré),
// titres numérotés, table des matières et renvois automatiques (champs Word).
'use strict';

const fs = require('fs');
const path = require('path');
const {
  Document, Packer, Paragraph, TextRun, ImageRun,
  Table, TableRow, TableCell, TableOfContents, SimpleField, Bookmark,
  Footer, PageNumber, AlignmentType, WidthType, BorderStyle,
  ShadingType, VerticalAlign, HeadingLevel, NumberFormat,
} = require('docx');

// ---------------------------------------------------------------- constantes
const MM = (mm) => Math.round(mm * 56.6929);        // mm -> twips
const MARGIN = MM(25);                               // 25 mm
const INDENT = MM(12.5);                             // alinéa 12,5 mm
const LINE_150 = 360;                                // interligne 1,5
const FONT = 'Times New Roman';
const SIZE_BODY = 24;                                // 12 pt (demi-points)
const SIZE_TABLE = 22;                               // 11 pt dans les tableaux
const SIZE_CAPTION = 22;                             // 11 pt légendes
const CONTENT_W_PX = 604;                            // 160 mm à 96 dpi
const FIG_DIR = path.join(__dirname, 'figures');
const OUT = path.join(__dirname, '..', 'thesis-v1.docx');

// ---------------------------------------------------------------- contenu
const MODULES = [
  '00-front', '01-intro', '015-organisme', '02-partie1', '03-partie2',
  '04-partie3', '05-partie4', '06-conclusion', '07-refs',
  '08-glossaire', '09-annexes',
].map((m) => require('./content/' + m + '.js'));

// ---------------------------------------------------------------- utilitaires
function pngSize(file) {
  const buf = fs.readFileSync(file);
  return { w: buf.readUInt32BE(16), h: buf.readUInt32BE(20) };
}

// Découpe un texte contenant {figRef:id}/{tabRef:id} en runs + champs REF.
const REF_RE = /\{(figRef|tabRef):([A-Za-z0-9]+)\}/g;
function textToRuns(text, opts = {}) {
  const runs = [];
  let last = 0, m;
  REF_RE.lastIndex = 0;
  while ((m = REF_RE.exec(text))) {
    if (m.index > last) runs.push(new TextRun({ text: text.slice(last, m.index), ...opts }));
    const label = m[1] === 'figRef' ? 'Figure' : 'Tableau';
    const mark = (m[1] === 'figRef' ? 'RefFig_' : 'RefTab_') + m[2];
    runs.push(new TextRun({ text: label + ' ', ...opts }));
    runs.push(new SimpleField(`REF ${mark} \\h`, '0'));
    last = m.index + m[0].length;
  }
  if (last < text.length) runs.push(new TextRun({ text: text.slice(last), ...opts }));
  return runs;
}

function para(text, opts = {}) {
  return new Paragraph({
    alignment: AlignmentType.JUSTIFIED,
    spacing: { line: LINE_150, after: 0 },
    indent: opts.noIndent ? undefined : { firstLine: INDENT },
    children: textToRuns(text, opts.run || {}),
    ...opts.par,
  });
}

// ---------------------------------------------------------------- titres numérotés
let partNo = 0, h2No = 0, h3No = 0;

function heading0(text) {
  return new Paragraph({
    heading: HeadingLevel.HEADING_1,
    pageBreakBefore: true,
    alignment: AlignmentType.CENTER,
    spacing: { before: 0, after: 360, line: LINE_150 },
    children: [new TextRun({ text })],
  });
}
function heading1(text) {
  partNo += 1; h2No = 0; h3No = 0;
  return new Paragraph({
    heading: HeadingLevel.HEADING_1,
    pageBreakBefore: true,
    alignment: AlignmentType.CENTER,
    spacing: { before: 0, after: 360, line: LINE_150 },
    children: [new TextRun({ text })],
  });
}
function heading2(text) {
  h2No += 1; h3No = 0;
  return new Paragraph({
    heading: HeadingLevel.HEADING_2,
    keepNext: true,
    spacing: { before: 360, after: 200, line: LINE_150 },
    children: [new TextRun({ text: `${partNo}.${h2No} ${text}` })],
  });
}
function heading2annex(text) {
  return new Paragraph({
    heading: HeadingLevel.HEADING_2,
    keepNext: true,
    spacing: { before: 360, after: 200, line: LINE_150 },
    children: [new TextRun({ text })],
  });
}
function heading3(text) {
  h3No += 1;
  return new Paragraph({
    heading: HeadingLevel.HEADING_3,
    keepNext: true,
    spacing: { before: 280, after: 160, line: LINE_150 },
    children: [new TextRun({ text: `${partNo}.${h2No}.${h3No} ${text}` })],
  });
}

// ---------------------------------------------------------------- figures
function figureBlock(fig) {
  const file = path.join(FIG_DIR, fig.file + '.png');
  const { w, h } = pngSize(file);
  const scale = Math.min(CONTENT_W_PX / w, 1) * (fig.screenshot ? 0.92 : 1);
  const out = [];
  out.push(new Paragraph({
    alignment: AlignmentType.CENTER,
    keepNext: true,
    spacing: { before: 240, after: 60, line: 240 },
    children: [new ImageRun({
      type: 'png',
      data: fs.readFileSync(file),
      transformation: { width: Math.round(w * scale), height: Math.round(h * scale) },
    })],
  }));
  // Légende EN-DESSOUS, italique, centrée : "Figure N – ..."
  out.push(new Paragraph({
    alignment: AlignmentType.CENTER,
    spacing: { before: 0, after: 240, line: 276 },
    children: [
      new TextRun({ text: 'Figure ', italics: true, size: SIZE_CAPTION }),
      new Bookmark({
        id: 'RefFig_' + fig.id,
        children: [new SimpleField('SEQ Figure \\* ARABIC', '0')],
      }),
      new TextRun({ text: ' \u2013 ' + fig.caption, italics: true, size: SIZE_CAPTION }),
    ],
  }));
  return out;
}

// ---------------------------------------------------------------- tableaux
const CELL_BORDER = { style: BorderStyle.SINGLE, size: 4, color: '7F7F7F' };
const BORDERS = {
  top: CELL_BORDER, bottom: CELL_BORDER, left: CELL_BORDER, right: CELL_BORDER,
  insideHorizontal: CELL_BORDER, insideVertical: CELL_BORDER,
};

function tableBlock(tab) {
  const out = [];
  // Légende AU-DESSUS, italique, centrée : "Tableau N – ..."
  out.push(new Paragraph({
    alignment: AlignmentType.CENTER,
    keepNext: true,
    spacing: { before: 240, after: 60, line: 276 },
    children: [
      new TextRun({ text: 'Tableau ', italics: true, size: SIZE_CAPTION }),
      new Bookmark({
        id: 'RefTab_' + tab.id,
        children: [new SimpleField('SEQ Tableau \\* ARABIC', '0')],
      }),
      new TextRun({ text: ' \u2013 ' + tab.caption, italics: true, size: SIZE_CAPTION }),
    ],
  }));
  const widths = tab.widths || tab.header.map(() => Math.floor(100 / tab.header.length));
  const mkCell = (text, isHeader, wPct) => new TableCell({
    width: { size: wPct, type: WidthType.PERCENTAGE },
    verticalAlign: VerticalAlign.CENTER,
    shading: isHeader ? { type: ShadingType.CLEAR, fill: 'E7E6E6' } : undefined,
    margins: { top: 60, bottom: 60, left: 100, right: 100 },
    children: [new Paragraph({
      alignment: isHeader ? AlignmentType.CENTER : AlignmentType.LEFT,
      spacing: { line: 240, after: 0 },
      children: [new TextRun({ text: String(text), bold: isHeader, size: SIZE_TABLE })],
    })],
  });
  const rows = [
    new TableRow({
      tableHeader: true,
      cantSplit: true,
      children: tab.header.map((c, i) => mkCell(c, true, widths[i])),
    }),
    ...tab.rows.map((r) => new TableRow({
      cantSplit: true,
      children: r.map((c, i) => mkCell(c, false, widths[i])),
    })),
  ];
  out.push(new Table({
    width: { size: 100, type: WidthType.PERCENTAGE },
    borders: BORDERS,
    rows,
  }));
  out.push(new Paragraph({ spacing: { after: 240, line: 240 }, children: [] }));
  return out;
}

// ---------------------------------------------------------------- sigles / refs / glossaire
function siglesBlock(pairs) {
  const noBorder = { style: BorderStyle.NONE, size: 0, color: 'FFFFFF' };
  const none = {
    top: noBorder, bottom: noBorder, left: noBorder, right: noBorder,
    insideHorizontal: noBorder, insideVertical: noBorder,
  };
  return [new Table({
    width: { size: 100, type: WidthType.PERCENTAGE },
    borders: none,
    rows: pairs.map(([abbr, def]) => new TableRow({
      cantSplit: true,
      children: [
        new TableCell({
          width: { size: 18, type: WidthType.PERCENTAGE },
          margins: { top: 40, bottom: 40, left: 0, right: 100 },
          children: [new Paragraph({
            spacing: { line: 276, after: 0 },
            children: [new TextRun({ text: abbr })],
          })],
        }),
        new TableCell({
          width: { size: 82, type: WidthType.PERCENTAGE },
          margins: { top: 40, bottom: 40, left: 100, right: 0 },
          children: [new Paragraph({
            alignment: AlignmentType.JUSTIFIED,
            spacing: { line: 276, after: 0 },
            children: [new TextRun({ text: def })],
          })],
        }),
      ],
    })),
  })];
}

function refsBlock(entries) {
  return entries.map((e) => new Paragraph({
    alignment: AlignmentType.JUSTIFIED,
    spacing: { line: 276, after: 80 },
    indent: { left: MM(10), hanging: MM(10) },
    children: [new TextRun({ text: e })],
  }));
}

function glossaireBlock(pairs) {
  return pairs.map(([term, def]) => new Paragraph({
    alignment: AlignmentType.JUSTIFIED,
    spacing: { line: 276, after: 80 },
    indent: { left: MM(8), hanging: MM(8) },
    children: [
      new TextRun({ text: term + ' : ', italics: true }),
      new TextRun({ text: def }),
    ],
  }));
}

// ---------------------------------------------------------------- page de titre
function titlePage(tp) {
  const line = (text, { size = 24, bold = false, caps = false, before = 0, after = 120 } = {}) =>
    new Paragraph({
      alignment: AlignmentType.CENTER,
      spacing: { before, after, line: 276 },
      children: [new TextRun({ text, size, bold, allCaps: caps })],
    });
  const rule = () => new Paragraph({
    alignment: AlignmentType.CENTER,
    spacing: { before: 60, after: 60 },
    border: { bottom: { style: BorderStyle.SINGLE, size: 6, color: '000000' } },
    children: [],
  });
  const out = [];
  out.push(line(tp.university, { size: 32, bold: true, before: 200 }));
  out.push(line(tp.faculty, { size: 26, bold: true }));
  out.push(line(tp.mention, { size: 24, after: 500 }));
  out.push(line(tp.docType, { size: 28, bold: true, before: 600 }));
  out.push(line(tp.docSubtype, { size: 24, after: 700 }));
  out.push(rule());
  out.push(line(tp.title, { size: 32, bold: true, before: 300, after: 200 }));
  out.push(line(tp.subtitle, { size: 26, bold: false, after: 300 }));
  out.push(rule());
  out.push(line(tp.author, { size: 24, before: 700 }));
  out.push(line(tp.supervisor, { size: 24 }));
  out.push(line(tp.year, { size: 24, before: 400, after: 400 }));
  for (const j of tp.jury) out.push(line(j, { size: 22, after: 60 }));
  return out;
}

// ---------------------------------------------------------------- traduction des blocs
function renderBlocks(blocks) {
  const out = [];
  for (const b of blocks) {
    if (b.titlePage) continue;                       // rendue à part
    else if (b.h0) out.push(heading0(b.h0));
    else if (b.h1) out.push(heading1(b.h1));
    else if (b.h2) out.push(heading2(b.h2));
    else if (b.h2annex) out.push(heading2annex(b.h2annex));
    else if (b.h3) out.push(heading3(b.h3));
    else if (b.p) out.push(para(b.p));
    else if (b.pNoIndent) out.push(para(b.pNoIndent, { noIndent: true }));
    else if (b.keywords) out.push(para(b.keywords, { noIndent: true, run: { italics: true } }));
    else if (b.sigles) out.push(...siglesBlock(b.sigles));
    else if (b.refs) out.push(...refsBlock(b.refs));
    else if (b.glossaire) out.push(...glossaireBlock(b.glossaire));
    else if (b.fig) out.push(...figureBlock(b.fig));
    else if (b.table) out.push(...tableBlock(b.table));
    else if (b.pagebreak) out.push(new Paragraph({ pageBreakBefore: true, children: [] }));
    else throw new Error('Bloc inconnu : ' + JSON.stringify(Object.keys(b)));
  }
  return out;
}

// ---------------------------------------------------------------- assemblage
const front = MODULES[0];
const tpBlock = front.find((b) => b.titlePage).titlePage;
const frontRest = front.filter((b) => !b.titlePage);
const bodyBlocks = MODULES.slice(1).flat();

const frontChildren = renderBlocks(frontRest);

// Table des matières + listes des figures et tableaux (champs Word).
const tocChildren = [
  heading0('TABLE DES MATIÈRES'),
  new TableOfContents('Table des matières', {
    hyperlink: true,
    headingStyleRange: '1-3',
  }),
  heading0('LISTE DES FIGURES'),
  new Paragraph({ children: [new SimpleField('TOC \\h \\z \\c "Figure"', '')] }),
  heading0('LISTE DES TABLEAUX'),
  new Paragraph({ children: [new SimpleField('TOC \\h \\z \\c "Tableau"', '')] }),
];

const bodyChildren = renderBlocks(bodyBlocks);

const footer = () => new Footer({
  children: [new Paragraph({
    alignment: AlignmentType.CENTER,
    spacing: { line: 240 },
    children: [new TextRun({ size: 22, children: [PageNumber.CURRENT] })],
  })],
});

const pageBase = {
  size: { width: MM(210), height: MM(297) },
  margin: { top: MARGIN, right: MARGIN, bottom: MARGIN, left: MARGIN },
};

const doc = new Document({
  creator: 'CGO',
  title: tpBlock.title,
  description: 'Mémoire de fin d\u2019études, version 1',
  features: { updateFields: true },
  styles: {
    default: {
      document: { run: { font: FONT, size: SIZE_BODY, color: '000000' } },
    },
    paragraphStyles: [
      {
        id: 'Heading1', name: 'Heading 1', basedOn: 'Normal', next: 'Normal', quickFormat: true,
        run: { font: FONT, size: 32, bold: true, color: '000000' },
        paragraph: { spacing: { before: 0, after: 360, line: LINE_150 } },
      },
      {
        id: 'Heading2', name: 'Heading 2', basedOn: 'Normal', next: 'Normal', quickFormat: true,
        run: { font: FONT, size: 28, bold: true, color: '000000' },
        paragraph: { spacing: { before: 360, after: 200, line: LINE_150 } },
      },
      {
        id: 'Heading3', name: 'Heading 3', basedOn: 'Normal', next: 'Normal', quickFormat: true,
        run: { font: FONT, size: 24, bold: true, italics: true, color: '000000' },
        paragraph: { spacing: { before: 280, after: 160, line: LINE_150 } },
      },
    ],
  },
  sections: [
    { // Page de titre, sans numérotation
      properties: { page: pageBase, titlePage: true },
      headers: {},
      footers: {},
      children: titlePage(tpBlock),
    },
    { // Pages liminaires : numérotation en chiffres romains
      properties: {
        page: { ...pageBase, pageNumbers: { start: 1, formatType: NumberFormat.LOWER_ROMAN } },
      },
      footers: { default: footer() },
      children: [...frontChildren, ...tocChildren],
    },
    { // Corps : numérotation arabe repartant à 1
      properties: {
        page: { ...pageBase, pageNumbers: { start: 1, formatType: NumberFormat.DECIMAL } },
      },
      footers: { default: footer() },
      children: bodyChildren,
    },
  ],
});

Packer.toBuffer(doc).then((buf) => {
  fs.writeFileSync(OUT, buf);
  const kb = (buf.length / 1024).toFixed(0);
  console.log(`OK ${OUT} (${kb} Ko)`);
}).catch((err) => { console.error(err); process.exit(1); });
