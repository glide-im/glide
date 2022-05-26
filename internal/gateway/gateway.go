package gateway

import (
	"errors"
	"github.com/glide-im/glide/pkg/gate"
	"github.com/glide-im/glide/pkg/logger"
	"github.com/glide-im/glide/pkg/messages"
	"github.com/panjf2000/ants/v2"
	"strconv"
	"sync"
)

var ErrClientClosed = errors.New("client closed")
var ErrClientNotExist = errors.New("client does not exist")

type Impl struct {
	gate.Gateway

	// clients is a map of all connected clients
	clients *clients

	// msgHandler client message handler
	msgHandler gate.MessageHandler

	// pool of ants, used to process messages concurrently.
	pool *ants.Pool
}

func NewServer() (*Impl, error) {
	ret := new(Impl)
	ret.clients = newClients()

	var err error
	ret.pool, err = ants.NewPool(50_0000,
		ants.WithNonblocking(true),
		ants.WithPanicHandler(func(i interface{}) {
			logger.E("%v", i)
		}),
		ants.WithPreAlloc(false),
	)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func (c *Impl) AddClient(cs gate.Client) {
	id := cs.GetInfo().ID
	c.clients.add(id.UID(), id.Device(), cs)
}

func (c *Impl) SetClientID(old, new_ gate.ID) error {

	device := new_.Device()

	tempDs := c.clients.get(old.UID())
	if tempDs == nil || tempDs.size() == 0 {
		return ErrClientNotExist
	}
	cli := tempDs.get(0)
	logged := c.clients.get(new_.UID())
	if logged != nil && logged.size() > 0 {
		// 多设备登录
		existing := logged.get(device)
		if existing != nil {
			_ = c.EnqueueMessage(new_, messages.NewMessage(0, messages.ActionNotifyKickOut, ""))
			existing.Exit()
			logged.remove(device)
		}
		if logged.size() > 0 {
			msg := "multi device login, device=" + strconv.FormatInt(device, 10)
			_ = cli.EnqueueMessage(messages.NewMessage(0, messages.ActionNotifyAccountLogin, msg))
		}
		logged.put(device, cli)
	} else {
		// 单设备登录
		c.clients.add(new_.UID(), device, cli)
	}
	cli.SetID(new_)
	// 删除临时 id
	c.clients.delete(old.UID(), 0)
	return nil
}

func (c *Impl) ExitClient(id gate.ID) error {
	cl := c.clients.get(id.UID())
	if cl == nil || cl.size() == 0 {
		return ErrClientNotExist
	}
	logDevice := cl.get(id.Device())
	if logDevice == nil {
		return ErrClientNotExist
	}

	logDevice.Exit()
	cl.remove(id.Device())
	return nil
}

// EnqueueMessage to the client with the specified uid and device, device: pass 0 express all device.
func (c *Impl) EnqueueMessage(id gate.ID, msg *messages.GlideMessage) error {

	var err error = nil
	ds := c.clients.get(id.UID())
	if ds == nil || ds.size() == 0 {
		return ErrClientNotExist
	}
	if id.Device() != 0 {
		d := ds.get(id.Device())
		if d == nil {
			return ErrClientNotExist
		}
		return c.enqueueMessage(d, msg)
	}
	ds.foreach(func(deviceId int64, cli gate.Client) {
		if id.Device() != 0 && deviceId != id.Device() {
			return
		}
		err = c.enqueueMessage(cli, msg)
	})
	return err
}

func (c *Impl) enqueueMessage(cli gate.Client, msg *messages.GlideMessage) error {
	if !cli.IsRunning() {
		return ErrClientClosed
	}
	err := c.pool.Submit(func() {
		_ = cli.EnqueueMessage(msg)
	})
	if err != nil {
		logger.E("message enqueue:%v", err)
		return err
	}
	return nil
}

func (c *Impl) IsOnline(id gate.ID) bool {
	ds := c.clients.get(id.UID())
	if ds == nil {
		return false
	}
	return ds.size() > 0
}

func (c *Impl) isDeviceOnline(uid, device int64) bool {
	ds := c.clients.get(uid)
	if ds == nil {
		return false
	}
	return ds.get(device) != nil
}

func (c *Impl) getClient(count int) []gate.Info {
	//goland:noinspection GoPreferNilSlice
	ret := []gate.Info{}
	ct := 0
	c.clients.m.RLock()
	for _, ds := range c.clients.clients {
		for _, d := range ds.ds {
			if !d.Logged() {
				continue
			}
			ret = append(ret, d.GetInfo())
			break
		}
		ct++
		if ct >= count {
			break
		}
	}
	c.clients.m.RUnlock()
	return ret
}

//////////////////////////////////////////////////////////////////////////////

type devices struct {
	ds map[int64]gate.Client
}

func (d *devices) put(device int64, cli gate.Client) {
	d.ds[device] = cli
}

func (d *devices) get(device int64) gate.Client {
	return d.ds[device]
}

func (d *devices) remove(device int64) {
	delete(d.ds, device)
}

func (d *devices) foreach(f func(device int64, c gate.Client)) {
	for k, v := range d.ds {
		f(k, v)
	}
}
func (d *devices) size() int {
	return len(d.ds)
}

type clients struct {
	m       sync.RWMutex
	clients map[int64]*devices
}

func newClients() *clients {
	ret := new(clients)
	ret.m = sync.RWMutex{}
	ret.clients = make(map[int64]*devices)
	return ret
}

func (g *clients) size() int {
	g.m.RLock()
	defer g.m.RUnlock()
	return len(g.clients)
}

func (g *clients) get(uid int64) *devices {
	g.m.RLock()
	defer g.m.RUnlock()
	cl, ok := g.clients[uid]
	if ok && cl.size() != 0 {
		return cl
	}
	return nil
}

func (g *clients) contains(uid int64) bool {
	g.m.RLock()
	defer g.m.RUnlock()
	_, ok := g.clients[uid]
	return ok
}

func (g *clients) add(uid int64, device int64, c gate.Client) {
	g.m.Lock()
	defer g.m.Unlock()
	cs, ok := g.clients[uid]
	if ok {
		cs.put(device, c)
	} else {
		d := &devices{map[int64]gate.Client{}}
		d.put(device, c)
		g.clients[uid] = d
	}
}

func (g *clients) delete(uid int64, device int64) {
	g.m.Lock()
	defer g.m.Unlock()
	d, ok := g.clients[uid]
	if ok {
		d.remove(device)
		if d.size() == 0 {
			delete(g.clients, uid)
		}
	}
}
