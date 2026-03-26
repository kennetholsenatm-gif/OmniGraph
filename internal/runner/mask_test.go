package runner

import (
	"bytes"
	"testing"
)

func TestRedact_longestFirst(t *testing.T) {
	data := []byte("prefix abcdef secret abc suffix")
	got := Redact(data, []string{"abc", "abcdef"}, nil)
	want := []byte("prefix *** secret *** suffix")
	if !bytes.Equal(got, want) {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestRedact_envStyleSecret(t *testing.T) {
	data := []byte("token=supersecret and supersecret again")
	got := Redact(data, []string{"supersecret"}, nil)
	want := []byte("token=*** and *** again")
	if !bytes.Equal(got, want) {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestRedact_skipsEmpty(t *testing.T) {
	data := []byte("hello")
	got := Redact(data, []string{"", "ell"}, nil)
	want := []byte("h***o")
	if !bytes.Equal(got, want) {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestRedact_customRepl(t *testing.T) {
	got := Redact([]byte("xabcy"), []string{"abc"}, []byte("[REDACTED]"))
	want := []byte("x[REDACTED]y")
	if !bytes.Equal(got, want) {
		t.Fatalf("got %q want %q", got, want)
	}
}
