package config

import "github.com/spf13/viper"

var (
	MySql    *MySqlConf
	WsServer *WsServerConf
)

type WsServerConf struct {
	ID        string
	Addr      string
	Port      int
	JwtSecret string
}

type ApiHttpConf struct {
	Addr string
	Port int
}

type IMRpcServerConf struct {
	Addr        string
	Port        int
	Network     string
	Etcd        []string
	Name        string
	EnableGroup bool
}

type MySqlConf struct {
	Host     string
	Port     int
	Username string
	Password string
	Db       string
	Charset  string
}

type RedisConf struct {
	Host     string
	Port     int
	Password string
	Db       int
}

func MustLoad() {

	viper.SetConfigName("config.toml")
	viper.SetConfigType("toml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")
	viper.AddConfigPath("./_config_local")
	viper.AddConfigPath("/etc/")
	viper.AddConfigPath("$HOME/.config/")

	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}
	c := struct {
		MySql       *MySqlConf
		Redis       *RedisConf
		WsServer    *WsServerConf
		ApiHttp     *ApiHttpConf
		IMRpcServer *IMRpcServerConf
	}{}

	err = viper.Unmarshal(&c)
	if err != nil {
		panic(err)
	}
	MySql = c.MySql
	WsServer = c.WsServer
}
