package controller

import (
	"net"

	"github.com/godbus/dbus/v5"
	"github.com/holoplot/go-avahi"
)

// MdnsClient manages services
type MdnsClient struct {
	server     *avahi.Server
	entryGroup *avahi.EntryGroup
	services   map[string]*Service
	records    map[string]*Record
}

type Service struct {
	hostname string
	port     int
}

type Record struct {
	hostname string
	address  net.IP
}

const (
	AVAHI_DNS_TYPE_A     = uint16(0x01)
	AVAHI_DNS_TYPE_NS    = uint16(0x02)
	AVAHI_DNS_TYPE_CNAME = uint16(0x05)
	AVAHI_DNS_TYPE_SOA   = uint16(0x06)
	AVAHI_DNS_TYPE_PTR   = uint16(0x0C)
	AVAHI_DNS_TYPE_HINFO = uint16(0x0D)
	AVAHI_DNS_TYPE_MX    = uint16(0x0F)
	AVAHI_DNS_TYPE_TXT   = uint16(0x10)
	AVAHI_DNS_TYPE_AAAA  = uint16(0x1C)
	AVAHI_DNS_TYPE_SRV   = uint16(0x21)

	AVAHI_DNS_CLASS_IN = uint16(0x01)

	AVAHI_PUBLISH_NO_REVERSE = 16

	AVAHI_PROTO_INET   = int32(0)  /**< IPv4 */
	AVAHI_PROTO_INET6  = int32(1)  /**< IPv6 */
	AVAHI_PROTO_UNSPEC = int32(-1) /**< Unspecified/all protocol(s) */

	AVAHI_IF_UNSPEC = int32(-1) /**< Unspecified/all interface(s) */
)

// NewClient creates MdnsClient
func NewClient() (*MdnsClient, error) {
	conn, err := dbus.SystemBus()
	if err != nil {
		return nil, err
	}

	server, err := avahi.ServerNew(conn)
	if err != nil {
		return nil, err
	}

	entryGroup, err := server.EntryGroupNew()
	if err != nil {
		return nil, err
	}

	return &MdnsClient{
		server, entryGroup, map[string]*Service{}, map[string]*Record{},
	}, nil
}

// Close closes the connection to avahi
func (m *MdnsClient) Close() {
	m.entryGroup.Reset()
	m.server.Close()
}

// RegisterService registers a new Service to avahi
func (m *MdnsClient) RegisterService(hostname string, port int) {
	service := &Service{
		hostname,
		port,
	}
	m.services[hostname] = service

	m.addService(service)
	m.commit()
}

// RegisterRecord registers a new record to avahi
func (m *MdnsClient) RegisterRecord(hostname string, address net.IP) error {
	record := &Record{
		hostname,
		address,
	}

	if oldRecord, ok := m.records[hostname]; ok {
		if record.address.String() != oldRecord.address.String() {
			delete(m.records, hostname)
			m.reload()
		} else {
			return nil
		}
	}

	m.records[hostname] = record

	err := m.addRecord(record)
	if err != nil {
		return err
	}
	return m.commit()
}

// Unregister removes the sevice
func (m *MdnsClient) Unregister(hostname string) error {
	delete(m.services, hostname)
	delete(m.records, hostname)

	return m.reload()
}

func (m *MdnsClient) reload() error {
	err := m.entryGroup.Reset()
	if err != nil {
		return err
	}

	for _, service := range m.services {
		err = m.addService(service)
		if err != nil {
			return err
		}
	}
	for _, record := range m.records {
		err = m.addRecord(record)
		if err != nil {
			return err
		}
	}

	return m.commit()
}

func (m *MdnsClient) addService(service *Service) error {
	return m.entryGroup.AddService(
		AVAHI_IF_UNSPEC,
		AVAHI_PROTO_UNSPEC,
		AVAHI_PUBLISH_NO_REVERSE,
		service.hostname,
		"_http._tcp",
		"",
		service.hostname,
		uint16(service.port),
		[][]byte{[]byte("")},
	)
}

func (m *MdnsClient) addRecord(record *Record) error {
	proto := AVAHI_PROTO_UNSPEC
	if len(record.address) == net.IPv4len {
		proto = AVAHI_PROTO_INET
	} else if len(record.address) == net.IPv6len {
		proto = AVAHI_PROTO_INET6
	}

	return m.entryGroup.AddAddress(
		AVAHI_IF_UNSPEC,
		proto,
		AVAHI_PUBLISH_NO_REVERSE,
		record.hostname,
		record.address.String(),
	)
}

func (m *MdnsClient) commit() error {
	b, err := m.entryGroup.IsEmpty()
	if err != nil {
		return err
	}

	if !b {
		return m.entryGroup.Commit()
	}

	return nil
}
