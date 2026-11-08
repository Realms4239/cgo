// Contrôle des captures : présence, dimensions, unicité, et un contenu
// non uniforme — une capture plate se voit, sans OCR.
import { createHash } from 'node:crypto'
import { mkdirSync, readFileSync, readdirSync, statSync } from 'node:fs'
import { join, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import { inflateSync } from 'node:zlib'

const defaultRoot = fileURLToPath(new URL('../../../docs/img/shots/2026-11-05', import.meta.url))
const root = resolve(process.argv[2] ?? defaultRoot)
mkdirSync(root, { recursive: true })

function pngInfo(bytes) {
  const signature = [137, 80, 78, 71, 13, 10, 26, 10]
  if (signature.some((b, i) => bytes[i] !== b)) throw new Error('signature PNG invalide')
  let offset = 8
  let width = 0, height = 0, bitDepth = 0, colorType = 0, interlace = 0
  const idat = []
  while (offset < bytes.length) {
    const length = bytes.readUInt32BE(offset)
    const type = bytes.toString('ascii', offset + 4, offset + 8)
    const dataOffset = offset + 8
    const data = bytes.subarray(dataOffset, dataOffset + length)
    if (type === 'IHDR') {
      width = data.readUInt32BE(0); height = data.readUInt32BE(4)
      bitDepth = data[8]; colorType = data[9]; interlace = data[12]
    } else if (type === 'IDAT') idat.push(Buffer.from(data))
    else if (type === 'IEND') break
    offset = dataOffset + length + 4
  }
  return { width, height, bitDepth, colorType, interlace, idat: Buffer.concat(idat) }
}

function bppFor(colorType) {
  if (colorType === 6) return 4
  if (colorType === 2) return 3
  if (colorType === 4) return 2
  if (colorType === 0) return 1
  throw new Error(`type PNG non supporté: ${colorType}`)
}

function pixels(bytes) {
  const info = pngInfo(bytes)
  if (info.bitDepth !== 8 || info.interlace !== 0) throw new Error('seuls les PNG 8 bits non entrelacés sont supportés')
  const bpp = bppFor(info.colorType)
  const stride = info.width * bpp
  const raw = inflateSync(info.idat)
  const out = Buffer.alloc(info.height * stride)
  let src = 0
  let prev = Buffer.alloc(stride)
  for (let y = 0; y < info.height; y++) {
    const filter = raw[src++]
    const line = Buffer.alloc(stride)
    for (let i = 0; i < stride; i++) {
      const x = raw[src++]
      const a = i >= bpp ? line[i - bpp] : 0
      const b = prev[i]
      const c = i >= bpp ? prev[i - bpp] : 0
      let v = x
      if (filter === 1) v = x + a
      else if (filter === 2) v = x + b
      else if (filter === 3) v = x + Math.floor((a + b) / 2)
      else if (filter === 4) {
        const p = a + b - c, pa = Math.abs(p - a), pb = Math.abs(p - b), pc = Math.abs(p - c)
        v = x + (pa <= pb && pa <= pc ? a : pb <= pc ? b : c)
      } else if (filter !== 0) throw new Error(`filtre PNG inconnu: ${filter}`)
      line[i] = v & 0xff
    }
    line.copy(out, y * stride)
    prev = line
  }
  return { info, out, bpp, stride }
}

function contentStats(p) {
  const { info: { width: w, height: h }, out, bpp, stride } = p
  const total = w * h
  const counts = new Map()
  for (let y = 0; y < h; y++) {
    for (let x = 0; x < w; x++) {
      const i = y * stride + x * bpp
      const key = `${out[i]},${out[i + 1]},${out[i + 2]}`
      counts.set(key, (counts.get(key) ?? 0) + 1)
    }
  }
  const [dominant] = [...counts.values()].sort((a, b) => b - a)
  return { unique: counts.size, dominantShare: dominant / total }
}

const files = readdirSync(root).filter(f => f.endsWith('.png')).sort()
if (files.length < 12) throw new Error(`captures insuffisantes: ${files.length} < 12`)
const hashes = new Set()
for (const file of files) {
  const path = join(root, file)
  const bytes = readFileSync(path)
  if (statSync(path).size < 2000) throw new Error(`capture trop légère: ${file}`)
  const digest = createHash('sha256').update(bytes).digest('hex')
  if (hashes.has(digest)) throw new Error(`capture dupliquée: ${file}`)
  hashes.add(digest)
  const p = pixels(bytes)
  const match = file.match(/-(890|1280|1920)\.png$/)
  if (match && p.info.width !== Number(match[1])) {
    throw new Error(`largeur ${p.info.width} ≠ ${match[1]} dans ${file}`)
  }
  const stats = contentStats(p)
  if (stats.unique < 20 || stats.dominantShare > 0.995) {
    throw new Error(`capture trop uniforme: ${file} (${stats.unique} couleurs, dominante ${(stats.dominantShare * 100).toFixed(1)} %)`)
  }
  console.log(`${file}: ${p.info.width}×${p.info.height}, ${stats.unique} couleurs, dominante ${(stats.dominantShare * 100).toFixed(1)} %, sha256 ${digest.slice(0, 12)}`)
}
console.log(`${files.length} captures: dimensions, unicité et contenu vérifiés`)
