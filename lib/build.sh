go build -o libuap.so -buildmode=c-shared main.go
gcc -o test main.c -L. -luap
