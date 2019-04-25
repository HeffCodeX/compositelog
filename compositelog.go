package compositelog

import (
	"io"
	"sync"
	"sync/atomic"
	"time"
)

const (
	CursorUp       = "\033[1A"
	EraseEndOfLine = "\033[2K"
	CarriageReturn = "\r"
	NewLine        = "\n"
)

type CompositeLog struct {
	errorsChan      chan error
	infoChan        chan string
	doneChan        chan bool
	requiredRedraws uint32
	wgMessages      sync.WaitGroup
	messagesLock    sync.Mutex
	configLock      sync.Mutex
	displayLock     sync.Mutex
	progressFrame   *progressFrame
	buffers         *buffers
}

const (
	errorsChanBufferSize = 100
	infoChanBufferSize   = 100
)

func New() *CompositeLog {
	return &CompositeLog{
		errorsChan:    make(chan error, errorsChanBufferSize),
		infoChan:      make(chan string, infoChanBufferSize),
		doneChan:      make(chan bool),
		progressFrame: newProgressFrame(),
		buffers:       &buffers{},
	}
}

func (l *CompositeLog) withConfigLock(do func()) {
	l.configLock.Lock()
	defer l.configLock.Unlock()

	do()
}

func (l *CompositeLog) withMessagesLock(do func()) {
	l.messagesLock.Lock()
	defer l.messagesLock.Unlock()

	do()
}

func (l *CompositeLog) waitRedraws(count uint32) {
	start := atomic.AddUint32(&l.requiredRedraws, count)

	for {
		time.Sleep(10 * time.Millisecond)
		now := atomic.LoadUint32(&l.requiredRedraws)
		if now <= start-count || now == 0 {
			break
		}
	}
}

func (l *CompositeLog) AttachProgressBar(bar *ProgressBar) int {
	var id int

	l.withConfigLock(func() {
		id = l.progressFrame.attachBar(bar)
	})

	return id
}

func (l *CompositeLog) DetachProgressBar(id int) {
	l.withConfigLock(func() {
		l.progressFrame.detachBar(id)
	})

	l.waitRedraws(1)
}

func (l *CompositeLog) DetachAllProgressBars() {
	l.withConfigLock(func() {
		l.progressFrame.detachAll()
	})

	l.waitRedraws(1)
}

func (l *CompositeLog) AddErrorWriter(writer io.Writer) {
	l.withConfigLock(func() {
		l.buffers.ErrorWriters = append(l.buffers.ErrorWriters, writer)
	})
}

func (l *CompositeLog) AddInfoWriter(writer io.Writer) {
	l.withConfigLock(func() {
		l.buffers.InfoWriters = append(l.buffers.InfoWriters, writer)
	})
}

func (l *CompositeLog) AddCompositeWriter(writer io.Writer) {
	l.withConfigLock(func() {
		l.buffers.CompositeWriters = append(l.buffers.CompositeWriters, writer)
	})
}

func (l *CompositeLog) WriteInfo(msg string) {
	l.withMessagesLock(func() {
		l.wgMessages.Add(1)
		l.infoChan <- msg
	})
}

func (l *CompositeLog) WriteError(err error) {
	l.withMessagesLock(func() {
		l.wgMessages.Add(1)
		l.errorsChan <- err
	})
}

func (l *CompositeLog) Display() {
	l.displayLock.Lock()
	go func() {
		defer l.displayLock.Unlock()

		refreshTicker := time.NewTicker(200 * time.Millisecond)

		for {
			loop := func() bool {
				done := false

				l.progressFrame.clear(&l.buffers.compositeBuffer)

				select {
				case err := <-l.errorsChan:
					l.buffers.writeError(err, true)
					l.wgMessages.Done()
				case msg := <-l.infoChan:
					l.buffers.writeInfo(msg, true)
					l.wgMessages.Done()
				case <-l.doneChan:
					refreshTicker.Stop()
					done = true
				case <-refreshTicker.C:
					break
				}

				if linesDrawn := l.progressFrame.draw(&l.buffers.compositeBuffer); linesDrawn > 0 {
					l.buffers.compositeBuffer.WriteString(NewLine)
				}

				l.buffers.flush()

				if atomic.LoadUint32(&l.requiredRedraws) > 0 {
					atomic.AddUint32(&l.requiredRedraws, ^uint32(0))
				}

				return done
			}

			var done bool
			l.withConfigLock(func() {
				done = loop()
			})
			if done {
				break
			}
		}
	}()
}

func (l *CompositeLog) WaitFlush() {
	l.withMessagesLock(func() {
		l.wgMessages.Wait()
		l.progressFrame.waitBars()
		l.waitRedraws(1)
	})
}

func (l *CompositeLog) Done() {
	l.WaitFlush()
	l.progressFrame.detachAll()
	l.doneChan <- true
	l.displayLock.Lock()
	l.displayLock.Unlock()
}
