package compositelog

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

func fmtProgressBar(s pbState, d ProgressBarDefinition) string {
	speed := float64(0)
	elapsedSeconds := s.timer.getElapsedSeconds()
	if elapsedSeconds > 0 {
		speed = float64(s.position) / float64(elapsedSeconds)
	}

	part := float64(s.position) / float64(d.Capacity)
	percent := 100 * part

	fillParticlesCount := int(math.Floor(10 * part))
	fillParticles := strings.Repeat("#", fillParticlesCount)

	emptyParticlesCount := 10 - fillParticlesCount
	emptyParticles := strings.Repeat(" ", emptyParticlesCount)

	bar := fmt.Sprintf("[%s%s] %6.2f%% | %s | %s", fillParticles, emptyParticles, percent, fmtPosition(s.position, d.Capacity), fmtSpeedometer(elapsedSeconds, speed))

	if len(s.beforeBar) > 0 {
		bar = fmt.Sprintf("%s\t%s", s.beforeBar, bar)
	}

	if len(s.afterBar) > 0 {
		bar = fmt.Sprintf("%s\t%s", bar, s.afterBar)
	}

	return bar
}

func fmtSpeedometer(elapsed uint64, speed float64) string {
	if elapsed == 0 || speed == 0 {
		return "???m ??s | ? i/s"
	}

	m := elapsed / 60
	s := elapsed % 60

	return fmt.Sprintf("%3dm %2ds | %.2f i/s", m, s, speed)
}

func fmtPosition(position, capacity int) string {
	str := strconv.Itoa(capacity)
	width := len(str)

	posFormat := fmt.Sprintf("%%%dd/%s", width, str)

	return fmt.Sprintf(posFormat, position)
}
