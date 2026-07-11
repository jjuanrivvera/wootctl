package commands

import (
	"fmt"
	"slices"
	"sort"
	"strings"

	"github.com/google/shlex"
	"github.com/spf13/cobra"

	"github.com/jjuanrivvera/wootctl/internal/config"
)

// ExpandAliases rewrites os.Args before cobra parses them: if the first argument names a
// user-defined alias AND is not a real built-in command, it is replaced by the alias's
// expansion. Built-ins always win, so an alias can never shadow `auth`, `conversations`, etc.
func ExpandAliases(args []string) []string {
	if len(args) == 0 {
		return args
	}
	name := args[0]
	if isBuiltinCommand(name) {
		return args
	}
	cfg, err := config.Load()
	if err != nil || cfg.Aliases == nil {
		return args
	}
	expansion, ok := cfg.Aliases[name]
	if !ok {
		return args
	}
	parts, err := shlex.Split(expansion)
	if err != nil || len(parts) == 0 {
		return args
	}
	return append(parts, args[1:]...)
}

// isBuiltinCommand reports whether name is a registered top-level command on a fresh tree.
func isBuiltinCommand(name string) bool {
	for _, c := range NewRootCmd().Commands() {
		if c.Name() == name || slices.Contains(c.Aliases, name) {
			return true
		}
	}
	return false
}

func init() {
	metaRegistrars = append(metaRegistrars, func(d *deps) *cobra.Command {
		aliasCmd := &cobra.Command{
			Use:   "alias",
			Short: "Manage user-defined command aliases",
			Long:  "Define shorthand commands. Aliases are expanded before parsing and can never shadow a built-in.",
		}

		setCmd := &cobra.Command{
			Use:   "set <name> <expansion>",
			Short: "Create or update an alias",
			Example: `  wootctl alias set open "conversations list --status open"
  wootctl alias set reply "messages create"`,
			Args: cobra.MinimumNArgs(2),
			RunE: func(cmd *cobra.Command, args []string) error {
				name := args[0]
				if isBuiltinCommand(name) {
					return fmt.Errorf("%q is a built-in command and cannot be aliased", name)
				}
				if err := config.ValidateProfileName(name); err != nil {
					return fmt.Errorf("invalid alias name: %w", err)
				}
				expansion := strings.Join(args[1:], " ")
				cfg, err := d.loadConfig()
				if err != nil {
					return err
				}
				if cfg.Aliases == nil {
					cfg.Aliases = map[string]string{}
				}
				cfg.Aliases[name] = expansion
				if err := cfg.Save(); err != nil {
					return err
				}
				fmt.Fprintf(cmd.OutOrStdout(), "alias %q → %q\n", name, expansion)
				return nil
			},
		}

		listCmd := &cobra.Command{
			Use:     "list",
			Aliases: []string{"ls"},
			Short:   "List aliases",
			Args:    cobra.NoArgs,
			RunE: func(cmd *cobra.Command, _ []string) error {
				cfg, err := d.loadConfig()
				if err != nil {
					return err
				}
				names := make([]string, 0, len(cfg.Aliases))
				for n := range cfg.Aliases {
					names = append(names, n)
				}
				sort.Strings(names)
				for _, n := range names {
					fmt.Fprintf(cmd.OutOrStdout(), "%s = %s\n", n, cfg.Aliases[n])
				}
				return nil
			},
		}

		removeCmd := &cobra.Command{
			Use:     "remove <name>",
			Aliases: []string{"rm", "delete"},
			Short:   "Remove an alias",
			Args:    cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				cfg, err := d.loadConfig()
				if err != nil {
					return err
				}
				if _, ok := cfg.Aliases[args[0]]; !ok {
					return fmt.Errorf("no such alias %q", args[0])
				}
				delete(cfg.Aliases, args[0])
				if err := cfg.Save(); err != nil {
					return err
				}
				fmt.Fprintf(cmd.OutOrStdout(), "removed alias %q\n", args[0])
				return nil
			},
		}

		aliasCmd.AddCommand(setCmd, listCmd, removeCmd)
		return aliasCmd
	})
}
