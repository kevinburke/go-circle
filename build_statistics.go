package circle

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/ssh/terminal"
)

const stepWidth = 45

var stepPadding = fmt.Sprintf("%%-%ds", stepWidth)

func isatty() bool {
	return terminal.IsTerminal(int(os.Stdout.Fd()))
}

var forceNonZeroTestVal = time.Duration(0)

// timeScaler returns a format string for the given Duration where all of the
// decimals will line up in the same column (fourth from the end).
func timeScaler(d time.Duration) string {
	if d == 0 && forceNonZeroTestVal != 0 {
		d = forceNonZeroTestVal
	}
	switch {
	case d == 0:
		return "0.0ms"
	case d >= time.Minute:
		mins := d / time.Minute
		d = d - mins*time.Minute
		s := strconv.FormatFloat(float64(d.Nanoseconds())/1e9, 'f', 0, 64)
		return strconv.Itoa(int(mins)) + "m" + fmt.Sprintf("%02s", s) + "s"
	case d >= time.Second:
		return strconv.FormatFloat(float64(d.Nanoseconds())/1e9, 'f', 1, 64) + "s"
	case d >= 50*time.Microsecond:
		return strconv.FormatFloat(float64(d.Nanoseconds())/1e9*1000, 'f', 0, 64) + "ms"
	case d >= time.Microsecond:
		return strconv.FormatFloat(float64(d.Nanoseconds())/1e9*1000*1000, 'f', 0, 64) + "µs"
	default:
		return strconv.FormatFloat(float64(d.Nanoseconds()), 'f', 0, 64) + "ns"
	}
}

// Statistics prints out statistics for the given build. If stdout is a TTY,
// failed builds will be surrounded by red ANSI escape sequences.
func (cb *CircleBuild) Statistics() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf(stepPadding, "Step"))
	l := stepWidth
	for i := uint8(0); i < cb.Parallel; i++ {
		b.WriteString(fmt.Sprintf("%-8d", i))
		l += 8
	}
	b.WriteString(fmt.Sprintf("\n%s\n", strings.Repeat("=", l)))
	for _, step := range cb.Steps {
		stepName := strings.Replace(step.Name, "\n", "\\n", -1)
		if len(stepName) > stepWidth-2 {
			stepName = fmt.Sprintf("%s… ", stepName[:(stepWidth-2)])
		} else {
			stepName = fmt.Sprintf(stepPadding, stepName)
		}
		b.WriteString(stepName)
		for _, action := range step.Actions {
			var dur time.Duration
			if time.Duration(action.Runtime) > time.Minute {
				dur = time.Duration(action.Runtime).Round(time.Second)
			} else {
				dur = time.Duration(action.Runtime).Round(time.Millisecond * 10)
			}
			if action.Failed() && isatty() {
				// color the output red
				fmt.Fprintf(&b, "\033[38;05;160m%8s\033[0m", timeScaler(dur))
			} else {
				fmt.Fprintf(&b, "%8s", timeScaler(dur))
			}
		}
		b.WriteString("\n")
	}
	return b.String()
}
