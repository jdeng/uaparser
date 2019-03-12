#include <stdio.h>
#include <stdlib.h>

extern char* ParseUserAgent(const char *);
extern void FreeUserAgent(char *);

int main(int argc, const char *argv[]) {
	const char *s = "Mozilla/5.0 (Linux; U; en-US) AppleWebKit/528.5+ (KHTML, like Gecko, Safari/528.5+) Version/4.0 Kindle/3.0 (screen 600×800; rotate)";
	if (argc > 1) s = argv[1];

	char *ua =  ParseUserAgent(s);
	printf("input: %s\nresult: %s\n", s, ua);
	FreeUserAgent(ua);
	return 0;
}

