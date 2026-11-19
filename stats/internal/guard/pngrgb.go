package guard

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"os"
	"strconv"
	"strings"
	"syscall"
	"unicode"
	"unicode/utf8"
)

// rgbImage is an 8-bit RGB raster: 3 bytes per pixel, row-major.
type rgbImage struct {
	W, H int
	Pix  []byte
}

func (im *rgbImage) at(x, y int) [3]byte {
	i := 3 * (y*im.W + x)
	return [3]byte{im.Pix[i], im.Pix[i+1], im.Pix[i+2]}
}

// decodeRGB reads a PNG as Pillow's convert("RGB") yields it: each pixel's
// stored colour channels, alpha dropped -- never composited or premultiplied,
// so a fully transparent pixel keeps its colour -- grey and palette expanded,
// a palette entry's tRNS alpha ignored. A 16-bit channel takes its high byte
// (design.md png-rgb-decode; Pillow itself clips 16-bit grey instead).
//
// color.RGBAModel is not used: it premultiplies, and a transparent pixel
// would read black. image/png returns *image.RGBA and *image.RGBA64 only for
// opaque data, where the premultiplied Pix equals the stored channels.
func decodeRGB(path string) (*rgbImage, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	src, err := png.Decode(f)
	if err != nil {
		return nil, err
	}
	b := src.Bounds()
	im := &rgbImage{W: b.Dx(), H: b.Dy(), Pix: make([]byte, 0, 3*b.Dx()*b.Dy())}
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			im.Pix = append(im.Pix, storedRGB(src, x, y)...)
		}
	}
	return im, nil
}

// storedRGB is one pixel's stored colour channels, 8 bits each.
func storedRGB(src image.Image, x, y int) []byte {
	switch s := src.(type) {
	case *image.NRGBA:
		i := s.PixOffset(x, y)
		return s.Pix[i : i+3]
	case *image.RGBA:
		i := s.PixOffset(x, y)
		return s.Pix[i : i+3]
	case *image.Gray:
		g := s.Pix[s.PixOffset(x, y)]
		return []byte{g, g, g}
	case *image.Gray16:
		g := s.Pix[s.PixOffset(x, y)] // big-endian: the high byte comes first
		return []byte{g, g, g}
	case *image.NRGBA64:
		i := s.PixOffset(x, y)
		return []byte{s.Pix[i], s.Pix[i+2], s.Pix[i+4]}
	case *image.RGBA64:
		i := s.PixOffset(x, y)
		return []byte{s.Pix[i], s.Pix[i+2], s.Pix[i+4]}
	case *image.Paletted:
		c := color.NRGBAModel.Convert(s.Palette[s.ColorIndexAt(x, y)]).(color.NRGBA)
		return []byte{c.R, c.G, c.B}
	}
	c := color.NRGBAModel.Convert(src.At(x, y)).(color.NRGBA)
	return []byte{c.R, c.G, c.B}
}

// encodeRGB writes im as an opaque PNG.
func encodeRGB(path string, im *rgbImage) error {
	out := image.NewRGBA(image.Rect(0, 0, im.W, im.H))
	for i := 0; i < im.W*im.H; i++ {
		copy(out.Pix[4*i:], im.Pix[3*i:3*i+3])
		out.Pix[4*i+3] = 255
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	if err := png.Encode(f, out); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}

// pilImageError is the text of the exception Pillow's Image.open(name).load()
// raised for the image at file, one helper for both image ports: an
// OSError's own text, Pillow's UnidentifiedImageError for a file not
// starting with the PNG signature — Pillow opened other formats, the ports
// read PNG only (design.md png-rgb-decode) — otherwise the Go decoder's
// error.
func pilImageError(err error, file, name string) string {
	// Python's open() refuses a NUL in a path with ValueError before any
	// system call; Go's reaches the kernel's EINVAL.
	if strings.IndexByte(name, 0) >= 0 {
		return "embedded null byte"
	}
	var errno syscall.Errno
	if errors.As(err, &errno) {
		return ppOSError(err, name)
	}
	if f, ferr := os.Open(file); ferr == nil {
		sig := make([]byte, 8)
		n, _ := io.ReadFull(f, sig)
		f.Close()
		if !bytes.Equal(sig[:n], []byte("\x89PNG\r\n\x1a\n")) {
			return "cannot identify image file " + ppRepr(name)
		}
	}
	return err.Error()
}

// ppIsWord is Python's Unicode `\w`: str.isalnum() or the underscore, read
// against Go's Unicode tables (unicode.Version) rather than the host Python's:
// a character a newer Unicode version made a letter or number is a word
// character to Python and not to Go, so the two can differ on those, and
// only ever by Go reporting a claim Python would not.
func ppIsWord(r rune) bool { return r == '_' || unicode.IsLetter(r) || unicode.IsNumber(r) }

// ppStrerror is `exc.strerror or exc`: the C library's text for the errno,
// which Go's own tables carry with the first letter lowered.
func ppStrerror(err error) string {
	var errno syscall.Errno
	if !errors.As(err, &errno) {
		return err.Error()
	}
	msg := errno.Error()
	r, n := utf8.DecodeRuneInString(msg)
	return string(unicode.ToUpper(r)) + msg[n:]
}

// ppOSError is str(exc) for an OSError raised on path:
// `[Errno N] <strerror>: '<path>'`.
func ppOSError(err error, path string) string {
	var errno syscall.Errno
	if !errors.As(err, &errno) {
		return err.Error()
	}
	return fmt.Sprintf("[Errno %d] %s: %s", int(errno), ppStrerror(err), ppRepr(path))
}

// ppRepr is Python's repr() of a str: single quotes unless the text holds a
// single quote and no double quote; backslash and quote escaped, and every
// character str.isprintable() rejects escaped as \xNN, \uNNNN or
// \UNNNNNNNN. strconv.IsPrint is the same category test (letters, marks,
// numbers, punctuation, symbols and the ASCII space), read against Go's
// Unicode tables rather than the host Python's. A byte that is not UTF-8 is
// the surrogate os.fsdecode made of it, \udcNN.
func ppRepr(s string) string {
	q := byte('\'')
	if strings.Contains(s, "'") && !strings.Contains(s, "\"") {
		q = '"'
	}
	var b strings.Builder
	b.WriteByte(q)
	for i := 0; i < len(s); {
		r, size := utf8.DecodeRuneInString(s[i:])
		i += size
		switch {
		case r == utf8.RuneError && size == 1:
			fmt.Fprintf(&b, `\udc%02x`, s[i-1])
		case r == '\\' || r == rune(q):
			b.WriteByte('\\')
			b.WriteRune(r)
		case r == '\n':
			b.WriteString(`\n`)
		case r == '\r':
			b.WriteString(`\r`)
		case r == '\t':
			b.WriteString(`\t`)
		case strconv.IsPrint(r):
			b.WriteRune(r)
		case r < 0x100:
			fmt.Fprintf(&b, `\x%02x`, r)
		case r < 0x10000:
			fmt.Fprintf(&b, `\u%04x`, r)
		default:
			fmt.Fprintf(&b, `\U%08x`, r)
		}
	}
	b.WriteByte(q)
	return b.String()
}
