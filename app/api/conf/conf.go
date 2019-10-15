package conf

import (
	"log"
	"os"

	"github.com/Goss-io/goss/lib"
	"github.com/Goss-io/goss/lib/cmd"
	"github.com/Goss-io/goss/lib/ini"
)

type Config struct {
	Node *nodeConfig
	Db   *dbConfig
	Base *baseConfig
}

type dbConfig struct {
	Host     string
	User     string
	Port     int
	Name     string
	Password string
}

type nodeConfig struct {
	IP    string
	Port  int
	Name  string
	Token string
	// StoreAddrs []string
}

type baseConfig struct {
	LogPath string //日志存放路径.
}

var Conf *Config

//Load .
func Load(cmd *cmd.Command) {
	iniPath := cmd.Conf
	if !lib.IsExists(cmd.Conf) {
		log.Println("配置文件不存在=>", iniPath)
		os.Exit(0)
		return
	}

	if err := ini.Load(iniPath); err != nil {
		log.Printf("%+v\n", err)
		return
	}

	cf := &Config{
		Db:   parseDbConfig(),
		Node: parseNodeConfig(cmd),
		Base: parseBaseConfig(),
	}

	Conf = cf
}

//node.
func parseNodeConfig(cmd *cmd.Command) *nodeConfig {
	masterip := ini.GetString("node_ip")
	if len(masterip) < 1 {
		log.Println("node_ip 不能为空")
		os.Exit(0)
	}

	masterport := ini.GetInt("node_port")
	if masterport < 1 {
		log.Println("node_port 不能为空")
		os.Exit(0)
	}

	name := ini.GetString("node_name")
	if len(name) < 1 {
		log.Println("node_name 不能为空")
		os.Exit(0)
	}

	token := ini.GetString("token")
	if len(token) < 1 {
		log.Println("token 不能为空")
		os.Exit(0)
	}
	// storeAddrs := strings.Split(nodeStoreAddr, ",")

	nodeconf := &nodeConfig{
		IP:    masterip,
		Port:  masterport,
		Name:  name,
		Token: token,
		// StoreAddrs: storeAddrs,
	}

	return nodeconf
}

//db.
func parseDbConfig() *dbConfig {
	dbHost := ini.GetString("db_host")
	if len(dbHost) < 1 {
		log.Println("db_host 不能为空")
		os.Exit(0)
	}

	dbUser := ini.GetString("db_user")
	if len(dbUser) < 1 {
		log.Println("db_user 不能为空")
		os.Exit(0)
	}

	dbPort := ini.GetInt("db_port")
	if dbPort < 1 {
		log.Println("db_port 不能为空")
		os.Exit(0)
	}

	dbName := ini.GetString("db_name")
	if len(dbName) < 1 {
		log.Println("db_name 不能为空")
		os.Exit(0)
	}

	dbPwd := ini.GetString("db_pwd")
	if len(dbPwd) < 1 {
		log.Println("db_pwd 不能为空")
		os.Exit(0)
	}

	dbconf := &dbConfig{
		Host:     dbHost,
		User:     dbUser,
		Port:     dbPort,
		Name:     dbName,
		Password: dbPwd,
	}
	return dbconf
}

//parseBaseConfig 解析基础配置.
func parseBaseConfig() *baseConfig {
	logpath := ini.GetString("log_path")
	if len(logpath) < 1 {
		log.Println("log_path 不能为空")
		os.Exit(0)
	}
	base := baseConfig{
		LogPath: logpath,
	}
	return &base
}
