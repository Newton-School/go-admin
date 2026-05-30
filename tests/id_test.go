package admin_test

import (
	"testing"

	. "github.com/Newton-School/go-admin"
)

func TestInt64IDCodecParsesAndFormats(t *testing.T) {
	codec := Int64ID()

	id, err := codec.Parse("42")
	if err != nil {
		t.Fatalf("parse int64 id: %v", err)
	}
	if id != 42 {
		t.Fatalf("expected 42, got %d", id)
	}
	if got := codec.Format(id); got != "42" {
		t.Fatalf("expected formatted id 42, got %q", got)
	}

	if _, err := codec.Parse("nope"); err == nil {
		t.Fatal("expected parse error for invalid int64 id")
	}
}

func TestStringIDCodecRejectsEmptyValues(t *testing.T) {
	codec := StringID()

	id, err := codec.Parse("abc-123")
	if err != nil {
		t.Fatalf("parse string id: %v", err)
	}
	if id != "abc-123" {
		t.Fatalf("expected abc-123, got %q", id)
	}
	if got := codec.Format(id); got != "abc-123" {
		t.Fatalf("expected formatted id abc-123, got %q", got)
	}

	if _, err := codec.Parse(""); err == nil {
		t.Fatal("expected empty id error")
	}
}
