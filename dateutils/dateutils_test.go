package dateutils

import (
	"testing"
	"time"
)

func TestDateUtils(t *testing.T) {
	layout := "2006-01-02"

	// IsValid
	if !IsValid("2026-05-25", layout) {
		t.Error("expected valid date format")
	}
	if IsValid("25/05/2026", layout) {
		t.Error("expected invalid date format")
	}

	// DateRelativeToToday
	// "2020-01-01" is in the past
	rel, err := DateRelativeToToday("2020-01-01", layout)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rel != -1 {
		t.Errorf("expected -1, got %d", rel)
	}

	_, err = DateRelativeToToday("invalid", layout)
	if err == nil {
		t.Error("expected error for invalid format")
	}

	// GetDaysBetween
	days, err := GetDaysBetween("2026-05-20", "2026-05-25", layout, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if days != 5 {
		t.Errorf("expected 5 days, got: %d", days)
	}

	daysWithEnd, err := GetDaysBetween("2026-05-20", "2026-05-25", layout, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if daysWithEnd != 6 {
		t.Errorf("expected 6 days, got: %d", daysWithEnd)
	}

	// AddDays
	added, err := AddDays("2026-05-20", layout, 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if added != "2026-05-25" {
		t.Errorf("expected 2026-05-25, got: %s", added)
	}

	// GetDayOfWeek
	// 2026-05-25 is a Monday
	day, err := GetDayOfWeek("2026-05-25", layout, "tr_TR")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if day != "Pazartesi" {
		t.Errorf("expected Pazartesi, got: %s", day)
	}

	// GetLastDayOfMonth
	lastDay, err := GetLastDayOfMonth("2026-02-15", layout)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if lastDay != "2026-02-28" { // 2026 is not a leap year
		t.Errorf("expected 2026-02-28, got: %s", lastDay)
	}

	// CalculateAge
	// If born 10 years ago today
	tenYearsAgo := time.Now().AddDate(-10, 0, 0).Format("2006-01-02")
	age, err := CalculateAge(tenYearsAgo, layout)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if age != 10 {
		t.Errorf("expected 10 years old, got: %d", age)
	}

	// GetQuarter
	q, err := GetQuarter("2026-05-25", layout)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if q != 2 { // May is in Q2
		t.Error("expected Q2")
	}

	// IsBusinessDay
	isBiz, err := IsBusinessDay("2026-05-25", layout) // Monday
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !isBiz {
		t.Error("expected Monday to be business day")
	}

	isBizSun, err := IsBusinessDay("2026-05-24", layout) // Sunday
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if isBizSun {
		t.Error("expected Sunday not to be business day")
	}

	// CountBusinessDays
	// From Monday 2026-05-18 to Friday 2026-05-22 should have 5 business days (inclusive)
	bizDays, err := CountBusinessDays("2026-05-18", "2026-05-22", layout, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if bizDays != 5 {
		t.Errorf("expected 5 business days, got: %d", bizDays)
	}

	// From Friday 2026-05-22 to Monday 2026-05-25 should have 2 business days (inclusive: Friday, Monday)
	bizDaysSpan, err := CountBusinessDays("2026-05-22", "2026-05-25", layout, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if bizDaysSpan != 2 {
		t.Errorf("expected 2 business days, got: %d", bizDaysSpan)
	}
}
