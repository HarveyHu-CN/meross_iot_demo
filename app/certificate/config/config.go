package config

import (
	"fmt"
	"meross_iot/library/configurator"
	"os"
	"path"
)

var configPath = make(map[string]string)

func Init()  {
	wd, err := os.Getwd()
	if err != nil {
		panic(fmt.Errorf("fatal error, fail to get work directory"))
	}
	//这里的路径设定，可运行文件必须放在cmd或同级目录下才行
	configPath["app"] = path.Clean(wd + "/../config/config.toml")
	configPath["global"] = path.Clean(wd + "/../../../config/config.toml")

	configurator.Load(configPath)
}

