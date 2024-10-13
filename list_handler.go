package main

import (
	"github.com/posteo/go-agentx"
	"github.com/posteo/go-agentx/pdu"
	"github.com/posteo/go-agentx/value"
)

// ListHandler is a helper that takes a list of oids and implements
// a default behaviour for that list.
type ListHandler struct {
	oids  []value.OID
	items map[string]*agentx.ListItem
}

// Add adds a list item for the provided oid and returns it.
func (l *ListHandler) Add(oid value.OID) *agentx.ListItem {
	if l.items == nil {
		l.items = make(map[string]*agentx.ListItem)
	}

	l.oids = append(l.oids, oid)
	item := &agentx.ListItem{}
	l.items[oid.String()] = item
	return item
}

// Get tries to find the provided oid and returns the corresponding value.
func (l *ListHandler) Get(oid value.OID) (value.OID, pdu.VariableType, interface{}, error) {
	if l.items == nil {
		return nil, pdu.VariableTypeNoSuchObject, nil, nil
	}

	item, ok := l.items[oid.String()]
	if ok {
		return oid, item.Type, item.Value, nil
	}
	return nil, pdu.VariableTypeNoSuchObject, nil, nil
}

// GetNext tries to find the value that follows the provided oid and returns it.
func (l *ListHandler) GetNext(from value.OID, includeFrom bool, to value.OID) (value.OID, pdu.VariableType, interface{}, error) {
	if l.items == nil {
		return nil, pdu.VariableTypeNoSuchObject, nil, nil
	}

	for _, oid := range l.oids {
		if oidWithin(oid, from, includeFrom, to) {
			return l.Get(oid)
		}
	}

	return nil, pdu.VariableTypeNoSuchObject, nil, nil
}

func oidWithin(oid value.OID, from value.OID, includeFrom bool, to value.OID) bool {
	fromCompare := compareOids(from, oid)
	toCompare := compareOids(to, oid)

	return (fromCompare == -1 || (fromCompare == 0 && includeFrom)) && (toCompare == 1)
}
