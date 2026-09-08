//go:build ignore

// Génère assets/logo.ico (+logo.png) : pixel-art programmatique, zéro
// dépendance — chaînette cyan (LIEN) sur fond nuit + point tricolore.
// Usage : go run ./tools/logo (écrit assets/, à commiter).
package main

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"fmt"
	"os"
)

const S = 48

type img struct{ px [S][S][3]byte }

func (m *img) set(x, y int, r, g, b byte) {
	if x < 0 || y < 0 || x >= S || y >= S {
		return
	}
	m.px[y][x] = [3]byte{r, g, b}
}

func (m *img) disc(cx, cy, r int, col [3]byte, fill bool) {
	for y := cy - r - 1; y <= cy+r+1; y++ {
		for x := cx - r - 1; x <= cx+r+1; x++ {
			dx, dy := x-cx, y-cy
			d := dx*dx + dy*dy
			if fill && d <= r*r {
				m.set(x, y, col[0], col[1], col[2])
			} else if !fill && d <= r*r && d >= (r-3)*(r-3) {
				m.set(x, y, col[0], col[1], col[2])
			}
		}
	}
}

func main() {
	var m img
	bg := [3]byte{0x0b, 0x0b, 0x0c}
	cyan := [3]byte{0x5a, 0xd3, 0xe3}
	red := [3]byte{0xe2, 0x27, 0x18}
	// fond nuit arrondi (coins adoucis)
	for y := 0; y < S; y++ {
		for x := 0; x < S; x++ {
			corner := (x < 8 && y < 8 && (8-x)*(8-x)+(8-y)*(8-y) > 64) ||
				(x >= S-8 && y < 8 && (x-(S-9))*(x-(S-9))+(8-y)*(8-y) > 64) ||
				(x < 8 && y >= S-8 && (8-x)*(8-x)+(y-(S-9))*(y-(S-9)) > 64) ||
				(x >= S-8 && y >= S-8 && (x-(S-9))*(x-(S-9))+(y-(S-9))*(y-(S-9)) > 64)
			if !corner {
				m.set(x, y, bg[0], bg[1], bg[2])
			}
		}
	}
	// deux anneaux entrelacés = le lien
	m.disc(19, 24, 10, cyan, false)
	m.disc(29, 24, 10, cyan, false)
	// point tricolore (identité)
	m.disc(38, 10, 3, red, true)

	// PNG minimal (RGBA, sans filtre) puis conteneur ICO
	var raw bytes.Buffer
	for y := 0; y < S; y++ {
		raw.WriteByte(0)
		for x := 0; x < S; x++ {
			p := m.px[y][x]
			// fond transparent hors arrondi
			if p == [3]byte{0, 0, 0} {
				raw.Write([]byte{0, 0, 0, 0})
			} else {
				raw.Write([]byte{p[0], p[1], p[2], 255}) // PNG = RGBA
			}
		}
	}
	var comp bytes.Buffer
	w := zlib.NewWriter(&comp)
	_, _ = w.Write(raw.Bytes())
	_ = w.Close()
	var ihdr bytes.Buffer
	binary.Write(&ihdr, binary.BigEndian, uint32(S))
	binary.Write(&ihdr, binary.BigEndian, uint32(S))
	ihdr.Write([]byte{8, 6, 0, 0, 0})
	png := func(tag string, b []byte) []byte {
		var o bytes.Buffer
		binary.Write(&o, binary.BigEndian, uint32(len(b)))
		o.WriteString(tag)
		o.Write(b)
		// crc32 IEEE (petit, sans table : bit à bit, 48x48 = rapide)
		crc := uint32(0xFFFFFFFF)
		for _, by := range append([]byte(tag), b...) {
			crc ^= uint32(by)
			for i := 0; i < 8; i++ {
				if crc&1 == 1 {
					crc = (crc >> 1) ^ 0xEDB88320
				} else {
					crc >>= 1
				}
			}
		}
		binary.Write(&o, binary.BigEndian, crc^0xFFFFFFFF)
		return o.Bytes()
	}
	var pngBytes bytes.Buffer
	pngBytes.Write([]byte{137, 80, 78, 71, 13, 10, 26, 10})
	pngBytes.Write(png("IHDR", ihdr.Bytes()))
	pngBytes.Write(png("IDAT", comp.Bytes()))
	pngBytes.Write(png("IEND", nil))

	os.MkdirAll("assets", 0755)
	os.WriteFile("assets/logo.png", pngBytes.Bytes(), 0644)
	// ICO : header + 1 entrée PNG-compressée (Vista+)
	var ico bytes.Buffer
	binary.Write(&ico, binary.LittleEndian, uint16(0))
	binary.Write(&ico, binary.LittleEndian, uint16(1))
	binary.Write(&ico, binary.LittleEndian, uint16(1))
	ico.Write([]byte{byte(S), byte(S), 0, 0})
	binary.Write(&ico, binary.LittleEndian, uint16(1))
	binary.Write(&ico, binary.LittleEndian, uint16(32))
	binary.Write(&ico, binary.LittleEndian, uint32(pngBytes.Len()))
	binary.Write(&ico, binary.LittleEndian, uint32(6+16))
	ico.Write(pngBytes.Bytes())
	os.WriteFile("assets/logo.ico", ico.Bytes(), 0644)
	fmt.Println("assets/logo.ico + logo.png OK")
}
