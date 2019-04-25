package compositelog

import (
	"sync"
	"sync/atomic"
	"time"
)

type ProgressBar struct {
	definition  ProgressBarDefinition
	processLock sync.Mutex
	eventsLock  sync.Mutex
	state       atomic.Value
	wgEvents    sync.WaitGroup
	eventsChan  chan pbEvent
}

type ProgressBarDefinition struct {
	Capacity int
}

type pbState struct {
	show      bool
	position  int
	beforeBar string
	afterBar  string
	timer     pbTimer
}

const eventsChanBufferSize = 100

func NewProgressBar(definition ProgressBarDefinition) *ProgressBar {
	pb := ProgressBar{
		definition: definition,
		state:      atomic.Value{},
		eventsChan: make(chan pbEvent, eventsChanBufferSize),
	}

	pb.Reset()

	return &pb
}

func (b *ProgressBar) Reset() {
	b.state.Store(pbState{
		show:      false,
		beforeBar: "",
		afterBar:  "",
		position:  0,
		timer: pbTimer{
			startTime:     time.Time{},
			endTime:       time.Time{},
			enableCounter: false,
		},
	})
}

func (b *ProgressBar) Show() {
	b.withEventsLock(func() {
		b.wgEvents.Add(1)
		b.eventsChan <- &pbEventShow{
			show: true,
		}
	})
}

func (b *ProgressBar) Hide() {
	b.withEventsLock(func() {
		b.wgEvents.Add(1)
		b.eventsChan <- &pbEventShow{
			show: false,
		}
	})
}

func (b *ProgressBar) IsShown() bool {
	return b.state.Load().(pbState).show
}

func (b *ProgressBar) Step(step int, beforeBar, afterBar string) {
	b.withEventsLock(func() {
		b.wgEvents.Add(1)
		b.eventsChan <- &pbEventStep{
			step:      step,
			beforeBar: beforeBar,
			afterBar:  afterBar,
		}
	})
}

func (b *ProgressBar) Set(position int, beforeBar, afterBar string) {
	b.withEventsLock(func() {
		b.wgEvents.Add(1)
		b.eventsChan <- &pbEventSet{
			position:  position,
			beforeBar: beforeBar,
			afterBar:  afterBar,
		}
	})
}

func (b *ProgressBar) Fill(beforeBar, afterBar string) {
	b.withEventsLock(func() {
		b.wgEvents.Add(1)
		b.eventsChan <- &pbEventSet{
			position:  b.definition.Capacity,
			beforeBar: beforeBar,
			afterBar:  afterBar,
		}
	})
}

func (b *ProgressBar) StartTimeCounter() {
	b.withEventsLock(func() {
		b.wgEvents.Add(1)
		b.eventsChan <- &pbEventEnableTimeCounter{
			enable: true,
		}
	})
}

func (b *ProgressBar) StopTimeCounter() {
	b.withEventsLock(func() {
		b.wgEvents.Add(1)
		b.eventsChan <- &pbEventEnableTimeCounter{
			enable: false,
		}
	})
}

func (b *ProgressBar) GetRow() string {
	state := b.state.Load().(pbState)

	return fmtProgressBar(state, b.definition)
}

func (b *ProgressBar) process() {
	b.processLock.Lock()
	go func() {
		defer b.processLock.Unlock()

		timerUpdate := time.NewTicker(time.Second)

		for {
			var event pbEvent
			state := b.state.Load().(pbState)

			select {
			case event = <-b.eventsChan:
				state = event.modifyState(state, b.definition)
			case <-timerUpdate.C:
				break
			}

			state.timer.update()
			b.state.Store(state)

			if event != nil {
				b.wgEvents.Done()

				if _, ok := event.(*pbEventDone); ok {
					timerUpdate.Stop()
					return
				}
			}
		}
	}()
}

func (b *ProgressBar) done() {
	b.withEventsLock(func() {
		b.wgEvents.Add(1)
		b.eventsChan <- &pbEventDone{}
	})
}

func (b *ProgressBar) waitEventsProcessed() {
	b.wgEvents.Wait()
}

func (b *ProgressBar) withEventsLock(do func()) {
	b.eventsLock.Lock()
	defer b.eventsLock.Unlock()

	do()
}
