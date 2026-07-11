package commands

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/spf13/cobra"

	"github.com/jjuanrivvera/wootctl/internal/api"
)

// reports wraps the read-only /api/v2 analytics endpoints plus v1 reporting_events. Every
// verb is a query-driven GET; --since/--until accept unix seconds or YYYY-MM-DD.
func init() {
	registerResource("", resourceSpec[api.Rec]{
		Use:     "reports",
		Aliases: []string{"report"},
		Short:   "Account analytics (v2 reports, summaries, reporting events)",
		New: func(c *api.Client) *api.Resource[api.Rec] {
			return api.NewResource[api.Rec](c, c.AccountPathV2("reports"))
		},
		Columns: []string{"value", "timestamp"},
		NoList:  true, NoGet: true, NoCreate: true, NoUpdate: true, NoDelete: true,
		Extra: []extraCommand{
			{Kind: kindRead, Build: reportsOverviewCmd},
			{Kind: kindRead, Build: reportsSummaryCmd},
			{Kind: kindRead, Build: reportsAccountConversationsCmd},
			{Kind: kindRead, Build: reportsAgentConversationsCmd},
			{Kind: kindRead, Build: reportsSimpleCmd("first-response-time-distribution", "reports/first_response_time_distribution", "First-response-time distribution buckets", nil)},
			{Kind: kindRead, Build: reportsInboxLabelMatrixCmd},
			{Kind: kindRead, Build: reportsOutgoingMessagesCountCmd},
			{Kind: kindRead, Build: reportsSummaryReportCmd("agent-summary", "agent")},
			{Kind: kindRead, Build: reportsSummaryReportCmd("channel-summary", "channel")},
			{Kind: kindRead, Build: reportsSummaryReportCmd("inbox-summary", "inbox")},
			{Kind: kindRead, Build: reportsSummaryReportCmd("team-summary", "team")},
			{Kind: kindRead, Build: reportsEventsCmd},
		},
	})
}

// timeParam accepts unix seconds or YYYY-MM-DD and normalizes to unix seconds, which every
// reports endpoint expects.
func timeParam(v string) (string, error) {
	if v == "" {
		return "", nil
	}
	if _, err := strconv.ParseInt(v, 10, 64); err == nil {
		return v, nil
	}
	t, err := time.Parse("2006-01-02", v)
	if err != nil {
		return "", fmt.Errorf("invalid time %q (want unix seconds or YYYY-MM-DD)", v)
	}
	return strconv.FormatInt(t.Unix(), 10), nil
}

// rangeFlags wires --since/--until and returns a query builder.
func rangeFlags(cmd *cobra.Command) func() (url.Values, error) {
	var since, until string
	cmd.Flags().StringVar(&since, "since", "", "range start: unix seconds or YYYY-MM-DD")
	cmd.Flags().StringVar(&until, "until", "", "range end: unix seconds or YYYY-MM-DD")
	return func() (url.Values, error) {
		q := url.Values{}
		s, err := timeParam(since)
		if err != nil {
			return nil, err
		}
		u, err := timeParam(until)
		if err != nil {
			return nil, err
		}
		if s != "" {
			q.Set("since", s)
		}
		if u != "" {
			q.Set("until", u)
		}
		return q, nil
	}
}

func reportsGet(d *deps, cmd *cobra.Command, c *api.Client, sub string, q url.Values, cols []string) (json.RawMessage, error) {
	var out json.RawMessage
	err := c.Send(cmd.Context(), http.MethodGet, c.AccountPathV2(sub), q, nil, &out)
	return out, err
}

