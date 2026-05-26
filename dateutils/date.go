package dateutils

import (
	"fmt"
	"strings"
	"time"
)

// IsValid checks if the date string conforms to the specified layout.
func IsValid(dateAsString, layout string) bool {
	_, err := time.Parse(layout, dateAsString)
	return err == nil
}

// DateRelativeToToday checks if the date is in the past (-1), today (0), or in the future (1).
func DateRelativeToToday(date, layout string) (int, error) {
	inputTime, err := time.Parse(layout, date)
	if err != nil {
		return 0, err
	}

	now := time.Now()
	// Strip time component to compare date-only
	inputDate := time.Date(inputTime.Year(), inputTime.Month(), inputTime.Day(), 0, 0, 0, 0, time.Local)
	todayDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)

	if inputDate.Before(todayDate) {
		return -1, nil
	} else if inputDate.Equal(todayDate) {
		return 0, nil
	}
	return 1, nil
}

// GetDaysBetween returns the number of days between two dates.
func GetDaysBetween(startDate, endDate, layout string, includeEndDate bool) (int, error) {
	start, err := time.Parse(layout, startDate)
	if err != nil {
		return 0, err
	}
	end, err := time.Parse(layout, endDate)
	if err != nil {
		return 0, err
	}

	if start.After(end) {
		return 0, nil
	}

	duration := end.Sub(start)
	days := int(duration.Hours() / 24)

	if includeEndDate {
		return days + 1, nil
	}
	return days, nil
}

// AddDays adds a number of days to a date string.
func AddDays(date, layout string, daysToAdd int) (string, error) {
	t, err := time.Parse(layout, date)
	if err != nil {
		return "", err
	}
	t = t.AddDate(0, 0, daysToAdd)
	return t.Format(layout), nil
}

var daysTR = map[string]string{
	"Monday":    "Pazartesi",
	"Tuesday":   "Salı",
	"Wednesday": "Çarşamba",
	"Thursday":  "Perşembe",
	"Friday":    "Cuma",
	"Saturday":  "Cumartesi",
	"Sunday":    "Pazar",
}

// GetDayOfWeek returns the name of the day of the week, with simple locale translation (supports "tr").
func GetDayOfWeek(date, layout string, locale string) (string, error) {
	t, err := time.Parse(layout, date)
	if err != nil {
		return "", err
	}

	englishDay := t.Weekday().String()
	if strings.HasPrefix(strings.ToLower(locale), "tr") {
		if trDay, ok := daysTR[englishDay]; ok {
			return trDay, nil
		}
	}
	return englishDay, nil
}

// GetLastDayOfMonth returns the date string corresponding to the last day of that month.
func GetLastDayOfMonth(date, layout string) (string, error) {
	t, err := time.Parse(layout, date)
	if err != nil {
		return "", err
	}

	// Go to first day of next month, then subtract one day
	firstOfNextMonth := time.Date(t.Year(), t.Month()+1, 1, 0, 0, 0, 0, t.Location())
	lastDay := firstOfNextMonth.AddDate(0, 0, -1)

	return lastDay.Format(layout), nil
}

// CalculateAge calculates age in years based on birthdate.
func CalculateAge(birthDate, layout string) (int, error) {
	birth, err := time.Parse(layout, birthDate)
	if err != nil {
		return 0, err
	}

	now := time.Now()
	if birth.After(now) {
		return 0, fmt.Errorf("birth date %s is in the future", birthDate)
	}

	years := now.Year() - birth.Year()

	// Adjust if birthday hasn't occurred yet this year
	birthMonthDay := birth.Month()*100 + time.Month(birth.Day())
	nowMonthDay := now.Month()*100 + time.Month(now.Day())
	if nowMonthDay < birthMonthDay {
		years--
	}
	return years, nil
}

// GetQuarter returns the quarter of the year (1, 2, 3, or 4).
func GetQuarter(date, layout string) (int, error) {
	t, err := time.Parse(layout, date)
	if err != nil {
		return 0, err
	}
	return int((t.Month()-1)/3) + 1, nil
}

// IsBusinessDay checks if the date is a weekday (Monday to Friday).
func IsBusinessDay(date, layout string) (bool, error) {
	t, err := time.Parse(layout, date)
	if err != nil {
		return false, err
	}
	return isBusinessDayTime(t), nil
}

func isBusinessDayTime(t time.Time) bool {
	wd := t.Weekday()
	return wd != time.Saturday && wd != time.Sunday
}

// CountBusinessDays counts weekdays between two dates.
func CountBusinessDays(startDate, endDate, layout string, includeEndDate bool) (int, error) {
	start, err := time.Parse(layout, startDate)
	if err != nil {
		return 0, err
	}
	end, err := time.Parse(layout, endDate)
	if err != nil {
		return 0, err
	}

	if start.After(end) {
		return 0, nil
	}

	count := 0
	current := start

	// Go date-only bounds
	endVal := time.Date(end.Year(), end.Month(), end.Day(), 0, 0, 0, 0, time.UTC)

	for {
		currentVal := time.Date(current.Year(), current.Month(), current.Day(), 0, 0, 0, 0, time.UTC)
		if currentVal.After(endVal) {
			break
		}
		if currentVal.Equal(endVal) && !includeEndDate {
			break
		}

		if isBusinessDayTime(current) {
			count++
		}
		current = current.AddDate(0, 0, 1)
	}

	return count, nil
}
