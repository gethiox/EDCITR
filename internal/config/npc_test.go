package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNPCMessagesAvailable(t *testing.T) {
	for _, msg := range NPCMessagesAvailable {
		t.Run(msg.Who, func(t *testing.T) {
			assert.True(t, msg.BaseLevel != LevelUndefined)
			assert.True(t, msg.Who != "")
			for _, subMsg := range msg.MsgTypes {
				t.Run(subMsg.Value, func(t *testing.T) {
					assert.True(t, subMsg.Level != LevelUndefined)
					assert.True(t, subMsg.Value != "")
				})
			}
		})
	}
}
