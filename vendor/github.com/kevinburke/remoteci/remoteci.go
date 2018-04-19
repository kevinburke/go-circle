package remoteci

import (
	"time"
)

func round(f float64) int {
	if f < 0 {
		return int(f - 0.5)
	}
	return int(f + 0.5)
}

// GetEffectiveCost returns the cost in cents to pay an average San
// Francisco-based engineer to wait for the amount of time specified by d.
func GetEffectiveCost(d time.Duration) int {
	// https://www.glassdoor.com/Salaries/san-francisco-software-engineer-salary-SRCH_IL.0,13_IM759_KO14,31.htm
	yearlySalaryCents := float64(110554 * 100)
	// Estimate fully loaded costs add 40%.
	fullyLoadedSalary := yearlySalaryCents * 1.4

	workingDays := float64(49 * 5)
	hoursInWorkday := float64(8)
	salaryPerHour := fullyLoadedSalary * float64(d) / (workingDays * hoursInWorkday * float64(time.Hour))
	return round(salaryPerHour)
}

// ShouldPrint returns true if we should print to the screen. It uses
// a heuristic so we print infrequently if the build is expected to finish
// a long time from now, and more frequently as we approach the end of the
// build.
func ShouldPrint(lastPrinted time.Time, previousBuildDuration, elapsedDuration time.Duration) bool {
	if lastPrinted.IsZero() {
		return true
	}
	var timeRemaining time.Duration
	if previousBuildDuration == 0 {
		if elapsedDuration > 5*time.Minute {
			timeRemaining = elapsedDuration + 1*time.Minute
		} else {
			timeRemaining = 5 * time.Minute // just guess
		}
	} else {
		timeRemaining = previousBuildDuration - elapsedDuration
	}
	var durToUse time.Duration
	switch {
	case timeRemaining > 25*time.Minute:
		durToUse = 3 * time.Minute
	case timeRemaining > 8*time.Minute:
		durToUse = 2 * time.Minute
	case timeRemaining > 5*time.Minute:
		durToUse = 30 * time.Second
	case timeRemaining > 3*time.Minute:
		durToUse = 20 * time.Second
	case timeRemaining > time.Minute:
		durToUse = 15 * time.Second
	default:
		durToUse = 10 * time.Second
	}
	now := time.Now()
	return lastPrinted.Add(durToUse).Before(now)
}
