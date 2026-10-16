package my_noire

import (
	"encoding/json/v2"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMarshal(t *testing.T) {
	c := NewRGBA(64.0, 128.0, 255.0, 1.0/255)

	data, err := json.Marshal(&c)
	assert.NoError(t, err)

	assert.Equal(t, []byte(`"#4080ff01"`), data)
}

func TestUnmarshal(t *testing.T) {
	var c Color
	err := json.Unmarshal([]byte(`"#4080ff01"`), &c)
	assert.NoError(t, err)

	assert.Equal(t, NewRGBA(64.0, 128.0, 255.0, 1.0/255), c)
}
