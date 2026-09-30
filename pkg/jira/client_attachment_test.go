package jira

import (
	"bytes"
	"errors"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"strings"
	"testing"

	"github.com/nikbrunner/lazyjira/pkg/internal/testkit"
)

func TestClient_AddAttachment_PostsMultipartFile(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name     string
		opts     ClientOpts
		path     string
		authPref string
	}{
		{"cloud", cloudOpts(), "/rest/api/3/issue/PLAT-9/attachments", "Basic "},
		{"server", serverOpts(), "/rest/api/2/issue/PLAT-9/attachments", "Bearer "},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			client, recorded := newRecordingClient(t, tc.opts, testkit.StubResponse{Status: http.StatusOK, Body: `[{"id":"1"}]`})

			if err := client.AddAttachment(t.Context(), "PLAT-9", "shot.png", []byte("PNGDATA")); err != nil {
				t.Fatalf("AddAttachment: %v", err)
			}

			testkit.AssertEqual(t, "method", recorded.Method, http.MethodPost)
			testkit.AssertEqual(t, "path", recorded.Path, tc.path)
			testkit.AssertEqual(t, "xsrf header", recorded.Header.Get("X-Atlassian-Token"), "no-check")
			if auth := recorded.Header.Get("Authorization"); !strings.HasPrefix(auth, tc.authPref) {
				t.Errorf("Authorization = %q, want %q prefix", auth, tc.authPref)
			}
			mediaType, params, err := mime.ParseMediaType(recorded.Header.Get("Content-Type"))
			if err != nil || mediaType != "multipart/form-data" {
				t.Fatalf("Content-Type = %q, want multipart/form-data", recorded.Header.Get("Content-Type"))
			}
			part, err := multipart.NewReader(bytes.NewReader(recorded.Body), params["boundary"]).NextPart()
			if err != nil {
				t.Fatalf("read part: %v", err)
			}
			testkit.AssertEqual(t, "field name", part.FormName(), "file")
			testkit.AssertEqual(t, "filename", part.FileName(), "shot.png")
			data, _ := io.ReadAll(part)
			testkit.AssertEqual(t, "body", string(data), "PNGDATA")
		})
	}
}

func TestClient_AddAttachment_WrapsAPIError(t *testing.T) {
	t.Parallel()
	client, _ := newRecordingClient(t, cloudOpts(), testkit.StubResponse{Status: http.StatusRequestEntityTooLarge, Body: `{"errorMessages":["too big"]}`})

	err := client.AddAttachment(t.Context(), "PLAT-9", "shot.png", []byte("x"))

	if err == nil {
		t.Fatal("expected error")
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) || apiErr.Status != http.StatusRequestEntityTooLarge {
		t.Errorf("err = %v, want a 413 APIError", err)
	}
}
