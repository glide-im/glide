package rpc

import (
	"context"
	"github.com/smallnest/rpcx/share"
)

type ExtraContext struct {
	context.Context
}

func NewContextFrom(c context.Context) *ExtraContext {
	return &ExtraContext{c}
}

func NewContext() *ExtraContext {
	return NewContextFrom(context.Background())
}

func (c *ExtraContext) PutReqExtra(k string, v string) *ExtraContext {
	mate := c.Context.Value(share.ReqMetaDataKey)
	if mate == nil {
		mate = map[string]string{}
		c.Context = context.WithValue(c.Context, share.ReqMetaDataKey, mate)
	}
	m := c.Context.Value(share.ReqMetaDataKey).(map[string]string)
	m[k] = v
	return c
}

func (c *ExtraContext) PutResExtra(k string, v string) *ExtraContext {
	mate := c.Context.Value(share.ResMetaDataKey)
	if mate == nil {
		mate = map[string]string{}
		c.Context = context.WithValue(c.Context, share.ResMetaDataKey, mate)
	}
	m := c.Context.Value(share.ResMetaDataKey).(map[string]string)
	m[k] = v
	return c
}

func (c *ExtraContext) GetReqExtra(k string) (string, bool) {
	mate := c.Context.Value(share.ReqMetaDataKey)
	if mate == nil {
		return "", false
	}
	m := c.Context.Value(share.ReqMetaDataKey).(map[string]string)
	v, ok := m[k]
	return v, ok
}

func (c *ExtraContext) GetResExtra(k string) (string, bool) {
	mate := c.Context.Value(share.ResMetaDataKey)
	if mate == nil {
		return "", false
	}
	m := c.Context.Value(share.ResMetaDataKey).(map[string]string)
	v, ok := m[k]
	return v, ok
}
