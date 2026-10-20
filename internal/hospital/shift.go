package hospital

import "time"

type Shift struct {
	startHour int
	endHour   int
}

func NewShift(startHour, endHour int) Shift {
	return Shift{
		startHour: startHour,
		endHour:   endHour,
	}
}

func (s Shift) IsActive(t time.Time) bool {
	hour := t.Hour()
	if s.startHour <= s.endHour {
		return hour >= s.startHour && hour < s.endHour
	}
	return hour >= s.startHour || hour < s.endHour
}
