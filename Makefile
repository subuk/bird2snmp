GO = go
SOURCES = $(shell ls *.go)

build: bird2snmp

bird2snmp: $(SOURCES)
	go build .

test:
	go test ./...

clean:
	rm -f bird2snmp
