package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"
	"sort"
	"strconv"
)

// DefaultPageSize is Chatwoot's server-fixed page size on paginated lists. It is only a
// hint for humans; the --all walker never uses it as a stop condition (page sizes differ
// per endpoint), see Resource.ListAll.
const DefaultPageSize = 25

// ListParams drives a single list request. Chatwoot paginates with a 1-based `page` query
// param; sorting and resource-specific filters ride Extra.
type ListParams struct {
	Page  int               // 1-based; 0 means "let the server default"
	Sort  string            // e.g. "name", "-last_activity_at" (resource-dependent)
	Extra map[string]string // resource-specific query params (status, inbox_id, q, …)
}

// values renders the params as url.Values, omitting zero/empty fields so the request
// stays minimal and deterministic.
func (p ListParams) values() url.Values {
	v := url.Values{}
	if p.Page > 0 {
		v.Set("page", strconv.Itoa(p.Page))
	}
	if p.Sort != "" {
		v.Set("sort", p.Sort)
	}
	// Sorted for a deterministic URL (and dry-run curl) — map order is random.
	keys := make([]string, 0, len(p.Extra))
	for k := range p.Extra {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		if val := p.Extra[k]; val != "" {
			v.Set(k, val)
		}
	}
	return v
}

// ListMeta carries the count metadata Chatwoot embeds in list envelopes. Different
// endpoints use different keys (conversations: all_count…, contacts: count/current_page,
// audit logs: total_entries/per_page) — one struct absorbs them all. current_page arrives
// as a string on some endpoints, hence FlexInt everywhere.
type ListMeta struct {
	Count           FlexInt `json:"count,omitempty"`
	CurrentPage     FlexInt `json:"current_page,omitempty"`
	TotalEntries    FlexInt `json:"total_entries,omitempty"`
	PerPage         FlexInt `json:"per_page,omitempty"`
	AllCount        FlexInt `json:"all_count,omitempty"`
	MineCount       FlexInt `json:"mine_count,omitempty"`
	UnassignedCount FlexInt `json:"unassigned_count,omitempty"`
	AssignedCount   FlexInt `json:"assigned_count,omitempty"`
}

// total returns the advertised total item count, or -1 when the envelope had none.
// all_count is deliberately NOT used: on filtered conversation lists it reports the
// unfiltered total, which would truncate the walk (DECISIONS.md #7).
func (m *ListMeta) total() int {
	if m == nil {
		return -1
	}
	if m.Count > 0 {
		return int(m.Count)
	}
	if m.TotalEntries > 0 {
		return int(m.TotalEntries)
	}
	return -1
}

// decodeList normalizes every list envelope Chatwoot uses into a typed slice + meta:
// a bare array, {"payload":[…]}, {"data":[…]}, {"data":{"meta":…,"payload":[…]}}
// (conversations), and single-array-key wrappers like {"audit_logs":[…]}.
func decodeList[T any](data []byte) ([]T, *ListMeta, error) {
	trim := bytes.TrimSpace(data)
	if len(trim) == 0 || string(trim) == "null" {
		return nil, nil, nil
	}
	if trim[0] == '[' {
		var items []T
		err := json.Unmarshal(trim, &items)
		return items, nil, err
	}
	if trim[0] != '{' {
		return nil, nil, fmt.Errorf("unrecognized list response (not JSON array or object)")
	}

	var env struct {
		Meta    *ListMeta       `json:"meta"`
		Payload json.RawMessage `json:"payload"`
		Data    json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(trim, &env); err != nil {
		return nil, nil, fmt.Errorf("decode list envelope: %w", err)
	}
	if p := bytes.TrimSpace(env.Payload); len(p) > 0 {
		if p[0] == '[' {
			var items []T
			err := json.Unmarshal(p, &items)
			return items, env.Meta, err
		}
		if p[0] == '{' { // {"payload":{"webhooks":[…]}} — the double-nested webhooks shape
			items, meta, err := decodeList[T](p)
			if err == nil {
				if meta == nil {
					meta = env.Meta
				}
				return items, meta, nil
			}
		}
	}
	if d := bytes.TrimSpace(env.Data); len(d) > 0 {
		if d[0] == '[' {
			var items []T
			err := json.Unmarshal(d, &items)
			return items, env.Meta, err
		}
		if d[0] == '{' { // {"data":{"meta":…,"payload":[…]}} — the conversations shape
			var inner struct {
				Meta    *ListMeta       `json:"meta"`
				Payload json.RawMessage `json:"payload"`
			}
			if err := json.Unmarshal(d, &inner); err != nil {
				return nil, nil, fmt.Errorf("decode nested list envelope: %w", err)
			}
			if arr := bytes.TrimSpace(inner.Payload); len(arr) > 0 && arr[0] == '[' {
				var items []T
				err := json.Unmarshal(arr, &items)
				meta := inner.Meta
				if meta == nil {
					meta = env.Meta
				}
				return items, meta, err
			}
		}
	}

	// Last resort: an envelope with exactly ONE array-valued key ({"audit_logs":[…]}).
	// "Exactly one" keeps this deterministic — two arrays would be a guess.
	var generic map[string]json.RawMessage
	if err := json.Unmarshal(trim, &generic); err != nil {
		return nil, nil, fmt.Errorf("decode list envelope: %w", err)
	}
	var arrKeys []string
	for k, v := range generic {
		if raw := bytes.TrimSpace(v); len(raw) > 0 && raw[0] == '[' {
			arrKeys = append(arrKeys, k)
		}
	}
	if len(arrKeys) == 1 {
		var items []T
		err := json.Unmarshal(generic[arrKeys[0]], &items)
		var meta ListMeta
		// Counts may sit beside the array (audit logs) rather than under "meta".
		_ = json.Unmarshal(trim, &meta)
		return items, &meta, err
	}
	return nil, nil, fmt.Errorf("unrecognized list envelope (keys with arrays: %d)", len(arrKeys))
}

// decodeOne unwraps single-object envelopes ({"payload":{…}} / {"data":{…}}) one level,
// else decodes the object directly. No Chatwoot model has a top-level payload/data field,
// so the probe cannot misfire on real records. Non-generic on purpose: callers hold `any`.
func decodeOne(data []byte, out any) error {
	trim := bytes.TrimSpace(data)
	if len(trim) == 0 || string(trim) == "null" {
		return nil
	}
	if trim[0] == '{' {
		var env struct {
			Payload json.RawMessage `json:"payload"`
			Data    json.RawMessage `json:"data"`
		}
		if json.Unmarshal(trim, &env) == nil {
			if p := bytes.TrimSpace(env.Payload); len(p) > 0 && p[0] == '{' {
				return json.Unmarshal(unwrapSingleton(p), out)
			}
			if d := bytes.TrimSpace(env.Data); len(d) > 0 && d[0] == '{' {
				return json.Unmarshal(unwrapSingleton(d), out)
			}
		}
	}
	return json.Unmarshal(trim, out)
}

// unwrapSingleton peels one more wrapper when an unwrapped envelope holds exactly ONE
// object-valued key ({"payload":{"webhook":{…}}} → the webhook). Real records always
// carry several fields, so a singleton object key can only be another wrapper.
func unwrapSingleton(obj []byte) []byte {
	var m map[string]json.RawMessage
	if json.Unmarshal(obj, &m) != nil || len(m) != 1 {
		return obj
	}
	for _, v := range m {
		if inner := bytes.TrimSpace(v); len(inner) > 0 && inner[0] == '{' {
			return inner
		}
	}
	return obj
}
