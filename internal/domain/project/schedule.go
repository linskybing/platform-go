package project

import (
	"fmt"
	"time"
)

type ScheduleWindow struct {
	Weekday int    `json:"weekday"`
	Start   string `json:"start"`
	End     string `json:"end"`
}

func IsTimeAllowed(windows []ScheduleWindow, now time.Time) (bool, error) {
	if len(windows) == 0 {
		return true, nil
	}

	weekday := int(now.Weekday())
	minutes := now.Hour()*60 + now.Minute()

	for _, w := range windows {
		if w.Weekday < 0 || w.Weekday > 6 {
			continue
		}
		startMin, err := parseClockMinutes(w.Start)
		if err != nil {
			return false, err
		}
		endMin, err := parseClockMinutes(w.End)
		if err != nil {
			return false, err
		}
		if startMin == endMin {
			if weekday == w.Weekday {
				return true, nil
			}
			continue
		}
		if endMin > startMin {
			if weekday == w.Weekday && minutes >= startMin && minutes < endMin {
				return true, nil
			}
			continue
		}
		// Overnight window (e.g., 22:00-06:00)
		if weekday == w.Weekday && minutes >= startMin {
			return true, nil
		}
		nextDay := (w.Weekday + 1) % 7
		if weekday == nextDay && minutes < endMin {
			return true, nil
		}
	}
	return false, nil
}

// Helper to fix compilation, ideally replaced by DB logic
func (p *Project) ScheduleWindowList() ([]ScheduleWindow, error) {
	// TODO: Implement conversion from ResourcePlan.WeekWindow or use DB constraints
	return nil, nil
}

func (p *Project) IsTimeAllowed(now time.Time) (bool, error) {
	windows, err := p.ScheduleWindowList()
	if err != nil {
		return false, err
	}
	return IsTimeAllowed(windows, now)
}

func parseClockMinutes(value string) (int, error) {
	if value == "" {
		return 0, fmt.Errorf("schedule window time is empty")
	}
	parsed, err := time.Parse("15:04", value)
	if err != nil {
		return 0, fmt.Errorf("invalid schedule window time: %s", value)
	}
	return parsed.Hour()*60 + parsed.Minute(), nil
}
