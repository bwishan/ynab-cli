package cli

import (
"encoding/json"

"github.com/spf13/cobra"

"github.com/bwishan/ynab-cli/internal/config"
)

func (a *App) registerPlanCommands() {
plansCmd := &cobra.Command{
Use:   "plans",
Short: "Manage budgets/plans",
}

listCmd := &cobra.Command{
Use:   "list",
Short: "List all plans",
RunE: func(cmd *cobra.Command, args []string) error {
includeAccounts, _ := cmd.Flags().GetBool("include-accounts")

data, err := a.client.GetPlans(includeAccounts)
if err != nil {
return err
}

headers := []string{"ID", "NAME", "LAST_MODIFIED"}
extractRows := func(data []byte) ([][]string, error) {
var resp struct {
Data struct {
Budgets []struct {
ID           string `json:"id"`
Name         string `json:"name"`
LastModified string `json:"last_modified_on"`
} `json:"budgets"`
} `json:"data"`
}
if err := json.Unmarshal(data, &resp); err != nil {
return nil, err
}
var rows [][]string
for _, b := range resp.Data.Budgets {
rows = append(rows, []string{b.ID, b.Name, b.LastModified})
}
return rows, nil
}

return a.printer.PrintResult(data, headers, extractRows)
},
}
listCmd.Flags().Bool("include-accounts", false, "Include accounts in response")

getCmd := &cobra.Command{
Use:   "get [plan-id]",
Short: "Get a specific plan",
RunE: func(cmd *cobra.Command, args []string) error {
var planIDArg string
if len(args) > 0 {
planIDArg = args[0]
}

planID, err := config.ResolvePlan(planIDArg)
if err != nil {
return err
}

lastKnowledge, _ := cmd.Flags().GetInt64("last-knowledge")

data, err := a.client.GetPlanByID(planID, lastKnowledge)
if err != nil {
return err
}

headers := []string{"ID", "NAME", "LAST_MODIFIED"}
extractRows := func(data []byte) ([][]string, error) {
var resp struct {
Data struct {
Budget struct {
ID           string `json:"id"`
Name         string `json:"name"`
LastModified string `json:"last_modified_on"`
} `json:"budget"`
} `json:"data"`
}
if err := json.Unmarshal(data, &resp); err != nil {
return nil, err
}
b := resp.Data.Budget
return [][]string{{b.ID, b.Name, b.LastModified}}, nil
}

return a.printer.PrintResult(data, headers, extractRows)
},
}
getCmd.Flags().Int64("last-knowledge", 0, "Server knowledge for delta sync")

settingsCmd := &cobra.Command{
Use:   "settings [plan-id]",
Short: "Get plan settings",
RunE: func(cmd *cobra.Command, args []string) error {
var planIDArg string
if len(args) > 0 {
planIDArg = args[0]
}

planID, err := config.ResolvePlan(planIDArg)
if err != nil {
return err
}

data, err := a.client.GetPlanSettings(planID)
if err != nil {
return err
}

headers := []string{"CURRENCY_FORMAT", "DATE_FORMAT"}
extractRows := func(data []byte) ([][]string, error) {
var resp struct {
Data struct {
Settings struct {
DateFormat struct {
Format string `json:"format"`
} `json:"date_format"`
CurrencyFormat struct {
ISOCode string `json:"iso_code"`
} `json:"currency_format"`
} `json:"settings"`
} `json:"data"`
}
if err := json.Unmarshal(data, &resp); err != nil {
return nil, err
}
s := resp.Data.Settings
return [][]string{{s.CurrencyFormat.ISOCode, s.DateFormat.Format}}, nil
}

return a.printer.PrintResult(data, headers, extractRows)
},
}

plansCmd.AddCommand(listCmd, getCmd, settingsCmd)
a.rootCmd.AddCommand(plansCmd)
}
