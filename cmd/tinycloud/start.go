package main

import (
	"net/http"
	"path/filepath"

	"github.com/urfave/cli"

	"github.com/lukasdietrich/tinycloud/database"
	"github.com/lukasdietrich/tinycloud/storage"
	"github.com/lukasdietrich/tinycloud/webdav"
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
			)

			db, err := database.Open(filepath.Join(data, databaseFile))
			if err != nil {
				return err
			}

			s, err := storage.New(data)
			if err != nil {
				return err
			}

			handler := webdav.New(&webdav.Config{
				Realm:    realm,
				Database: db,
				Storage:  s,
			})

			if ctx.Bool("tls") {
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
