package webp

import (
	"bytes"
	"image"
	"image/color"
	"testing"
)

var benchmarkDecodedImage image.Image

func BenchmarkDecodeFavicon64(b *testing.B) {
	benchmarkDecodeNRGBA(b, 64, 64)
}

func BenchmarkDecodeLarge1024(b *testing.B) {
	benchmarkDecodeNRGBA(b, 1024, 1024)
}

func benchmarkDecodeNRGBA(b *testing.B, width, height int) {
	b.Helper()

	source := image.NewNRGBA(image.Rect(0, 0, width, height))
	for y := range height {
		for x := range width {
			source.SetNRGBA(x, y, color.NRGBA{
				R: uint8(x),
				G: uint8(y),
				B: uint8(x ^ y),
				A: uint8(128 + (x+y)%128),
			})
		}
	}

	var encoded bytes.Buffer
	if err := EncodeLossless(&encoded, source); err != nil {
		b.Fatalf("encode benchmark fixture: %v", err)
	}
	data := encoded.Bytes()

	b.ReportAllocs()
	b.SetBytes(int64(len(data)))
	b.ResetTimer()
	for range b.N {
		decoded, err := Decode(bytes.NewReader(data))
		if err != nil {
			b.Fatalf("decode benchmark fixture: %v", err)
		}
		benchmarkDecodedImage = decoded
	}
}
