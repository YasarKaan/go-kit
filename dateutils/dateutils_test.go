package dateutils

import (
	"testing"
	"time"
)

func TestDateUtilsAutoDetect(t *testing.T) {
	// IsValid
	if !IsValid("2026-05-25") {
		t.Error("expected valid ISO8601 date")
	}
	if !IsValid("25/05/2026") {
		t.Error("expected valid DMY date")
	}
	if IsValid("invalid-date") {
		t.Error("expected invalid")
	}

	// DateRelativeToToday
	// "2020-01-01" is in the past
	rel, err := DateRelativeToToday("2020-01-01")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rel != -1 {
		t.Errorf("expected -1, got %d", rel)
	}

	// GetDaysBetween
	days, err := GetDaysBetween("2026-05-20", "2026-05-25", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if days != 5 {
		t.Errorf("expected 5 days, got: %d", days)
	}

	daysWithEnd, err := GetDaysBetween("2026-05-20", "2026-05-25", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if daysWithEnd != 6 {
		t.Errorf("expected 6 days, got: %d", daysWithEnd)
	}

	// AddDays (checks preservation of format layout)
	addedDash, err := AddDays("2026-05-20", 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if addedDash != "2026-05-25" {
		t.Errorf("expected 2026-05-25, got: %s", addedDash)
	}

	addedSlash, err := AddDays("20/05/2026", 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if addedSlash != "25/05/2026" {
		t.Errorf("expected 25/05/2026, got: %s", addedSlash)
	}

	// GetDayOfWeek
	// 2026-05-25 is a Monday
	day, err := GetDayOfWeek("2026-05-25", "tr_TR")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if day != "Pazartesi" {
		t.Errorf("expected Pazartesi, got: %s", day)
	}

	// GetLastDayOfMonth
	lastDay, err := GetLastDayOfMonth("2026-02-15")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if lastDay != "2026-02-28" { // 2026 is not a leap year
		t.Errorf("expected 2026-02-28, got: %s", lastDay)
	}

	// CalculateAge
	tenYearsAgo := time.Now().AddDate(-10, 0, 0).Format("2006-01-02")
	age, err := CalculateAge(tenYearsAgo)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if age != 10 {
		t.Errorf("expected 10 years old, got: %d", age)
	}

	// GetQuarter
	q, err := GetQuarter("2026-05-25")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if q != 2 { // May is in Q2
		t.Error("expected Q2")
	}

	// IsBusinessDay
	isBiz, err := IsBusinessDay("2026-05-25") // Monday
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !isBiz {
		t.Error("expected Monday to be business day")
	}

	isBizSun, err := IsBusinessDay("2026-05-24") // Sunday
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if isBizSun {
		t.Error("expected Sunday not to be business day")
	}

	// CountBusinessDays
	// From Monday 2026-05-18 to Friday 2026-05-22 should have 5 business days (inclusive)
	bizDays, err := CountBusinessDays("2026-05-18", "2026-05-22", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if bizDays != 5 {
		t.Errorf("expected 5 business days, got: %d", bizDays)
	}

	// From Friday 2026-05-22 to Monday 2026-05-25 should have 2 business days (inclusive: Friday, Monday)
	bizDaysSpan, err := CountBusinessDays("2026-05-22", "2026-05-25", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if bizDaysSpan != 2 {
		t.Errorf("expected 2 business days, got: %d", bizDaysSpan)
	}
}
