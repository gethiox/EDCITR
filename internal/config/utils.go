package config

import (
	"reflect"

	"github.com/gethiox/EDCITR/internal/app/journal"
)

func (c *EDCITRConfig) IsDirty(savedConfig *EDCITRConfig) bool {
	if savedConfig == nil {
		return false
	}
	return !reflect.DeepEqual(*c, *savedConfig)
}

func (c *EDCITRConfig) IsReprocessingRequired(processedConfig *EDCITRConfig) bool {
	if processedConfig == nil {
		return false
	}

	if c.Viewports.Count != processedConfig.Viewports.Count {
		return true
	}

	if c.Filters.NotificationsBlacklistMode != processedConfig.Filters.NotificationsBlacklistMode {
		return true
	}
	if c.Filters.NpcMessagesBlacklistMode != processedConfig.Filters.NpcMessagesBlacklistMode {
		return true
	}

	if c.Stats.AlwaysResetSession != processedConfig.Stats.AlwaysResetSession {
		return true
	}

	if !mapsEqual(c.Viewports.ViewportChannelBitmask, processedConfig.Viewports.ViewportChannelBitmask) {
		return true
	}
	if !filtersEqual(c.Filters.Notifications, processedConfig.Filters.Notifications) {
		return true
	}
	if !filtersEqual(c.Filters.NpcMessages, processedConfig.Filters.NpcMessages) {
		return true
	}

	return false
}

func (c *EDCITRConfig) DeepCopy() *EDCITRConfig {
	clone := *c // Shallow copy struct fields

	// Deep copy slices
	clone.Viewports.Appearance.HighlightWords = make([]string, len(c.Viewports.Appearance.HighlightWords))
	copy(clone.Viewports.Appearance.HighlightWords, c.Viewports.Appearance.HighlightWords)

	// Deep copy ViewportChannelBitmask map
	clone.Viewports.ViewportChannelBitmask = make(map[int]journal.MessageChannel, len(c.Viewports.ViewportChannelBitmask))
	for k, v := range c.Viewports.ViewportChannelBitmask {
		clone.Viewports.ViewportChannelBitmask[k] = v
	}

	// Deep copy Notification maps
	clone.Audio.ChannelSoundPaths = make(map[journal.MessageChannel]string, len(c.Audio.ChannelSoundPaths))
	for k, v := range c.Audio.ChannelSoundPaths {
		clone.Audio.ChannelSoundPaths[k] = v
	}

	clone.Audio.ChannelVolume = make(map[journal.MessageChannel]float32, len(c.Audio.ChannelVolume))
	for k, v := range c.Audio.ChannelVolume {
		clone.Audio.ChannelVolume[k] = v
	}

	clone.Filters.Notifications = copyFiltersMap(c.Filters.Notifications)
	clone.Filters.NpcMessages = copyFiltersMap(c.Filters.NpcMessages)

	return &clone
}

func copyFiltersMap(src map[string]Filter) map[string]Filter {
	if src == nil {
		return nil
	}

	dst := make(map[string]Filter, len(src))
	for k, filter := range src {
		// Copy the outer struct and Base
		newFilter := Filter{
			Base: filter.Base,
		}

		// Deep copy the inner map if it exists
		if filter.SubTypes != nil {
			newFilter.SubTypes = make(map[string]*FilterInfo, len(filter.SubTypes))
			for subK, subV := range filter.SubTypes {
				newFilter.SubTypes[subK] = subV
			}
		}
		dst[k] = newFilter
	}
	return dst
}

func mapsEqual[K comparable, V comparable](a, b map[K]V) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if b[k] != v {
			return false
		}
	}
	return true
}

func filtersEqual(a, b map[string]Filter) bool {
	if len(a) != len(b) {
		return false
	}
	for k, filterA := range a {
		filterB, ok := b[k]
		if !ok {
			return false
		}

		// Compare the Base struct
		if filterA.Base != filterB.Base {
			return false
		}

		// Compare SubTypes map lengths
		if len(filterA.SubTypes) != len(filterB.SubTypes) {
			return false
		}

		// Compare individual SubTypes
		for subK, subA := range filterA.SubTypes {
			if filterB.SubTypes[subK] != subA {
				return false
			}
		}
	}
	return true
}
