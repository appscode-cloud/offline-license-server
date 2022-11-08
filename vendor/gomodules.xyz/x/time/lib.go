package time

import (
	"fmt"
	"time"
)

// A WeekendAdjustment specifies whether to move before/after/keep the same date.
type WeekendAdjustment int

const (
	NoChange WeekendAdjustment = iota
	Before
	After
)

var longWeekendAdjustment = []string{
	"NoChange",
	"Before",
	"After",
}

// String returns the English name of the day ("Sunday", "Monday", ...).
func (d WeekendAdjustment) String() string {
	if NoChange <= d && d <= After {
		return longWeekendAdjustment[d]
	}
	return fmt.Sprintf("%sWeekendAdjustment(%d)", `%!`, d)
}

func AdjustForWeekend(now time.Time, adj WeekendAdjustment) time.Time {
	d := now.Weekday()

	if d == time.Friday {
		switch adj {
		case NoChange:
			return now
		case Before:
			return now.AddDate(0, 0, -1)
		case After:
			return now.AddDate(0, 0, 3)
		}
	} else if d == time.Saturday {
		switch adj {
		case NoChange:
			return now
		case Before:
			return now.AddDate(0, 0, -2)
		case After:
			return now.AddDate(0, 0, 2)
		}
	} else if d == time.Sunday {
		switch adj {
		case NoChange:
			return now
		case Before:
			return now.AddDate(0, 0, -3)
		case After:
			return now.AddDate(0, 0, 1)
		}
	}
	return now
}
