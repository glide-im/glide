package im_server

import (
	"github.com/glide-im/glide/pkg/gate"
	"github.com/rcrowley/go-metrics"
)

func newMeter(name string) metrics.Meter {
	meter := metrics.NewMeter()
	_ = metrics.Register(name, meter)
	return meter
}

func newCounter(name string) metrics.Counter {
	c := metrics.NewCounter()
	_ = metrics.Register(name, c)
	return c
}

func newGauge(name string) metrics.Gauge {
	g := metrics.NewGauge()
	_ = metrics.Register(name, g)
	return g
}

func newHistogram(name string, size int, alpha float64) metrics.Histogram {
	h := metrics.NewHistogram(metrics.NewExpDecaySample(size, alpha))
	_ = metrics.Register(name, h)
	return h
}

type MessageMetrics struct {
	MessageInMeter  metrics.Meter
	MessageOutMeter metrics.Meter
	OutCounter      metrics.Counter
	FailsCounter    metrics.Counter
	InCounter       metrics.Counter
	InHistogram     metrics.Histogram
	OutHistogram    metrics.Histogram
}

func NewMessageMetrics() *MessageMetrics {
	m := &MessageMetrics{
		MessageInMeter:  newMeter("message.in"),
		MessageOutMeter: newMeter("message.out"),
		InCounter:       newCounter("message.in"),
		OutCounter:      newCounter("message.out"),
		FailsCounter:    newCounter("message.fails"),
		InHistogram:     newHistogram("message.in", 1024, 0.015),
		OutHistogram:    newHistogram("message.out", 1024, 0.015),
	}
	return m
}

func (m *MessageMetrics) In() {
	m.InCounter.Inc(1)
	m.MessageInMeter.Mark(1)
}

func (m *MessageMetrics) Out() {
	m.MessageOutMeter.Mark(1)
	m.OutCounter.Inc(1)
}

func (m *MessageMetrics) OutFailed() {
	m.FailsCounter.Inc(1)
}

type ConnectionMetrics struct {
	ConnectionCounter metrics.Counter
	LoginCounter      metrics.Counter
	OnlineTempCounter metrics.Counter
	MaxOnline         metrics.Gauge
	AliveMeter        metrics.Meter
	AliveLoggedMeter  metrics.Meter

	AliveTempH   metrics.Histogram
	AliveLoggedH metrics.Histogram
}

func NewConnectionMetrics() *ConnectionMetrics {
	m := &ConnectionMetrics{
		ConnectionCounter: newCounter("conn.all"),
		LoginCounter:      newCounter("conn.login"),
		OnlineTempCounter: newCounter("conn.online.temp"),
		MaxOnline:         newGauge("conn.online"),
		AliveMeter:        newMeter("conn.alive"),
		AliveLoggedMeter:  newMeter("conn.logged"),
		AliveTempH:        newHistogram("conn.alive.temp", 1024, 0.015),
		AliveLoggedH:      newHistogram("conn.alive.logged", 1024, 0.015),
	}
	return m
}

func (c *ConnectionMetrics) Connected() {
	c.ConnectionCounter.Inc(1)
	c.OnlineTempCounter.Inc(1)
	conn := c.ConnectionCounter.Count()
	c.AliveMeter.Mark(c.ConnectionCounter.Count())
	if conn > c.MaxOnline.Value() {
		c.MaxOnline.Update(conn)
	}
}

func (c *ConnectionMetrics) Login() {
	c.LoginCounter.Inc(1)
	c.AliveLoggedMeter.Mark(c.LoginCounter.Count())
	c.OnlineTempCounter.Dec(1)
}

func (c *ConnectionMetrics) Exit(info gate.Info) {
	c.ConnectionCounter.Dec(1)
	c.AliveMeter.Mark(c.ConnectionCounter.Count())
	if info.ID.IsTemp() {
		c.OnlineTempCounter.Dec(1)
		c.AliveTempH.Update(c.OnlineTempCounter.Count())
	} else {
		c.LoginCounter.Dec(1)
		c.AliveLoggedH.Update(c.LoginCounter.Count())
		c.AliveLoggedMeter.Mark(c.LoginCounter.Count())
	}
}
