package doctor

import (
	sandboxport "github.com/jamessawle/sbxflow/internal/ports/sandbox"
)

// CommandOutput is the captured output from one Docker Sandboxes inspection.
type CommandOutput = sandboxport.Output

// Inspector is the read-only sandbox port used by doctor.
type Inspector = sandboxport.Inspector

// NewDefaultChecks returns the production diagnostic checks in execution
// order. Their infrastructure dependency is supplied separately.
func NewDefaultChecks() []Check {
	return []Check{compatibilityCheck{}, diagnosticsCheck{}, networkPolicyCheck{}, kitSourcesCheck{}}
}
