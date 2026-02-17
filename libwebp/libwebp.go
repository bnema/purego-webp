package libwebp

import (
	"errors"
	"fmt"
	"unsafe"

	lowlevel "github.com/bnema/purego-webp/internal/libwebp"
)

var (
	ErrInvalidData      = errors.New("libwebp: invalid webp data")
	ErrDecodeFailed     = errors.New("libwebp: decode failed")
	ErrEncodeFailed     = errors.New("libwebp: encode failed")
	ErrInvalidDimension = errors.New("libwebp: invalid dimensions")
	ErrInvalidStride    = errors.New("libwebp: invalid stride")
	ErrBufferTooSmall   = errors.New("libwebp: output buffer too small")
)

type VP8StatusCode int32

const (
	VP8StatusOK              VP8StatusCode = 0
	VP8StatusOutOfMemory     VP8StatusCode = 1
	VP8StatusInvalidParam    VP8StatusCode = 2
	VP8StatusBitstreamError  VP8StatusCode = 3
	VP8StatusUnsupportedFeat VP8StatusCode = 4
	VP8StatusSuspended       VP8StatusCode = 5
	VP8StatusUserAbort       VP8StatusCode = 6
	VP8StatusNotEnoughData   VP8StatusCode = 7
)

type BitstreamFeatures struct {
	Width        int
	Height       int
	HasAlpha     bool
	HasAnimation bool
	Format       int
}

type DecBuffer = lowlevel.WebPDecBuffer
type DecoderOptions = lowlevel.WebPDecoderOptions
type DecoderConfig = lowlevel.WebPDecoderConfig
type Config = lowlevel.WebPConfig
type MemoryWriter = lowlevel.WebPMemoryWriter
type Picture = lowlevel.WebPPicture

const (
	PresetDefault = 0
	PresetPicture = 1
	PresetPhoto   = 2
	PresetDrawing = 3
	PresetIcon    = 4
	PresetText    = 5

	HintDefault = 0
	HintPicture = 1
	HintPhoto   = 2
	HintGraph   = 3
	HintLast    = 4

	ModeRGB      = 0
	ModeRGBA     = 1
	ModeBGR      = 2
	ModeBGRA     = 3
	ModeARGB     = 4
	ModeRGBA4444 = 5
	ModeRGB565   = 6
	ModergbA     = 7
	ModebgrA     = 8
	ModeArgb     = 9
	ModergbA4444 = 10
	ModeYUV      = 11
	ModeYUVA     = 12
	ModeLast     = 13
)

func Available() bool {
	return lowlevel.Available()
}

func Version() (decoder uint32, encoder uint32, err error) {
	if err := lowlevel.EnsureLoaded(); err != nil {
		return 0, 0, err
	}

	return uint32(lowlevel.WebPGetDecoderVersion()), uint32(lowlevel.WebPGetEncoderVersion()), nil
}

func WebPGetInfo(data []byte) (width, height int, ok bool, err error) {
	if err := lowlevel.EnsureLoaded(); err != nil {
		return 0, 0, false, err
	}
	if len(data) == 0 {
		return 0, 0, false, nil
	}

	var w, h int32
	ret := lowlevel.WebPGetInfo(&data[0], uintptr(len(data)), &w, &h)
	return int(w), int(h), ret != 0, nil
}

func WebPGetFeatures(data []byte) (features BitstreamFeatures, status VP8StatusCode, err error) {
	if err := lowlevel.EnsureLoaded(); err != nil {
		return BitstreamFeatures{}, 0, err
	}
	if len(data) == 0 {
		return BitstreamFeatures{}, VP8StatusNotEnoughData, nil
	}

	var raw lowlevel.WebPBitstreamFeatures
	status = VP8StatusCode(lowlevel.WebPGetFeaturesInternal(&data[0], uintptr(len(data)), &raw, lowlevel.WebPDecoderABIVersion))

	return BitstreamFeatures{
		Width:        int(raw.Width),
		Height:       int(raw.Height),
		HasAlpha:     raw.HasAlpha != 0,
		HasAnimation: raw.HasAnimation != 0,
		Format:       int(raw.Format),
	}, status, nil
}

func WebPInitDecBuffer(buffer *DecBuffer) (ok bool, err error) {
	if err := lowlevel.EnsureLoaded(); err != nil {
		return false, err
	}
	if buffer == nil {
		return false, ErrInvalidData
	}

	return lowlevel.WebPInitDecBufferInternal(buffer, lowlevel.WebPDecoderABIVersion) != 0, nil
}

