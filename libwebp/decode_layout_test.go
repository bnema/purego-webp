package libwebp

import (
	"errors"
	"math"
	"testing"
)

func TestCheckedDecodeLayout(t *testing.T) {
	tests := []struct {
		name                 string
		width, height, bpp   int
		wantStride, wantSize int
		wantErr              error
	}{
		{name: "normal", width: 3, height: 2, bpp: 4, wantStride: 12, wantSize: 24},
		{name: "non-positive width", width: 0, height: 1, bpp: 4, wantErr: ErrInvalidDimension},
		{name: "non-positive height", width: 1, height: 0, bpp: 4, wantErr: ErrInvalidDimension},
		{name: "MaxInt32 stride", width: math.MaxInt32 / 4, height: 1, bpp: 4, wantStride: math.MaxInt32 - 3, wantSize: math.MaxInt32 - 3},
		{name: "stride exceeds C int", width: math.MaxInt32/4 + 1, height: 1, bpp: 4, wantErr: ErrInvalidStride},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stride, size, err := checkedDecodeLayout(tt.width, tt.height, tt.bpp)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("checkedDecodeLayout(%d, %d, %d) error = %v, want %v", tt.width, tt.height, tt.bpp, err, tt.wantErr)
			}
			if stride != tt.wantStride || size != tt.wantSize {
				t.Fatalf("checkedDecodeLayout() = (%d, %d), want (%d, %d)", stride, size, tt.wantStride, tt.wantSize)
			}
		})
	}
}

func TestCheckedDecodeLayoutRejectsIntOverflow(t *testing.T) {
	maxInt := int(^uint(0) >> 1)
	if _, _, err := checkedDecodeLayout(maxInt/4+1, 1, 4); err == nil {
		t.Fatal("width * 4 overflow was accepted")
	}
	if _, _, err := checkedDecodeLayout(1, maxInt/4+1, 4); err == nil {
		t.Fatal("stride * height overflow was accepted")
	}
}

func TestWebPDecodeRGBAIntoRejectsUndersizedOutput(t *testing.T) {
	data, err := WebPEncodeLosslessRGBA([]byte{1, 2, 3, 4, 5, 6, 7, 8}, 2, 1, 8)
	if err != nil {
		t.Fatalf("encode fixture: %v", err)
	}
	_, _, err = WebPDecodeRGBAInto(data, make([]byte, 7), 8)
	if !errors.Is(err, ErrBufferTooSmall) {
		t.Fatalf("WebPDecodeRGBAInto() error = %v, want %v", err, ErrBufferTooSmall)
	}
}
