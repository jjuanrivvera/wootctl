package commands

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"slices"
	"strings"

	"github.com/spf13/cobra"
)

func init() {
	metaRegistrars = append(metaRegistrars, func(d *deps) *cobra.Command {
		var data string
		var query []string
		cmd := &cobra.Command{
			Use:   "api <METHOD> <PATH> [-d body] [-q key=value ...]",
			Short: "Send a raw authenticated request (escape hatch)",
			Long: `Call any Chatwoot endpoint directly. The path is relative to the instance root
(e.g. api/v1/accounts/1/labels, platform/api/v1/users, public/api/v1/inboxes/…);
the credential class is chosen from the path exactly like first-class commands
(platform/* uses the platform token; public/* sends none).

This is the documented escape hatch for anything wootctl does not wrap as a
first-class command. It honors --dry-run, -o/--output, and --jq like every other
command. Non-GET methods are never auto-retried.`,
			Example: `  wootctl api GET api/v1/accounts/1/conversations/42
  wootctl api GET api/v2/accounts/1/reports/summary -q since=1719800000 -q until=1722400000
  wootctl api POST api/v1/accounts/1/labels -d '{"title":"vip","color":"#0055ff"}'
  wootctl api DELETE api/v1/accounts/1/labels/9 --dry-run`,
			Args: cobra.ExactArgs(2),
			RunE: func(cmd *cobra.Command, args []string) error {
				method := strings.ToUpper(args[0])
				valid := []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodHead, http.MethodOptions}
				if !slices.Contains(valid, method) {
					return fmt.Errorf("invalid method %q (want one of %s)", args[0], strings.Join(valid, "|"))
				}
				path := strings.TrimLeft(args[1], "/")

				q := url.Values{}
				for _, kv := range query {
					k, v, ok := strings.Cut(kv, "=")
					if !ok {
						return fmt.Errorf("invalid -q %q (want key=value)", kv)
					}
					q.Add(k, v)
				}

				var body *bytes.Reader
				if data != "" {
					raw, err := readDataArg(cmd, data)
					if err != nil {
						return err
					}
					body = bytes.NewReader(raw)
				}

				c, _, err := d.getAPIClient(false)
				if err != nil {
					return err
				}
				var status int
				var respBody []byte
				if body == nil {
					status, _, respBody, err = c.Raw(cmd.Context(), method, path, q, nil)
				} else {
					status, _, respBody, err = c.Raw(cmd.Context(), method, path, q, body)
				}
				if err != nil {
					return err
				}
				if status == 0 { // dry-run
					return nil
				}
				if len(respBody) == 0 {
					if !d.gf.quiet {
						fmt.Fprintf(cmd.OutOrStdout(), "HTTP %d (empty body)\n", status)
					}
					return nil
				}
				if json.Valid(respBody) {
					return d.render(cmd, json.RawMessage(respBody), nil)
				}
				// Non-JSON (the CSAT survey page, exports): print raw so pipes still work.
				_, err = cmd.OutOrStdout().Write(respBody)
				return err
			},
		}
		cmd.Flags().StringVarP(&data, "data", "d", "", "JSON body: inline, @file, or - for stdin")
		cmd.Flags().StringArrayVarP(&query, "query", "q", nil, "query parameter key=value (repeatable)")
		return annotate(cmd, kindWrite) // raw calls may mutate; the guard gates by METHOD (§3b.6)
	})
}