func WebPFreeDecBuffer(buffer *DecBuffer) error {
	if err := lowlevel.EnsureLoaded(); err != nil {
		return err
	}
	if buffer == nil {
		return nil
	}

	lowlevel.WebPFreeDecBuffer(buffer)
	return nil
}

func WebPInitDecoderConfig(config *DecoderConfig) (ok bool, err error) {
	if err := lowlevel.EnsureLoaded(); err != nil {
		return false, err
	}
	if config == nil {
		return false, ErrInvalidData
	}

	return lowlevel.WebPInitDecoderConfigInternal(config, lowlevel.WebPDecoderABIVersion) != 0, nil
}

func WebPValidateDecoderConfig(config *DecoderConfig) (ok bool, err error) {
	if err := lowlevel.EnsureLoaded(); err != nil {
		return false, err
	}
	if config == nil {
		return false, ErrInvalidData
	}

	return lowlevel.WebPValidateDecoderConfig(config) != 0, nil
}

func WebPConfigInit(config *Config) (ok bool, err error) {
	if err := lowlevel.EnsureLoaded(); err != nil {
		return false, err
	}
	if config == nil {
		return false, ErrInvalidData
	}

	return lowlevel.WebPConfigInitInternal(config, PresetDefault, 75, lowlevel.WebPEncoderABIVersion) != 0, nil
}

func WebPConfigPreset(config *Config, preset int32, quality float32) (ok bool, err error) {
	if err := lowlevel.EnsureLoaded(); err != nil {
		return false, err
	}
	if config == nil {
		return false, ErrInvalidData
	}

	return lowlevel.WebPConfigInitInternal(config, preset, quality, lowlevel.WebPEncoderABIVersion) != 0, nil
}

func WebPConfigLosslessPreset(config *Config, level int32) (ok bool, err error) {
	if err := lowlevel.EnsureLoaded(); err != nil {
		return false, err
	}
	if config == nil {
		return false, ErrInvalidData
	}

	return lowlevel.WebPConfigLosslessPreset(config, level) != 0, nil
}

func WebPValidateConfig(config *Config) (ok bool, err error) {
	if err := lowlevel.EnsureLoaded(); err != nil {
		return false, err
	}
	if config == nil {
		return false, ErrInvalidData
	}

	return lowlevel.WebPValidateConfig(config) != 0, nil
}

func WebPPictureInit(picture *Picture) (ok bool, err error) {
	if err := lowlevel.EnsureLoaded(); err != nil {
		return false, err
	}
	if picture == nil {
		return false, ErrInvalidData
	}

	return lowlevel.WebPPictureInitInternal(picture, lowlevel.WebPEncoderABIVersion) != 0, nil
}

func WebPMemoryWriterInit(writer *MemoryWriter) error {
	if err := lowlevel.EnsureLoaded(); err != nil {
		return err
	}
	if writer == nil {
		return ErrInvalidData
	}

	lowlevel.WebPMemoryWriterInit(writer)
	return nil
}

func WebPMemoryWriterClear(writer *MemoryWriter) error {
	if err := lowlevel.EnsureLoaded(); err != nil {
		return err
	}
	if writer == nil {
		return nil
	}

	lowlevel.WebPMemoryWriterClear(writer)
	return nil
}

func WebPEncode(config *Config, picture *Picture) (ok bool, err error) {
	if err := lowlevel.EnsureLoaded(); err != nil {
		return false, err
	}
	if config == nil || picture == nil {
		return false, ErrInvalidData
	}

	return lowlevel.WebPEncode(config, picture) != 0, nil
}

func WebPINewDecoder(outputBuffer *DecBuffer) (uintptr, error) {
	if err := lowlevel.EnsureLoaded(); err != nil {
		return 0, err
	}

	idec := lowlevel.WebPINewDecoder(outputBuffer)
	if idec == 0 {
		return 0, ErrDecodeFailed
	}

	return idec, nil
}

func WebPINewRGB(csp int32, outputBuffer []byte, outputStride int32) (uintptr, error) {
	if err := lowlevel.EnsureLoaded(); err != nil {
		return 0, err
	}

	ptr, size := ptrAndSize(outputBuffer)
	idec := lowlevel.WebPINewRGB(csp, ptr, size, outputStride)
	if idec == 0 {
		return 0, ErrDecodeFailed
	}

	return idec, nil
}

