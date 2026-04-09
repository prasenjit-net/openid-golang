package session

import (
	"testing"
	"time"

	"github.com/prasenjit-net/openid-golang/pkg/models"
)

func TestUserSession_IsAuthenticated(t *testing.T) {
	tests := []struct {
		name    string
		session *models.UserSession
		want    bool
	}{
		{
			name: "authenticated and not expired",
			session: &models.UserSession{
				UserID:    "user123",
				ExpiresAt: time.Now().Add(1 * time.Hour),
			},
			want: true,
		},
		{
			name: "expired session",
			session: &models.UserSession{
				UserID:    "user123",
				ExpiresAt: time.Now().Add(-1 * time.Hour),
			},
			want: false,
		},
		{
			name: "no user ID",
			session: &models.UserSession{
				UserID:    "",
				ExpiresAt: time.Now().Add(1 * time.Hour),
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.session.IsAuthenticated(); got != tt.want {
				t.Errorf("IsAuthenticated() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUserSession_IsAuthTimeFresh(t *testing.T) {
	tests := []struct {
		name     string
		authTime time.Time
		maxAge   int
		want     bool
	}{
		{
			name:     "fresh auth within max_age",
			authTime: time.Now().Add(-30 * time.Second),
			maxAge:   60,
			want:     true,
		},
		{
			name:     "stale auth exceeds max_age",
			authTime: time.Now().Add(-120 * time.Second),
			maxAge:   60,
			want:     false,
		},
		{
			name:     "max_age zero should return false",
			authTime: time.Now(),
			maxAge:   0,
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session := &models.UserSession{
				AuthTime: tt.authTime,
			}
			if got := session.IsAuthTimeFresh(tt.maxAge); got != tt.want {
				t.Errorf("IsAuthTimeFresh() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGenerateSessionID(t *testing.T) {
	// Test that generated IDs are unique
	id1, err := generateSessionID()
	if err != nil {
		t.Fatalf("generateSessionID() error = %v", err)
	}

	id2, err := generateSessionID()
	if err != nil {
		t.Fatalf("generateSessionID() error = %v", err)
	}

	if id1 == id2 {
		t.Error("generateSessionID() produced duplicate IDs")
	}

	if len(id1) == 0 {
		t.Error("generateSessionID() produced empty ID")
	}
}
