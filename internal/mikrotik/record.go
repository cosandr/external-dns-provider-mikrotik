package mikrotik

import (
	"fmt"
	"math"
	"time"

	log "github.com/sirupsen/logrus"
	"sigs.k8s.io/external-dns/endpoint"
)

// DNSRecord represents a MikroTik DNS record
// https://help.mikrotik.com/docs/display/ROS/DNS#DNS-DNSStatic
type DNSRecord struct {
	ID             string `json:".id,omitempty"`
	Address        string `json:"address,omitempty"`
	CName          string `json:"cname,omitempty"`
	ForwardTo      string `json:"forward-to,omitempty"`
	MXExchange     string `json:"mx-exchange,omitempty"`
	Name           string `json:"name"`
	SrvPort        string `json:"srv-port,omitempty"`
	SrvTarget      string `json:"srv-target,omitempty"`
	Text           string `json:"text,omitempty"`
	Type           string `json:"type"`
	AddressList    string `json:"address-list,omitempty"`
	Comment        string `json:"comment,omitempty"`
	Disabled       string `json:"disabled,omitempty"`
	MatchSubdomain string `json:"match-subdomain,omitempty"`
	MXPreference   string `json:"mx-preference,omitempty"`
	NS             string `json:"ns,omitempty"`
	Regexp         string `json:"regexp,omitempty"`
	SrvPriority    string `json:"srv-priority,omitempty"`
	SrvWeight      string `json:"srv-wright,omitempty"`
	TTL            string `json:"ttl,omitempty"`
}

// NewDNSRecord converts an ExternalDNS Endpoint to a Mikrotik DNSRecord
func NewRecordFromEndpoint(endpoint *endpoint.Endpoint, defaultTtl endpoint.TTL, defaultComment string) (*DNSRecord, error) {
	log.Debugf("converting ExternalDNS endpoint: %v", endpoint)
	tmpTtl := defaultTtl
	if endpoint.RecordTTL.IsConfigured() {
		tmpTtl = endpoint.RecordTTL
	}

	ttl, err := time.ParseDuration(fmt.Sprintf("%ds", tmpTtl))
	if err != nil {
		log.Errorf("Cannot parse TTL: %v", err)
		return nil, err
	}

	record := DNSRecord{
		Name:    endpoint.DNSName,
		Type:    endpoint.RecordType,
		Comment: defaultComment,
		TTL:     ttl.String(),
	}

	switch endpoint.RecordType {
	case "A", "AAAA":
		record.Address = endpoint.Targets[0]
	case "CNAME":
		record.CName = endpoint.Targets[0]
	case "TXT":
		record.Text = endpoint.Targets[0]

	default:
		return nil, fmt.Errorf("unsupported DNS type: %s", endpoint.RecordType)
	}
	log.Debugf("converted Mikrotik DNS Record: %v", record)

	return &record, nil
}

func NewEndpointFromRecord(record DNSRecord) (*endpoint.Endpoint, error) {
	log.Debugf("converting Mikrotik DNS record: %v", record)

	dur, err := time.ParseDuration(record.TTL)
	if err != nil {
		return nil, err
	}

	recType := record.Type
	if recType == "" {
		recType = "A"
	}

	ep := endpoint.Endpoint{
		DNSName:    record.Name,
		RecordType: recType,
		RecordTTL:  endpoint.TTL(math.Round(dur.Seconds())),
		// TODO: ProviderSpecific
	}
	switch record.Type {
	case "", "A", "AAAA": // "" means A record because mikrotik is weird like that... :P
		ep.Targets = endpoint.NewTargets(record.Address)
	case "CNAME":
		ep.Targets = endpoint.NewTargets(record.CName)
	case "TXT":
		ep.Targets = endpoint.NewTargets(record.Text)

	default:
		return nil, fmt.Errorf("unsupported DNS type: %s", record.Type)
	}
	log.Debugf("converted ExternalDNS endpoint: %v", ep)

	return &ep, nil
}
