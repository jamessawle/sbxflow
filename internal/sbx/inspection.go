package sbx

import "context"

// Locate resolves the Docker Sandboxes executable on PATH.
func (c Client) Locate() (string, error) {
	return c.Commands.LookPath("sbx")
}

// Version returns the captured `sbx version` result.
func (c Client) Version(ctx context.Context, executable string) Output {
	return c.Commands.Run(ctx, executable, "version")
}

// Diagnose returns Docker's machine-readable diagnostics.
func (c Client) Diagnose(ctx context.Context, executable string) Output {
	return c.Commands.Run(ctx, executable, "diagnose", "--output", "json")
}

// ListPolicies returns all active and inactive policies as JSON.
func (c Client) ListPolicies(ctx context.Context, executable string) Output {
	return c.Commands.Run(ctx, executable, "policy", "ls", "--json", "--include-inactive")
}

// GetKitAllowedSources returns the effective kit source setting as JSON.
func (c Client) GetKitAllowedSources(ctx context.Context, executable string) Output {
	return c.Commands.Run(ctx, executable, "settings", "get", "--json", "kit.allowedSources")
}
