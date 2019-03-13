package uaparser

import (
	"strings"
)

func (ua *UserAgent) Check(s string) bool {
	c := func(k ...string) bool {
		for _, x := range k {
			if strings.Contains(s, x) {
				return true
			}
		}
		return false
	}

	if c("ipad") && ua.Device.Name == "" {
		return false
	}

	if c("iphone") && ua.Device.Name == "" {
		return false
	}

	if c("apple_tv", "appletv", "apple tv") && ua.Device.Name != "appletv" {
		return false
	}

	if c("tvos") && ua.OS.Name != "tvos" {
		return false
	}

	if c("tv") && ua.DeviceType != SmartTV && ua.Device.Name == "" {
		return false
	}

	if c("tablet") && ua.DeviceType != Tablet {
		return false
	}

	return true
}
