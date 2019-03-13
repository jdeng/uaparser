package uaparser

import (
	"fmt"
	"strings"
)

type section struct {
	name, version string
}

type comment []*section
type skip section
type items []interface{}
type product struct {
	*section
	comment
}

func newSection(s string) *section {
	xs := strings.SplitN(s, "/", 2)
	var name, version string
	name = strings.TrimSpace(xs[0])
	if len(xs) == 2 {
		version = strings.TrimSpace(xs[1])
	}
	return &section{name: name, version: version}
}

func parseComment(s string) comment {
	var out comment
	xs := strings.Split(s, ";")
	for _, x := range xs {
		x = strings.TrimSpace(x)
		if x == "" {
			continue
		}
		out = append(out, newSection(x))
	}
	return out
}

func parse(ua string) items {
	var result items

	inComment, inSkip := 0, 0
	lastPos := -1
	for i, c := range ua {
		if inComment == 0 && inSkip == 0 {
			if c == ' ' {
				if lastPos >= 0 {
					result = append(result, newSection(ua[lastPos:i]))
				}
				lastPos = -1
			} else if c == '(' {
				if lastPos >= 0 {
					result = append(result, newSection(ua[lastPos:i]))
				}

				inComment += 1
				lastPos = i + 1
			} else if c == '[' {
				inSkip += 1
				lastPos = i + 1
			} else if lastPos < 0 {
				if c != ',' && c != ';' {
					lastPos = i
				}
			}
		} else if inComment > 0 {
			if c == ')' {
				inComment -= 1
				if inComment == 0 {
					if lastPos >= 0 {
						result = append(result, parseComment(ua[lastPos:i]))
					}
					lastPos = -1
				}
			} else if c == '(' {
				inComment += 1
			}
		} else if inSkip > 0 {
			if c == ']' {
				inSkip -= 1
				if inSkip == 0 {
					result = append(result, &skip{name: ua[lastPos:i]})
				}
				lastPos = -1
			} else if c == '[' {
				inSkip += 1
			}
		}
	}

	if lastPos >= 0 {
		result = append(result, newSection(ua[lastPos:]))
	}

	return result
}

const (
	UnknownDevice = iota
	Phone
	Tablet
	SmartTV
	SetTop
	Console
	Desktop
	Wearable
)

type Component struct {
	Name, Version string
	priority      int
}

func (c *Component) use(sec *section, reco *recognizer) bool {
	if reco == nil {
		c.Name = sec.name
		c.Version = sec.version
		return true
	}

	if reco.priority > c.priority {
		c.Name = sec.name
		c.Version = sec.version
		c.priority = reco.priority
		return true
	}

	return false
}

const (
	UNKNOWN = iota
	OS
	BROWSER
	DEVICE
	ENGINE
	LANGUAGE
	SKIP
)

const (
	IN_PRODUCT = 0x01
	IN_COMMENT = 0x02
	IN_BOTH    = IN_PRODUCT | IN_COMMENT
)

type recognizer struct {
	typ        int
	priority   int
	deviceType int
	rewrite    string
	prefix     string
	handler    func(ua *UserAgent, reco *recognizer, sec *section) bool
}

type UserAgent struct {
	DeviceType                  int
	OS, Browser, Device, Engine Component
	Language                    string

	rv      string
	tags    map[string]string
	mobile  bool
	webview bool
	mozilla string
}

func (ua *UserAgent) ShortName() string {
	return fmt.Sprintf("%d;%s;%s;%s", ua.DeviceType, ua.Device.Name, ua.OS.Name, ua.Browser.Name)
}

