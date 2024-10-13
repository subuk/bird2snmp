package bird

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// isNumeric checks if a byte is character for number
func isNumeric(b byte) bool {
	return b >= byte('0') && b <= byte('9')
}

// trimDupSpace trims duplicate whitespace
func trimDupSpace(s string) string {
	headTailWhitespace := regexp.MustCompile(`^[\s\p{Zs}]+|[\s\p{Zs}]+$`)
	innerWhitespace := regexp.MustCompile(`[\s\p{Zs}]{2,}`)
	return innerWhitespace.ReplaceAllString(headTailWhitespace.ReplaceAllString(s, ""), " ")
}

// isIP checks if a string is an IPv4 or IPv6 address
func isIP(s string) bool {
	return net.ParseIP(s) != nil
}

// Daemon stores a BIRD socket connection
type Daemon struct {
	Conn net.Conn
}

// Protocol stores a BIRD protocol
type Protocol struct {
	Name  string
	Proto string
	Table string
	State string
	Since time.Time
	Info  string
}

type Route struct {
	Prefix      string
	Interface   string
	Protocol    string
	AddressType string
	Since       time.Time
	Weight      int
}

// New returns a new Daemon
func New(socket string) (*Daemon, error) {
	conn, err := net.Dial("unix", socket)
	if err != nil {
		return nil, err
	}
	return &Daemon{Conn: conn}, nil
}

// Close closes the BIRD socket connection
func (d *Daemon) Close() error {
	return d.Conn.Close()
}

// Read a line from bird socket, removing preceding status number, output it. Returns if there are more lines.
func (d *Daemon) Read(w io.Writer) bool {
	// Read from socket byte by byte, until reaching newline character
	c := make([]byte, 1024, 1024)
	pos := 0
	for {
		if pos >= 1024 {
			break
		}
		_, err := d.Conn.Read(c[pos : pos+1])
		if err != nil {
			panic(err)
		}
		if c[pos] == byte('\n') {
			break
		}
		pos++
	}

	c = c[:pos+1]

	// Remove preceding status numbers
	if pos > 4 && isNumeric(c[0]) && isNumeric(c[1]) && isNumeric(c[2]) && isNumeric(c[3]) {
		// There is a status number at beginning, remove it (first 5 bytes)
		if w != nil && pos > 6 {
			pos = 5
			if _, err := w.Write(c[pos:]); err != nil {
				panic(err)
			}
		}
		return c[0] != byte('0') && c[0] != byte('8') && c[0] != byte('9')
	} else {
		if w != nil {
			if _, err := w.Write(c[1:]); err != nil {
				panic(err)
			}
		}
		return true
	}
}

// ReadString reads the full BIRD response as a string
func (d *Daemon) ReadString() (string, error) {
	var buf bytes.Buffer
	for d.Read(&buf) {
	}
	if r := recover(); r != nil {
		return "", fmt.Errorf("%s", r)
	}
	return buf.String(), nil
}

// Write a command to BIRD
func (d *Daemon) Write(command string) {
	d.Conn.Write([]byte(strings.TrimRight(command, "\n") + "\n"))
}

// Protocols gets a slice of parsed protocols
func (d *Daemon) Protocols() ([]Protocol, error) {
	d.Write("show protocols")
	protocolsString, err := d.ReadString()
	if err != nil {
		return nil, err
	}
	return ParseShowProtocols(protocolsString)
}

// Routes gets a slice of parsed routes
func (d *Daemon) Routes(table string) ([]Route, error) {
	d.Write("show route table " + table)
	protocolsString, err := d.ReadString()
	if err != nil {
		return nil, err
	}

	var routes []Route
	headerString := "Table " + table + ":"
	currentRoute := ""
	seenHeader := false
	lines := strings.Split(strings.TrimSuffix(protocolsString, "\n"), "\n")
	for i := 0; i < len(lines); i++ {
		line := trimDupSpace(lines[i])
		// Skip header
		if seenHeader {
			parts := strings.Split(line, " ")
			// This really should never be true, but just in-case
			if parts[0] != "dev" {
				if isIP(strings.Split(parts[0], "/")[0]) {
					currentRoute = parts[0]
				} else {
					// Lets just always have the prefix in front to standardize the layout
					parts = append([]string{currentRoute}, parts...)
				}

				// Build base Route struct
				route := Route{
					Prefix:      currentRoute,
					Interface:   "",
					Protocol:    parts[2][1:], // Remove the leading [
					AddressType: parts[1],
				}

				// Lets get the time
				timeVal, err := time.Parse("2006-01-02 15:04:05", parts[3]+" "+strings.Split(parts[4], "]")[0])
				if err != nil {
					return nil, err
				}
				route.Since = timeVal

				// Lets get the weight
				weightIndex := 6
				if parts[weightIndex-1] == "from" {
					weightIndex += 2
				}
				if parts[weightIndex-1] != "*" {
					weightIndex = weightIndex - 1
				}
				weight, err := strconv.Atoi(strings.Split(parts[weightIndex][1:len(parts[weightIndex])-1], "/")[0])
				if err != nil {
					return nil, err
				}
				route.Weight = weight

				// Lets get the interface name
				i++
				interfaceLine := trimDupSpace(lines[i])
				if strings.Contains(interfaceLine, "dev") {
					route.Interface = interfaceLine[4:] // Remove the leading dev[SPACE]
				}
				routes = append(routes, route)
			}
		} else if strings.Contains(line, headerString) {
			seenHeader = true
		}
	}
	return routes, nil
}
