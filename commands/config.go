package commands

import (
	"fmt"
	"sort"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/jjuanrivvera/wootctl/internal/config"
)

func init() {
	metaRegistrars = append(metaRegistrars, func(d *deps) *cobra.Command {
		configCmd := &cobra.Command{
			Use:   "config",
			Short: "Inspect and edit wootctl configuration",
			Long:  "View the config file, switch profiles, and set per-profile options. Secrets live in the keyring and are never shown here.",
		}
		configCmd.AddCommand(configPathCmd(d), configViewCmd(d), configSetCmd(d), configUseCmd(d), configListProfilesCmd(d))
		return configCmd
	})
}

func configPathCmd(_ *deps) *cobra.Command {
	return &cobra.Command{
		Use:   "path",
		Short: "Print the config file path",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			p, err := config.Path()
			if err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), p)
			return nil
		},
	}
}

func configViewCmd(d *deps) *cobra.Command {
	return &cobra.Command{
		Use:     "view",
		Aliases: []string{"show"},
		Short:   "Show the current configuration (secrets redacted)",
		Example: `  wootctl config view
  wootctl config view -o json`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := d.loadConfig()
			if err != nil {
				return err
			}
			// Tokens are never in the config file — note where they live so `view` is
			// self-explanatory.
			view := map[string]any{
				"config_path":     cfg.FilePath(),
				"current_profile": cfg.CurrentProfile,
				"profiles":        cfg.Profiles,
				"aliases":         cfg.Aliases,
				"token_storage":   "OS keyring with encrypted-file fallback (run `wootctl auth status` to verify)",
			}
			return d.render(cmd, mustJSON(view), nil)
		},
	}
}

func configSetCmd(d *deps) *cobra.Command {
	return &cobra.Command{
		Use:   "set <key> <value>",
		Short: "Set a per-profile option (keys: base_url, account_id, rps)",
		Example: `  wootctl config set base_url https://app.chatwoot.com
  wootctl config set account_id 2
  wootctl --profile staging config set rps 2`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			key, value := args[0], args[1]
			cfg, err := d.loadConfig()
			if err != nil {
				return err
			}
			profileName := cfg.ResolveProfileName(d.gf.profile)
			prof, _ := cfg.Profile(profileName)
			switch key {
			case "base_url", "base-url":
				if err := config.ValidateBaseURL(value); err != nil {
					return err
				}
				prof.BaseURL = value
			case "account_id", "account-id":
				prof.AccountID = value
			case "rps":
				f, err := strconv.ParseFloat(value, 64)
				if err != nil || f <= 0 {
					return fmt.Errorf("rps must be a positive number, got %q", value)
				}
				prof.Rps = f
			default:
				return fmt.Errorf("unknown config key %q (supported: base_url, account_id, rps)", key)
			}
			if err := cfg.SetProfile(profileName, prof); err != nil {
				return err
			}
			if err := cfg.Save(); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "set %s=%s on profile %q\n", key, value, profileName)
			return nil
		},
	}
}

func configUseCmd(d *deps) *cobra.Command {
	return &cobra.Command{
		Use:   "use <profile>",
		Short: "Switch the active profile",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := d.loadConfig()
			if err != nil {
				return err
			}
			if _, ok := cfg.Profile(args[0]); !ok {
				return fmt.Errorf("no such profile %q — create it with `wootctl --profile %s auth login`", args[0], args[0])
			}
			cfg.CurrentProfile = args[0]
			if err := cfg.Save(); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "now using profile %q\n", args[0])
			return nil
		},
	}
}

func configListProfilesCmd(d *deps) *cobra.Command {
	return &cobra.Command{
		Use:     "list-profiles",
		Aliases: []string{"profiles"},
		Short:   "List configured profiles",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := d.loadConfig()
			if err != nil {
				return err
			}
			names := cfg.ProfileNames()
			sort.Strings(names)
			rows := make([]map[string]any, 0, len(names))
			for _, n := range names {
				p, _ := cfg.Profile(n)
				rows = append(rows, map[string]any{
					"profile":    n,
					"current":    n == cfg.CurrentProfile,
					"base_url":   p.BaseURL,
					"account_id": p.AccountID,
					"email":      p.Email,
				})
			}
			return d.render(cmd, mustJSON(rows), []string{"profile", "current", "base_url", "account_id", "email"})
		},
	}
}