func WebPINewYUV(luma []byte, lumaStride int32, u []byte, uStride int32, v []byte, vStride int32) (uintptr, error) {
	if err := lowlevel.EnsureLoaded(); err != nil {
		return 0, err
	}

	lumaPtr, lumaSize := ptrAndSize(luma)
	uPtr, uSize := ptrAndSize(u)
	vPtr, vSize := ptrAndSize(v)

	idec := lowlevel.WebPINewYUV(lumaPtr, lumaSize, lumaStride, uPtr, uSize, uStride, vPtr, vSize, vStride)
	if idec == 0 {
		return 0, ErrDecodeFailed
	}

	return idec, nil
}

func WebPINewYUVA(luma []byte, lumaStride int32, u []byte, uStride int32, v []byte, vStride int32, a []byte, aStride int32) (uintptr, error) {
	if err := lowlevel.EnsureLoaded(); err != nil {
		return 0, err
	}

	lumaPtr, lumaSize := ptrAndSize(luma)
	uPtr, uSize := ptrAndSize(u)
	vPtr, vSize := ptrAndSize(v)
	aPtr, aSize := ptrAndSize(a)

	idec := lowlevel.WebPINewYUVA(lumaPtr, lumaSize, lumaStride, uPtr, uSize, uStride, vPtr, vSize, vStride, aPtr, aSize, aStride)
	if idec == 0 {
		return 0, ErrDecodeFailed
	}

	return idec, nil
}

func WebPIDelete(idec uintptr) error {
	if err := lowlevel.EnsureLoaded(); err != nil {
		return err
	}
	if idec == 0 {
		return nil
	}

	lowlevel.WebPIDelete(idec)
	return nil
}

func WebPIAppend(idec uintptr, data []byte) (VP8StatusCode, error) {
	if err := lowlevel.EnsureLoaded(); err != nil {
		return 0, err
	}
	if idec == 0 || len(data) == 0 {
		return VP8StatusInvalidParam, ErrInvalidData
	}

	return VP8StatusCode(lowlevel.WebPIAppend(idec, &data[0], uintptr(len(data)))), nil
}

func WebPIUpdate(idec uintptr, data []byte) (VP8StatusCode, error) {
	if err := lowlevel.EnsureLoaded(); err != nil {
		return 0, err
	}
	if idec == 0 || len(data) == 0 {
		return VP8StatusInvalidParam, ErrInvalidData
	}

	return VP8StatusCode(lowlevel.WebPIUpdate(idec, &data[0], uintptr(len(data)))), nil
}

func WebPIDecode(data []byte, config *DecoderConfig) (uintptr, error) {
	if err := lowlevel.EnsureLoaded(); err != nil {
		return 0, err
	}
	if len(data) == 0 {
		return 0, ErrInvalidData
	}

	idec := lowlevel.WebPIDecode(&data[0], uintptr(len(data)), config)
	if idec == 0 {
		return 0, ErrDecodeFailed
	}

	return idec, nil
}

func WebPIDecodedArea(idec uintptr, left, top, width, height *int32) (*DecBuffer, error) {
	if err := lowlevel.EnsureLoaded(); err != nil {
		return nil, err
	}
	if idec == 0 {
		return nil, ErrInvalidData
	}

	buf := lowlevel.WebPIDecodedArea(idec, left, top, width, height)
	if buf == nil {
		return nil, ErrDecodeFailed
	}

	return buf, nil
}

func WebPIDecGetRGB(idec uintptr, lastY, width, height, stride *int32) (uintptr, error) {
	if err := lowlevel.EnsureLoaded(); err != nil {
		return 0, err
	}
	if idec == 0 {
		return 0, ErrInvalidData
	}

	ptr := lowlevel.WebPIDecGetRGB(idec, lastY, width, height, stride)
	if ptr == nil {
		return 0, ErrDecodeFailed
	}

	return uintptr(unsafe.Pointer(ptr)), nil
}

func WebPIDecGetYUVA(idec uintptr, lastY *int32, u, v, a **byte, width, height, stride, uvStride, aStride *int32) (uintptr, error) {
	if err := lowlevel.EnsureLoaded(); err != nil {
		return 0, err
	}
	if idec == 0 {
		return 0, ErrInvalidData
	}

	ptr := lowlevel.WebPIDecGetYUVA(idec, lastY, u, v, a, width, height, stride, uvStride, aStride)
	if ptr == nil {
		return 0, ErrDecodeFailed
	}

	return uintptr(unsafe.Pointer(ptr)), nil
}

