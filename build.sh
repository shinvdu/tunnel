#!/bin/bash
export GO111MODULE=on

if ! [[ -d bin ]]; then
  mkdir bin
fi

# Linux Build
GOOS=linux GOARCH=amd64 go build -o ./bin/tun-linux-amd64 ./tun.go

# LinuxARM Build
GOOS=linux GOARCH=arm64 go build -o ./bin/tun-linux-arm64 ./tun.go

# macOS Build
GOOS=darwin GOARCH=amd64 go build -o ./bin/tun-darwin-amd64 ./tun.go

# Windows Build
GOOS=windows GOARCH=amd64 go build -o ./bin/tun-windows-amd64.exe ./tun.go

echo "DONE!!!"