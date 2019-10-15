package conf

import (
	"log"
	"os"

	"github.com/Goss-io/Goss/lib"

	"github.com/Goss-io/Goss/lib/cmd"
	"github.com/Goss-io/Goss/lib/ini"
)

type Config struct {
	Node *nodeConfig
	Base *baseConfig
}

type nodeConfig struct {
	IP      string
	Port    int
	WebPort int
	Name    string
	Token   string
}

type baseConfig struct {
	LogPath string
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
		Node: parseNodeConfig(cmd),
		Base: parseBaseConfig(),
	}

	Conf = cf
}

//node.
func parseNodeConfig(cmd *cmd.Command) *nodeConfig {
	name := ini.GetString("node_name")
	if len(name) < 1 {
		log.Println("node_name 不能为空")
		os.Exit(0)
	}

	storeip := ini.GetString("node_ip")
	if len(storeip) < 1 {
		log.Println("node_ip 不能为空")
		os.Exit(0)
	}
	storeport := ini.GetInt("node_port")
	if storeport < 1 {
		log.Println("node_port 不能为空")
		os.Exit(0)
	}
	webport := ini.GetInt("node_web_port")
	if webport < 1 {
		log.Println("node_web_port 不能为空")
		os.Exit(0)
	}
	token := ini.GetString("token")
	if len(token) < 1 {
		log.Println("token 不能为空")
		os.Exit(0)
	}

	nodeconf := &nodeConfig{
		IP:      storeip,
		Port:    storeport,
		WebPort: webport,
		Name:    name,
		Token:   token,
	}

	return nodeconf
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
