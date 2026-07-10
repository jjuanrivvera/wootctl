package api

import "encoding/json"

// Rec is the record type every resource handle carries. Chatwoot is a fast-moving
// self-hosted API that adds response fields every release; decoding into raw JSON means
// -o json/-o yaml never silently drop fields a typed struct didn't know about
// (DECISIONS.md #19). Typed structs are used only where cwctl itself consumes fields.
type Rec = json.RawMessage

// --- application API (account-scoped unless noted) ---

func (c *Client) AgentBots() *Resource[Rec] { return NewResource[Rec](c, c.AccountPath("agent_bots")) }
func (c *Client) Agents() *Resource[Rec]    { return NewResource[Rec](c, c.AccountPath("agents")) }
func (c *Client) AuditLogs() *Resource[Rec] { return NewResource[Rec](c, c.AccountPath("audit_logs")) }
func (c *Client) AutomationRules() *Resource[Rec] {
	return NewResource[Rec](c, c.AccountPath("automation_rules"))
}
func (c *Client) CannedResponses() *Resource[Rec] {
	return NewResource[Rec](c, c.AccountPath("canned_responses"))
}
func (c *Client) Contacts() *Resource[Rec] {
	return NewResource[Rec](c, c.AccountPath("contacts")).WithUpdateMethod("PUT")
}
func (c *Client) Conversations() *Resource[Rec] {
	return NewResource[Rec](c, c.AccountPath("conversations"))
}
func (c *Client) CustomAttributes() *Resource[Rec] {
	return NewResource[Rec](c, c.AccountPath("custom_attribute_definitions"))
}
func (c *Client) CustomFilters() *Resource[Rec] {
	return NewResource[Rec](c, c.AccountPath("custom_filters"))
}
func (c *Client) Inboxes() *Resource[Rec] { return NewResource[Rec](c, c.AccountPath("inboxes")) }
func (c *Client) Labels() *Resource[Rec]  { return NewResource[Rec](c, c.AccountPath("labels")) }
func (c *Client) Portals() *Resource[Rec] { return NewResource[Rec](c, c.AccountPath("portals")) }
func (c *Client) Teams() *Resource[Rec]   { return NewResource[Rec](c, c.AccountPath("teams")) }
func (c *Client) Webhooks() *Resource[Rec] {
	return NewResource[Rec](c, c.AccountPath("webhooks"))
}

// --- platform API (platform app token; selected by path prefix) ---

func (c *Client) PlatformAccounts() *Resource[Rec] {
	return NewResource[Rec](c, "platform/api/v1/accounts")
}
func (c *Client) PlatformAgentBots() *Resource[Rec] {
	return NewResource[Rec](c, "platform/api/v1/agent_bots")
}
func (c *Client) PlatformUsers() *Resource[Rec] {
	return NewResource[Rec](c, "platform/api/v1/users")
}
