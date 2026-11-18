package main

import (
	"bytes"
	"encoding/binary"
	"image"
	"image/color"
	"image/png"
	"log"
	"math"
	"os"
	"path/filepath"
)

const (
	iconBase    = 256.0
	supersample = 8
)

type point struct {
	x, y float64
}

type iconImage struct {
	size int
	data []byte
}

var (
	teal  = color.NRGBA{R: 15, G: 118, B: 110, A: 255}
	paper = color.NRGBA{R: 255, G: 255, B: 255, A: 255}
)

func main() {
	sizes := []int{16, 20, 24, 32, 40, 48, 64, 128, 256}
	images := make([]iconImage, 0, len(sizes))
	for _, size := range sizes {
		img := render(size)
		images = append(images, iconImage{size: size, data: encodeDIB(img)})
		if size == 256 {
			var buf bytes.Buffer
			if err := png.Encode(&buf, img); err != nil {
				log.Fatal(err)
			}
			if err := os.MkdirAll("assets", 0755); err != nil {
				log.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join("assets", "openmd-256.png"), buf.Bytes(), 0644); err != nil {
				log.Fatal(err)
			}
		}
		if size == 64 {
			var buf bytes.Buffer
			if err := png.Encode(&buf, img); err != nil {
				log.Fatal(err)
			}
			if err := os.MkdirAll(filepath.Join("internal", "render", "web"), 0755); err != nil {
				log.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join("internal", "render", "web", "favicon.png"), buf.Bytes(), 0644); err != nil {
				log.Fatal(err)
			}
		}
	}

	if err := os.MkdirAll("assets", 0755); err != nil {
		log.Fatal(err)
	}
	if err := writeICO(filepath.Join("assets", "openmd.ico"), images); err != nil {
		log.Fatal(err)
	}
}

func render(size int) *image.NRGBA {
	dim := size * supersample
	scale := float64(dim) / iconBase
	img := image.NewRGBA(image.Rect(0, 0, dim, dim))

	const border = 14.0
	fillRoundRect(img, 4*scale, 4*scale, 252*scale, 252*scale, 52*scale, teal)
	fillRoundRect(
		img,
		(4+border)*scale,
		(4+border)*scale,
		(252-border)*scale,
		(252-border)*scale,
		(52-border)*scale,
		paper,
	)

	m := []point{
		{78 * scale, 176 * scale},
		{78 * scale, 80 * scale},
		{128 * scale, 132 * scale},
		{178 * scale, 80 * scale},
		{178 * scale, 176 * scale},
	}
	strokePolyline(img, m, 30*scale, teal)

	return downsample(img, size)
}

func fillRoundRect(img *image.RGBA, x0, y0, x1, y1, radius float64, c color.NRGBA) {
	startX := int(math.Floor(x0))
	startY := int(math.Floor(y0))
	endX := int(math.Ceil(x1))
	endY := int(math.Ceil(y1))
	for y := startY; y < endY; y++ {
		for x := startX; x < endX; x++ {
			px := float64(x) + 0.5
			py := float64(y) + 0.5
			if insideRoundRect(px, py, x0, y0, x1, y1, radius) {
				img.Set(x, y, c)
			}
		}
	}
}

func insideRoundRect(px, py, x0, y0, x1, y1, radius float64) bool {
	if px < x0 || px > x1 || py < y0 || py > y1 {
		return false
	}
	cx := math.Max(x0+radius, math.Min(px, x1-radius))
	cy := math.Max(y0+radius, math.Min(py, y1-radius))
	dx := px - cx
	dy := py - cy
	return dx*dx+dy*dy <= radius*radius
}

func strokePolyline(img *image.RGBA, pts []point, width float64, c color.NRGBA) {
	radius := width / 2
	for i := 0; i < len(pts)-1; i++ {
		strokeSegment(img, pts[i], pts[i+1], radius, c)
	}
	for _, p := range pts {
		fillCircle(img, p.x, p.y, radius, c)
	}
}

func strokeSegment(img *image.RGBA, a, b point, radius float64, c color.NRGBA) {
	startX := int(math.Floor(math.Min(a.x, b.x) - radius))
	startY := int(math.Floor(math.Min(a.y, b.y) - radius))
	endX := int(math.Ceil(math.Max(a.x, b.x) + radius))
	endY := int(math.Ceil(math.Max(a.y, b.y) + radius))
	for y := startY; y < endY; y++ {
		for x := startX; x < endX; x++ {
			px := float64(x) + 0.5
			py := float64(y) + 0.5
			if distanceToSegment(px, py, a, b) <= radius {
				img.Set(x, y, c)
			}
		}
	}
}

