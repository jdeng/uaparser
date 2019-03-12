package uaparser

type ps struct {
	priority int
	source   int
}

type psd struct {
	priority   int
	source     int
	deviceType int
}

var knownTags = map[string]string{
	"ctv":      "smarttv",
	"mobile":   "",
	"tablet":   "",
	"stb":      "",
	"smarttv":  "",
	"smart_tv": "smarttv",
}

var browsers = map[string]ps{
	"safari":         ps{1, IN_PRODUCT},
	"mobile safari":  ps{1, IN_PRODUCT},
	"msie":           ps{2, IN_COMMENT},
	"firefox":        ps{2, IN_PRODUCT},
	"opera":          ps{2, IN_BOTH},
	"chrome":         ps{2, IN_PRODUCT},
	"dalvik":         ps{2, IN_PRODUCT},
	"edge":           ps{3, IN_PRODUCT},
	"silk":           ps{3, IN_BOTH},
	"fxios":          ps{3, IN_BOTH},
	"lg browser":     ps{3, IN_PRODUCT},
	"applecoremedia": ps{2, IN_PRODUCT},
	"netflix":        ps{3, IN_PRODUCT},
}

var engines = map[string]ps{
	"applewebkit": ps{1, IN_PRODUCT},
	"trident":     ps{1, IN_COMMENT},
	"gecko":       ps{1, IN_COMMENT},
}

var oses = map[string]ps{
	"linux":      ps{1, IN_BOTH},
	"freebsd":    ps{1, IN_BOTH},
	"android":    ps{2, IN_BOTH},
	"windows_nt": ps{1, IN_COMMENT},
	"ios":        ps{2, IN_BOTH},
	//	"web0s":      ps{2, IN_BOTH},
}

var devices = map[string]psd{
	"iphone": psd{2, IN_BOTH, Phone},
	"ipad":   psd{2, IN_BOTH, Tablet},
	"ipod":   psd{1, IN_BOTH, Phone},

	"roku":     psd{1, IN_BOTH, SmartTV},
	"googletv": psd{1, IN_BOTH, SmartTV},

	"xbox":     psd{1, IN_BOTH, Console},
	"xbox one": psd{2, IN_BOTH, Console},
	"crkey":    psd{2, IN_PRODUCT, SmartTV},

	"kindle": psd{2, IN_PRODUCT, Tablet},
}

var skips = map[string]ps{
	"u":           ps{0, IN_COMMENT},
	"x11":         ps{0, IN_COMMENT},
	"ubuntu":      ps{0, IN_COMMENT},
	"compatible":  ps{0, IN_COMMENT},
	"ppc":         ps{0, IN_COMMENT},
	"arm":         ps{0, IN_COMMENT},
	"touch":       ps{0, IN_COMMENT},
	"macintosh":   ps{0, IN_COMMENT},
	"x64":         ps{0, IN_COMMENT},
	"win64":       ps{0, IN_COMMENT},
	"wow64":       ps{0, IN_COMMENT},
	"like gecko":  ps{0, IN_PRODUCT},
	"like chrome": ps{0, IN_PRODUCT},

	"microsoft": ps{0, IN_COMMENT},
	"iemobile":  ps{0, IN_COMMENT},
	"nokia":     ps{0, IN_COMMENT},
}

var (
	commentRecognizers       = make(map[string]*recognizer)
	productRecognizers       = make(map[string]*recognizer)
	commentPrefixRecognizers []*recognizer
	productPrefixRecognizers []*recognizer
)

func addPrefix(source int, prefix string, name string, priority int, deviceType int, handler func(*UserAgent, *recognizer, *section) bool) {
	if (source & IN_COMMENT) != 0 {
		commentPrefixRecognizers = append(commentPrefixRecognizers, &recognizer{
			prefix:     prefix,
			rewrite:    name,
			priority:   priority,
			deviceType: deviceType,
			handler:    handler,
		})
	}
	if (source & IN_PRODUCT) != 0 {
		productPrefixRecognizers = append(productPrefixRecognizers, &recognizer{
			prefix:     prefix,
			rewrite:    name,
			priority:   priority,
			deviceType: deviceType,
			handler:    handler,
		})
	}
}

