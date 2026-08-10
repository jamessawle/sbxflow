package validation

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jamessawle/sbxflow/internal/configuration"
	"github.com/jamessawle/sbxflow/internal/sbx"
)

// LocalKitResult reports Docker's validation outcome for one local target.
type LocalKitResult struct {
	Target      configuration.LocalKit
	Valid       bool
	Diagnostics string
	Err         error
}

func validateLocalKits(ctx context.Context, targets []configuration.LocalKit, client sbx.Client) []LocalKitResult {
	if len(targets) == 0 {
		return nil
	}
	paths := make([]string, 0, len(targets))
	for _, target := range targets {
		paths = append(paths, target.Path)
	}
	outputs, err := client.ValidateKits(ctx, paths)
	if err != nil {
		return []LocalKitResult{{Err: fmt.Errorf("local kit validation could not be completed: sbx is unavailable: %w", err)}}
	}

	results := make([]LocalKitResult, 0, len(targets))
	for index, target := range targets {
		output := outputs[index]
		diagnostics := strings.TrimSpace(strings.Join(nonempty(string(output.Stderr), string(output.Stdout)), "\n"))
		result := LocalKitResult{Target: target, Valid: output.Err == nil, Diagnostics: diagnostics}
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
