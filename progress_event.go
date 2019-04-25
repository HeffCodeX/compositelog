package compositelog

import "time"

type pbEvent interface {
	modifyState(s pbState, d ProgressBarDefinition) pbState
}

type pbEventSetStartTime struct {
	time time.Time
}

func (e *pbEventSetStartTime) modifyState(s pbState, d ProgressBarDefinition) pbState {
	s.timer.startTime = e.time

	return s
}

type pbEventEnableTimeCounter struct {
	enable bool
}

func (e *pbEventEnableTimeCounter) modifyState(s pbState, d ProgressBarDefinition) pbState {
	s.timer.enableCounter = e.enable

	return s
}

type pbEventSet struct {
	position            int
	beforeBar, afterBar string
}

func (e *pbEventSet) modifyState(s pbState, d ProgressBarDefinition) pbState {
	position := e.position
	if position > d.Capacity {
		position = d.Capacity
	}

	s.position = position
	s.beforeBar = e.beforeBar
	s.afterBar = e.afterBar

	return s
}

type pbEventStep struct {
	step                int
	beforeBar, afterBar string
}

func (e *pbEventStep) modifyState(s pbState, d ProgressBarDefinition) pbState {
	position := s.position + e.step
	if position > d.Capacity {
		position = d.Capacity
	}

	s.position = position
	s.beforeBar = e.beforeBar
	s.afterBar = e.afterBar

	return s
}

type pbEventShow struct {
	show bool
}

func (e *pbEventShow) modifyState(s pbState, d ProgressBarDefinition) pbState {
	s.show = e.show

	return s
}

type pbEventDone struct {
}

func (e *pbEventDone) modifyState(s pbState, d ProgressBarDefinition) pbState {
	s.show = false

	return s
}
