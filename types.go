package circle

// Circle has some types that we want to decode to normal Go types, define/do
// that here.

import (
	"bytes"
	"encoding/json"
	"net/url"
	"time"
)

// Unmarshallable URL
type URL struct {
	*url.URL
}

func (oururl *URL) UnmarshalJSON(b []byte) error {
	// extra hop here to strip the leading/trailing quotes
	var s string
	err := json.Unmarshal(b, &s)
	if err != nil {
		return err
	}
	u, err := url.Parse(s)
	if err != nil {
		return err
	}
	*oururl = URL{u}
	return nil
}

type CircleDuration time.Duration

var null = []byte("null")

func (cd *CircleDuration) UnmarshalJSON(b []byte) error {
	if bytes.Equal(b, null) {
		*cd = CircleDuration(time.Duration(-1))
		return nil
	}
	var d time.Duration
	err := json.Unmarshal(b, &d)
	if err != nil {
		return err
	}
	*cd = CircleDuration(d * time.Millisecond)
	return nil
}
