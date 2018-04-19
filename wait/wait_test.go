package wait

import (
	"io"
	"net/http"
	"testing"
	"time"
)

func makeRequest(client http.Client, method, uri string) (io.ReadCloser, error) {
	req, err := http.NewRequest(method, uri, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Add("User-Agent", "wait-test")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}

func TestHttpError(t *testing.T) {
	client := http.Client{
		Timeout: 200 * time.Millisecond,
	}
	_, err := makeRequest(client, "GET", "http://localhost:11233")
	if !isHttpError(err) {
		t.Fatalf("expected err to be http error, was %s", err)
	}

	_, err = makeRequest(client, "GET", "https://httpbin.org/delay/2")
	if !isHttpError(err) {
		t.Fatalf("expected err to be http error, was %s", err)
	}
}

func TestShorterString(t *testing.T) {
	minTipLength := getShorterString("1d79f2b877c86ac0964f3fe69a0171926aa6f1d8", "1d79f2b87")
	expectedMinTipLength := 9
	if minTipLength != expectedMinTipLength {
		t.Errorf("expected half hour cost to be %d, was %d", expectedMinTipLength, minTipLength)
	}

	minTipLength = getShorterString("1d79f2b877c86ac0964f3fe69a0171926aa6f1d8", "1d79f2b")
	expectedMinTipLength = 7
	if minTipLength != expectedMinTipLength {
		t.Errorf("expected half hour cost to be %d, was %d", expectedMinTipLength, minTipLength)
	}

	minTipLength = getShorterString("1d79f", "1d79f2b87")
	expectedMinTipLength = 5
	if minTipLength != expectedMinTipLength {
		t.Errorf("expected half hour cost to be %d, was %d", expectedMinTipLength, minTipLength)
	}
}
