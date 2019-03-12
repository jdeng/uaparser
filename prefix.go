package uaparser

import (
	"strings"
)

func handle_rv(ua *UserAgent, reco *recognizer, sec *section) bool {
	ua.rv = strings.TrimSpace(strings.TrimPrefix(sec.name, reco.prefix))
	return true
}

// Mozilla/5.0 (compatible; MSIE 10.0; Windows NT 6.2; Trident/6.0)
func handle_browser_version(ua *UserAgent, reco *recognizer, sec *section) bool {
	sec.version = strings.TrimSpace(strings.TrimPrefix(sec.name, reco.prefix))
	sec.name = reco.rewrite
	ua.Browser.use(sec, reco)
	return true
}

// Mozilla/5.0 (CrKey armv7l 1.4.15250)
func handle_device_version(ua *UserAgent, reco *recognizer, sec *section) bool {
	sec.version = strings.TrimSpace(strings.TrimPrefix(sec.name, reco.prefix))
	sec.name = reco.rewrite
	if ua.Device.use(sec, reco) {
		if reco.deviceType > 0 {
			ua.DeviceType = reco.deviceType
		}
	}

	return true
}

// Mozilla/5.0 (Windows NT 6.3; Trident/7.0; .NET4.0E; .NET4.0C; rv:11.0)
func handle_os_version(ua *UserAgent, reco *recognizer, sec *section) bool {
	sec.version = strings.TrimSpace(strings.TrimPrefix(sec.name, reco.prefix))
	sec.name = reco.rewrite
	ua.OS.use(sec, reco)
	return true
}

// Mozilla/5.0 (iPhone; CPU iPhone OS 8_3 like Mac OS X)
func handle_ios(ua *UserAgent, reco *recognizer, sec *section) bool {
	name := strings.TrimPrefix(sec.name, reco.prefix)
	name = strings.TrimSuffix(name, "like mac os x")
	sec.version = strings.TrimSpace(name)
	sec.name = reco.rewrite
	if sec.name == "" {
		sec.name = "ios"
	}
	ua.OS.use(sec, reco)
	return true
}