func WebPIDecGetYUV(idec uintptr, lastY *int32, u, v **byte, width, height, stride, uvStride *int32) (uintptr, error) {
	var a *byte
	return WebPIDecGetYUVA(idec, lastY, u, v, &a, width, height, stride, uvStride, nil)
}

func WebPDecode(data []byte, config *DecoderConfig) (status VP8StatusCode, err error) {
	if err := lowlevel.EnsureLoaded(); err != nil {
		return 0, err
	}
	if len(data) == 0 {
		return VP8StatusNotEnoughData, nil
	}
	if config == nil {
		return VP8StatusInvalidParam, ErrInvalidData
	}

	return VP8StatusCode(lowlevel.WebPDecode(&data[0], uintptr(len(data)), config)), nil
}

func WebPDecodeRGBA(data []byte) (pix []byte, width, height, stride int, err error) {
	return decodeToOwnedBuffer(data, 4, lowlevel.WebPDecodeRGBA)
}

func WebPDecodeARGB(data []byte) (pix []byte, width, height, stride int, err error) {
	return decodeToOwnedBuffer(data, 4, lowlevel.WebPDecodeARGB)
}

func WebPDecodeBGRA(data []byte) (pix []byte, width, height, stride int, err error) {
	return decodeToOwnedBuffer(data, 4, lowlevel.WebPDecodeBGRA)
}

func WebPDecodeRGB(data []byte) (pix []byte, width, height, stride int, err error) {
	return decodeToOwnedBuffer(data, 3, lowlevel.WebPDecodeRGB)
}

func WebPDecodeBGR(data []byte) (pix []byte, width, height, stride int, err error) {
	return decodeToOwnedBuffer(data, 3, lowlevel.WebPDecodeBGR)
}

func WebPDecodeRGBAInto(data []byte, outputBuffer []byte, outputStride int) (width, height int, err error) {
	return decodeInto(data, outputBuffer, outputStride, 4, lowlevel.WebPDecodeRGBAInto)
}

func WebPDecodeARGBInto(data []byte, outputBuffer []byte, outputStride int) (width, height int, err error) {
	return decodeInto(data, outputBuffer, outputStride, 4, lowlevel.WebPDecodeARGBInto)
}

func WebPDecodeBGRAInto(data []byte, outputBuffer []byte, outputStride int) (width, height int, err error) {
	return decodeInto(data, outputBuffer, outputStride, 4, lowlevel.WebPDecodeBGRAInto)
}

func WebPDecodeRGBInto(data []byte, outputBuffer []byte, outputStride int) (width, height int, err error) {
	return decodeInto(data, outputBuffer, outputStride, 3, lowlevel.WebPDecodeRGBInto)
}

func WebPDecodeBGRInto(data []byte, outputBuffer []byte, outputStride int) (width, height int, err error) {
	return decodeInto(data, outputBuffer, outputStride, 3, lowlevel.WebPDecodeBGRInto)
}

func WebPDecodeYUV(data []byte) (y, u, v []byte, width, height, yStride, uvStride int, err error) {
	if err := lowlevel.EnsureLoaded(); err != nil {
		return nil, nil, nil, 0, 0, 0, 0, err
	}
	if len(data) == 0 {
		return nil, nil, nil, 0, 0, 0, 0, ErrInvalidData
	}

	var w, h int32
	var uPtr, vPtr *byte
	var ys, uvs int32
	yPtr := lowlevel.WebPDecodeYUV(&data[0], uintptr(len(data)), &w, &h, &uPtr, &vPtr, &ys, &uvs)
	if yPtr == nil {
		return nil, nil, nil, 0, 0, 0, 0, ErrDecodeFailed
	}
	defer lowlevel.WebPFree(uintptr(unsafe.Pointer(yPtr)))

	width = int(w)
	height = int(h)
	if width <= 0 || height <= 0 {
		return nil, nil, nil, 0, 0, 0, 0, ErrInvalidDimension
	}

	yStride = int(ys)
	uvStride = int(uvs)
	uvHeight := (height + 1) / 2

	y = make([]byte, yStride*height)
	u = make([]byte, uvStride*uvHeight)
	v = make([]byte, uvStride*uvHeight)

	copy(y, unsafe.Slice(yPtr, len(y)))
	copy(u, unsafe.Slice(uPtr, len(u)))
	copy(v, unsafe.Slice(vPtr, len(v)))

	return y, u, v, width, height, yStride, uvStride, nil
}

