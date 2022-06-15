package subscription_impl

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPermission_Allows(t *testing.T) {
	p := PermRead | PermWrite
	assert.True(t, p.allows(MaskPermRead, MaskPermWrite))
}

func TestPermission_AllowsFalse(t *testing.T) {
	p := PermRead | PermWrite
	assert.False(t, p.allows(MaskPermAdmin))
	assert.False(t, p.allows(MaskPermWrite, MaskPermAdmin))
}

func TestPermission_Denies(t *testing.T) {
	p := PermRead | PermWrite
	assert.True(t, p.denies(MaskPermAdmin))
}

func TestPermission_DeniesFalse(t *testing.T) {
	p := PermRead | PermWrite
	assert.False(t, p.denies(MaskPermRead))
}
