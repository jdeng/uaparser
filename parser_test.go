package uaparser

import (
	//	"fmt"
	"testing"
)

type tcase struct {
	in  string
	out string
}

func TestParse(t *testing.T) {
	cases := []tcase{
		tcase{"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.11 (KHTML, like Gecko) Chrome/23.0.1271.97 Safari/537.11",
			"6;;linux;chrome"},
		tcase{"Mozilla/5.0 (Windows NT 6.1; WOW64; Trident/7.0; rv:11.0) like Gecko", "6;;windows_nt;msie"},
		//Roku
		tcase{"Roku/DVP-9.0 (289.00E04144A)", "3;roku;;"},
		tcase{"com.google.android.youtube/14.08.55(Linux; U; Android 6.0; es_US; M4 SS4457 Build/MRA58K) gzip,gzip(gfe)", "1;m4 ss4457;android;"},
		tcase{"com.google.ios.youtube/14.07.7 (iPhone11,8; U; CPU iOS 12_1_4 like Mac OS X; en_US)", "1;iphone;ios;"},

		//"Mozilla/5.0 (RokuOS) Cobalt/9.174384-gold (unlike Gecko) Starboard/4, Roku_OTT_MC2/9.0 (Roku, 3900X, Wireless),gzip(gfe)"

		tcase{"Roku/DVP-9.0 (519.00E04142A),gzip(gfe)", ""},
		//Android 8.0.0 (samsung; SM-A600FN; Sky Go Android PR17.3.3-1100)
	}

	for i, x := range cases {
		out := Parse(x.in).ShortName()
		if out != x.out {
			t.Errorf("%d: %s, expected: %s, got: %s\n", i, x.in, x.out, out)
		}
	}
}