func WebPDecodeYUVInto(data []byte, luma []byte, lumaStride int, u []byte, uStride int, v []byte, vStride int) (width, height int, err error) {
	if err := lowlevel.EnsureLoaded(); err != nil {
		return 0, 0, err
	}
	if len(data) == 0 {
		return 0, 0, ErrInvalidData
	}
	if len(luma) == 0 || len(u) == 0 || len(v) == 0 {
		return 0, 0, ErrBufferTooSmall
	}

	w, h, ok, err := WebPGetInfo(data)
	if err != nil {
		return 0, 0, err
	}
	if !ok {
		return 0, 0, ErrInvalidData
	}
	if lumaStride < w {
		return 0, 0, ErrInvalidStride
	}
	uvWidth := (w + 1) / 2
	uvHeight := (h + 1) / 2
	if uStride < uvWidth || vStride < uvWidth {
		return 0, 0, ErrInvalidStride
	}
	if len(luma) < lumaStride*h || len(u) < uStride*uvHeight || len(v) < vStride*uvHeight {
		return 0, 0, ErrBufferTooSmall
	}

	ptr := lowlevel.WebPDecodeYUVInto(
		&data[0],
		uintptr(len(data)),
		&luma[0],
		uintptr(len(luma)),
		int32(lumaStride),
		&u[0],
		uintptr(len(u)),
		int32(uStride),
		&v[0],
		uintptr(len(v)),
		int32(vStride),
	)
	if ptr == nil {
		return 0, 0, ErrDecodeFailed
	}

	return w, h, nil
}

func WebPEncodeRGBA(rgba []byte, width, height, stride int, quality float32) ([]byte, error) {
	return encodeWithQuality(rgba, width, height, stride, 4, quality, lowlevel.WebPEncodeRGBA)
}

func WebPEncodeBGRA(bgra []byte, width, height, stride int, quality float32) ([]byte, error) {
	return encodeWithQuality(bgra, width, height, stride, 4, quality, lowlevel.WebPEncodeBGRA)
}

func WebPEncodeRGB(rgb []byte, width, height, stride int, quality float32) ([]byte, error) {
	return encodeWithQuality(rgb, width, height, stride, 3, quality, lowlevel.WebPEncodeRGB)
}

func WebPEncodeBGR(bgr []byte, width, height, stride int, quality float32) ([]byte, error) {
	return encodeWithQuality(bgr, width, height, stride, 3, quality, lowlevel.WebPEncodeBGR)
}

func WebPEncodeLosslessRGBA(rgba []byte, width, height, stride int) ([]byte, error) {
	return encodeLossless(rgba, width, height, stride, 4, lowlevel.WebPEncodeLosslessRGBA)
}

func WebPEncodeLosslessBGRA(bgra []byte, width, height, stride int) ([]byte, error) {
	return encodeLossless(bgra, width, height, stride, 4, lowlevel.WebPEncodeLosslessBGRA)
}

func WebPEncodeLosslessRGB(rgb []byte, width, height, stride int) ([]byte, error) {
	return encodeLossless(rgb, width, height, stride, 3, lowlevel.WebPEncodeLosslessRGB)
}

func WebPEncodeLosslessBGR(bgr []byte, width, height, stride int) ([]byte, error) {
	return encodeLossless(bgr, width, height, stride, 3, lowlevel.WebPEncodeLosslessBGR)
}

type decodeFunc func(data *byte, dataSize uintptr, width *int32, height *int32) *byte

type decodeIntoFunc func(data *byte, dataSize uintptr, outputBuffer *byte, outputBufferSize uintptr, outputStride int32) *byte

