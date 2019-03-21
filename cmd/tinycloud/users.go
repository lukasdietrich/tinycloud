package main

import (
	"errors"
	"fmt"
	"path/filepath"
	"sort"

	"github.com/urfave/cli"
	"gopkg.in/AlecAivazis/survey.v1"

	"github.com/lukasdietrich/tinycloud"
)

func users() cli.Command {
	var (
		users    = make(tinycloud.Users)
		filename string
	)

	return cli.Command{
		Name: "users",
		Before: func(ctx *cli.Context) error {
			filename = filepath.Join(ctx.GlobalString("data"), "users.json")
			users.Load(filename)

			return nil
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
							Name:     "name",
							Prompt:   &survey.Input{Message: "Name:"},
							Validate: survey.ComposeValidators(survey.Required, uniqueName(users)),
						},
						{
							Name:     "pass",
							Prompt:   &survey.Password{Message: "Password:"},
							Validate: survey.MinLength(5),
						},
					}, &answers)

					if err != nil {
						return err
					}

					users.Put(answers.Name, answers.Pass)
					return users.Save(filename)
				},
			},
			{
				Name: "update",
				Action: func(ctx *cli.Context) error {
					return nil
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
							Name:     "name",
							Prompt:   &survey.Input{Message: "Name:"},
							Validate: survey.ComposeValidators(existsName(users)),
						},
						{
							Name:   "confirm",
							Prompt: &survey.Select{Message: "Are you sure?", Options: []string{"No", "Yes"}},
						},
					}, &answers)

					if err != nil {
						return err
					}

					if answers.Confirm == "Yes" {
						delete(users, answers.Name)
						return users.Save(filename)
					}

					return nil
				},
			},
			{
				Name: "list",
				Action: func(ctx *cli.Context) error {
					var names []string

					for name, _ := range users {
						names = append(names, name)
					}

					sort.Strings(names)

					fmt.Printf("%d users:\n", len(users))

					for _, name := range names {
						fmt.Printf("- %s\n", name)
					}

					return nil
				},
			},
		},
	}
}

func uniqueName(users tinycloud.Users) survey.Validator {
	return func(v interface{}) error {
		name := v.(string)

		for user, _ := range users {
			if user == name {
				return errors.New("name already taken")
			}
		}

		return nil
	}
}

func existsName(users tinycloud.Users) survey.Validator {
	return func(v interface{}) error {
		if _, ok := users[v.(string)]; !ok {
			return errors.New("unknown username")
		}

		return nil
	}
}
