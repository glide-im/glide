package im_server

import (
	"errors"
	"github.com/glide-im/glide/pkg/client"
	"github.com/glide-im/glide/pkg/conn"
	"github.com/glide-im/glide/pkg/logger"
	"github.com/glide-im/glide/pkg/messages"
	"github.com/panjf2000/ants/v2"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

var ErrClientClosed = errors.New("client closed")
var ErrClientNotExist = errors.New("client does not exist")

var pool *ants.Pool

func init() {
	var err error
	pool, err = ants.NewPool(50_0000,
		ants.WithNonblocking(true),
		ants.WithPanicHandler(func(i interface{}) {
			logger.E("%v", i)
		}),
		ants.WithPreAlloc(false),
	)
	if err != nil {
		panic(err)
	}
}

type DefaultClientManager struct {
	clients      *clients
	clientOnline int64
	messageSent  int64
	maxOnline    int64
	startAt      int64
}

func NewDefaultManager() *DefaultClientManager {
	ret := new(DefaultClientManager)
	ret.clients = newClients()
	ret.startAt = time.Now().Unix()
	return ret
}

// ClientConnected 当一个用户连接建立后, 由该方法创建 Client 实例 Client 并管理该连接, 返回该由连接创建客户端的标识 id
// 返回的标识 id 是一个临时 id, 后续连接认证后会改变
func (c *DefaultClientManager) ClientConnected(conn conn.Connection) int64 {

	// 获取一个临时 uid 标识这个连接
	connUid := int64(0)
	ret := NewClient(conn, c, nil)
	ret.SetID("")
	c.clients.add(connUid, 0, ret)

	// 开始处理连接的消息
	ret.Run()
	return connUid
}

func (c *DefaultClientManager) AddClient(id client.ID, cs client.Client) {
	c.clients.add(id.UID(), 0, cs)
	atomic.AddInt64(&c.clientOnline, 1)
}

func (c *DefaultClientManager) SigIn(old, new_ client.ID) error {

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

	max := atomic.LoadInt64(&c.maxOnline)
	current := atomic.AddInt64(&c.clientOnline, 1)
	if max < current {
		atomic.StoreInt64(&c.maxOnline, current)
	}
	return nil
}

func (c *DefaultClientManager) Logout(id client.ID) error {
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
	atomic.AddInt64(&c.clientOnline, -1)
	return nil
}

// EnqueueMessage to the client with the specified uid and device, device: pass 0 express all device.
func (c *DefaultClientManager) EnqueueMessage(id client.ID, msg *messages.GlideMessage) error {
	atomic.AddInt64(&c.messageSent, 1)

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
	ds.foreach(func(deviceId int64, cli client.Client) {
		if id.Device() != 0 && deviceId != id.Device() {
			return
		}
		err = c.enqueueMessage(cli, msg)
	})
	return err
}

func (c *DefaultClientManager) enqueueMessage(cli client.Client, msg *messages.GlideMessage) error {
	if !cli.IsRunning() {
		return ErrClientClosed
	}
	err := pool.Submit(func() {
		_ = cli.EnqueueMessage(msg)
	})
	if err != nil {
		logger.E("message enqueue:%v", err)
		return err
	}
	return nil
}

func (c *DefaultClientManager) IsOnline(id client.ID) bool {
	ds := c.clients.get(id.UID())
	if ds == nil {
		return false
	}
	return ds.size() > 0
}

func (c *DefaultClientManager) isDeviceOnline(uid, device int64) bool {
	ds := c.clients.get(uid)
	if ds == nil {
		return false
	}
	return ds.get(device) != nil
}

func (c *DefaultClientManager) getClient(count int) []client.Info {
	//goland:noinspection GoPreferNilSlice
	ret := []client.Info{}
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

func (c *DefaultClientManager) GetManagerInfo() client.ServerInfo {
	return client.ServerInfo{
		Online:      atomic.LoadInt64(&c.clientOnline),
		MaxOnline:   atomic.LoadInt64(&c.maxOnline),
		MessageSent: atomic.LoadInt64(&c.messageSent),
		StartAt:     c.startAt,
	}
}

//////////////////////////////////////////////////////////////////////////////

type devices struct {
	ds map[int64]client.Client
}

func (d *devices) put(device int64, cli client.Client) {
	d.ds[device] = cli
}

func (d *devices) get(device int64) client.Client {
	return d.ds[device]
}

func (d *devices) remove(device int64) {
	delete(d.ds, device)
}

func (d *devices) foreach(f func(device int64, c client.Client)) {
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

func (g *clients) add(uid int64, device int64, c client.Client) {
	g.m.Lock()
	defer g.m.Unlock()
	cs, ok := g.clients[uid]
	if ok {
		cs.put(device, c)
	} else {
		d := &devices{map[int64]client.Client{}}
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
