// Package webp provides an idiomatic Go image API on top of libwebp bindings.
package webp

import (
	"image"
	"image/color"
	"io"

	"github.com/bnema/purego-webp/libwebp"
)

type EncodeOptions struct {
	Quality  float32
	Lossless bool
}

func init() {
	image.RegisterFormat("webp", "RIFF????WEBPVP8", Decode, DecodeConfig)
}

// Decode reads a WebP image from r and returns it as image.Image.
func Decode(r io.Reader) (image.Image, error) {
	b, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	pix, w, h, stride, err := libwebp.WebPDecodeRGBA(b)
	if err != nil {
		return nil, err
	}

	img := image.NewNRGBA(image.Rect(0, 0, w, h))
	if stride == img.Stride {
		copy(img.Pix, pix)
		return img, nil
	}

	for y := 0; y < h; y++ {
		srcStart := y * stride
		dstStart := y * img.Stride
		copy(img.Pix[dstStart:dstStart+img.Stride], pix[srcStart:srcStart+img.Stride])
	}

	return img, nil
}

// DecodeConfig returns image metadata for a WebP image from r.
func DecodeConfig(r io.Reader) (image.Config, error) {
	b, err := io.ReadAll(r)
	if err != nil {
		return image.Config{}, err
	}

	w, h, ok, err := libwebp.WebPGetInfo(b)
	if err != nil {
		return image.Config{}, err
	}
	if !ok {
		return image.Config{}, libwebp.ErrInvalidData
	}

	return image.Config{
		ColorModel: color.NRGBAModel,
		Width:      w,
		Height:     h,
	}, nil
}

// Encode writes src as WebP to w using the provided options.
func Encode(w io.Writer, src image.Image, opts *EncodeOptions) error {
	nrgba := toNRGBA(src)

	if opts != nil && opts.Lossless {
		enc, err := libwebp.WebPEncodeLosslessRGBA(nrgba.Pix, nrgba.Rect.Dx(), nrgba.Rect.Dy(), nrgba.Stride)
		if err != nil {
			return err
		}
		_, err = w.Write(enc)
		return err
	}

	quality := float32(75)
	if opts != nil && opts.Quality > 0 {
		quality = opts.Quality
	}

	enc, err := libwebp.WebPEncodeRGBA(nrgba.Pix, nrgba.Rect.Dx(), nrgba.Rect.Dy(), nrgba.Stride, quality)
	if err != nil {
		return err
	}

	_, err = w.Write(enc)
	return err
}

// EncodeLossless writes src as lossless WebP to w.
func EncodeLossless(w io.Writer, src image.Image) error {
	return Encode(w, src, &EncodeOptions{Lossless: true})
}

func toNRGBA(src image.Image) *image.NRGBA {
	if nrgba, ok := src.(*image.NRGBA); ok {
		return nrgba
	}

	b := src.Bounds()
	nrgba := image.NewNRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			nrgba.SetNRGBA(x-b.Min.X, y-b.Min.Y, color.NRGBAModel.Convert(src.At(x, y)).(color.NRGBA))
		}
	}

	return nrgba
}
