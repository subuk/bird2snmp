package bird

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Routes struct {
	Imported  int
	Filtered  int
	Exported  int
	Preferred int
}

type BGP struct {
	NeighborAddress string
	NeighborAS      int
	LocalAS         int
	NeighborID      string
}

type ProtocolState struct {
	Name   string
	Proto  string
	Table  string
	State  string
	Since  time.Time
	Info   string
	Routes *Routes
	BGP    *BGP
}

func trimRepeatingSpace(s string) string {
	space := regexp.MustCompile(`\s+`)
	return space.ReplaceAllString(s, " ")
}

func parseBGP(s string) (*BGP, error) {
	out := &BGP{
		NeighborAddress: "",
		NeighborAS:      -1,
		LocalAS:         -1,
		NeighborID:      "",
	}

	if !strings.Contains(s, "BGP state:") {
		return nil, nil
	}

	addressRegex := regexp.MustCompile(`(.*)Neighbor address:(.*)`)
	address := trimRepeatingSpace(
		trimDupSpace(
			addressRegex.FindString(s),
		),
	)
	out.NeighborAddress = strings.Split(address, "Neighbor address: ")[1]

	neighborASRegex := regexp.MustCompile(`(.*)Neighbor AS:(.*)`)
	neighborAS := trimRepeatingSpace(
		trimDupSpace(
			neighborASRegex.FindString(s),
		),
	)
	neighborAS = strings.Split(neighborAS, "Neighbor AS: ")[1]
	neighborASInt, err := strconv.ParseInt(neighborAS, 10, 32)
	if err != nil {
		return nil, err
	}
	out.NeighborAS = int(neighborASInt)

	localASRegex := regexp.MustCompile(`(.*)Local AS:(.*)`)
	localAS := trimRepeatingSpace(
		trimDupSpace(
			localASRegex.FindString(s),
		),
	)
	localAS = strings.Split(localAS, "Local AS: ")[1]
	localASInt, err := strconv.ParseInt(localAS, 10, 32)
	if err != nil {
		return nil, err
	}
	out.LocalAS = int(localASInt)

	neighborIDRegex := regexp.MustCompile(`(.*)Neighbor ID:(.*)`)
	neighborID := trimRepeatingSpace(
		trimDupSpace(
			neighborIDRegex.FindString(s),
		),
	)
	neighborIDParts := strings.Split(neighborID, "Neighbor ID: ")
	if len(neighborIDParts) > 1 {
		out.NeighborID = neighborIDParts[1]
	}

	return out, nil
}

func parseRoutes(s string) (*Routes, error) {
	out := &Routes{
		Imported:  -1,
		Filtered:  -1,
		Exported:  -1,
		Preferred: -1,
	}

	routesRegex := regexp.MustCompile(`(.*)Routes:(.*)`)
	routes := routesRegex.FindString(s)
	routes = trimDupSpace(routes)
	routes = trimRepeatingSpace(routes)

	routeTokens := strings.Split(routes, "Routes: ")
	if len(routeTokens) < 2 {
		return out, nil
	}

	routesParts := strings.Split(routeTokens[1], ", ")

	for r := range routesParts {
		parts := strings.Split(routesParts[r], " ")
		num, err := strconv.ParseInt(parts[0], 10, 32)
		if err != nil {
			return nil, err
		}
		switch parts[1] {
		case "imported":
			out.Imported = int(num)
		case "filtered":
			out.Filtered = int(num)
		case "exported":
			out.Exported = int(num)
		case "preferred":
			out.Preferred = int(num)
		}
	}

	return out, nil
}

// ParseOne parses a single protocol
func ParseOne(p string) (*ProtocolState, error) {
	// Remove lines that start with BIRD
	birdRegex := regexp.MustCompile(`BIRD.*ready.*`)
	p = birdRegex.ReplaceAllString(p, "")
	tableHeaderRegex := regexp.MustCompile(`Name.*Info`)
	p = tableHeaderRegex.ReplaceAllString(p, "")

	// Remove leading and trailing newlines
	p = strings.Trim(p, "\n")
	header := strings.Split(p, "\n")[0]
	header = trimRepeatingSpace(header)
	headerParts := strings.Split(header, " ")

	if len(headerParts) < 6 {
		return nil, fmt.Errorf("invalid header len %d: %+v (%s)", len(headerParts), headerParts, header)
	}

	// Parse since timestamp
	since, err := time.Parse(time.DateTime, headerParts[4]+" "+headerParts[5])
	if err != nil {
		return nil, err
	}

	// Parse header
	protocolState := &ProtocolState{
		Name:  headerParts[0],
		Proto: headerParts[1],
		Table: headerParts[2],
		State: headerParts[3],
		Since: since,
		Info:  trimDupSpace(strings.Join(headerParts[6:], " ")),
	}

	routes, err := parseRoutes(p)
	if err != nil {
		return nil, err
	}
	protocolState.Routes = routes

	bgp, err := parseBGP(p)
	if err != nil {
		return nil, err
	}
	protocolState.BGP = bgp

	return protocolState, nil
}

// Parse parses a list of protocols
func Parse(p string) ([]*ProtocolState, error) {
	protocols := strings.Split(p, "\n\n")
	protocolStates := make([]*ProtocolState, len(protocols))
	for i, protocol := range protocols {
		protocolState, err := ParseOne(protocol)
		if err != nil {
			return nil, err
		}
		protocolStates[i] = protocolState
	}
	return protocolStates, nil
}

// ParseShowProtocols parses the output of `show protocols`
// bird must be configured to use the iso long timeformat. Other timeformats are currently not supported
func ParseShowProtocols(protocolsString string) ([]Protocol, error) {
	var protocols []Protocol
	for _, line := range strings.Split(strings.TrimSuffix(protocolsString, "\n"), "\n") {
		line = trimDupSpace(line)
		// Skip header
		if !(strings.Contains(line, "Name Proto Table") || strings.Contains(line, "ready.")) {
			parts := strings.Split(line, " ")
			if len(parts) < 6 {
				continue
			}
			info := strings.Join(parts[6:], " ")
			establishedSince := parts[4] + " " + parts[5]
			layout := "2006-01-02 15:04:05"
			timeVal, err := time.Parse(layout, establishedSince)
			if err != nil {
				return nil, err
			}
			protocols = append(protocols, Protocol{
				Name:  parts[0],
				Proto: parts[1],
				Table: parts[2],
				State: parts[3],
				Since: timeVal,
				Info:  info,
			})
		}
	}
	return protocols, nil
}
