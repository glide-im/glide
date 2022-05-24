package im_server

import (
	"github.com/glide-im/glide/pkg/client"
	"time"
)

var cacheServerInfo *client.ServerInfo = nil
var cacheInfoExpired = time.Now()

func (c *DefaultClientManager) GetServerInfo(count int) *client.ServerInfo {
	if cacheInfoExpired.After(time.Now()) {
		return cacheServerInfo
	}
	cacheInfoExpired = time.Now().Add(time.Second * 5)
	info := c.GetManagerInfo()
	cacheServerInfo = &info
	cacheServerInfo.OnlineCli = c.getClient(count)
	return cacheServerInfo
}
