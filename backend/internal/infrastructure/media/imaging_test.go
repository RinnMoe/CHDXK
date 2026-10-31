package media

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"image"
	"image/color"
	"image/png"
	"testing"

	"jcourse/internal/domain/canteen"
)

func testPNG(t *testing.T, width, height int) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			img.Set(x, y, color.RGBA{R: 10, G: 20, B: 30, A: 255})
		}
	}
	var buffer bytes.Buffer
	if err := png.Encode(&buffer, img); err != nil {
		t.Fatalf("encode png: %v", err)
	}
	return buffer.Bytes()
}

func TestProcessImage(t *testing.T) {
	raw := testPNG(t, 12, 8)
	processed, err := (Processor{}).Process(raw, 1000)
	if err != nil {
		t.Fatalf("Process() = %v", err)
	}
	if processed.MIME != "image/png" || processed.Width != 12 || processed.Height != 8 {
		t.Fatalf("processed = %+v", processed)
	}
	sum := sha256.Sum256(processed.Data)
	if processed.SHA256 != hex.EncodeToString(sum[:]) {
		t.Fatalf("sha256 = %s", processed.SHA256)
	}
}

func TestProcessImageRejects(t *testing.T) {
	if _, err := (Processor{}).Process([]byte("not an image"), 1000); err != canteen.ErrMediaType {
		t.Fatalf("text err = %v", err)
	}
	raw := testPNG(t, 20, 20)
	if _, err := (Processor{}).Process(raw, 10); err != canteen.ErrMediaDimension {
		t.Fatalf("dimension err = %v", err)
	}
}
