package main

import (
	"fmt"
	"github.com/glide-im/glide/im_service/client"
	"github.com/glide-im/glide/pkg/auth/jwt_auth"
	"github.com/glide-im/glide/pkg/rpc"
	"github.com/glide-im/glide/pkg/subscription"
	"github.com/glide-im/glide/pkg/subscription/subscription_impl"
)

var imServiceRpcCli *client.Client

func initialize() {
	var err error
	imServiceRpcCli, err = client.NewClient(&rpc.ClientOptions{
		Addr: "127.0.0.1",
		Port: 8092,
		Name: "im_rpc_server",
	})
	if err != nil {
		panic(err)
	}
}

func main() {

	//ExampleUserLogin("1")

	initialize()
	ExampleCreateGroup("c1", "1")
	ExampleAddMember("c1", "2")
}

func ExampleUserLogin(uid string) {
	// 用户输入账号密码, 查询到 uid, 用 uid 生成 jwt token, 返回给客户端, 客户端使用该 token 登录聊天服务
	token, err := jwt_auth.NewAuthorizeImpl("secret").GetToken(&jwt_auth.JwtAuthInfo{
		UID:         uid,
		Device:      "0",
		ExpiredHour: 10,
	})
	if err != nil {
		panic(err)
	}
	fmt.Println(token)
}

func ExampleCreateGroup(channelId, adminId string) {

	info := &subscription.ChanInfo{
		ID:      subscription.ChanID(channelId),
		Muted:   false,
		Blocked: false,
		Closed:  false,
	}
	// 创建频道
	err := imServiceRpcCli.CreateChannel(info.ID, info)
	if err != nil {
		panic(err)
	}

	// 添加创建者到频道
	err = imServiceRpcCli.Subscribe(subscription.ChanID(channelId), subscription.SubscriberID(adminId),
		&subscription_impl.SubscriberOptions{Perm: subscription_impl.PermAdmin})
	if err != nil {
		panic(err)
	}
	// 邀请其他成员
	// ...
}

func ExampleAddMember(channelId, memberId string) {

	err := imServiceRpcCli.Subscribe(subscription.ChanID(channelId), subscription.SubscriberID(memberId),
		&subscription_impl.SubscriberOptions{Perm: subscription_impl.PermRead})
	if err != nil {
		panic(err)
	}
}
