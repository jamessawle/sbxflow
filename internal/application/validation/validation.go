// Package validation coordinates repository configuration validation.
package validation

import (
	"context"

	"github.com/jamessawle/sbxflow/internal/domain/configuration"
	sandboxport "github.com/jamessawle/sbxflow/internal/ports/sandbox"
)

// Report is the complete repository configuration validation state.
type Report = configuration.Validation

// Validator coordinates configuration-domain resolution with the external
// local-kit validation capability.
type Validator struct {
	Configurations configuration.ConfigurationResolver
	LocalKits      sandboxport.KitValidator
}

// NewValidator constructs the repository validation use case.
func NewValidator(configurations configuration.ConfigurationResolver, localKits sandboxport.KitValidator) Validator {
	return Validator{Configurations: configurations, LocalKits: localKits}
}

// Run validates the repository containing start.
func (v Validator) Run(ctx context.Context, start string) Report {
	resolution := v.Configurations.Resolve(start)
	report := Report{
		Declaration: resolution.Declaration,
		Linked:      resolution.Linked,
		Errors:      append([]error(nil), resolution.Errors...),
		Warnings:    append([]string(nil), resolution.Warnings...),
	}
	if !resolution.Valid() {
		return report
	}
	report.LocalKits = validateLocalKits(ctx, resolution.LocalKits, v.LocalKits)
	for _, result := range report.LocalKits {
		if result.Err != nil {
			report.Errors = append(report.Errors, result.Err)
		}
	}
	return report
}
