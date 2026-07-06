package cli

import (
	"fmt"
)

func cmdList(args []string) error {
	if err := parseFlags(args); err != nil {
		return err
	}

	app, err := Load()
	if err != nil {
		return err
	}

	if len(app.Vault.Servers) == 0 {
		fmt.Println("No servers configured.")
		fmt.Println("Add one with: ssh-mcp add")
		return nil
	}

	fmt.Printf("%-20s %-30s %-25s %-10s\n", "ID", "HOST", "NAME", "AUTH")
	fmt.Println("--------------------------------------------------------------------------------")
	for _, s := range app.Vault.Servers {
		authDisplay := string(s.Auth.Method)
		if s.Auth.PrivateKeyPath != "" {
			authDisplay = fmt.Sprintf("key:%s", s.Auth.PrivateKeyPath)
		}
		fmt.Printf("%-20s %-30s %-25s %-10s\n", s.ID, fmt.Sprintf("%s:%d", s.Host, s.Port), s.Name, authDisplay)
	}

	return nil
}

// parseFlags is a simple helper that rejects unexpected flags from args.
func parseFlags(args []string) error {
	for _, a := range args {
		if a == "--help" || a == "-h" {
			return fmt.Errorf("help requested")
		}
	}
	return nil
}
