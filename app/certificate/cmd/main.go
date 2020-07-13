package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"meross_iot/app/certificate/config"
	"meross_iot/library/cache/redis"
	"meross_iot/library/configurator"
	"meross_iot/library/db/mysql"
	"meross_iot/library/logger"
)

const (
	AppName = "certificate"
)

func main()  {
	config.Init()
	c := mysql.NewConfig()
	configurator.Is("global").UnmarshalKey("mainDb", c)
	//fmt.Printf("%+v\n", c)
	mysql.New(c)
	cc := redis.NewConfig()
	configurator.Is("global").UnmarshalKey("mainCache", cc)
	fmt.Printf("%+v\n", cc)
	redis.New(cc)
	logger.Init(AppName, zerolog.ErrorLevel)
	r := gin.Default()
	//http.InitRouter(r)
	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}

