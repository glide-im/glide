package group_subscription

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPermission_Allows(t *testing.T) {
	p := PermRead | PermWrite
	assert.True(t, p.Allows(PermRead, PermWrite))
}

func TestPermission_AllowsFalse(t *testing.T) {
	p := PermRead | PermWrite
	assert.False(t, p.Allows(PermRead, PermAdmin))
}

func TestPermission_Denies(t *testing.T) {
	p := PermRead | PermWrite
	assert.True(t, p.Denies(PermAdmin))
}

func TestPermission_DeniesFalse(t *testing.T) {
	p := PermRead | PermWrite
	assert.False(t, p.Denies(PermRead))
}
