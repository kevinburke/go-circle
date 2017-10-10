package types

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
)

func ExampleNullString() {
	n := NullString{Valid: true, String: "foo"}
	json.NewEncoder(os.Stdout).Encode(n)
	// Output: "foo"
}

func TestString(t *testing.T) {
	var ns NullString
	str := []byte("\"foo\"")
	err := json.Unmarshal(str, &ns)
	assertNotError(t, err, "")
	assertEquals(t, ns.Valid, true)
	assertEquals(t, ns.String, "foo")
}

func TestNullString(t *testing.T) {
	var ns NullString
	str := []byte("null")
	err := json.Unmarshal(str, &ns)
	assertNotError(t, err, "")
	assertEquals(t, ns.Valid, false)
}

func TestStringMarshal(t *testing.T) {
	ns := NullString{
		Valid:  true,
		String: "foo bar",
	}
	b, err := json.Marshal(ns)
	assertNotError(t, err, "")
	assertEquals(t, string(b), "\"foo bar\"")
}

func TestStringMarshalNull(t *testing.T) {
	ns := NullString{
		Valid:  false,
		String: "",
	}
	b, err := json.Marshal(ns)
	assertNotError(t, err, "")
	assertEquals(t, string(b), "null")
}

var byteTests = []struct {
	in  string
	err string
	out Bits
}{
	{"5bit", "", 5 * Bit},
	{"5kB", "", 5 * Kilobyte},
	{"1.3kB", "", 1300 * Byte},
	{"1300B", "", 1300 * Byte},
	{"1300", "types: missing unit in input 1300", 0},
}

func TestParseBits(t *testing.T) {
	for _, tt := range byteTests {
		out, err := ParseBits(tt.in)
		hdr := fmt.Sprintf("ParseBits(%q):", tt.in)
		if err == nil {
			if tt.err != "" {
				t.Errorf("%s expected to get error %v, got a valid result", hdr, tt.err)
				continue
			}
			if out != tt.out {
				t.Errorf("%s got %v, want %v", hdr, out, tt.out)
			}
		} else {
			if tt.err == "" {
				t.Errorf("%s didn't expect error but got %v", hdr, err)
				continue
			}
			if err.Error() != tt.err {
				t.Errorf("%s got error %v, want %v", hdr, err, tt.err)
				continue
			}
		}
	}
}

func TestBytes(t *testing.T) {
	b := 13 * Bit
	if b.Bytes() != 1.625 {
		t.Errorf("13 bits should be 1.625 bytes, got %v", b.Bytes())
	}
}

func TestKilobytes(t *testing.T) {
	b := 3*Kilobyte + 10*Byte
	if got := b.Kilobytes(); got != 3.01 {
		t.Errorf("3.01kB should be 3.01 bytes, got %v", got)
	}
}

func TestBitsString(t *testing.T) {
	b := 0 * Bit
	if got := b.String(); got != "0" {
		t.Errorf("0 bits should be 0, got %q", got)
	}
	b = 7 * Bit
	if got := b.String(); got != "7bit" {
		t.Errorf("7 bits should be 7bit, got %q", got)
	}
	b = 9 * Bit
	if got := b.String(); got != "1.125B" {
		t.Errorf("9 bits should be 1.125B, got %q", got)
	}
	b = 7380*Kilobyte + 871*Byte
	if got := b.String(); got != "7.38MB" {
		t.Errorf("25015kB should be 25.015MB, got %q", got)
	}
	b = 25015 * Kilobyte
	if got := b.String(); got != "25.015MB" {
		t.Errorf("25015kB should be 25.015MB, got %q", got)
	}
	b = -25015 * Kilobyte
	if got := b.String(); got != "-25.015MB" {
		t.Errorf("25015kB should be 25.015MB, got %q", got)
	}
	b = -123*Megabyte - 15*Kilobyte
	if got := b.String(); got != "-123.015MB" {
		t.Errorf("should be -123.015MB, got %q", got)
	}
	b = 1*Exabyte + 15*Petabyte
	if got := b.String(); got != "1.015EB" {
		t.Errorf("should be 1.015EB, got %q", got)
	}
}
