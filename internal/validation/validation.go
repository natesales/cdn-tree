package validation

import (
	"github.com/go-playground/validator/v10"
	"github.com/miekg/dns"
)

// Register registers custom DNS validation handlers with the validator
func Register(validate *validator.Validate) error {
	return validate.RegisterValidation("dns-rr", validateDnsLabel)
}

// Validation functions

// validateDnsLabel
func validateDnsLabel(fl validator.FieldLevel) bool {
	_, err := dns.NewRR(fl.Field().String()) // ignore resulting RR
	return err != nil
}
