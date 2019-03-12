go build -o libuap.a -buildmode=c-archive main.go
gcc -o test main.c libuap.a
