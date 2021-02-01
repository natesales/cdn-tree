package main

import (
	"fmt"
	"github.com/miekg/dns"
	"io/ioutil"
	"log"
)

func main() {
	zones := []string{"example.com"}

	for _, zone := range zones {
		key := &dns.DNSKEY{
			Hdr:       dns.RR_Header{Name: dns.Fqdn(zone), Class: dns.ClassINET, Ttl: 3600, Rrtype: dns.TypeDNSKEY},
			Algorithm: dns.ECDSAP256SHA256, Flags: 257, Protocol: 3,
		}

		priv, err := key.Generate(256)
		if err != nil {
			log.Fatal(err)
		}

		ds := key.ToDS(dns.SHA256)

		base := fmt.Sprintf("K%s+%03d+%05d", key.Header().Name, key.Algorithm, key.KeyTag())
		if err := ioutil.WriteFile(base+".key", []byte(key.String()+"\n"), 0644); err != nil {
			log.Fatal(err)
		}
		if err := ioutil.WriteFile(base+".private", []byte(key.PrivateKeyString(priv)), 0600); err != nil {
			log.Fatal(err)
		}

		fmt.Println(ds.String())
		fmt.Printf("Key tag: %d\n", ds.KeyTag)
		fmt.Printf("Algo: %d\n", ds.Algorithm)
		fmt.Printf("Digest Type: %d\n", ds.DigestType)
		fmt.Printf("Digest: %s\n", ds.Digest)
	}
}
