package compositelog

import "time"

type pbTimer struct {
	startTime     time.Time
	endTime       time.Time
	enableCounter bool
}

func (t *pbTimer) update() {
	if !t.enableCounter {
		return
	}

	now := time.Now()
	t.endTime = now

	if t.startTime.IsZero() {
		t.startTime = now
	}
}

func (t *pbTimer) getElapsedSeconds() uint64 {
	if t.startTime.IsZero() || t.endTime.IsZero() {
		return 0
	}

	if t.startTime.UnixNano() > t.endTime.UnixNano() {
		return 0
	}

	delta := t.endTime.UnixNano() - t.startTime.UnixNano()

	return uint64(time.Duration(delta).Round(time.Second).Seconds())
}
