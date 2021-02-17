// Package validation provides functions for struct data validation
package validation

import (
	"github.com/go-playground/validator/v10"
	"github.com/miekg/dns"

	"github.com/natesales/cdn-tree/internal/util"
)

// Register registers custom DNS validation handlers with the validator
func Register(validate *validator.Validate) error {
	// Map of validation functions
	validators := map[string]func(fl validator.FieldLevel) bool{
		"dns-rr": validateDnsLabel,
		"region": validateRegion,
	}

	// Register all validators
	for name, function := range validators {
		err := validate.RegisterValidation(name, function)
		if err != nil {
			return err
		}
	}

	return nil // nil error
}

// Validation functions

// validateDnsLabel attempts to create a DNS RR and returns if it succeeded or not
func validateDnsLabel(fl validator.FieldLevel) bool {
	_, err := dns.NewRR(fl.Field().String()) // ignore resulting RR
	return err != nil
}

// validateRegion validates that a provided region string is one of the permitted values
func validateRegion(fl validator.FieldLevel) bool {
	return util.Includes([]string{"us-west", "us-central", "eu-west", "eu-central"}, fl.Field().String())
}
