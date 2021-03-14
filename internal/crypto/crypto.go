// Package crypto provides functions and types for cryptographic operations (DNSSEC, TLS, ARGON2)
package crypto

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/subtle"
	"fmt"
	"log"
	"math/big"

	"github.com/go-acme/lego/v4/certcrypto"
	"github.com/go-acme/lego/v4/certificate"
	"github.com/go-acme/lego/v4/challenge/http01"
	"github.com/go-acme/lego/v4/lego"
	"github.com/go-acme/lego/v4/registration"
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

// ACME Client process

// AcmeUser implements acme.User
type AcmeUser struct {
	Email        string
	Registration *registration.Resource
	key          crypto.PrivateKey
}

func (u *AcmeUser) GetEmail() string {
	return u.Email
}

func (u AcmeUser) GetRegistration() *registration.Resource {
	return u.Registration
}

func (u *AcmeUser) GetPrivateKey() crypto.PrivateKey {
	return u.key
}

// NewCertRequest requests a new TLS certificate
func NewCertRequest(domain string) {
	// Create a user. New accounts need an email and private key to start.
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		log.Fatal(err)
	}

	myUser := AcmeUser{
		Email: "you@yours.com",
		key:   privateKey,
	}

	config := lego.NewConfig(&myUser)

	//config.CADirURL = "https://acme-v02.api.letsencrypt.org/directory"
	config.CADirURL = "https://acme-staging-v02.api.letsencrypt.org/directory"
	config.Certificate.KeyType = certcrypto.RSA2048

	// A client facilitates communication with the CA server.
	client, err := lego.NewClient(config)
	if err != nil {
		log.Fatal(err)
	}

	err = client.Challenge.SetHTTP01Provider(http01.NewProviderServer("", "5001"))
	if err != nil {
		log.Fatal(err)
	}

	// New users will need to register
	reg, err := client.Registration.Register(registration.RegisterOptions{TermsOfServiceAgreed: true})
	if err != nil {
		log.Fatal(err)
	}
	myUser.Registration = reg

	request := certificate.ObtainRequest{
		Domains: []string{domain},
		Bundle:  true,
	}

	certificates, err := client.Certificate.Obtain(request)
	if err != nil {
		log.Fatal(err)
	}

	// Each certificate comes back with the cert bytes, the bytes of the client's private key, and a certificate URL.
	fmt.Printf("%#v\n", certificates)
}
