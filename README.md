# purego-webp

Pure Go bindings for `libwebp` using [purego](https://github.com/ebitengine/purego), with a C-first API and an idiomatic Go wrapper API.

## Goals

- No cgo required
- Keep low-level API close to `libwebp`
- Provide a clean high-level Go API for app usage
- Fast cross-compilation
- Small, focused package surface

## Packages

- `libwebp`: C-first API surface (`WebPGetInfo`, `WebPDecodeRGBA`, `WebPEncodeRGBA`, `WebPFree` behavior wrapped safely)
- `webp`: idiomatic `image`/`io` APIs (`Decode`, `DecodeConfig`, `Encode`, `EncodeLossless`)
- `internal/libwebp`: dynamic loading + symbol registration via purego

## Current status

Template, generator, and broad libwebp API coverage are in place.

Coverage snapshot:

- Implemented generated low-level bindings for decode, incremental decode, simple encode, advanced encode config, picture APIs, and memory-writer APIs.
- Remaining header-level names are mostly inline helper wrappers that route to `*Internal` symbols; these are exposed in `libwebp` as Go wrappers.

## Runtime requirement

`libwebp` must be installed on the host system at runtime (for example `libwebp.so*` on Linux).

## Examples

### High-level Go API

```go
package main

import (
	"bytes"
	"image"
	"image/color"
	"log"
	"os"

	webpimg "github.com/bnema/purego-webp/webp"
)

func main() {
	src := image.NewNRGBA(image.Rect(0, 0, 32, 32))
	src.Set(0, 0, color.NRGBA{R: 255, G: 0, B: 0, A: 255})

	var buf bytes.Buffer
	if err := webpimg.Encode(&buf, src, &webpimg.EncodeOptions{Quality: 80}); err != nil {
		log.Fatal(err)
	}

	decoded, err := webpimg.Decode(bytes.NewReader(buf.Bytes()))
	if err != nil {
		log.Fatal(err)
	}
	_ = decoded

	f, err := os.Create("out.webp")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	if _, err := f.Write(buf.Bytes()); err != nil {
		log.Fatal(err)
	}
}
```

### C-first API

```go
package main

import (
	"log"
	"os"

	"github.com/bnema/purego-webp/libwebp"
)

func main() {
	data, err := os.ReadFile("image.webp")
	if err != nil {
		log.Fatal(err)
	}

	width, height, ok, err := libwebp.WebPGetInfo(data)
	if err != nil {
		log.Fatal(err)
	}
	if !ok {
		log.Fatal("invalid webp")
	}

	rgba, _, _, stride, err := libwebp.WebPDecodeRGBA(data)
	if err != nil {
		log.Fatal(err)
	}

	encoded, err := libwebp.WebPEncodeRGBA(rgba, width, height, stride, 75)
	if err != nil {
		log.Fatal(err)
	}

	_ = encoded
}
```

Also available in `libwebp` now:

- Decode variants: `WebPDecodeARGB`, `WebPDecodeBGRA`, `WebPDecodeRGB`, `WebPDecodeBGR`, `WebPDecodeRGBAInto`
- Decode config/incremental: `WebPInitDecBuffer`, `WebPInitDecoderConfig`, `WebPIAppend`, `WebPIUpdate`, `WebPIDecGetRGB`, `WebPIDecGetYUVA`
- Encode variants: `WebPEncodeRGB`, `WebPEncodeBGR`, `WebPEncodeBGRA`
- Lossless variants: `WebPEncodeLosslessRGB`, `WebPEncodeLosslessBGR`, `WebPEncodeLosslessBGRA`, `WebPEncodeLosslessRGBA`
- Advanced encode/config: `WebPConfigInit`, `WebPConfigPreset`, `WebPConfigLosslessPreset`, `WebPValidateConfig`, `WebPEncode`

## Notes

- This repo includes a template + generator for low-level symbol bindings.
- Edit `gen/spec.json`, then run `./gen.sh` (or `go generate ./...`).
- Generated file: `internal/libwebp/generated_symbols.go`.
