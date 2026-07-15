package webp

import (
	"bytes"
	"image"
	"image/color"
	"testing"

	"github.com/bnema/purego-webp/libwebp"
)

func testWebP(t testing.TB) ([]byte, *image.NRGBA) {
	t.Helper()
	img := image.NewNRGBA(image.Rect(0, 0, 3, 2))
	pixels := []color.NRGBA{
		{R: 0x10, G: 0x20, B: 0x30, A: 0x40},
		{R: 0x50, G: 0x60, B: 0x70, A: 0x80},
		{R: 0x90, G: 0xa0, B: 0xb0, A: 0xc0},
		{R: 0xd0, G: 0xe0, B: 0xf0, A: 0xff},
		{R: 0x01, G: 0x23, B: 0x45, A: 0x67},
		{R: 0x89, G: 0xab, B: 0xcd, A: 0xef},
	}
	for i, p := range pixels {
		img.SetNRGBA(i%3, i/3, p)
	}
	data, err := libwebp.WebPEncodeLosslessRGBA(img.Pix, 3, 2, img.Stride)
	if err != nil {
		t.Fatalf("encode fixture: %v", err)
	}
	return data, img
}

func TestDecodeFinalNRGBA(t *testing.T) {
	data, want := testWebP(t)
	gotImage, err := Decode(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("Decode() error = %v", err)
	}
	got, ok := gotImage.(*image.NRGBA)
	if !ok {
		t.Fatalf("Decode() type = %T, want *image.NRGBA", gotImage)
	}
	if got.Stride != want.Stride {
		t.Fatalf("Stride = %d, want %d", got.Stride, want.Stride)
	}
	if !bytes.Equal(got.Pix, want.Pix) {
		t.Fatalf("Pix = %x, want %x", got.Pix, want.Pix)
	}
}

func TestDecodeMalformedAndTruncated(t *testing.T) {
	if _, err := Decode(bytes.NewReader([]byte("not a webp"))); err == nil {
		t.Fatal("Decode(malformed) succeeded")
	}
	data, _ := testWebP(t)
	if _, err := Decode(bytes.NewReader(data[:len(data)/2])); err == nil {
		t.Fatal("Decode(truncated) succeeded")
	}
}

func TestDecodeNRGBALayout(t *testing.T) {
	stride, size, err := decodeNRGBALayout(3, 2)
	if err != nil || stride != 12 || size != 24 {
		t.Fatalf("decodeNRGBALayout(3, 2) = (%d, %d, %v), want (12, 24, nil)", stride, size, err)
	}
	for _, dimensions := range [][2]int{{0, 1}, {1, 0}, {-1, 1}, {1, -1}} {
		if _, _, err := decodeNRGBALayout(dimensions[0], dimensions[1]); err == nil {
			t.Fatalf("decodeNRGBALayout(%d, %d) accepted non-positive dimensions", dimensions[0], dimensions[1])
		}
	}
}

func TestDecodeNRGBALayoutLargeDoesNotAllocate(t *testing.T) {
	if allocations := testing.AllocsPerRun(100, func() {
		stride, size, err := decodeNRGBALayout(1<<20, 1<<10)
		if err != nil || stride != 1<<22 || size != 1<<32 {
			t.Fatalf("decodeNRGBALayout() = (%d, %d, %v)", stride, size, err)
		}
	}); allocations != 0 {
		t.Fatalf("decodeNRGBALayout allocated %v times", allocations)
	}
}

func BenchmarkDecode64x64(b *testing.B) {
	img := image.NewNRGBA(image.Rect(0, 0, 64, 64))
	for y := 0; y < 64; y++ {
		for x := 0; x < 64; x++ {
			img.SetNRGBA(x, y, color.NRGBA{R: uint8(x), G: uint8(y), B: uint8(x ^ y), A: uint8(255 - x)})
		}
	}
	data, err := libwebp.WebPEncodeLosslessRGBA(img.Pix, 64, 64, img.Stride)
	if err != nil {
		b.Fatalf("encode fixture: %v", err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		decoded, err := Decode(bytes.NewReader(data))
		if err != nil {
			b.Fatal(err)
		}
		if decoded.Bounds().Dx() != 64 {
			b.Fatal("unexpected decoded width")
		}
	}
}