func init() {
	for k, v := range oses {
		if (v.source & IN_PRODUCT) != 0 {
			productRecognizers[k] = &recognizer{typ: OS, priority: v.priority}
		}
		if (v.source & IN_COMMENT) != 0 {
			commentRecognizers[k] = &recognizer{typ: OS, priority: v.priority}
		}
	}

	for k, v := range browsers {
		if (v.source & IN_PRODUCT) != 0 {
			productRecognizers[k] = &recognizer{typ: BROWSER, priority: v.priority}
		}
		if (v.source & IN_COMMENT) != 0 {
			commentRecognizers[k] = &recognizer{typ: BROWSER, priority: v.priority}
		}
	}

	for k, v := range engines {
		if (v.source & IN_PRODUCT) != 0 {
			productRecognizers[k] = &recognizer{typ: ENGINE, priority: v.priority}
		}
		if (v.source & IN_COMMENT) != 0 {
			commentRecognizers[k] = &recognizer{typ: ENGINE, priority: v.priority}
		}

	}
	for k, v := range devices {
		if (v.source & IN_PRODUCT) != 0 {
			productRecognizers[k] = &recognizer{typ: DEVICE, priority: v.priority, deviceType: v.deviceType}
		}
		if (v.source & IN_COMMENT) != 0 {
			commentRecognizers[k] = &recognizer{typ: DEVICE, priority: v.priority, deviceType: v.deviceType}
		}

	}

	for _, v := range languages {
		commentRecognizers[v] = &recognizer{typ: LANGUAGE}
	}

	for k, v := range skips {
		if (v.source & IN_PRODUCT) != 0 {
			productRecognizers[k] = &recognizer{typ: SKIP, priority: v.priority}
		}
		if (v.source & IN_COMMENT) != 0 {
			commentRecognizers[k] = &recognizer{typ: SKIP, priority: v.priority}
		}

	}

	addPrefix(IN_COMMENT, "windows nt ", "windows_nt", 1, 0, handle_os_version)
	addPrefix(IN_COMMENT, "linux ", "linux", 1, 0, handle_os_version)
	addPrefix(IN_COMMENT, "android ", "android", 2, 0, handle_os_version)
	addPrefix(IN_COMMENT, "windows phone", "windows_phone", 1, 0, handle_os_version)
	addPrefix(IN_COMMENT, "intel mac os x ", "macosx", 1, 0, handle_os_version)
	addPrefix(IN_COMMENT, "cros ", "chromeos", 1, 0, handle_os_version)

	addPrefix(IN_COMMENT, "crkey ", "chromecast", 1, SmartTV, handle_device_version)
	addPrefix(IN_COMMENT, "apple tv", "apple_tv", 1, SmartTV, handle_device_version)
	addPrefix(IN_COMMENT, "playstation 4", "ps4", 1, Console, handle_device_version)

	addPrefix(IN_PRODUCT, "roku ", "roku", 2, SmartTV, handle_device_version)
	addPrefix(IN_PRODUCT, "rokudvp-", "roku", 2, SmartTV, handle_device_version)
	addPrefix(IN_COMMENT, "googletv ", "googletv", 2, SmartTV, handle_device_version)
	addPrefix(IN_PRODUCT, "lg netcast.tv", "lgtv", 2, SmartTV, handle_device_version)

	addPrefix(IN_COMMENT, "msie ", "msie", 2, 0, handle_browser_version)

	addPrefix(IN_COMMENT, "rv:", "", 1, 0, handle_rv)

	addPrefix(IN_COMMENT, "cpu iphone os ", "", 1, 0, handle_ios)
	addPrefix(IN_COMMENT, "cpu os ", "", 1, 0, handle_ios)
	addPrefix(IN_COMMENT, "ios ", "", 1, 0, handle_ios)
}
