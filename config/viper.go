package config

import "github.com/spf13/viper"

var (
	Common    *CommonConf
	MySql     *MySqlConf
	WsServer  *WsServerConf
	IMService *IMRpcServerConf
	Redis     *RedisConf
	Kafka     *KafkaConf
)

type CommonConf struct {
	StoreOfflineMessage bool
	StoreMessageHistory bool
	SecretKey           string
}

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
	Addr    string
	Port    int
	Network string
	Etcd    []string
	Name    string
}

type KafkaConf struct {
	Address []string
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
	viper.AddConfigPath("./_config_local")
	viper.AddConfigPath("./config")
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
		IMRpcServer *IMRpcServerConf
		CommonConf  *CommonConf
		KafkaConf   *KafkaConf
	}{}

	err = viper.Unmarshal(&c)
	if err != nil {
		panic(err)
	}
	MySql = c.MySql
	WsServer = c.WsServer
	IMService = c.IMRpcServer
	Common = c.CommonConf
	Redis = c.Redis
	Kafka = c.KafkaConf

	if Common == nil {
		panic("CommonConf is nil")
	}
	if c.MySql == nil {
		panic("mysql config is nil")
	}
	if c.WsServer == nil {
		panic("ws server config is nil")
	}
	if c.IMRpcServer == nil {
		panic("im rpc server config is nil")
	}
}
