package main

import (
	"fmt"
	"github.com/miekg/dns"
	"log"
)

// Key stores all attributes for a DNSSEC signing key
type Key struct {
	Base           string // base key filename prefix
	Key            string // DNSKEY
	Private        string // private key
	DSKeyTag       int    // DS key tag
	DSAlgo         int    // DS algorithm
	DSDigestType   int    // DS digest type
	DSDigest       string // DS digest
	DSRecordString string // full DS record in zone file format
}

// NewKey generates a new DNSSEC signing key for a zone
func NewKey(zone string) Key {
	key := &dns.DNSKEY{
		Hdr:       dns.RR_Header{Name: dns.Fqdn(zone), Class: dns.ClassINET, Ttl: 3600, Rrtype: dns.TypeDNSKEY},
		Algorithm: dns.ECDSAP256SHA256, Flags: 257, Protocol: 3,
	}

	priv, err := key.Generate(256)
	if err != nil {
		log.Fatal(err)
	}

	ds := key.ToDS(dns.SHA256)

	return Key{
		Base:           fmt.Sprintf("K%s+%03d+%05d", key.Header().Name, key.Algorithm, key.KeyTag()),
		Key:            key.String(),
		Private:        key.PrivateKeyString(priv),
		DSKeyTag:       int(ds.KeyTag),
		DSAlgo:         int(ds.Algorithm),
		DSDigestType:   int(ds.DigestType),
		DSDigest:       ds.Digest,
		DSRecordString: ds.String(),
	}
}