func reportsOverviewCmd(d *deps) *cobra.Command {
	var metric, rtype, id string
	cmd := &cobra.Command{
		Use:   "overview",
		Short: "Time-series statistics for a metric (account/agent/inbox/label/team)",
		Example: `  wootctl reports overview --metric conversations_count --type account --since 2026-06-01 --until 2026-07-01
  wootctl reports overview --metric avg_first_response_time --type inbox --id 3 --since 2026-06-01`,
		Args: cobra.NoArgs,
	}
	rng := rangeFlags(cmd)
	cmd.Flags().StringVar(&metric, "metric", "", "conversations_count | incoming_messages_count | outgoing_messages_count | avg_first_response_time | avg_resolution_time | resolutions_count | bot_resolutions_count | bot_handoffs_count | reply_time")
	cmd.Flags().StringVar(&rtype, "type", "account", "account | agent | inbox | label | team")
	cmd.Flags().StringVar(&id, "id", "", "object id when type != account")
	_ = cmd.MarkFlagRequired("metric")
	cmd.RunE = runE(d, false, []string{"value", "timestamp"}, func(cmd *cobra.Command, c *api.Client, _ []string) (json.RawMessage, error) {
		q, err := rng()
		if err != nil {
			return nil, err
		}
		q.Set("metric", metric)
		q.Set("type", rtype)
		if id != "" {
			q.Set("id", id)
		}
		return reportsGet(d, cmd, c, "reports", q, nil)
	})
	return cmd
}

func reportsSummaryCmd(d *deps) *cobra.Command {
	var rtype, id string
	cmd := &cobra.Command{
		Use:     "summary",
		Short:   "Aggregate statistics for a range (conversations, response times, resolutions)",
		Example: "  wootctl reports summary --since 2026-06-01 --until 2026-07-01",
		Args:    cobra.NoArgs,
	}
	rng := rangeFlags(cmd)
	cmd.Flags().StringVar(&rtype, "type", "account", "account | agent | inbox | label | team")
	cmd.Flags().StringVar(&id, "id", "", "object id when type != account")
	cmd.RunE = runE(d, false, nil, func(cmd *cobra.Command, c *api.Client, _ []string) (json.RawMessage, error) {
		q, err := rng()
		if err != nil {
			return nil, err
		}
		q.Set("type", rtype)
		if id != "" {
			q.Set("id", id)
		}
		return reportsGet(d, cmd, c, "reports/summary", q, nil)
	})
	return cmd
}

func reportsAccountConversationsCmd(d *deps) *cobra.Command {
	return &cobra.Command{
		Use:     "account-conversations",
		Short:   "Account-level open/unattended conversation metrics",
		Example: "  wootctl reports account-conversations",
		Args:    cobra.NoArgs,
		RunE: runE(d, false, nil, func(cmd *cobra.Command, c *api.Client, _ []string) (json.RawMessage, error) {
			q := url.Values{"type": {"account"}}
			return reportsGet(d, cmd, c, "reports/conversations", q, nil)
		}),
	}
}

func reportsAgentConversationsCmd(d *deps) *cobra.Command {
	var userID string
	cmd := &cobra.Command{
		Use:     "agent-conversations",
		Short:   "Per-agent conversation metrics (all agents, or one with --user-id)",
		Example: "  wootctl reports agent-conversations --user-id 7",
		Args:    cobra.NoArgs,
		RunE: runE(d, false, nil, func(cmd *cobra.Command, c *api.Client, _ []string) (json.RawMessage, error) {
			q := url.Values{"type": {"agent"}}
			if userID != "" {
				q.Set("user_id", userID)
			}
			return reportsGet(d, cmd, c, "reports/conversations", q, nil)
		}),
	}
	cmd.Flags().StringVar(&userID, "user-id", "", "limit to one agent")
	return cmd
}

// reportsSimpleCmd covers range-only endpoints (first_response_time_distribution).
func reportsSimpleCmd(use, sub, short string, cols []string) func(d *deps) *cobra.Command {
	return func(d *deps) *cobra.Command {
		cmd := &cobra.Command{
			Use:     use,
			Short:   short,
			Example: "  wootctl reports " + use + " --since 2026-06-01 --until 2026-07-01",
			Args:    cobra.NoArgs,
		}
		rng := rangeFlags(cmd)
		cmd.RunE = runE(d, false, cols, func(cmd *cobra.Command, c *api.Client, _ []string) (json.RawMessage, error) {
			q, err := rng()
			if err != nil {
				return nil, err
			}
			return reportsGet(d, cmd, c, sub, q, nil)
		})
		return cmd
	}
}

