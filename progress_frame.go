package compositelog

import (
	"bytes"
	"strings"
)

type progressFrame struct {
	bars        []barItem
	drawOnce    []string
	lastBarID   uint
	frameHeight uint
}

type barItem struct {
	id  int
	bar *ProgressBar
}

// TODO: implement vertical moving bars in frame
// TODO: align all bars with various width in frame

func newProgressFrame() *progressFrame {
	return &progressFrame{
		bars:        []barItem{},
		lastBarID:   0,
		frameHeight: 0,
	}
}

func (f *progressFrame) attachBar(bar *ProgressBar) int {
	lastBarID := 0
	if len(f.bars) > 0 {
		lastBarID = f.bars[len(f.bars)-1].id
	}

	newBarID := lastBarID + 1

	f.bars = append(f.bars, barItem{
		id:  newBarID,
		bar: bar,
	})

	bar.process()

	return newBarID
}

func (f *progressFrame) detachBar(barID int) {
	idx := -1
	for i, item := range f.bars {
		if item.id == barID {
			item.bar.done()
			item.bar.withEventsLock(func() {
				item.bar.waitEventsProcessed()
				f.drawOnce = append(f.drawOnce, item.bar.GetRow())
			})

			idx = i
			break
		}
	}

	if idx == -1 {
		return
	}

	copy(f.bars[idx:], f.bars[idx+1:])
	f.bars[len(f.bars)-1] = barItem{}
	f.bars = f.bars[:len(f.bars)-1]
}

func (f *progressFrame) detachAll() {
	for _, item := range f.bars {
		item.bar.done()
		item.bar.withEventsLock(func() {
			item.bar.waitEventsProcessed()
			f.drawOnce = append(f.drawOnce, item.bar.GetRow())
		})
	}

	f.bars = []barItem{}
}

func (f *progressFrame) clear(buffer *bytes.Buffer) {
	for i := uint(0); i < f.frameHeight; i++ {
		buffer.WriteString(CarriageReturn + CursorUp + EraseEndOfLine)
	}
	f.frameHeight = 0
}

func (f *progressFrame) draw(buffer *bytes.Buffer) uint { // TODO: check terminal width
	var barRows []string

	for _, item := range f.bars {
		if item.bar.IsShown() {
			barRows = append(barRows, item.bar.GetRow())
		}
	}

	drawCRLF := false
	if len(f.drawOnce) > 0 && len(barRows) == 0 {
		drawCRLF = true
	}

	drawRows := append(f.drawOnce, barRows...)
	f.drawOnce = make([]string, 0)

	buffer.WriteString(strings.Join(drawRows, CarriageReturn+NewLine))
	if drawCRLF {
		buffer.WriteString(CarriageReturn + NewLine)
	}

	f.frameHeight = uint(len(barRows))

	return f.frameHeight
}

func (f *progressFrame) waitBars() {
	for _, item := range f.bars {
		item.bar.withEventsLock(func() {
			item.bar.waitEventsProcessed()
		})
	}
}
