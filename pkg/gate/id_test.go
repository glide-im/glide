package gate

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGenTempID(t *testing.T) {

	id, err := GenTempID("test")

	assert.Nil(t, err)
	assert.NotEmpty(t, id)
	assert.True(t, id.IsTemp())
}
