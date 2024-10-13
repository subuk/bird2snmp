package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/natesales/go-bird"
	"github.com/posteo/go-agentx"
	"github.com/posteo/go-agentx/pdu"
	"github.com/posteo/go-agentx/value"
)

var (
	oidBgp                       = value.OID{1, 3, 6, 1, 2, 1, 15}
	oidBgpVersion                = value.OID{1, 3, 6, 1, 2, 1, 15, 1}
	oidBgpLocalAs                = value.OID{1, 3, 6, 1, 2, 1, 15, 2}
	oidBgpPeerState              = value.OID{1, 3, 6, 1, 2, 1, 15, 3, 1, 2}
	oidBgpPeerRemoteAddr         = value.OID{1, 3, 6, 1, 2, 1, 15, 3, 1, 7}
	oidBgpPeerFsmEstablishedTime = value.OID{1, 3, 6, 1, 2, 1, 15, 3, 1, 16}
	oidBgpIdentifier             = value.OID{1, 3, 6, 1, 2, 1, 15, 4}
)

// 1.3.6.1.2.1.15
type BirdBGPHandler struct {
	bird  *bird.Daemon
	birdT time.Time
	mu    *sync.RWMutex
	data  *ListHandler
}

func NewBirdBGPHandler(birdSocketPath string) (*BirdBGPHandler, error) {
	d, err := bird.New(birdSocketPath)
	if err != nil {
		return nil, fmt.Errorf("failed to connect bird: %w", err)
	}
	d.Read(nil)
	handler := &BirdBGPHandler{bird: d, mu: &sync.RWMutex{}}
	if err := handler.Refresh(); err != nil {
		return nil, fmt.Errorf("failed to refresh bird stats: %w", err)
	}
	return handler, nil
}

var bgpStateToInt = map[string]int32{
	"Established": 6,
	"Active":      3,
	"Connect":     2,
	"Idle":        1,
	"Down":        1,
	"Passive":     1,
}

func (h *BirdBGPHandler) Refresh() error {

	h.bird.Write("show status")
	showStatusString, err := h.bird.ReadString()
	if err != nil {
		return err
	}
	status := ParseShowStatus(showStatusString)

	h.bird.Write("show protocols all")
	protocolsAllString, err := h.bird.ReadString()
	if err != nil {
		return err
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	protocols := ParseShowProtocolsAll(protocolsAllString)
	h.birdT = time.Now().UTC()
	h.data = &ListHandler{}

	var item *agentx.ListItem
	item = h.data.Add(oidBgpVersion)
	item.Type = pdu.VariableTypeOctetString
	item.Value = "4"

	item = h.data.Add(append(oidBgpLocalAs, 0))
	item.Type = pdu.VariableTypeInteger
	item.Value = int32(protocols[0].LocalAs)

	for _, proto := range protocols {
		item = h.data.Add(append(oidBgpPeerState, ipToOid(proto.NeighborAddress)...))
		item.Type = pdu.VariableTypeInteger
		item.Value = bgpStateToInt[proto.State]
	}
	for _, proto := range protocols {
		item = h.data.Add(append(oidBgpPeerRemoteAddr, ipToOid(proto.NeighborAddress)...))
		item.Type = pdu.VariableTypeIPAddress
		item.Value = proto.NeighborAddress.To4()
	}
	for _, proto := range protocols {
		item = h.data.Add(append(oidBgpPeerFsmEstablishedTime, ipToOid(proto.NeighborAddress)...))
		item.Type = pdu.VariableTypeGauge32
		if proto.Up {
			item.Value = uint32(time.Since(proto.Since).Seconds())
		} else {
			item.Value = uint32(0)
		}
	}
	item = h.data.Add(append(oidBgpIdentifier, 0))
	item.Type = pdu.VariableTypeIPAddress
	item.Value = status.RouterId.To4()
	return nil
}

func (h *BirdBGPHandler) Register(priority byte, client *agentx.Client) error {
	session, err := client.Session()
	if err != nil {
		return fmt.Errorf("failed to initialize agentx session: %w", err)
	}
	session.Handler = h
	if err := session.Register(priority, oidBgp); err != nil {
		return fmt.Errorf("failed to register agentx session: %w", err)
	}
	return nil
}

func (h *BirdBGPHandler) Get(oid value.OID) (value.OID, pdu.VariableType, interface{}, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.data.Get(oid)
}

func (h *BirdBGPHandler) GetNext(from value.OID, includeFrom bool, to value.OID) (value.OID, pdu.VariableType, interface{}, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	// log.Printf("[TRACE] request from=%v includeFrom=%v to=%v", from, includeFrom, to)
	repOid, repType, repV, err := h.data.GetNext(from, includeFrom, to)
	// log.Printf("[TRACE] response oid=%v type=%s value=%v err=%v", repOid, repType, repV, err)
	return repOid, repType, repV, err
}
