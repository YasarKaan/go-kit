package dateutils

import (
	"strings"
	"time"
)

// convertJavaLayout translates standard Java date/time format strings (like yyyy-MM-dd)
// into Go time package layouts (like 2006-01-02).
func convertJavaLayout(javaLayout string) string {
	r := strings.NewReplacer(
		"yyyy", "2006",
		"yy", "06",
		"MM", "01",
		"dd", "02",
		"HH", "15",
		"mm", "04",
		"ss", "05",
	)
	return r.Replace(javaLayout)
}

// IsValid checks if the date string conforms to the specified layout.
func IsValid(dateAsString, dateFormat string) bool {
	layout := convertJavaLayout(dateFormat)
	_, err := time.Parse(layout, dateAsString)
	return err == nil
}

// DateRelativeToToday checks if the date is in the past (-1), today (0), or in the future (1).
func DateRelativeToToday(date, dateFormat string) int {
	layout := convertJavaLayout(dateFormat)
	inputTime, err := time.Parse(layout, date)
	if err != nil {
		return -2
	}

	now := time.Now()
	// Strip time component to compare date-only (like LocalDate)
	inputDate := time.Date(inputTime.Year(), inputTime.Month(), inputTime.Day(), 0, 0, 0, 0, time.Local)
	todayDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)

	if inputDate.Before(todayDate) {
		return -1
	} else if inputDate.Equal(todayDate) {
		return 0
	} else if inputDate.After(todayDate) {
		return 1
	}
	return -2
}

// GetDaysBetween returns the number of days between two dates.
func GetDaysBetween(startDate, endDate, dateFormat string, includeEndDate bool) int {
	layout := convertJavaLayout(dateFormat)
	start, err1 := time.Parse(layout, startDate)
	end, err2 := time.Parse(layout, endDate)
	if err1 != nil || err2 != nil {
		return 0
	}

	if start.After(end) {
		return 0
	}

	duration := end.Sub(start)
	days := int(duration.Hours() / 24)

	if includeEndDate {
		return days + 1
	}
	return days
}

// AddDays adds a number of days to a date string.
func AddDays(date, dateFormat string, daysToAdd int64) string {
	layout := convertJavaLayout(dateFormat)
	t, err := time.Parse(layout, date)
	if err != nil {
		return ""
	}
	t = t.AddDate(0, 0, int(daysToAdd))
	return t.Format(layout)
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
func GetDayOfWeek(date, dateFormat string, locale string) string {
	layout := convertJavaLayout(dateFormat)
	t, err := time.Parse(layout, date)
	if err != nil {
		return ""
	}

	englishDay := t.Weekday().String()
	if strings.HasPrefix(strings.ToLower(locale), "tr") {
		if trDay, ok := daysTR[englishDay]; ok {
			return trDay
		}
	}
	return englishDay
}

// GetLastDayOfMonth returns the date string corresponding to the last day of that month.
func GetLastDayOfMonth(date, dateFormat string) string {
	layout := convertJavaLayout(dateFormat)
	t, err := time.Parse(layout, date)
	if err != nil {
		return ""
	}

	// Go to first day of next month, then subtract one day
	firstOfNextMonth := time.Date(t.Year(), t.Month()+1, 1, 0, 0, 0, 0, t.Location())
	lastDay := firstOfNextMonth.AddDate(0, 0, -1)

	return lastDay.Format(layout)
}

// CalculateAge calculates age in years based on birthdate.
func CalculateAge(birthDate, dateFormat string) int {
	layout := convertJavaLayout(dateFormat)
	birth, err := time.Parse(layout, birthDate)
	if err != nil {
		return 0
	}

	now := time.Now()
	if birth.After(now) {
		return 0
	}

	years := now.Year() - birth.Year()

	// Adjust if birthday hasn't occurred yet this year
	birthMonthDay := birth.Month()*100 + time.Month(birth.Day())
	nowMonthDay := now.Month()*100 + time.Month(now.Day())
	if nowMonthDay < birthMonthDay {
		years--
	}
	return years
}

// GetQuarter returns the quarter of the year (1, 2, 3, or 4).
func GetQuarter(date, dateFormat string) int {
	layout := convertJavaLayout(dateFormat)
	t, err := time.Parse(layout, date)
	if err != nil {
		return 0
	}
	return int((t.Month()-1)/3) + 1
}

// IsBusinessDay checks if the date is a weekday (Monday to Friday).
func IsBusinessDay(date, dateFormat string) bool {
	layout := convertJavaLayout(dateFormat)
	t, err := time.Parse(layout, date)
	if err != nil {
		return false
	}
	return isBusinessDayTime(t)
}

func isBusinessDayTime(t time.Time) bool {
	wd := t.Weekday()
	return wd != time.Saturday && wd != time.Sunday
}

// CountBusinessDays counts weekdays between two dates.
func CountBusinessDays(startDate, endDate, dateFormat string, includeEndDate bool) int {
	layout := convertJavaLayout(dateFormat)
	start, err1 := time.Parse(layout, startDate)
	end, err2 := time.Parse(layout, endDate)
	if err1 != nil || err2 != nil {
		return 0
	}

	if start.After(end) {
		return 0
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

	return count
}
