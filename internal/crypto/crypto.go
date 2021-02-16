package crypto

import (
	"crypto/rand"
	"crypto/subtle"
	"crypto/tls"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"os"
	"path/filepath"

	"golang.org/x/crypto/acme/autocert"

	"github.com/miekg/dns"
	"golang.org/x/crypto/argon2"
)

// DNSSECKey stores all attributes for a DNSSEC signing key
type DNSSECKey struct {
	Base           string `json:"base"`           // base key filename prefix
	Key            string `json:"key"`            // DNSKEY
	Private        string `json:"private"`        // private key
	DSKeyTag       int    `json:"dskeytag"`       // DS key tag
	DSAlgo         int    `json:"dsalgo"`         // DS algorithm
	DSDigestType   int    `json:"dsdigesttype"`   // DS digest type
	DSDigest       string `json:"dsdigest"`       // DS digest
	DSRecordString string `json:"dsrecordstring"` // full DS record in zone file format
}

// NewKey generates a new DNSSEC signing key for a zone
func NewKey(zone string) DNSSECKey {
	key := &dns.DNSKEY{
		Hdr:       dns.RR_Header{Name: dns.Fqdn(zone), Class: dns.ClassINET, Ttl: 3600, Rrtype: dns.TypeDNSKEY},
		Algorithm: dns.ECDSAP256SHA256, Flags: 257, Protocol: 3,
	}

	priv, err := key.Generate(256)
	if err != nil {
		log.Fatal(err)
	}

	ds := key.ToDS(dns.SHA256)

	return DNSSECKey{
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

// RandomString returns a securely generated random string
func RandomString() string {
	length := 48
	const letters = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	ret := make([]byte, length)
	for i := 0; i < length; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			log.Fatalf("system RNG error: %v\n", err)
		}
		ret[i] = letters[num.Int64()]
	}

	return string(ret)
}

// argon2IDKey computes an argon2 hash by given input and salt
func argon2IDKey(input []byte, salt []byte) []byte {
	return argon2.IDKey(input, salt, 1, 64*1024, 4, 32)
}

// PasswordHash hashes a plaintext string and returns an argon2 hash byte slice
func PasswordHash(plaintext string) ([]byte, error) {
	// Generate a random 16-byte salt
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return make([]byte, 0), err
	}

	hash := argon2IDKey([]byte(plaintext), salt)

	return append(salt, hash...), nil
}

// ValidHash validates a hash and provided plaintext password
func ValidHash(payload []byte, plaintext string) bool {
	salt := payload[:16]
	hash := payload[16:]

	providedHash := argon2IDKey([]byte(plaintext), salt)
	return subtle.ConstantTimeCompare(hash, providedHash) == 1
}

// AcmeValidationHandler creates an HTTP ACME validation server
func AcmeValidationHandler(domain string) {
	// Create the cache directory
	dir := filepath.Join(os.TempDir(), "autocert-cache")
	if err := os.MkdirAll(dir, 0700); err == nil {
		log.Fatal(err)
	}

	// create the autocert.Manager with domains and path to the cache
	certManager := autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(domain),
		Cache:      autocert.DirCache(dir),
	}

	// create the server itself
	server := &http.Server{
		Addr:      ":5001",
		TLSConfig: &tls.Config{GetCertificate: certManager.GetCertificate},
	}

	// serve HTTPS
	log.Printf("Serving http/https for: %+v", domain)
	log.Fatal(server.ListenAndServe())
}
