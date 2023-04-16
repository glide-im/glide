

- [go 实现的一个简单客户端](https://github.com/glide-im/glide_cli)

本项目是 glide 的单体服务实现, 包含了长连接网关, 群聊功能, 并提供了管理长连接及群聊的接口客户端.

本项目依赖 `glide`, 关于 `glide` 更多信息请查看此项目文档.

`internal` 包中为本项目的业务实现, `pkg` 包中为提供给外部系统的接口及 rpc 客户端实现.

`im-service` 需要配合接口项目 `api` 使用, 本项目只提供长连接的消息收发, 登录鉴权等需要通过 `api` 实现, `api` 通过 本项目提供的 rpc 接口进行交互, 例如给指定用户设置 id,
给指定id用户推送消息, 断开指定链接, 判断用户是否在线等.

本项目需要在消息收发时保存消息, 且消息直接通过 mysql 持久化, 离线消息通过 redis 保存.

## 简单介绍

> 2022年9月22日17:49:47

这个项目是一个单体聊天服务器, 包含群聊和单聊功能, 启用聊天历史(StoreMessageHistory.StoreMessageHistory=true)的情况下需要依赖 mysql, 启用离线(
StoreMessageHistory.StoreOfflineMessage=true)则还需要 redis, 如果关闭离线和历史消息则不要要配置也可运行, 配置文件中的 [Redis] 和 [MySql] 不需要配置即可.

客户端连接到服务后, 会收到一条 `Action` 为 `hello` 的消息, 里面包含了一些配置的一个临时 id 用来标记当前客户端, 这个 id 在每次连接到服务时都不一样.

如果客户端需要登录, 鉴权, 需要发送 `Action` 为 `api.auth` 消息进行鉴权, 具体协议查看源码. 鉴权后一个连接即会绑定一个用户id `uid`, 设备 id `device`, 鉴权 token 类型为 jwt.

聊天服务不提供用户, 用户关系, 群等管理功能, 只处理消息转发和消息保存等和消息推送转发流程相关业务, 其他功能在 http 接口项目 `api` 中有具体的实现及演示.

客户端鉴权, 管理等功能通过 HTTP API 实现, 后端 HTTP 服务再通过 `im_service` 提供的 RPC 接口进行踢人, 推送消息, 管理群等功能, 项目根目录下 `pkg` 包中列出了所有聊天服务提供的接口 及 rpc
客户端实现, 外围管理服务直接依赖本项目, 再通过 pkg 中的 rpc 客户端即可对聊天服务管理.

## 运行

> 2022年9月28日17:48:30

- [配置文件](../config/config.toml)
- [程序入口](../cmd/im_service/main.go)
- [如何调用本服务提供的RPC接口](../example/client/rpc_client_example.go)

## 测试一下发消息

> 2022年9月28日17:37:46

为了方便测试发消息等服务端的流程, 可以使用用 go 实现的一个简单客户端 [glide-cli](https://github.com/glide-im/glide_cli) 进行消息收发.
`glide_cli` 中 `example` 目录中有一个简单的例子发消息, 使用时只需要将 jwt 的 secret 配置与 `im_service` 一致即可, uid 可以随意填, 这样就免去了麻烦 API 鉴权过程,
只需要运行一个 `im_service` 即可进行消息收发的测试.