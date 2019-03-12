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
				inComment += 1
				lastPos = i + 1
			} else if c == '[' {
				inSkip += 1
				lastPos = i + 1
			} else if lastPos < 0 {
				lastPos = i
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
	Console
	Wearable
	Desktop
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
			prefix += items[j].(*section).name
			items[j] = nil
		}
		if prefix != "" {
			section := items[end].(*section)
			section.name = prefix + " " + section.name
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
		} else {
			fmt.Printf("ignoring %#v\n", item)
		}
		i += 1
	}

	if len(products) == 0 {
		fmt.Printf("no products: %#v\n", items)
		return ua
	}

	var xProducts, xComments []*section
	if products[0].name == "mozilla" {
		ua.mozilla = products[0].version
	} else {
		sec := products[0].section
		if !ua.try(sec, 0, true) {
			tryPrefix(ua, productPrefixRecognizers, sec)
		}
	}

	for _, sec := range products[0].comment {
		if ua.try(sec, 0, false) {
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
	if ua.Device.Name == "apple_tv" && ua.OS.Name == "ios" {
		ua.OS.Name = "tvos"
	}

	if ua.OS.Name == "linux" {
		if strings.HasPrefix(ua.OS.Version, "smarttv") {
			ua.tags["smarttv"] = ""
		}
	} else if ua.OS.Name == "android" { // android: last comment is device id
		if ua.Device.Name == "" && len(xComments) > 0 {
			sec := xComments[len(xComments)-1]
			xComments = xComments[:len(xComments)-1]

			sec.name = strings.TrimSuffix(sec.name, " build")
			ua.Device.use(sec, nil)
		}
	} else if ua.OS.Name == "windows_nt" {
		if ua.Browser.Name == "" && ua.rv != "" {
			ua.Browser.Name = "msie"
			ua.Browser.Version = ua.rv
		}
	}

	//tagging
	if _, ok := ua.tags["smarttv"]; ok {
		ua.DeviceType = SmartTV
	} else if _, ok := ua.tags["ctv"]; ok && ua.OS.Name == "android" {
		ua.DeviceType = SmartTV
	} else if _, ok := ua.tags["tablet"]; ok {
		ua.DeviceType = Tablet
	} else if _, ok := ua.tags["mobile"]; ok {
		ua.mobile = true
	}

	// second phase after tagging
	if ua.OS.Name == "android" {
		if ua.DeviceType == UnknownDevice {
			ua.DeviceType = Phone
		}
	} else if ua.Browser.Name == "safari" && ua.mobile {
		ua.Browser.Name = "mobile safari"
	}

	if ua.DeviceType == UnknownDevice && !ua.mobile {
		if ua.OS.Name == "linux" || ua.OS.Name == "windows_nt" || ua.OS.Name == "macosx" {
			ua.DeviceType = Desktop
		}
	}

	//	fmt.Printf("unrecognized: %#v, %#v\n", xProducts, xComments)
	return ua
}
