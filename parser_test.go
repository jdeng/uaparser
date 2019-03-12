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
		tcase{"Roku/DVP-9.0+(519.00E04142A)", "3;roku;;"},
		tcase{"Roku/DVP-9.0 (289.00E04144A)", "3;roku;;"},
	}

	for i, x := range cases {
		out := Parse(x.in).ShortName()
		if out != x.out {
			t.Errorf("%d: %s, expected: %s, got: %s\n", i, x.in, x.out, out)
		}
	}
}
