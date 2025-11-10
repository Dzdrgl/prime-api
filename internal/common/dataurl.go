package common

import (
	"encoding/base64"
	"errors"
	"fmt"
	"regexp"
)

var (
	ErrInvalidDataURL = errors.New("invalid_data_url")
)

var dataURLRe = regexp.MustCompile(`^data:(image\/(?:jpeg|png|webp));base64,([A-Za-z0-9+/=]+)$`)

func ParseImageDataURL(s string, maxBytes int) (string, []byte, error) {
	m := dataURLRe.FindStringSubmatch(s)
	if len(m) != 3 {
		return "", nil, ErrInvalidDataURL
	}
	mime := m[1]
	raw := m[2]
	// hızlı kaba sınama: base64 ~ %33 overhead, gerçek sınırı decode sonrası uygula
	dec, err := base64.StdEncoding.DecodeString(raw)
	if err != nil {
		return "", nil, ErrInvalidDataURL
	}
	if maxBytes > 0 && len(dec) > maxBytes {
		return "", nil, fmt.Errorf("image too large: %d bytes", len(dec))
	}
	return mime, dec, nil
}
