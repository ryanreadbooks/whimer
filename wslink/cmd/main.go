package main

import (
	"flag"

	"github.com/zeromicro/go-zero/core/service"
)

var configFile = flag.String("f", "etc/wslink.yaml", "the config file")

func main() {
	flag.Parse()

	group := service.NewServiceGroup()
	defer group.Stop()
	group.Start()
}
