all: esxi-disk-copy
archs: linux_amd64 macos_amd64 windows_amd64

esxi-disk-copy: *.go
	go build -o esxi-disk-copy *.go

linux_amd64: *.go
	GOOS=linux GOARCH=amd64 go build -o esxi-disk-copy.linux_amd64 *.go

macos_amd64: *.go
	GOOS=darwin GOARCH=amd64 go build -o esxi-disk-copy.macos_amd64 *.go

windows_amd64: *.go
	GOOS=windows GOARCH=amd64 go build -o esxi-disk-copy.windows_amd64.exe *.go

clean:
	rm -f esxi-disk-copy esxi-disk-copy.exe esxi-disk-copy.linux_amd64 esxi-disk-copy.macos_amd64 esxi-disk-copy.windows_amd64.exe
