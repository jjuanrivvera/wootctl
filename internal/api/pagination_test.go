package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testRec struct {
	ID   ID     `json:"id"`
	Name string `json:"name"`
}

func TestDecodeList_Envelopes(t *testing.T) {
	cases := []struct {
		name      string
		body      string
		wantLen   int
		wantMeta  bool
		wantTotal int
	}{
		{"bare array", `[{"id":1},{"id":2}]`, 2, false, -1},
		{"payload wrapper", `{"payload":[{"id":1}]}`, 1, false, -1},
		{"payload with meta", `{"meta":{"count":40,"current_page":"1"},"payload":[{"id":1}]}`, 1, true, 40},
		{"data array", `{"data":[{"id":1}]}`, 1, false, -1},
		{"conversations shape", `{"data":{"meta":{"mine_count":2,"all_count":7},"payload":[{"id":9}]}}`, 1, true, -1},
		{"single array key (audit logs)", `{"audit_logs":[{"id":1}],"current_page":1,"per_page":15,"total_entries":31}`, 1, true, 31},
		{"double-nested payload (webhooks)", `{"payload":{"webhooks":[{"id":1},{"id":2}]}}`, 2, false, -1},
		{"empty array", `[]`, 0, false, -1},
		{"empty body", ``, 0, false, -1},
		{"null", `null`, 0, false, -1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			items, meta, err := decodeList[testRec](([]byte)(tc.body))
			require.NoError(t, err)
			assert.Len(t, items, tc.wantLen)
			if tc.wantMeta {
				require.NotNil(t, meta)
				assert.Equal(t, tc.wantTotal, meta.total())
			} else if tc.wantTotal >= 0 {
				t.Fatalf("test case inconsistent")
			}
		})
	}
}

func TestDecodeList_Rejects(t *testing.T) {
	for name, body := range map[string]string{
		"scalar":          `42`,
		"two array keys":  `{"a":[],"b":[]}`,
		"object no array": `{"id":1}`,
		"bad json":        `{`,
	} {
		t.Run(name, func(t *testing.T) {
			_, _, err := decodeList[testRec]([]byte(body))
			assert.Error(t, err)
		})
	}
}

func TestDecodeOne_Envelopes(t *testing.T) {
	for name, body := range map[string]string{
		"direct":            `{"id":5,"name":"x"}`,
		"payload":           `{"payload":{"id":5,"name":"x"}}`,
		"data":              `{"data":{"id":5,"name":"x"}}`,
		"singleton wrapper": `{"payload":{"webhook":{"id":5,"name":"x"}}}`,
	} {
		t.Run(name, func(t *testing.T) {
			var rec testRec
			require.NoError(t, decodeOne([]byte(body), &rec))
			assert.Equal(t, ID("5"), rec.ID)
			assert.Equal(t, "x", rec.Name)
		})
	}
	var rec testRec
	require.NoError(t, decodeOne([]byte(``), &rec))     // empty body is a no-op
	require.NoError(t, decodeOne([]byte(`null`), &rec)) // null too
	assert.Error(t, decodeOne([]byte(`{`), &rec))
}

func TestListMeta_Total(t *testing.T) {
	assert.Equal(t, -1, (*ListMeta)(nil).total())
	assert.Equal(t, 12, (&ListMeta{Count: 12}).total())
	assert.Equal(t, 31, (&ListMeta{TotalEntries: 31}).total())
	// all_count deliberately ignored: it reports the UNFILTERED total on filtered lists.
	assert.Equal(t, -1, (&ListMeta{AllCount: 99}).total())
}

func TestListParams_Values(t *testing.T) {
	v := ListParams{Page: 2, Sort: "-name", Extra: map[string]string{"status": "open", "empty": ""}}.values()
	assert.Equal(t, "2", v.Get("page"))
	assert.Equal(t, "-name", v.Get("sort"))
	assert.Equal(t, "open", v.Get("status"))
	_, has := v["empty"]
	assert.False(t, has, "empty extras must be omitted")

	zero := ListParams{}.values()
	assert.Empty(t, zero, "zero params render an empty query")
}
