package cli

import (
	"flag"
	"fmt"
	"os"
)

func cmdRemove(args []string) error {
	fs := flag.NewFlagSet("remove", flag.ExitOnError)
	id := fs.String("id", "", "Server identifier to remove")
	fs.Parse(args)

	if *id == "" {
		fmt.Fprintln(os.Stderr, "Required flag: --id")
		return fmt.Errorf("missing --id")
	}

	app, err := Load()
	if err != nil {
		return err
	}

	for i, s := range app.Vault.Servers {
		if s.ID == *id {
			app.Vault.Servers = append(app.Vault.Servers[:i], app.Vault.Servers[i+1:]...)
			if err := app.Save(); err != nil {
				return err
			}
			fmt.Printf("Server '%s' removed.\n", *id)
			return nil
		}
	}

	return fmt.Errorf("server '%s' not found", *id)
}
