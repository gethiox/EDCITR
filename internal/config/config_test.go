package config

import (
	"encoding/json/v2"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultSaveLoad(t *testing.T) {
	cfg := *GetDefaultConfig()

	data, err := json.Marshal(cfg)
	if err != nil {
		t.Fatal(err)
	}

	var cfg2 EDCITRConfig
	err = json.Unmarshal(data, &cfg2)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, cfg, cfg2)
}
