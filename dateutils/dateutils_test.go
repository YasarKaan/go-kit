package dateutils

import (
	"testing"
	"time"
)

func TestDateUtils(t *testing.T) {
	// IsValid
	if !IsValid("2026-05-25", "yyyy-MM-dd") {
		t.Error("expected valid date format")
	}
	if IsValid("25/05/2026", "yyyy-MM-dd") {
		t.Error("expected invalid date format")
	}

	// DateRelativeToToday
	// "2020-01-01" is in the past
	if DateRelativeToToday("2020-01-01", "yyyy-MM-dd") != -1 {
		t.Error("expected date to be in the past")
	}

	// GetDaysBetween
	days := GetDaysBetween("2026-05-20", "2026-05-25", "yyyy-MM-dd", false)
	if days != 5 {
		t.Errorf("expected 5 days, got: %d", days)
	}

	daysWithEnd := GetDaysBetween("2026-05-20", "2026-05-25", "yyyy-MM-dd", true)
	if daysWithEnd != 6 {
		t.Errorf("expected 6 days, got: %d", daysWithEnd)
	}

	// AddDays
	added := AddDays("2026-05-20", "yyyy-MM-dd", 5)
	if added != "2026-05-25" {
		t.Errorf("expected 2026-05-25, got: %s", added)
	}

	// GetDayOfWeek
	// 2026-05-25 is a Monday
	day := GetDayOfWeek("2026-05-25", "yyyy-MM-dd", "tr_TR")
	if day != "Pazartesi" {
		t.Errorf("expected Pazartesi, got: %s", day)
	}

	// GetLastDayOfMonth
	lastDay := GetLastDayOfMonth("2026-02-15", "yyyy-MM-dd")
	if lastDay != "2026-02-28" { // 2026 is not a leap year
		t.Errorf("expected 2026-02-28, got: %s", lastDay)
	}

	// CalculateAge
	// If born 10 years ago today
	tenYearsAgo := time.Now().AddDate(-10, 0, 0).Format("2006-01-02")
	age := CalculateAge(tenYearsAgo, "yyyy-MM-dd")
	if age != 10 {
		t.Errorf("expected 10 years old, got: %d", age)
	}

	// GetQuarter
	if GetQuarter("2026-05-25", "yyyy-MM-dd") != 2 { // May is in Q2
		t.Error("expected Q2")
	}

	// IsBusinessDay
	if !IsBusinessDay("2026-05-25", "yyyy-MM-dd") { // Monday
		t.Error("expected Monday to be business day")
	}
	if IsBusinessDay("2026-05-24", "yyyy-MM-dd") { // Sunday
		t.Error("expected Sunday not to be business day")
	}

	// CountBusinessDays
	// From Monday 2026-05-18 to Friday 2026-05-22 should have 5 business days (inclusive)
	bizDays := CountBusinessDays("2026-05-18", "2026-05-22", "yyyy-MM-dd", true)
	if bizDays != 5 {
		t.Errorf("expected 5 business days, got: %d", bizDays)
	}

	// From Friday 2026-05-22 to Monday 2026-05-25 should have 2 business days (inclusive: Friday, Monday)
	bizDaysSpan := CountBusinessDays("2026-05-22", "2026-05-25", "yyyy-MM-dd", true)
	if bizDaysSpan != 2 {
		t.Errorf("expected 2 business days, got: %d", bizDaysSpan)
	}
}