func reportsInboxLabelMatrixCmd(d *deps) *cobra.Command {
	var inboxIDs, labelIDs []string
	cmd := &cobra.Command{
		Use:     "inbox-label-matrix",
		Short:   "Conversation counts as an inbox × label matrix",
		Example: "  wootctl reports inbox-label-matrix --since 2026-06-01 --inbox-ids 1,3",
		Args:    cobra.NoArgs,
	}
	rng := rangeFlags(cmd)
	cmd.Flags().StringSliceVar(&inboxIDs, "inbox-ids", nil, "restrict to these inbox ids")
	cmd.Flags().StringSliceVar(&labelIDs, "label-ids", nil, "restrict to these label ids")
	cmd.RunE = runE(d, false, nil, func(cmd *cobra.Command, c *api.Client, _ []string) (json.RawMessage, error) {
		q, err := rng()
		if err != nil {
			return nil, err
		}
		for _, v := range inboxIDs {
			q.Add("inbox_ids[]", v)
		}
		for _, v := range labelIDs {
			q.Add("label_ids[]", v)
		}
		return reportsGet(d, cmd, c, "reports/inbox_label_matrix", q, nil)
	})
	return cmd
}

func reportsOutgoingMessagesCountCmd(d *deps) *cobra.Command {
	var groupBy string
	cmd := &cobra.Command{
		Use:     "outgoing-messages-count",
		Short:   "Outgoing message counts grouped by day/week/month/year",
		Example: "  wootctl reports outgoing-messages-count --group-by day --since 2026-06-01",
		Args:    cobra.NoArgs,
	}
	rng := rangeFlags(cmd)
	cmd.Flags().StringVar(&groupBy, "group-by", "", "day | week | month | year")
	_ = cmd.MarkFlagRequired("group-by")
	cmd.RunE = runE(d, false, nil, func(cmd *cobra.Command, c *api.Client, _ []string) (json.RawMessage, error) {
		q, err := rng()
		if err != nil {
			return nil, err
		}
		q.Set("group_by", groupBy)
		return reportsGet(d, cmd, c, "reports/outgoing_messages_count", q, nil)
	})
	return cmd
}

// reportsSummaryReportCmd covers the four /summary_reports/<kind> endpoints.
func reportsSummaryReportCmd(use, kind string) func(d *deps) *cobra.Command {
	return func(d *deps) *cobra.Command {
		var businessHours bool
		cmd := &cobra.Command{
			Use:     use,
			Short:   "Summary report grouped by " + kind,
			Example: "  wootctl reports " + use + " --since 2026-06-01 --until 2026-07-01",
			Args:    cobra.NoArgs,
		}
		rng := rangeFlags(cmd)
		cmd.Flags().BoolVar(&businessHours, "business-hours", false, "compute times within business hours only")
		cmd.RunE = runE(d, false, nil, func(cmd *cobra.Command, c *api.Client, _ []string) (json.RawMessage, error) {
			q, err := rng()
			if err != nil {
				return nil, err
			}
			if businessHours {
				q.Set("business_hours", "true")
			}
			return reportsGet(d, cmd, c, "summary_reports/"+kind, q, nil)
		})
		return cmd
	}
}

func reportsEventsCmd(d *deps) *cobra.Command {
	var inboxID, userID, name string
	cmd := &cobra.Command{
		Use:     "events",
		Short:   "List account reporting events (first response, resolutions, …)",
		Example: "  wootctl reports events --name first_response --since 2026-06-01",
		Args:    cobra.NoArgs,
	}
	rng := rangeFlags(cmd)
	cmd.Flags().StringVar(&inboxID, "inbox-id", "", "only this inbox")
	cmd.Flags().StringVar(&userID, "user-id", "", "only this agent")
	cmd.Flags().StringVar(&name, "name", "", "event name: conversation_creation | first_response | conversation_resolved | …")
	cmd.RunE = runListE(d, false, []string{"id", "name", "value", "user_id", "created_at"}, func(cmd *cobra.Command, c *api.Client, _ []string) (json.RawMessage, error) {
		q, err := rng()
		if err != nil {
			return nil, err
		}
		for k, v := range map[string]string{"inbox_id": inboxID, "user_id": userID, "name": name} {
			if v != "" {
				q.Set(k, v)
			}
		}
		if d.gf.page > 0 {
			q.Set("page", itoa(d.gf.page))
		}
		var out json.RawMessage
		err = c.Send(cmd.Context(), http.MethodGet, c.AccountPath("reporting_events"), q, nil, &out)
		return out, err
	})
	return cmd
}
