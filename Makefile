GO = go
SOURCES = $(shell ls *.go)

build: bird2snmp

bird2snmp: $(SOURCES)
	go build -o bin/bird2snmp .

test:
	go test ./...

release-binaries:
	GOARCH=mipsle GOOS=linux go build -o bin/release/bird2snmp.linux.mipsle
	GOARCH=mips   GOOS=linux go build -o bin/release/bird2snmp.linux.mips
	GOARCH=amd64  GOOS=linux go build -o bin/release/bird2snmp.linux.amd64
	GOARCH=arm64  GOOS=linux go build -o bin/release/bird2snmp.linux.arm64

clean:
	rm -rf bin/
