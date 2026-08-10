package lifecycle

import "github.com/jamessawle/sbxflow/internal/domain/configuration"

// Target is the declaration identity used by lifecycle operations.
type Target = configuration.LifecycleTarget

// TargetResolver resolves the nearest repository lifecycle target.
type TargetResolver = configuration.TargetResolver
