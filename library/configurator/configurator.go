package configurator

import (
	"fmt"
	"github.com/spf13/viper"
	"path"
	"strings"
)

type entity struct {
	name string
	path string
	v *viper.Viper
}

const CONF_EXT  = "toml"

var container = make(map[string]*entity)

func Load(confPaths map[string]string) {
	for name, confPath := range confPaths  {
		if _, ok := container[name]; ok {
			panic("Config entry [" + name + "] is already loaded\n")
		}
		container[name] = &entity{
			name : name,
			path : confPath,
			v : viper.New(),
		}
		confPath = path.Clean(confPath)
		ext := strings.TrimLeft(path.Ext(confPath), ".")
		if ext != CONF_EXT {
			panic(fmt.Errorf("Fatal error config file ext is not toml: [%s]\n", ext))
		}
		container[name].v.SetConfigType(CONF_EXT)
		container[name].v.AddConfigPath(path.Dir(confPath))
		container[name].v.SetConfigName(path.Base(confPath))
		err := container[name].v.ReadInConfig()
		if err != nil {
			panic(fmt.Errorf("Viper fatal error when loading config file : name [%s] path[%s] err[%s] \n",
				name, confPath, err))
		}
	}
}

func Is(name string) *viper.Viper {
	e, ok := container[name]
	if !ok {
		return nil
	}
	return e.v
}

func (e *entity) Get(key string) interface{} {
	return e.v.Get(key)
}


