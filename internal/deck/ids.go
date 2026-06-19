package deck

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/binary"
	mathrand "math/rand"
	"strings"
	"time"
	"unicode"
)

const randomIDBytes = 12

// NewDeckID returns a branded deck ID with a URL-safe random suffix.
func NewDeckID() string {
	return newID("deck")
}

// NewSlideID returns a branded slide ID with a URL-safe random suffix.
func NewSlideID() string {
	return newID("slide")
}

// Slugify lowers a title and converts non-alphanumeric runs to hyphens.
func Slugify(title string) string {
	var b strings.Builder
	b.Grow(len(title))
	lastHyphen := false
	for _, r := range title {
		switch {
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			b.WriteRune(unicode.ToLower(r))
			lastHyphen = false
		case !lastHyphen:
			b.WriteByte('-')
			lastHyphen = true
		}
	}
	slug := strings.Trim(b.String(), "-")
	if slug == "" {
		return "deck"
	}
	return slug
}

func newID(prefix string) string {
	b := make([]byte, randomIDBytes)
	if _, err := rand.Read(b); err != nil {
		seed := time.Now().UnixNano()
		if len(b) >= 8 {
			binary.LittleEndian.PutUint64(b[:8], uint64(seed))
		}
		r := mathrand.New(mathrand.NewSource(seed))
		for i := 8; i < len(b); i++ {
			b[i] = byte(r.Intn(256))
		}
	}
	return prefix + "_" + base64.RawURLEncoding.EncodeToString(b)
}