func decodeInto(data []byte, outputBuffer []byte, outputStride int, bytesPerPixel int, fn decodeIntoFunc) (width, height int, err error) {
	if err := lowlevel.EnsureLoaded(); err != nil {
		return 0, 0, err
	}
	if len(data) == 0 {
		return 0, 0, ErrInvalidData
	}
	if len(outputBuffer) == 0 {
		return 0, 0, ErrBufferTooSmall
	}

	w, h, ok, err := WebPGetInfo(data)
	if err != nil {
		return 0, 0, err
	}
	if !ok {
		return 0, 0, ErrInvalidData
	}
	if outputStride < w*bytesPerPixel {
		return 0, 0, ErrInvalidStride
	}
	required := outputStride * h
	if len(outputBuffer) < required {
		return 0, 0, ErrBufferTooSmall
	}

	ptr := fn(&data[0], uintptr(len(data)), &outputBuffer[0], uintptr(len(outputBuffer)), int32(outputStride))
	if ptr == nil {
		return 0, 0, ErrDecodeFailed
	}

	return w, h, nil
}

func decodeToOwnedBuffer(data []byte, bytesPerPixel int, fn decodeFunc) (pix []byte, width, height, stride int, err error) {
	if err := lowlevel.EnsureLoaded(); err != nil {
		return nil, 0, 0, 0, err
	}
	if len(data) == 0 {
		return nil, 0, 0, 0, ErrInvalidData
	}

	var w, h int32
	ptr := fn(&data[0], uintptr(len(data)), &w, &h)
	if ptr == nil {
		return nil, 0, 0, 0, ErrDecodeFailed
	}
	defer lowlevel.WebPFree(uintptr(unsafe.Pointer(ptr)))

	width = int(w)
	height = int(h)
	if width <= 0 || height <= 0 {
		return nil, 0, 0, 0, ErrInvalidDimension
	}

	stride = width * bytesPerPixel
	bufLen := stride * height
	pix = make([]byte, bufLen)
	copy(pix, unsafe.Slice(ptr, bufLen))

	return pix, width, height, stride, nil
}

type encodeLossyFunc func(pix *byte, width int32, height int32, stride int32, quality float32, output **byte) uintptr

func encodeWithQuality(pix []byte, width, height, stride, bytesPerPixel int, quality float32, fn encodeLossyFunc) ([]byte, error) {
	if err := lowlevel.EnsureLoaded(); err != nil {
		return nil, err
	}
	if err := validatePixelInput(pix, width, height, stride, bytesPerPixel); err != nil {
		return nil, err
	}

	var out *byte
	size := fn(&pix[0], int32(width), int32(height), int32(stride), quality, &out)
	if size == 0 || out == nil {
		return nil, ErrEncodeFailed
	}
	defer lowlevel.WebPFree(uintptr(unsafe.Pointer(out)))

	b := make([]byte, int(size))
	copy(b, unsafe.Slice(out, int(size)))

	return b, nil
}

type encodeLosslessFunc func(pix *byte, width int32, height int32, stride int32, output **byte) uintptr

func encodeLossless(pix []byte, width, height, stride, bytesPerPixel int, fn encodeLosslessFunc) ([]byte, error) {
	if err := lowlevel.EnsureLoaded(); err != nil {
		return nil, err
	}
	if err := validatePixelInput(pix, width, height, stride, bytesPerPixel); err != nil {
		return nil, err
	}

	var out *byte
	size := fn(&pix[0], int32(width), int32(height), int32(stride), &out)
	if size == 0 || out == nil {
		return nil, ErrEncodeFailed
	}
	defer lowlevel.WebPFree(uintptr(unsafe.Pointer(out)))

	b := make([]byte, int(size))
	copy(b, unsafe.Slice(out, int(size)))

	return b, nil
}

func validatePixelInput(pix []byte, width, height, stride, bytesPerPixel int) error {
	if width <= 0 || height <= 0 {
		return ErrInvalidDimension
	}
	if stride < width*bytesPerPixel {
		return ErrInvalidStride
	}
	required := stride * height
	if len(pix) < required {
		return fmt.Errorf("libwebp: pixel buffer too small: got=%d need>=%d", len(pix), required)
	}
	return nil
}

func WebPIsPremultipliedMode(mode int) bool {
	return mode == ModergbA || mode == ModebgrA || mode == ModeArgb || mode == ModergbA4444
}

func WebPIsAlphaMode(mode int) bool {
	return mode == ModeRGBA || mode == ModeBGRA || mode == ModeARGB || mode == ModeRGBA4444 || mode == ModeYUVA || WebPIsPremultipliedMode(mode)
}

func WebPIsRGBMode(mode int) bool {
	return mode < ModeYUV
}

func ptrAndSize(b []byte) (*byte, uintptr) {
	if len(b) == 0 {
		return nil, 0
	}
	return &b[0], uintptr(len(b))
}
