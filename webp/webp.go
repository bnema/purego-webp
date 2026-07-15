// Package webp provides an idiomatic Go image API on top of libwebp bindings.
package webp

import (
	"errors"
	"image"
	"image/color"
	"io"
	"math"

	"github.com/bnema/purego-webp/libwebp"
)

type EncodeOptions struct {
	Quality  float32
	Lossless bool
}

const maxDecodedImageBytes = 1 << 30

var errDecodedImageTooLarge = errors.New("webp: decoded image exceeds size limit")

func init() {
	image.RegisterFormat("webp", "RIFF????WEBPVP8", Decode, DecodeConfig)
}

// Decode reads a WebP image from r and returns it as image.Image.
func Decode(r io.Reader) (image.Image, error) {
	b, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	w, h, ok, err := libwebp.WebPGetInfo(b)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, libwebp.ErrInvalidData
	}
	stride, size, err := decodeNRGBALayout(w, h)
	if err != nil {
		return nil, err
	}
	if size > maxDecodedImageBytes {
		return nil, errDecodedImageTooLarge
	}

	img := image.NewNRGBA(image.Rect(0, 0, w, h))
	if img.Stride != stride || len(img.Pix) != size {
		return nil, errDecodedImageTooLarge
	}
	if err := libwebp.WebPDecodeRGBAIntoWithInfo(b, img.Pix, img.Stride, w, h); err != nil {
		return nil, err
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

// decodeNRGBALayout verifies the Go allocation and C int stride constraints
// without allocating the output buffer.
func decodeNRGBALayout(width, height int) (stride, size int, err error) {
	if width <= 0 || height <= 0 {
		return 0, 0, libwebp.ErrInvalidDimension
	}
	if width > math.MaxInt32/4 {
		return 0, 0, libwebp.ErrInvalidStride
	}
	stride = width * 4
	if height > int(^uint(0)>>1)/stride {
		return 0, 0, libwebp.ErrInvalidDimension
	}
	return stride, stride * height, nil
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
