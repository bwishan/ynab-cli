package cli

import (
	"encoding/json"
)

func (a *App) registerUserCommands() {
	cmd := &Command{
		Name:        "user",
		Description: "User account information",
		Subcommands: map[string]*Subcommand{
			"get": {
				Name:        "get",
				Description: "Get the authenticated user",
				Usage:       "",
				Run: func(ctx *Context, args []string) error {
					data, err := ctx.Client.GetUser()
					if err != nil {
						return err
					}

					headers := []string{"ID"}
					extractRows := func(data []byte) ([][]string, error) {
						var resp struct {
							Data struct {
								User struct {
									ID string `json:"id"`
								} `json:"user"`
							} `json:"data"`
						}
						if err := json.Unmarshal(data, &resp); err != nil {
							return nil, err
						}
						return [][]string{{resp.Data.User.ID}}, nil
					}

					return ctx.Printer.PrintResult(data, headers, extractRows)
				},
			},
		},
	}
	a.registerCommand(cmd)
}
