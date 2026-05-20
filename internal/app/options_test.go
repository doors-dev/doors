package app

import "testing"

func TestOptionsDefaults(t *testing.T) {
	o := Options{}
	o.initDefaults()
	if o.ID != "doors" {
		t.Fatalf("unexpected default server id: %q", o.ID)
	}
	if o.Conf.ServerSessionCookiePrefix != "" {
		t.Fatalf("unexpected default session cookie prefix: %q", o.Conf.ServerSessionCookiePrefix)
	}
}
