package version

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestString(t *testing.T) {
	Version, Commit, Date = "1.2.3", "abc1234", "2026-06-29"
	s := String()
	assert.True(t, strings.HasPrefix(s, "wootctl 1.2.3"))
	assert.Contains(t, s, "abc1234")
}

func TestGet(t *testing.T) {
	Version = "9.9.9"
	info := Get()
	assert.Equal(t, "9.9.9", info.Version)
}
