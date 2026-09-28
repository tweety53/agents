package guard

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"hash/crc32"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

// rawPNG builds a single-IDAT PNG from unfiltered rows, for the colour types
// and bit depths image/png's encoder never writes (1-bit grey, grey+alpha).
func rawPNG(t *testing.T, w, h int, depth, colorType byte, rows [][]byte) []byte {
	t.Helper()
	var out bytes.Buffer
	out.WriteString("\x89PNG\r\n\x1a\n")
	chunk := func(typ string, data []byte) {
		_ = binary.Write(&out, binary.BigEndian, uint32(len(data)))
		body := append([]byte(typ), data...)
		out.Write(body)
		_ = binary.Write(&out, binary.BigEndian, crc32.ChecksumIEEE(body))
	}
	ihdr := make([]byte, 13)
	binary.BigEndian.PutUint32(ihdr[0:], uint32(w))
	binary.BigEndian.PutUint32(ihdr[4:], uint32(h))
	ihdr[8], ihdr[9] = depth, colorType
	chunk("IHDR", ihdr)
	var z bytes.Buffer
	zw := zlib.NewWriter(&z)
	for _, r := range rows {
		_, _ = zw.Write(append([]byte{0}, r...)) // filter type 0: none
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	chunk("IDAT", z.Bytes())
	chunk("IEND", nil)
	return out.Bytes()
}

func encodePNG(t *testing.T, im image.Image) []byte {
	t.Helper()
	var b bytes.Buffer
	if err := png.Encode(&b, im); err != nil {
		t.Fatal(err)
	}
	return b.Bytes()
}

func TestDecodeRGB(t *testing.T) {
	t.Parallel()

	nrgba := image.NewNRGBA(image.Rect(0, 0, 2, 1))
	nrgba.SetNRGBA(0, 0, color.NRGBA{200, 100, 50, 0})
	nrgba.SetNRGBA(1, 0, color.NRGBA{10, 20, 30, 128})

	rgba := image.NewRGBA(image.Rect(0, 0, 2, 1))
	rgba.SetRGBA(0, 0, color.RGBA{1, 2, 3, 255})
	rgba.SetRGBA(1, 0, color.RGBA{250, 128, 7, 255})

	gray := image.NewGray(image.Rect(0, 0, 2, 1))
	gray.SetGray(0, 0, color.Gray{7})
	gray.SetGray(1, 0, color.Gray{250})

	gray16 := image.NewGray16(image.Rect(0, 0, 2, 1))
	gray16.SetGray16(0, 0, color.Gray16{0x12ff})
	gray16.SetGray16(1, 0, color.Gray16{0xab01})

	nrgba64 := image.NewNRGBA64(image.Rect(0, 0, 2, 1))
	nrgba64.SetNRGBA64(0, 0, color.NRGBA64{0x10ff, 0x20ff, 0x30ff, 0})
	nrgba64.SetNRGBA64(1, 0, color.NRGBA64{0xa001, 0xb001, 0xc001, 0x8000})

	rgba64 := image.NewRGBA64(image.Rect(0, 0, 1, 1))
	rgba64.SetRGBA64(0, 0, color.RGBA64{0x01ff, 0x80ff, 0xfe00, 0xffff})

	palOpaque := image.NewPaletted(image.Rect(0, 0, 2, 1),
		color.Palette{color.NRGBA{9, 8, 7, 255}, color.NRGBA{90, 80, 70, 255}})
	palOpaque.SetColorIndex(1, 0, 1)

	palAlpha := image.NewPaletted(image.Rect(0, 0, 2, 1),
		color.Palette{color.NRGBA{200, 150, 100, 0}, color.NRGBA{60, 40, 20, 77}})
	palAlpha.SetColorIndex(1, 0, 1)

	cases := []struct {
		name string
		png  []byte
		want [][3]byte // row-major
	}{
		{"NRGBA alpha 0 and 128 keeps colour", encodePNG(t, nrgba), [][3]byte{{200, 100, 50}, {10, 20, 30}}},
		{"RGBA opaque", encodePNG(t, rgba), [][3]byte{{1, 2, 3}, {250, 128, 7}}},
		{"Gray", encodePNG(t, gray), [][3]byte{{7, 7, 7}, {250, 250, 250}}},
		{"Gray16 high byte", encodePNG(t, gray16), [][3]byte{{0x12, 0x12, 0x12}, {0xab, 0xab, 0xab}}},
		{"NRGBA64 high bytes", encodePNG(t, nrgba64), [][3]byte{{0x10, 0x20, 0x30}, {0xa0, 0xb0, 0xc0}}},
		{"RGBA64 opaque high bytes", encodePNG(t, rgba64), [][3]byte{{0x01, 0x80, 0xfe}}},
		{"Paletted opaque", encodePNG(t, palOpaque), [][3]byte{{9, 8, 7}, {90, 80, 70}}},
		{"Paletted translucent entry keeps RGB", encodePNG(t, palAlpha), [][3]byte{{200, 150, 100}, {60, 40, 20}}},
		{"1-bit grey", rawPNG(t, 3, 1, 1, 0, [][]byte{{0b10100000}}), [][3]byte{{255, 255, 255}, {0, 0, 0}, {255, 255, 255}}},
		{"grey+alpha 8-bit keeps grey", rawPNG(t, 2, 1, 8, 4, [][]byte{{90, 0, 180, 128}}), [][3]byte{{90, 90, 90}, {180, 180, 180}}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			p := filepath.Join(t.TempDir(), "in.png")
			if err := os.WriteFile(p, c.png, 0o644); err != nil {
				t.Fatal(err)
			}
			im, err := decodeRGB(p)
			if err != nil {
				t.Fatal(err)
			}
			if im.W*im.H != len(c.want) || len(im.Pix) != 3*len(c.want) {
				t.Fatalf("size %dx%d, %d bytes; want %d pixels", im.W, im.H, len(im.Pix), len(c.want))
			}
			for i, w := range c.want {
				if got := im.at(i%im.W, i/im.W); got != w {
					t.Errorf("pixel %d: got %v, want %v", i, got, w)
				}
			}
		})
	}

	t.Run("encodeRGB round-trips", func(t *testing.T) {
		t.Parallel()
		p := filepath.Join(t.TempDir(), "out.png")
		in := &rgbImage{W: 2, H: 1, Pix: []byte{1, 2, 3, 250, 128, 7}}
		if err := encodeRGB(p, in); err != nil {
			t.Fatal(err)
		}
		got, err := decodeRGB(p)
		if err != nil {
			t.Fatal(err)
		}
		if got.W != 2 || got.H != 1 || !bytes.Equal(got.Pix, in.Pix) {
			t.Fatalf("got %+v, want %+v", got, in)
		}
	})

	for name, body := range map[string][]byte{
		"truncated file": encodePNG(t, rgba)[:40],
		"non-PNG file":   []byte("GIF89a not a png"),
	} {
		t.Run(name+" errors", func(t *testing.T) {
			t.Parallel()
			p := filepath.Join(t.TempDir(), "bad.png")
			if err := os.WriteFile(p, body, 0o644); err != nil {
				t.Fatal(err)
			}
			if _, err := decodeRGB(p); err == nil {
				t.Fatal("decodeRGB succeeded, want an error")
			}
		})
	}
}
