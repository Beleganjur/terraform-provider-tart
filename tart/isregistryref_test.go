package tart

import "testing"

func TestIsRegistryRef(t *testing.T) {
	cases := []struct{
		in string
		want bool
		name string
	}{
		{"", false, "empty"},
		{"local", false, "local name"},
		{"debian-13", false, "dash local"},
		{"img:tag", true, "has colon tag"},
		{"org/img", true, "has slash"},
		{"org/img:tag", true, "org slash and tag"},
		{"ghcr.io/org/img:tag", true, "fqdn"},
		{"https://example.com/img", true, "url scheme"},
		{"http://example.com/img:tag", true, "url scheme with tag"},
	}
	for _, tc := range cases {
		if got := isRegistryRef(tc.in); got != tc.want {
			t.Errorf("%s: isRegistryRef(%q)=%v, want %v", tc.name, tc.in, got, tc.want)
		}
	}
}
