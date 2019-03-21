package main

import (
	"errors"
	"log"
	"net/http"
	"path/filepath"

	"github.com/urfave/cli"

	"github.com/lukasdietrich/tinycloud"
)

func start() cli.Command {
	return cli.Command{
		Name: "start",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:   "addr",
				EnvVar: "TINYCLOUD_ADDRESS",
			},
			cli.StringFlag{
				Name:   "realm",
				EnvVar: "TINYCLOUD_REALM",
			},
			cli.BoolFlag{
				Name:   "tls",
				EnvVar: "TINYCLOUD_TLS",
			},
			cli.StringFlag{
				Name:   "certFile",
				EnvVar: "TINYCLOUD_CERTFILE",
			},
			cli.StringFlag{
				Name:   "keyFile",
				EnvVar: "TINYCLOUD_KEYFILE",
			},
		},
		Action: func(ctx *cli.Context) error {
			var (
				data  = ctx.GlobalString("data")
				realm = ctx.String("realm")
				users = make(tinycloud.Users)
			)

			users.Load(filepath.Join(data, "users.json"))

			if len(users) == 0 {
				return errors.New("no users configured")
			}

			handler := tinycloud.New(&tinycloud.Config{
				Users:  users,
				Folder: data,
				Realm:  realm,
			})

			if ctx.Bool("tls") {
				log.Printf("with tls")
				return http.ListenAndServeTLS(
					ctx.String("addr"),
					ctx.String("certFile"),
					ctx.String("keyFile"),
					handler)
			} else {
				return http.ListenAndServe(
					ctx.String("addr"),
					handler)
			}
		},
	}
}
