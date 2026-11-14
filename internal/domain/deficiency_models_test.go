package domain_test

import (
	"testing"
	"time"

	"homolog/internal/domain"
)

func TestCalculateDeadlineAndUrgency(t *testing.T) {
	now := time.Date(2026, 11, 14, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name        string
		dueDate     time.Time
		wantBadge   string
		wantMinDays int
		wantMaxDays int
	}{
		{
			name:        "Critical deadline (7 days left)",
			dueDate:     now.AddDate(0, 0, 7),
			wantBadge:   "critical",
			wantMinDays: 6,
			wantMaxDays: 7,
		},
		{
			name:        "Urgent deadline (20 days left)",
			dueDate:     now.AddDate(0, 0, 20),
			wantBadge:   "urgent",
			wantMinDays: 19,
			wantMaxDays: 20,
		},
		{
			name:        "Warning deadline (45 days left)",
			dueDate:     now.AddDate(0, 0, 45),
			wantBadge:   "warning",
			wantMinDays: 44,
			wantMaxDays: 45,
		},
		{
			name:        "Normal deadline (90 days left)",
			dueDate:     now.AddDate(0, 0, 90),
			wantBadge:   "normal",
			wantMinDays: 89,
			wantMaxDays: 90,
		},
		{
			name:        "Expired deadline (-2 days)",
			dueDate:     now.AddDate(0, 0, -2),
			wantBadge:   "expired",
			wantMinDays: -3,
			wantMaxDays: -1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			days, badge := domain.CalculateDeadlineAndUrgency(tc.dueDate, now)
			if badge != tc.wantBadge {
				t.Errorf("badge attendu '%s', obtenu '%s'", tc.wantBadge, badge)
			}
			if days < tc.wantMinDays || days > tc.wantMaxDays {
				t.Errorf("jours attendus entre %d et %d, obtenu %d", tc.wantMinDays, tc.wantMaxDays, days)
			}
		})
	}
}

func TestComputeMetrics(t *testing.T) {
	now := time.Date(2026, 11, 14, 12, 0, 0, 0, time.UTC)

	letter := domain.ANPPDeficiencyLetter{
		ID:      "test-letter-1",
		DueDate: now.AddDate(0, 0, 10),
		Status:  domain.DeficiencyLetterStatusOpen,
		Tasks: []domain.ANPPDeficiencyTask{
			{
				ID:      "t1",
				DueDate: now.AddDate(0, 0, 5),
				Status:  domain.DeficiencyTaskStatusAnswered,
			},
			{
				ID:      "t2",
				DueDate: now.AddDate(0, 0, 8),
				Status:  domain.DeficiencyTaskStatusValidatedDT,
			},
			{
				ID:      "t3",
				DueDate: now.AddDate(0, 0, 10),
				Status:  domain.DeficiencyTaskStatusPending,
			},
		},
	}

	letter.ComputeMetrics(now)

	if letter.TotalTasksCount != 3 {
		t.Errorf("TotalTasksCount attendu 3, obtenu %d", letter.TotalTasksCount)
	}
	if letter.AnsweredTasksCount != 2 {
		t.Errorf("AnsweredTasksCount attendu 2, obtenu %d", letter.AnsweredTasksCount)
	}
	if letter.ValidatedTasksCount != 1 {
		t.Errorf("ValidatedTasksCount attendu 1, obtenu %d", letter.ValidatedTasksCount)
	}
	if letter.UrgencyBadge != "critical" {
		t.Errorf("UrgencyBadge attendu 'critical', obtenu '%s'", letter.UrgencyBadge)
	}
	if letter.Status != domain.DeficiencyLetterStatusPartiallyAnswered {
		t.Errorf("Status attendu '%s', obtenu '%s'", domain.DeficiencyLetterStatusPartiallyAnswered, letter.Status)
	}
}
