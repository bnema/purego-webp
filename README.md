# purego-webp

Pure Go bindings for `libwebp` using [purego](https://github.com/ebitengine/purego), inspired by [kolesa-team/go-webp](https://github.com/kolesa-team/go-webp).

## Goals

- No cgo required
- API shape close to `go-webp`
- Fast cross-compilation
- Small, focused package surface

## Planned package layout

- `webp`: top-level encode/decode APIs
- `decoder`: decode options and helpers
- `encoder`: encode options and presets
- `internal/libwebp`: low-level symbol loading and wrappers

## Current status

Repository template scaffolded. Public API is in place, implementation is pending.

## Example (target API)

```go
package main

import (
	"log"
	"os"

	"github.com/jwijenbergh/purego-webp/decoder"
	"github.com/jwijenbergh/purego-webp/webp"
)

func main() {
	f, err := os.Open("image.webp")
	if err != nil {
		log.Fatal(err)
	}

	img, err := webp.Decode(f, &decoder.Options{})
	if err != nil {
		log.Fatal(err)
	}

	_ = img
}
```
