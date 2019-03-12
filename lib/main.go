package main

/*
#include <stdlib.h>
*/
import "C"

import "unsafe"
import "github.com/jdeng/uaparser"

//export ParseUserAgent
func ParseUserAgent(s *C.char) *C.char {
	return C.CString(uaparser.Parse(C.GoString(s)).ShortName())
}

//export FreeUserAgent
func FreeUserAgent(s *C.char) {
	C.free(unsafe.Pointer(s))
}

func main() {}
