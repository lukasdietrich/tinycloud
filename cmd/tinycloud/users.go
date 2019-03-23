package main

import (
	"errors"
	"log"
	"path/filepath"
	"strconv"

	"github.com/urfave/cli"
	"gopkg.in/AlecAivazis/survey.v1"

	"github.com/lukasdietrich/tinycloud/database"
	"github.com/lukasdietrich/tinycloud/storage"
)

func users() cli.Command {
	var (
		db *database.DB
		s  *storage.Storage
	)

	return cli.Command{
		Name: "users",
		Before: func(ctx *cli.Context) (err error) {
			data := ctx.GlobalString("data")

			s, err = storage.New(data)
			if err != nil {
				return
			}

			db, err = database.Open(filepath.Join(data, databaseFile))
			return
		},
		Subcommands: []cli.Command{
			{
				Name: "add",
				Action: func(ctx *cli.Context) error {
					var answers struct {
						Name string
						Pass string
					}

					err := survey.Ask([]*survey.Question{
						{
							Name:   "name",
							Prompt: &survey.Input{Message: "Name:"},
							Validate: survey.ComposeValidators(
								survey.Required,
								uniqueName(db.Users()),
							),
						},
						{
							Name:   "pass",
							Prompt: &survey.Password{Message: "Password:"},
							Validate: survey.ComposeValidators(
								survey.Required,
								survey.MinLength(5),
							),
						},
					}, &answers)

					if err != nil {
						return err
					}

					err = db.Users().Add(answers.Name, answers.Pass)
					if err != nil {
						return err
					}

					return s.MkdirAll(
						s.Resolve(storage.Users, answers.Name, "/"),
						0700,
					)
				},
			},
			{
				Name: "update",
				Action: func(ctx *cli.Context) error {
					var answers struct {
						Name string
						Pass string
					}

					err := survey.Ask([]*survey.Question{
						{
							Name:   "name",
							Prompt: &survey.Input{Message: "Name:"},
							Validate: survey.ComposeValidators(
								survey.Required,
								existsName(db.Users()),
							),
						},
						{
							Name:   "pass",
							Prompt: &survey.Password{Message: "Password:"},
							Validate: survey.ComposeValidators(
								survey.Required,
								survey.MinLength(5),
							),
						},
					}, &answers)

					if err != nil {
						return err
					}

					return db.Users().Update(answers.Name, answers.Pass)
				},
			},
			{
				Name: "delete",
				Action: func(ctx *cli.Context) error {
					var answers struct {
						Name    string
						Confirm string
					}

					err := survey.Ask([]*survey.Question{
						{
							Name:   "name",
							Prompt: &survey.Input{Message: "Name:"},
							Validate: survey.ComposeValidators(
								survey.Required,
								existsName(db.Users()),
							),
						},
						{
							Name: "confirm",
							Prompt: &survey.Select{
								Message: "Are you sure?",
								Options: []string{"No", "Yes"},
							},
						},
					}, &answers)

					if err != nil {
						return err
					}

					if answers.Confirm != "Yes" {
						return nil
					}

					err = s.RemoveAll(
						s.Resolve(storage.Users, answers.Name, "/"),
					)

					if err != nil {
						return err
					}

					return db.Users().Delete(answers.Name)
				},
			},
			{
				Name: "list",
				Action: func(ctx *cli.Context) error {
					list, err := db.Users().List()
					if err != nil {
						return err
					}

					log.Printf("there are (%d) users:", len(list))

					for _, name := range list {
						log.Printf("- %s", strconv.Quote(name))
					}

					return nil
				},
			},
		},
	}
}

func uniqueName(users database.Users) survey.Validator {
	return func(v interface{}) error {
		name := v.(string)

		if exists, err := users.Exists(name); err != nil {
			return err
		} else if exists {
			return errors.New("name already taken")
		}

		return nil
	}
}

func existsName(users database.Users) survey.Validator {
	return func(v interface{}) error {
		name := v.(string)

		if exists, err := users.Exists(name); err != nil {
			return err
		} else if !exists {
			return errors.New("unknown username")
		}

		return nil
	}
}
