build:
	GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o scrubber_linux_amd64 .
	GOOS=darwin GOARCH=amd64 go build -ldflags "-s -w" -o scrubber_darwin_amd64 .
	GOOS=windows GOARCH=amd64 go build -ldflags "-s -w" -o scrubber_windows.exe .