func (ua *UserAgent) try(sec *section, pos int, isProduct bool) bool {
	var reco *recognizer
	var ok bool
	if isProduct {
		reco, ok = productRecognizers[sec.name]
	} else {
		reco, ok = commentRecognizers[sec.name]
	}
	if ok {
		if reco.handler == nil {
			switch reco.typ {
			case BROWSER:
				ua.Browser.use(sec, reco)
			case ENGINE:
				ua.Engine.use(sec, reco)
			case OS:
				ua.OS.use(sec, reco)
			case DEVICE:
				if ua.Device.use(sec, reco) {
					if reco.deviceType > 0 {
						ua.DeviceType = reco.deviceType
					}
				}
			case LANGUAGE:
				ua.Language = sec.name
			case SKIP:
			}
			return true
		} else {
			if reco.handler(ua, reco, sec) {
				return true
			}
		}
	}

	return false
}

func Parse(s string) *UserAgent {
	s = strings.ToLower(s)
	s = strings.Replace(s, "+", " ", -1)
	items := parse(s)

	lastPos := 0
	mergeItems := func(end int) {
		if lastPos < 0 {
			return
		}
		prefix := ""
		for j := lastPos; j < end; j += 1 {
			if prefix != "" {
				prefix += " "
			}
			if sec, ok := items[j].(*section); ok {
				prefix += sec.name
			}
			items[j] = nil
		}
		if prefix != "" {
			if sec, ok := items[end].(*section); ok {
				sec.name = prefix + " " + sec.name
			}
		}
		lastPos = -1
	}

	tryPrefix := func(ua *UserAgent, recognizers []*recognizer, sec *section) bool {
		used := false
		for _, reco := range recognizers {
			if reco.prefix == "" || reco.handler == nil {
				continue
			}

			if strings.HasPrefix(sec.name, reco.prefix) {
				if reco.handler(ua, reco, sec) {
					used = true
					break
				}
			}
		}
		return used
	}

	var firstTag string
	if len(items) > 0 {
		if sec, ok := items[0].(*section); ok {
			firstTag = strings.Trim(sec.name, "\"")
		}
	}

	//merge items
	for i, item := range items {
		switch item.(type) {
		case *section:
			if lastPos < 0 {
				lastPos = i
			}
			section := item.(*section)
			if section.version != "" {
				mergeItems(i)
			}
		case comment:
			if i > 0 {
				mergeItems(i - 1)
			}
			lastPos = -1
		case skip:
			if i > 0 {
				mergeItems(i - 1)
			}
			lastPos = -1
		}
	}

	//TODO: handle non section last item
	if lastPos > 0 {
		mergeItems(len(items) - 1)
	}

	ua := &UserAgent{tags: make(map[string]string)}
	if len(items) == 0 {
		fmt.Printf("no items: %#v\n", items)
		return ua
	}

	// convert to products
	var products []product
	for i := 0; i < len(items); {
		item := items[i]
		if item == nil {
			i += 1
			continue
		}

		if sec, ok := item.(*section); ok {
			if i+1 < len(items) {
				if com, ok := items[i+1].(comment); ok {
					products = append(products, product{sec, com})
					i += 2
					continue
				}
			}
			products = append(products, product{sec, nil})
		} else if com, ok := item.(comment); ok {
			if i == 0 {
				products = append(products, product{nil, com})

			} else {
				//				fmt.Printf("ignoring %#v\n", item)
			}

		}
		i += 1
	}

	if len(products) == 0 {
		fmt.Printf("no products: %#v: %s\n", items, s)
		return ua
	}

	var xProducts, xComments []*section
	if sec := products[0].section; sec != nil {
		if sec.name == "mozilla" {
			ua.mozilla = sec.version
		} else {
			if !ua.try(sec, 0, true) {
				tryPrefix(ua, productPrefixRecognizers, sec)
			}
		}
	}

	//only use the first non-empty comment
	var comments comment
	for i := 0; i < len(products); i += 1 {
		comments = products[i].comment
		if comments != nil {
			break
		}
	}

	for _, sec := range comments {
		if ua.try(sec, 0, false) {
			continue
		}

		if sec.name == "mobile" {
			ua.mobile = true
			continue
		}

		if sec.name == "wv" && ua.OS.Name == "android" {
			ua.webview = true
			continue
		}

		if t, ok := knownTags[sec.name]; ok {
			if t == "" {
				t = sec.name
			}
			ua.tags[t] = sec.version
			continue
		}

		if tryPrefix(ua, commentPrefixRecognizers, sec) {
			continue
		}

		xComments = append(xComments, sec)
	}

	for i := 1; i < len(products); i += 1 {
		sec := products[i].section
		if ua.try(sec, i, true) {
			continue
		}

		if sec.name == "mobile" {
			ua.mobile = true
		}

		if t, ok := knownTags[sec.name]; ok {
			if t == "" {
				t = sec.name
			}
			ua.tags[t] = sec.version
			continue
		}

		if tryPrefix(ua, commentPrefixRecognizers, sec) {
			continue
		}

		xProducts = append(xProducts, sec)
	}

	// extra rules
	switch ua.OS.Name {
	case "linux":
		if strings.HasPrefix(ua.OS.Version, "smarttv") {
			ua.DeviceType = SmartTV
		}
	case "android":
		// android: last comment is device id
		if ua.Device.Name == "" && len(xComments) > 0 {
			sec := xComments[len(xComments)-1]
			xComments = xComments[:len(xComments)-1]

			sec.name = strings.TrimSuffix(sec.name, " build")
			ua.Device.use(sec, nil)
		}

		if _, ok := ua.tags["ctv"]; ok {
			ua.DeviceType = SmartTV
		}

		if ua.Engine.Name == "exoplayerlib" && ua.Browser.Name == "" {
			ua.Browser.Name = "exoplayerapp"
		}

		//popular devices
		if strings.HasPrefix(ua.Device.Name, "aft") { // Amazon Fire TV
			ua.DeviceType = SmartTV
		} else if strings.HasPrefix(ua.Device.Name, "kf") { // Amazon Kindle Fire
			ua.DeviceType = Tablet
		} else if strings.HasSuffix(ua.Device.Name, "tv") {
			ua.DeviceType = SmartTV
		}

	case "windows_nt":
		if ua.Browser.Name == "" && ua.rv != "" {
			ua.Browser.Name = "msie"
			ua.Browser.Version = ua.rv
		}
		//TODO: normalize windows

	case "tvos":
		if ua.Device.Name == "" {
			ua.Device.Name = "appletv"
		}
		ua.DeviceType = SmartTV

	case "rokuos": //Cobalt
		ua.DeviceType = SmartTV
		ua.Device.Name = "roku"
		ua.OS.Name = ""
	}

	//tagging
	if _, ok := ua.tags["tablet"]; ok {
		ua.DeviceType = Tablet
	} else if _, ok := ua.tags["mobile"]; ok {
		ua.mobile = true
	}

	// iOS apps
	if _, ok := ua.tags["cfnetwork"]; ok && ua.OS.Name == "darwin" && ua.Device.Name == "" {
		if strings.HasPrefix(firstTag, "appletv") || strings.HasSuffix(firstTag, "tvos") {
			ua.OS.Name = "tvos"
			ua.Device.Name = "appletv"
			ua.DeviceType = SmartTV
		} else {
			ua.mobile = true
			ua.OS.Name = "ios"
			if strings.HasPrefix(firstTag, "ipad") {
				ua.DeviceType = Tablet
				ua.Device.Name = "ipad"
			} else {
				ua.DeviceType = Phone
				if strings.HasPrefix(firstTag, "iphone") {
					ua.Device.Name = "iphone"
				}
			}
		}
	}

	// second phase after tagging
	if ua.Engine.Name == "cobalt" {
		for _, p := range xProducts {
			if strings.HasPrefix(p.name, "_") {
				names := strings.Split(p.name[1:], "_")
				switch names[0] {
				case "ott":
					fallthrough
				case "atv":
					fallthrough
				case "tv":
					ua.DeviceType = SmartTV
				case "stb":
					ua.DeviceType = SetTop
				case "game":
					ua.DeviceType = Console
				}

				if ua.Device.Name == "" {
					if len(names) >= 2 && names[1] != "" {
						ua.Device.Name = names[1]
					} else {
						ua.Device.Name = names[0]
					}
				}

				if ua.OS.Name == "darwin" && ua.Device.Name == "ott" {
					ua.Device.Name = "appletv"
				}
			}
		}
	}

	if ua.OS.Name == "android" {
		if ua.DeviceType == UnknownDevice {
			ua.DeviceType = Phone
		}
	} else if ua.Browser.Name == "safari" && ua.mobile {
		ua.Browser.Name = "mobile safari"
	}

	if ua.DeviceType == UnknownDevice {
		switch ua.Browser.Name {
		case "hbbtv":
			fallthrough
		case "youviewhtml":
			ua.DeviceType = SmartTV

		//LG Browser/8.00.00 (webOS.TV-2017), _TV_M2R/05.80.02 (LG, 43LJ5500-SA, wireless),gzip(gfe)
		case "lg browser":
			for _, p := range xProducts {
				if strings.HasPrefix(p.name, "_tv_") || strings.HasPrefix(p.name, "lg netcast.tv") || strings.HasPrefix(p.name, "webos.tv") || strings.HasPrefix(p.name, "lg simplesmart.tv") || strings.HasPrefix(p.name, "lg netcast.media") {
					ua.DeviceType = SmartTV
				}
			}

		case "opr": //opera
			if _, ok := ua.tags["omi"]; ok && ua.OS.Name == "linux" && (strings.HasPrefix(ua.OS.Version, "armv") || ua.OS.Version == "mips") {
				ua.DeviceType = SmartTV
			}
		}

		if ua.OS.Name == "ios" {
			ua.DeviceType = Phone
		} else if ua.OS.Name == "tizen" {
			ua.DeviceType = SmartTV
		}
	}

	if ua.DeviceType == UnknownDevice && !ua.mobile && ua.Browser.Name != "" {
		if ua.OS.Name == "windows_nt" || ua.OS.Name == "macosx" || ua.OS.Name == "chromeos" {
			ua.DeviceType = Desktop
		} else if ua.OS.Name == "linux" {
			if strings.HasPrefix(ua.OS.Version, "x86_64") || ua.OS.Version == "i686" {
				ua.DeviceType = Desktop
			}
		}

	}

	switch ua.Device.Name {
	case "roku 3":
		ua.Device.Name = "roku"
	case "playstation 4":
		ua.Device.Name = "ps4"
	case "playstation 3":
		ua.Device.Name = "ps3"
	case "crkey":
		ua.Device.Name = "chromecast"
	case "nintendo wiiu":
		ua.Device.Name = "wiiu"
	case "nintendo switch":
		ua.Device.Name = "switch"
	case "xboxone":
		ua.Device.Name = "xbox"
	case "xbox one":
		ua.Device.Name = "xbox"
	case "xbox360":
		ua.Device.Name = "xbox"
	}

	if ua.DeviceType == UnknownDevice && firstTag != "" {
		name := firstTag
		if x, ok := knownProducts[name]; ok {
			ua.DeviceType = x.DeviceType
			if ua.Device.Name == "" {
				ua.Device.Name = x.Device.Name
			}
		} else {
			if strings.HasPrefix(name, "appletv") {
				ua.Device.Name = "appletv"
				ua.DeviceType = SmartTV
			} else if strings.HasPrefix(name, "iphone") {
				ua.OS.Name = "ios"
				ua.Device.Name = "iphone"
				ua.DeviceType = Phone
			} else if strings.HasPrefix(name, "ipod") {
				ua.OS.Name = "ios"
				ua.Device.Name = "iphone"
				ua.DeviceType = Phone
			} else if strings.HasPrefix(name, "ipad") {
				ua.OS.Name = "ios"
				ua.Device.Name = "ipad"
				ua.DeviceType = Tablet
			} else if strings.HasPrefix(name, "androidtv") {
				ua.Device.Name = "androidtv"
				ua.DeviceType = SmartTV
			} else if strings.HasPrefix(name, "android") {
				ua.OS.Name = "android"
				ua.DeviceType = Phone
			}
		}
	}

	return ua
}
