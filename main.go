package main

import (
	"github.com/nisainan/shadowproxy/confer"
	"github.com/nisainan/shadowproxy/server"
	"github.com/urfave/cli"
	"log"
	"os"
	"os/signal"
)

func main() {
	log.SetFlags(log.Lshortfile | log.LstdFlags)
	app := cli.NewApp()
	app.Name = "github.com/nisainan/shadowproxy"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "c",
			Value: "config.yaml",
			Usage: "config file url",
		},
	}
	app.Before = func(c *cli.Context) error {
		return confer.InitConfer(c.String("c"))
	}
	app.Action = func(c *cli.Context) error {
		return server.ListenAndServe()
	}
	err := app.Run(os.Args)
	if err != nil {
		panic("app run error: " + err.Error())
	}
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, os.Kill)
	<-quit
}
