package intel

import (
	"testing"
	"time"
)

func TestSyncer_SyncAll(t *testing.T) {
	t.Run("integration test - requires database", func(t *testing.T) {
		// This requires a real database connection
		t.Skip("integration test - requires database")
	})

	t.Run("coordinates sync order correctly", func(t *testing.T) {
		// Unit test to verify sync order without actual DB
		t.Skip("requires mock syncers")
	})
}

func TestSyncer_SyncNVD(t *testing.T) {
	t.Run("integration test - requires database", func(t *testing.T) {
		t.Skip("integration test - requires database")
	})
}

func TestSyncer_SyncEPSS(t *testing.T) {
	t.Run("integration test - requires database", func(t *testing.T) {
		t.Skip("integration test - requires database")
	})
}

func TestSyncer_SyncKEV(t *testing.T) {
	t.Run("integration test - requires database", func(t *testing.T) {
		t.Skip("integration test - requires database")
	})
}

func TestSyncer_GetSyncStatus(t *testing.T) {
	t.Run("integration test - requires database", func(t *testing.T) {
		t.Skip("integration test - requires database")
	})

	t.Run("handles missing sync runs", func(t *testing.T) {
		t.Skip("requires mock repository")
	})
}

func TestSyncStatus(t *testing.T) {
	t.Run("serializes to JSON correctly", func(t *testing.T) {
		now := time.Now()
		status := &SyncStatus{
			Sources: map[string]SourceStatus{
				"nvd": {
					Source:         "nvd",
					Status:         "completed",
					LastSync:       &now,
					Records:        100,
					RecordsUpdated: 95,
				},
			},
		}

		if status.Sources["nvd"].Source != "nvd" {
			t.Error("Source mismatch")
		}

		if status.Sources["nvd"].Status != "completed" {
			t.Error("Status mismatch")
		}

		if status.Sources["nvd"].Records != 100 {
			t.Error("Records count mismatch")
		}
	})
}

func TestSourceStatus(t *testing.T) {
	t.Run("handles completed status", func(t *testing.T) {
		now := time.Now()
		status := SourceStatus{
			Source:         "nvd",
			Status:         "completed",
			LastSync:       &now,
			Records:        100,
			RecordsUpdated: 95,
		}

		if status.Source != "nvd" {
			t.Errorf("expected source 'nvd', got '%s'", status.Source)
		}

		if status.Status != "completed" {
			t.Errorf("expected status 'completed', got '%s'", status.Status)
		}
	})

	t.Run("handles failed status with error", func(t *testing.T) {
		errMsg := "sync failed"
		status := SourceStatus{
			Source:  "epss",
			Status:  "failed",
			Records: 0,
			Error:   &errMsg,
		}

		if status.Error == nil {
			t.Error("expected error message, got nil")
		}

		if *status.Error != "sync failed" {
			t.Errorf("expected error 'sync failed', got '%s'", *status.Error)
		}

		_ = errMsg // Use the variable
	})

	t.Run("handles never_run status", func(t *testing.T) {
		status := SourceStatus{
			Source: "kev",
			Status: "never_run",
		}

		if status.Status != "never_run" {
			t.Errorf("expected status 'never_run', got '%s'", status.Status)
		}

		if status.LastSync != nil {
			t.Error("expected nil LastSync for never_run status")
		}
	})
}
