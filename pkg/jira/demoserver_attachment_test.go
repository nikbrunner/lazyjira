//go:build demo

package jira

import (
	"testing"
)

func TestDemoServer_AcceptsAttachmentUpload(t *testing.T) {
	t.Parallel()
	srv, err := NewDemoServer()
	if err != nil {
		t.Fatalf("NewDemoServer: %v", err)
	}
	t.Cleanup(func() { _ = srv.Close() })
	client := NewClientWithOpts(ClientOpts{Host: srv.URL, Email: "demo@lazyjira.dev", Token: "demo", IsCloud: true})

	if err := client.AddAttachment(t.Context(), "PLAT-1", "shot.png", []byte("PNGDATA")); err != nil {
		t.Fatalf("AddAttachment against the demo server: %v", err)
	}
}