func fillCircle(img *image.RGBA, cx, cy, radius float64, c color.NRGBA) {
	startX := int(math.Floor(cx - radius))
	startY := int(math.Floor(cy - radius))
	endX := int(math.Ceil(cx + radius))
	endY := int(math.Ceil(cy + radius))
	for y := startY; y < endY; y++ {
		for x := startX; x < endX; x++ {
			dx := float64(x) + 0.5 - cx
			dy := float64(y) + 0.5 - cy
			if dx*dx+dy*dy <= radius*radius {
				img.Set(x, y, c)
			}
		}
	}
}

func distanceToSegment(px, py float64, a, b point) float64 {
	dx := b.x - a.x
	dy := b.y - a.y
	if dx == 0 && dy == 0 {
		return math.Hypot(px-a.x, py-a.y)
	}
	t := ((px-a.x)*dx + (py-a.y)*dy) / (dx*dx + dy*dy)
	t = math.Max(0, math.Min(1, t))
	return math.Hypot(px-(a.x+t*dx), py-(a.y+t*dy))
}

func downsample(src *image.RGBA, size int) *image.NRGBA {
	dst := image.NewNRGBA(image.Rect(0, 0, size, size))
	factor := supersample
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			var sumR, sumG, sumB, sumA uint64
			for sy := 0; sy < factor; sy++ {
				for sx := 0; sx < factor; sx++ {
					r, g, b, a := src.At(x*factor+sx, y*factor+sy).RGBA()
					sumR += uint64(r)
					sumG += uint64(g)
					sumB += uint64(b)
					sumA += uint64(a)
				}
			}
			n := uint64(factor * factor)
			a := sumA / n
			var r, g, b uint64
			if sumA > 0 {
				r = sumR * 0xffff / sumA
				g = sumG * 0xffff / sumA
				b = sumB * 0xffff / sumA
			}
			dst.SetNRGBA(x, y, color.NRGBA{
				R: uint8(r >> 8),
				G: uint8(g >> 8),
				B: uint8(b >> 8),
				A: uint8(a >> 8),
			})
		}
	}
	return dst
}

func writeICO(path string, images []iconImage) error {
	var buf bytes.Buffer
	buf.Write([]byte{0, 0, 1, 0})
	if err := binary.Write(&buf, binary.LittleEndian, uint16(len(images))); err != nil {
		return err
	}

	offset := 6 + 16*len(images)
	for _, image := range images {
		width := byte(image.size)
		height := byte(image.size)
		if image.size == 256 {
			width = 0
			height = 0
		}
		buf.WriteByte(width)
		buf.WriteByte(height)
		buf.WriteByte(0)
		buf.WriteByte(0)
		_ = binary.Write(&buf, binary.LittleEndian, uint16(1))
		_ = binary.Write(&buf, binary.LittleEndian, uint16(32))
		_ = binary.Write(&buf, binary.LittleEndian, uint32(len(image.data)))
		_ = binary.Write(&buf, binary.LittleEndian, uint32(offset))
		offset += len(image.data)
	}
	for _, image := range images {
		buf.Write(image.data)
	}
	return os.WriteFile(path, buf.Bytes(), 0644)
}

func encodeDIB(img *image.NRGBA) []byte {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
	maskRowBytes := ((width + 31) / 32) * 4

	var buf bytes.Buffer
	_ = binary.Write(&buf, binary.LittleEndian, uint32(40))
	_ = binary.Write(&buf, binary.LittleEndian, int32(width))
	_ = binary.Write(&buf, binary.LittleEndian, int32(height*2))
	_ = binary.Write(&buf, binary.LittleEndian, uint16(1))
	_ = binary.Write(&buf, binary.LittleEndian, uint16(32))
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0))
	_ = binary.Write(&buf, binary.LittleEndian, uint32(width*height*4))
	_ = binary.Write(&buf, binary.LittleEndian, int32(0))
	_ = binary.Write(&buf, binary.LittleEndian, int32(0))
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0))
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0))

	for y := height - 1; y >= 0; y-- {
		for x := 0; x < width; x++ {
			c := img.NRGBAAt(x, y)
			buf.WriteByte(c.B)
			buf.WriteByte(c.G)
			buf.WriteByte(c.R)
			buf.WriteByte(c.A)
		}
	}
	buf.Write(make([]byte, maskRowBytes*height))
	return buf.Bytes()
}
