package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"othello/model"
	"othello/router"
	"syscall"
)

var (
	httpPort string
	database string
)

func init() {
	flag.StringVar(&httpPort, "port", "8081", "指定HTTP运行端口")
	flag.StringVar(&database, "db", "test", "指定数据库")
}

func main() {
	r := router.Router()
	flag.Parse()
	if err := model.Init(database); err != nil {
		log.Fatal(err)
	}

	errs := make(chan error)

	go func() {
		ch := make(chan os.Signal)
		signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
		errs <- fmt.Errorf("%s", <-ch)
	}()

	go func() {
		errs <- r.Run(":" + httpPort)
	}()

	log.Println("Exited: ", <-errs)
}
