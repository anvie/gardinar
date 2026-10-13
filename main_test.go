package main

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestWebhookHandler(t *testing.T) {
	cfg := &Config{SecretKey: "secret", Tasks: map[string]string{"deploy": "deploy-command"}}
	tests := []struct {
		name       string
		key        string
		body       string
		run        TaskRunner
		wantStatus int
		wantBody   string
	}{
		{"success", "secret", `{"task":"deploy","source_dir":"/tmp","params":["restart"]}`, func(string, string, []string) (string, error) { return "", nil }, http.StatusOK, `{"result":"success","state":"succeeded"}`},
		{"task failure", "secret", `{"task":"deploy"}`, func(string, string, []string) (string, error) { return "sensitive output", errors.New("exit status 1") }, http.StatusInternalServerError, `{"state":"failed","failed_step":"task","error":"task execution failed"}`},
		{"unauthorized", "wrong", `{"task":"deploy"}`, nil, http.StatusUnauthorized, `{"error":"unauthorized"}`},
		{"malformed payload", "secret", `{`, nil, http.StatusBadRequest, `{"error":"invalid webhook payload"}`},
		{"missing task", "secret", `{"task":"missing"}`, nil, http.StatusBadRequest, `{"error":"task not found"}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			run := tt.run
			if run == nil {
				run = func(string, string, []string) (string, error) { t.Fatal("runner should not be called"); return "", nil }
			}
			req := httptest.NewRequest(http.MethodPost, "/webhook", strings.NewReader(tt.body))
			req.Header.Set("X-SECRET-KEY", tt.key)
			rec := httptest.NewRecorder()
			NewWebhookHandler(cfg, run).ServeHTTP(rec, req)
			if rec.Code != tt.wantStatus || strings.TrimSpace(rec.Body.String()) != tt.wantBody {
				t.Fatalf("status/body = %d/%s, want %d/%s", rec.Code, rec.Body.String(), tt.wantStatus, tt.wantBody)
			}
		})
	}
}
