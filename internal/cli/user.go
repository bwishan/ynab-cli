package cli

import (
"encoding/json"

"github.com/spf13/cobra"
)

func (a *App) registerUserCommands() {
userCmd := &cobra.Command{
Use:   "user",
Short: "User account information",
}

getCmd := &cobra.Command{
Use:   "get",
Short: "Get the authenticated user",
RunE: func(cmd *cobra.Command, args []string) error {
data, err := a.client.GetUser()
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

return a.printer.PrintResult(data, headers, extractRows)
},
}

userCmd.AddCommand(getCmd)
a.rootCmd.AddCommand(userCmd)
}
