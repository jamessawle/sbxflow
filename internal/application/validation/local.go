package validation

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jamessawle/sbxflow/internal/domain/configuration"
	sandboxport "github.com/jamessawle/sbxflow/internal/ports/sandbox"
)

func validateLocalKits(ctx context.Context, targets []configuration.LocalKit, validator sandboxport.KitValidator) []configuration.LocalKitValidation {
	if len(targets) == 0 {
		return nil
	}
	paths := make([]string, 0, len(targets))
	for _, target := range targets {
		paths = append(paths, target.Path)
	}
	outputs, err := validator.ValidateKits(ctx, paths)
	if err != nil {
		return []configuration.LocalKitValidation{{Err: fmt.Errorf("local kit validation could not be completed: sbx is unavailable: %w", err)}}
	}

	results := make([]configuration.LocalKitValidation, 0, len(targets))
	for index, target := range targets {
		output := outputs[index]
		diagnostics := strings.TrimSpace(strings.Join(nonempty(string(output.Stderr), string(output.Stdout)), "\n"))
		result := configuration.LocalKitValidation{Target: target, Valid: output.Err == nil, Diagnostics: diagnostics}
		if output.Err != nil {
			if errors.Is(output.Err, context.DeadlineExceeded) {
				result.Err = fmt.Errorf("local kit %s/%s validation timed out: %w", target.Source, target.Kit, output.Err)
			} else {
				result.Err = fmt.Errorf("local kit %s/%s was rejected by sbx: %w", target.Source, target.Kit, output.Err)
			}
		}
		results = append(results, result)
	}
	return results
}

func nonempty(values ...string) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			result = append(result, strings.TrimSpace(value))
		}
	}
	return result
}
