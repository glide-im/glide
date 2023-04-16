package main

import (
	"github.com/glide-im/glide/pkg/gate"
	"github.com/glide-im/glide/pkg/messages"
	"github.com/glide-im/glide/pkg/rpc"
	"github.com/glide-im/glide/pkg/subscription/subscription_impl"
	"github.com/glide-im/im-service/pkg/client"
)

/// 如何控制消息服务器, 用户连接, 发布订阅接口
/// 用于 HTTP API 接口等非 IM 业务配合使用

func main() {
	// 消息网关接口(用户管理用户连接, 用户状态)
	RpcGatewayClientExample()

	// 发布订阅(群聊)
	RpcSubscriberClientExample()
}

func RpcGatewayClientExample() {

	// 消息网关 RPC 客户端配置
	options := &rpc.ClientOptions{
		Addr: "127.0.0.1",
		Port: 8092,
		Name: "im_rpc_server",
	}

	// 创建消息网关接口客户端
	cli, err := client.NewGatewayRpcImpl(options)
	// 长时间不用完记得关闭
	defer cli.Close()
	if err != nil {
		panic(err)
	}

	// 给网关中指定 id 的链接推送一条消息 (例如加好友通知, 多设备登录通知等等)
	err = cli.EnqueueMessage(gate.NewID2("1"), messages.NewEmptyMessage())
	if err != nil {
		panic(err)
	}

	// 设置网关中连接新 id
	err = cli.SetClientID(gate.NewID2("1"), gate.NewID2("2"))
	if err != nil {
		panic(err)
	}

	// 断开 uid 为 1 的设备 1
	// 单体情况, 网关 id 传空即可
	_ = cli.ExitClient(gate.NewID("", "1", "1"))

	// 获取某个用户是否在线
	cli.IsOnline(gate.NewID2("1"))
}

func RpcSubscriberClientExample() {
	options := &rpc.ClientOptions{
		Addr: "127.0.0.1",
		Port: 8092,
		Name: "im_rpc_server",
	}
	cli, err := client.NewSubscriptionRpcImpl(options)
	defer cli.Close()
	if err != nil {
		panic(err)
	}

	//err = cli.CreateChannel("1", &subscription.ChanInfo{
	//	ID:   "1",
	//	Type: 0,
	//})
	//if err != nil {
	//	panic(err)
	//}

	// 用户订阅某个频道的消息(用户上线, 开始接受群消息)
	err = cli.Subscribe("1", "1", &subscription_impl.SubscriberOptions{
		Perm: subscription_impl.PermRead | subscription_impl.PermWrite,
	})
	if err != nil {
		panic(err)
	}

	// 移除指定 id 频道 (解散群, 删除频道等)
	_ = cli.RemoveChannel("1")

	msg := &subscription_impl.PublishMessage{
		From:    "1",
		Seq:     1,
		Type:    subscription_impl.TypeMessage,
		Message: messages.NewMessage(0, "1", &messages.ChatMessage{}),
	}
	// 推送消息到指定频道 (发送一条系统消息, 群通知等)
	err = cli.Publish("1", msg)
	if err != nil {
		panic(err)
	}
}
