package media

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"image"
	"image/jpeg"
	"image/png"
	"net/http"

	"jcourse/internal/domain/canteen"
)

// Processor implements canteen.ImageProcessor with the standard library only.
type Processor struct{}

func (Processor) Process(raw []byte, maxPixels int) (canteen.ProcessedImage, error) {
	if len(raw) == 0 {
		return canteen.ProcessedImage{}, canteen.ErrMediaType
	}
	mime := http.DetectContentType(raw)
	if mime != "image/jpeg" && mime != "image/png" {
		return canteen.ProcessedImage{}, canteen.ErrMediaType
	}
	config, _, err := image.DecodeConfig(bytes.NewReader(raw))
	if err != nil || config.Width <= 0 || config.Height <= 0 {
		return canteen.ProcessedImage{}, canteen.ErrMediaType
	}
	if maxPixels > 0 && config.Width*config.Height > maxPixels {
		return canteen.ProcessedImage{}, canteen.ErrMediaDimension
	}
	decoded, _, err := image.Decode(bytes.NewReader(raw))
	if err != nil {
		return canteen.ProcessedImage{}, canteen.ErrMediaType
	}

	var buffer bytes.Buffer
	switch mime {
	case "image/jpeg":
		if err := jpeg.Encode(&buffer, decoded, &jpeg.Options{Quality: 85}); err != nil {
			return canteen.ProcessedImage{}, err
		}
	case "image/png":
		if err := png.Encode(&buffer, decoded); err != nil {
			return canteen.ProcessedImage{}, err
		}
	}
	sum := sha256.Sum256(buffer.Bytes())
	return canteen.ProcessedImage{
		Data:   buffer.Bytes(),
		MIME:   mime,
		Width:  config.Width,
		Height: config.Height,
		SHA256: hex.EncodeToString(sum[:]),
	}, nil
}

var _ canteen.ImageProcessor = Processor{}
