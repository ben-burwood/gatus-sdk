package gatussdk

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestGetAnnouncements(t *testing.T) {
	tests := []struct {
		name           string
		mockResponse   string
		mockStatusCode int
		expected       []Announcement
		expectError    bool
	}{
		{
			name:           "announcements configured",
			mockResponse:   `{"oidc":false,"authenticated":true,"announcements":[{"timestamp":"2025-11-07T14:00:00Z","type":"outage","message":"Database is down"},{"timestamp":"2025-11-01T09:30:00Z","type":"operational","message":"Maintenance completed","archived":true}]}`,
			mockStatusCode: http.StatusOK,
			expected: []Announcement{
				{
					Timestamp: time.Date(2025, 11, 7, 14, 0, 0, 0, time.UTC),
					Type:      AnnouncementTypeOutage,
					Message:   "Database is down",
				},
				{
					Timestamp: time.Date(2025, 11, 1, 9, 30, 0, 0, time.UTC),
					Type:      AnnouncementTypeOperational,
					Message:   "Maintenance completed",
					Archived:  true,
				},
			},
		},
		{
			name:           "no announcements configured",
			mockResponse:   `{"oidc":false,"authenticated":true,"announcements":[]}`,
			mockStatusCode: http.StatusOK,
			expected:       []Announcement{},
		},
		{
			name:           "gatus version without announcement support",
			mockResponse:   `{"oidc":false,"authenticated":true}`,
			mockStatusCode: http.StatusOK,
			expected:       []Announcement{},
		},
		{
			name:           "server error",
			mockResponse:   `internal error`,
			mockStatusCode: http.StatusInternalServerError,
			expectError:    true,
		},
		{
			name:           "invalid JSON",
			mockResponse:   `{"announcements":`,
			mockStatusCode: http.StatusOK,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/api/v1/config" {
					t.Errorf("unexpected path: %s", r.URL.Path)
				}
				if r.Method != http.MethodGet {
					t.Errorf("unexpected method: %s", r.Method)
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatusCode)
				w.Write([]byte(tt.mockResponse))
			}))
			defer server.Close()

			client := NewClient(server.URL)
			announcements, err := client.GetAnnouncements(context.Background())

			if tt.expectError {
				if err == nil {
					t.Fatal("expected error but got none")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if announcements == nil {
				t.Fatal("expected non-nil slice")
			}
			if len(announcements) != len(tt.expected) {
				t.Fatalf("got %d announcements, want %d", len(announcements), len(tt.expected))
			}
			for i, want := range tt.expected {
				got := announcements[i]
				if !got.Timestamp.Equal(want.Timestamp) || got.Type != want.Type || got.Message != want.Message || got.Archived != want.Archived {
					t.Errorf("announcement[%d] = %+v, want %+v", i, got, want)
				}
			}
		})
	}
}

func TestGetAnnouncements_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	_, err := client.GetAnnouncements(context.Background())

	var apiErr *APIError
	if !errors.As(err, &apiErr) || apiErr.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404 APIError, got %v", err)
	}
}
