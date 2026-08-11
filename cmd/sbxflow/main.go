package main

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"time"

	"github.com/jamessawle/sbxflow/internal/adapters/inbound/cli"
	"github.com/jamessawle/sbxflow/internal/adapters/outbound/declaration"
	"github.com/jamessawle/sbxflow/internal/adapters/outbound/sbx"
	"github.com/jamessawle/sbxflow/internal/application/doctor"
	"github.com/jamessawle/sbxflow/internal/application/lifecycle"
	applicationvalidation "github.com/jamessawle/sbxflow/internal/application/validation"
	"github.com/jamessawle/sbxflow/internal/domain/configuration"
	buildinfo "github.com/jamessawle/sbxflow/internal/ports/buildInfo"
	declarationport "github.com/jamessawle/sbxflow/internal/ports/declaration"
	sandboxport "github.com/jamessawle/sbxflow/internal/ports/sandbox"
)

const (
	doctorCommandTimeout  = 15 * time.Second
	sandboxCommandTimeout = 30 * time.Second
)

// main initializes the command-line application and executes it with configured
// declaration, validation, diagnostic, and sandbox lifecycle runners. It preserves
// subprocess exit codes and uses status 1 for other execution failures.
func main() {
	declarations := declaration.NewRepository()
	doctorSandboxes := sbx.NewClient(doctorCommandTimeout)
	lifecycleSandboxes := sbx.NewClient(sandboxCommandTimeout)
	var declarationLoader declarationport.Loader = declarations
	var localKitResolver declarationport.LocalKitResolver = declarations
	var targetResolver declarationport.TargetResolver = declarations
	var doctorInspector sandboxport.Inspector = doctorSandboxes
	var localKitValidator sandboxport.KitValidator = lifecycleSandboxes
	var upSandboxes interface {
		sandboxport.StateLookup
		sandboxport.Creator
		sandboxport.Remover
		sandboxport.Runner
		sandboxport.NetworkPolicy
	} = lifecycleSandboxes
	var downSandboxes interface {
		sandboxport.Lookup
		sandboxport.Stopper
	} = lifecycleSandboxes
	var destroySandboxes interface {
		sandboxport.Lookup
		sandboxport.Remover
		sandboxport.NetworkPolicy
	} = lifecycleSandboxes
	configurationResolver := configuration.Resolver{
		Declarations: declarationLoader,
		LocalPaths:   localKitResolver,
	}
	validationRunner := applicationvalidation.NewValidator(configurationResolver, localKitValidator)
	doctorRunner := doctor.NewRunner(doctorInspector, doctor.NewDefaultChecks()...)
	upRunner := lifecycle.UpRunner{Validation: validationRunner, Sandboxes: upSandboxes}
	downRunner := lifecycle.DownRunner{Targets: targetResolver, Sandboxes: downSandboxes}
	destroyRunner := lifecycle.DestroyRunner{Targets: targetResolver, Sandboxes: destroySandboxes}

	err := cli.Execute(
		context.Background(),
		os.Args[1:],
		cli.Streams{In: os.Stdin, Out: os.Stdout, Err: os.Stderr},
		buildinfo.Current(),
		doctorRunner,
		validationRunner,
		upRunner,
		downRunner,
		destroyRunner,
	)
	if err != nil {
		var exitError *exec.ExitError
		if errors.As(err, &exitError) {
			os.Exit(exitError.ExitCode())
		}
		os.Exit(1)
	}
}
