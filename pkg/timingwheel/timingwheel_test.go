package timingwheel

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"strconv"
	"testing"
	"time"
)

func TestNewTimingWheel(t *testing.T) {

	tw := NewTimingWheel(time.Millisecond*100, 3, 20)
	assert.NotNil(t, tw)
	assert.Equal(t, 20, tw.wheel.slot.len)
	assert.Equal(t, 400, tw.wheel.slotCap)
}

func (s *slot) tasks() [][]*Task {
	sl := s
	for sl.index != 0 {
		sl = sl.next
	}
	var t [][]*Task

	t = append(t, sl.valueArray())
	sl = sl.next
	if sl.index != 0 {
		t = append(t, sl.valueArray())
		sl = sl.next
	}
	t = append(t, sl.valueArray())
	return t
}

func (w *wheel) status() string {
	var s []string
	sl := w.slot
	for ; sl.index != sl.len-1; sl = sl.next {

	}
	for i := 0; i != sl.len; i++ {
		sl = sl.next
		if sl.index == w.slot.index {
			s = append(s, strconv.Itoa(i))
			continue
		}
		if sl.isEmpty() {
			s = append(s, "_")
		} else {
			s = append(s, "#")
		}
	}

	var ts []string
	for _, tasks := range w.slot.tasks() {
		var tt []string
		for _, t := range tasks {
			tt = append(tt, strconv.Itoa(t.offset))
		}
		ts = append(ts, fmt.Sprintf("%v", tt))
	}

	return fmt.Sprintf("%v %v %d", s, ts, w.remain)
}

func sleepRndMilleSec(start int32, end int32) {
	n := rand.Int31n(end - start)
	n = start + n
	time.Sleep(time.Duration(n) * time.Millisecond)
}
