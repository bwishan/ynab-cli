package cli

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/bwishan/ynab-cli/internal/config"
)

func (a *App) registerPayeeCommands() {
	// payees command
	payeesCmd := &Command{
		Name:        "payees",
		Description: "Manage payees",
		Subcommands: map[string]*Subcommand{
			"list": {
				Name:        "list",
				Description: "List all payees for a plan",
				Usage:       "<plan-id> [--last-knowledge N]",
				Run: func(ctx *Context, args []string) error {
					var planIDArg string
					var lastKnowledge int64

					for i := 0; i < len(args); i++ {
						switch {
						case args[i] == "--last-knowledge" && i+1 < len(args):
							v, err := strconv.ParseInt(args[i+1], 10, 64)
							if err != nil {
								return fmt.Errorf("invalid --last-knowledge value: %s", args[i+1])
							}
							lastKnowledge = v
							i++
						case strings.HasPrefix(args[i], "--last-knowledge="):
							val := strings.TrimPrefix(args[i], "--last-knowledge=")
							v, err := strconv.ParseInt(val, 10, 64)
							if err != nil {
								return fmt.Errorf("invalid --last-knowledge value: %s", val)
							}
							lastKnowledge = v
						case !strings.HasPrefix(args[i], "--"):
							if planIDArg == "" {
								planIDArg = args[i]
							}
						}
					}

					planID, err := config.ResolvePlan(planIDArg)
					if err != nil {
						return err
					}

					data, err := ctx.Client.GetPayees(planID, lastKnowledge)
					if err != nil {
						return err
					}

					headers := []string{"ID", "NAME", "TRANSFER_ACCOUNT_ID", "DELETED"}
					extractRows := func(data []byte) ([][]string, error) {
						var resp struct {
							Data struct {
								Payees []struct {
									ID                string `json:"id"`
									Name              string `json:"name"`
									TransferAccountID string `json:"transfer_account_id"`
									Deleted           bool   `json:"deleted"`
								} `json:"payees"`
							} `json:"data"`
						}
						if err := json.Unmarshal(data, &resp); err != nil {
							return nil, err
						}
						var rows [][]string
						for _, p := range resp.Data.Payees {
							rows = append(rows, []string{
								p.ID,
								p.Name,
								p.TransferAccountID,
								fmt.Sprintf("%t", p.Deleted),
							})
						}
						return rows, nil
					}

					return ctx.Printer.PrintResult(data, headers, extractRows)
				},
			},
			"get": {
				Name:        "get",
				Description: "Get a specific payee",
				Usage:       "<plan-id> <payee-id>",
				Run: func(ctx *Context, args []string) error {
					var positional []string
					for _, arg := range args {
						if !strings.HasPrefix(arg, "--") {
							positional = append(positional, arg)
						}
					}

					if len(positional) < 2 {
						return fmt.Errorf("usage: ynab payees get <plan-id> <payee-id>")
					}

					planID, err := config.ResolvePlan(positional[0])
					if err != nil {
						return err
					}

					data, err := ctx.Client.GetPayeeByID(planID, positional[1])
					if err != nil {
						return err
					}

					headers := []string{"ID", "NAME", "TRANSFER_ACCOUNT_ID", "DELETED"}
					extractRows := func(data []byte) ([][]string, error) {
						var resp struct {
							Data struct {
								Payee struct {
									ID                string `json:"id"`
									Name              string `json:"name"`
									TransferAccountID string `json:"transfer_account_id"`
									Deleted           bool   `json:"deleted"`
								} `json:"payee"`
							} `json:"data"`
						}
						if err := json.Unmarshal(data, &resp); err != nil {
							return nil, err
						}
						p := resp.Data.Payee
						return [][]string{{
							p.ID,
							p.Name,
							p.TransferAccountID,
							fmt.Sprintf("%t", p.Deleted),
						}}, nil
					}

					return ctx.Printer.PrintResult(data, headers, extractRows)
				},
			},
			"create": {
				Name:        "create",
				Description: "Create a new payee",
				Usage:       "<plan-id> --name NAME",
				Run: func(ctx *Context, args []string) error {
					var planIDArg string
					var name string

					for i := 0; i < len(args); i++ {
						switch {
						case args[i] == "--name" && i+1 < len(args):
							name = args[i+1]
							i++
						case strings.HasPrefix(args[i], "--name="):
							name = strings.TrimPrefix(args[i], "--name=")
						case !strings.HasPrefix(args[i], "--"):
							if planIDArg == "" {
								planIDArg = args[i]
							}
						}
					}

					if name == "" {
						return fmt.Errorf("--name is required")
					}

					planID, err := config.ResolvePlan(planIDArg)
					if err != nil {
						return err
					}

					body := map[string]interface{}{
						"payee": map[string]interface{}{
							"name": name,
						},
					}

					data, err := ctx.Client.CreatePayee(planID, body)
					if err != nil {
						return err
					}

					headers := []string{"ID", "NAME"}
					extractRows := func(data []byte) ([][]string, error) {
						var resp struct {
							Data struct {
								Payee struct {
									ID   string `json:"id"`
									Name string `json:"name"`
								} `json:"payee"`
							} `json:"data"`
						}
						if err := json.Unmarshal(data, &resp); err != nil {
							return nil, err
						}
						p := resp.Data.Payee
						return [][]string{{p.ID, p.Name}}, nil
					}

					return ctx.Printer.PrintResult(data, headers, extractRows)
				},
			},
			"update": {
				Name:        "update",
				Description: "Update a payee",
				Usage:       "<plan-id> <payee-id> --name NAME",
				Run: func(ctx *Context, args []string) error {
					var positional []string
					var name string

					for i := 0; i < len(args); i++ {
						switch {
						case args[i] == "--name" && i+1 < len(args):
							name = args[i+1]
							i++
						case strings.HasPrefix(args[i], "--name="):
							name = strings.TrimPrefix(args[i], "--name=")
						case !strings.HasPrefix(args[i], "--"):
							positional = append(positional, args[i])
						}
					}

					if len(positional) < 2 {
						return fmt.Errorf("usage: ynab payees update <plan-id> <payee-id> --name NAME")
					}
					if name == "" {
						return fmt.Errorf("--name is required")
					}

					planID, err := config.ResolvePlan(positional[0])
					if err != nil {
						return err
					}

					body := map[string]interface{}{
						"payee": map[string]interface{}{
							"name": name,
						},
					}

					data, err := ctx.Client.UpdatePayee(planID, positional[1], body)
					if err != nil {
						return err
					}

					headers := []string{"ID", "NAME"}
					extractRows := func(data []byte) ([][]string, error) {
						var resp struct {
							Data struct {
								Payee struct {
									ID   string `json:"id"`
									Name string `json:"name"`
								} `json:"payee"`
							} `json:"data"`
						}
						if err := json.Unmarshal(data, &resp); err != nil {
							return nil, err
						}
						p := resp.Data.Payee
						return [][]string{{p.ID, p.Name}}, nil
					}

					return ctx.Printer.PrintResult(data, headers, extractRows)
				},
			},
		},
	}
	a.registerCommand(payeesCmd)

	// payee-locations command
	locationsCmd := &Command{
		Name:        "payee-locations",
		Description: "Manage payee locations",
		Subcommands: map[string]*Subcommand{
			"list": {
				Name:        "list",
				Description: "List all payee locations for a plan",
				Usage:       "<plan-id>",
				Run: func(ctx *Context, args []string) error {
					var planIDArg string
					for _, arg := range args {
						if !strings.HasPrefix(arg, "--") {
							if planIDArg == "" {
								planIDArg = arg
							}
						}
					}

					planID, err := config.ResolvePlan(planIDArg)
					if err != nil {
						return err
					}

					data, err := ctx.Client.GetPayeeLocations(planID)
					if err != nil {
						return err
					}

					headers := []string{"ID", "PAYEE_ID", "LATITUDE", "LONGITUDE", "DELETED"}
					extractRows := func(data []byte) ([][]string, error) {
						var resp struct {
							Data struct {
								PayeeLocations []struct {
									ID        string `json:"id"`
									PayeeID   string `json:"payee_id"`
									Latitude  string `json:"latitude"`
									Longitude string `json:"longitude"`
									Deleted   bool   `json:"deleted"`
								} `json:"payee_locations"`
							} `json:"data"`
						}
						if err := json.Unmarshal(data, &resp); err != nil {
							return nil, err
						}
						var rows [][]string
						for _, l := range resp.Data.PayeeLocations {
							rows = append(rows, []string{
								l.ID,
								l.PayeeID,
								l.Latitude,
								l.Longitude,
								fmt.Sprintf("%t", l.Deleted),
							})
						}
						return rows, nil
					}

					return ctx.Printer.PrintResult(data, headers, extractRows)
				},
			},
			"get": {
				Name:        "get",
				Description: "Get a specific payee location",
				Usage:       "<plan-id> <location-id>",
				Run: func(ctx *Context, args []string) error {
					var positional []string
					for _, arg := range args {
						if !strings.HasPrefix(arg, "--") {
							positional = append(positional, arg)
						}
					}

					if len(positional) < 2 {
						return fmt.Errorf("usage: ynab payee-locations get <plan-id> <location-id>")
					}

					planID, err := config.ResolvePlan(positional[0])
					if err != nil {
						return err
					}

					data, err := ctx.Client.GetPayeeLocationByID(planID, positional[1])
					if err != nil {
						return err
					}

					headers := []string{"ID", "PAYEE_ID", "LATITUDE", "LONGITUDE", "DELETED"}
					extractRows := func(data []byte) ([][]string, error) {
						var resp struct {
							Data struct {
								PayeeLocation struct {
									ID        string `json:"id"`
									PayeeID   string `json:"payee_id"`
									Latitude  string `json:"latitude"`
									Longitude string `json:"longitude"`
									Deleted   bool   `json:"deleted"`
								} `json:"payee_location"`
							} `json:"data"`
						}
						if err := json.Unmarshal(data, &resp); err != nil {
							return nil, err
						}
						l := resp.Data.PayeeLocation
						return [][]string{{
							l.ID,
							l.PayeeID,
							l.Latitude,
							l.Longitude,
							fmt.Sprintf("%t", l.Deleted),
						}}, nil
					}

					return ctx.Printer.PrintResult(data, headers, extractRows)
				},
			},
			"list-by-payee": {
				Name:        "list-by-payee",
				Description: "List locations for a specific payee",
				Usage:       "<plan-id> <payee-id>",
				Run: func(ctx *Context, args []string) error {
					var positional []string
					for _, arg := range args {
						if !strings.HasPrefix(arg, "--") {
							positional = append(positional, arg)
						}
					}

					if len(positional) < 2 {
						return fmt.Errorf("usage: ynab payee-locations list-by-payee <plan-id> <payee-id>")
					}

					planID, err := config.ResolvePlan(positional[0])
					if err != nil {
						return err
					}

					data, err := ctx.Client.GetPayeeLocationsByPayee(planID, positional[1])
					if err != nil {
						return err
					}

					headers := []string{"ID", "PAYEE_ID", "LATITUDE", "LONGITUDE", "DELETED"}
					extractRows := func(data []byte) ([][]string, error) {
						var resp struct {
							Data struct {
								PayeeLocations []struct {
									ID        string `json:"id"`
									PayeeID   string `json:"payee_id"`
									Latitude  string `json:"latitude"`
									Longitude string `json:"longitude"`
									Deleted   bool   `json:"deleted"`
								} `json:"payee_locations"`
							} `json:"data"`
						}
						if err := json.Unmarshal(data, &resp); err != nil {
							return nil, err
						}
						var rows [][]string
						for _, l := range resp.Data.PayeeLocations {
							rows = append(rows, []string{
								l.ID,
								l.PayeeID,
								l.Latitude,
								l.Longitude,
								fmt.Sprintf("%t", l.Deleted),
							})
						}
						return rows, nil
					}

					return ctx.Printer.PrintResult(data, headers, extractRows)
				},
			},
		},
	}
	a.registerCommand(locationsCmd)
}
