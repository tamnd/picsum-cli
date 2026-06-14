package picsum

import (
	"testing"
)

// These tests are offline: they exercise the URI driver's pure string functions.
// HTTP behaviour is covered in picsum_test.go.

func TestDomainInfo(t *testing.T) {
	info := Domain{}.Info()
	if info.Scheme != "picsum" {
		t.Errorf("Scheme = %q, want picsum", info.Scheme)
	}
	if len(info.Hosts) == 0 || info.Hosts[0] != Host {
		t.Errorf("Hosts = %v, want [%s]", info.Hosts, Host)
	}
	if info.Identity.Binary != "picsum" {
		t.Errorf("Identity.Binary = %q, want picsum", info.Identity.Binary)
	}
}

func TestClassify(t *testing.T) {
	_, _, err := Domain{}.Classify("")
	if err == nil {
		t.Error("expected error for empty input, got nil")
	}

	typ, id, err := Domain{}.Classify("42")
	if err != nil || typ != "image" || id != "42" {
		t.Errorf("Classify = (%q, %q, %v), want (image, 42, nil)", typ, id, err)
	}

	typ, id, err = Domain{}.Classify("  0  ")
	if err != nil || typ != "image" || id != "0" {
		t.Errorf("Classify trimmed = (%q, %q, %v), want (image, 0, nil)", typ, id, err)
	}
}

func TestLocate(t *testing.T) {
	got, err := Domain{}.Locate("image", "42")
	want := "https://picsum.photos/id/42/info"
	if err != nil || got != want {
		t.Errorf("Locate = (%q, %v), want (%q, nil)", got, err, want)
	}
}

func TestLocateUnknownType(t *testing.T) {
	_, err := Domain{}.Locate("unknown", "foo")
	if err == nil {
		t.Error("expected error for unknown type, got nil")
	}
}
