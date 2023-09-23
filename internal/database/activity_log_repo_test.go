package database

import "testing"

func TestReportBusyTimes(t *testing.T) {
	repo := setupTestDB()
	defer teardownTestDB(repo)

	busyHours, err := repo.ReportBusyTimes()
	if err != nil {
		t.Fatalf("error getting busy times report: %v", err)
	}
	if len(busyHours) == 0 {
		t.Fatalf("expected some busy hours, got none")
	}

	for hour, count := range busyHours {
		t.Logf("Busy hour: %v", hour)
		t.Logf("Busy count: %v", count)
	}
}
