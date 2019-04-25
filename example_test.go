package compositelog

import (
	"os"
	"sync"
	"testing"
	"time"
)

func testLoop(wg *sync.WaitGroup, bar *ProgressBar, iterations uint, wait time.Duration) {
	bar.StartTimeCounter()

	for i := uint(0); i < iterations; i++ {
		bar.Step(1, "A", "B")
		time.Sleep(wait)
	}

	bar.StopTimeCounter()
	wg.Done()
}

func TestExample(t *testing.T) {
	c := New()

	c.Display()
	c.AddCompositeWriter(os.Stdout)
	c.WriteInfo("Start")

	b1 := NewProgressBar(ProgressBarDefinition{Capacity: 100})
	b2 := NewProgressBar(ProgressBarDefinition{Capacity: 200})
	b3 := NewProgressBar(ProgressBarDefinition{Capacity: 300})

	bID1 := c.AttachProgressBar(b1)
	bID2 := c.AttachProgressBar(b2)
	bID3 := c.AttachProgressBar(b3)

	b1.Show()
	b2.Show()
	b3.Show()

	wg := new(sync.WaitGroup)
	wg.Add(3)
	go testLoop(wg, b1, 100, 10*time.Millisecond)
	go testLoop(wg, b2, 200, 10*time.Millisecond)
	go testLoop(wg, b3, 300, 10*time.Millisecond)

	ticker := time.NewTicker(2 * time.Second)
	go func() {
		for {
			<-ticker.C
			c.WriteInfo("write something")
		}
	}()

	wg.Wait()
	ticker.Stop()

	c.DetachProgressBar(bID1)
	c.WriteInfo("Detach bar 1")

	b2.Hide()
	c.WriteInfo("Hide bar 2")

	c.WriteInfo("Before show and detach bar 2")
	b2.Show()
	c.DetachProgressBar(bID2)
	c.WaitFlush()
	c.WriteInfo("After show and detach bar 2")

	c.DetachProgressBar(bID3)
	c.WriteInfo("Detach bar 3")

	c.WaitFlush()
	c.WriteInfo("Done")
	c.Done()
}
