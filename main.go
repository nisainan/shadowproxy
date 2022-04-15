package main

import (
	"github.com/urfave/cli"
	"log"
	"os"
	"os/signal"
	"playproxy/confer"
	"playproxy/server"
)

func main() {
	log.SetFlags(log.Lshortfile | log.LstdFlags)
	app := cli.NewApp()
	app.Name = "playproxy"
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
