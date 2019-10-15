package cmd

import (
	"flag"
	"log"
	"os"
)

func (c *Command) parse() {

	var confPath string
	flag.StringVar(&confPath, "conf", "", "conf")

	flag.Parse()

	//conf.
	c.Conf = confPath
	if confPath == "" {
		log.Println("请指定配置文件 -conf")
		os.Exit(0)
	}
}

type Command struct {
	Conf string
}

func New() *Command {
	command := &Command{}
	command.parse()

	return command
}
