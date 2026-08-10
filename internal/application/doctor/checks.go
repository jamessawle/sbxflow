package doctor

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/jamessawle/sbxflow/internal/sbx"
)

const (
	// The v0.35.0 command contracts were verified against the released binary.
	// Docker's current v0.37.x CLI reference retains all four contracts. Exclude
	// the next minor until its pre-1.0 schemas have been deliberately verified.
	SupportedSbxMinimum          = "v0.35.0"
	SupportedSbxMaximumExclusive = "v0.38.0"
	defaultCommandTimeout        = 15 * time.Second
)

var sbxVersionPattern = regexp.MustCompile(`(?m)^sbx version:\s*v?(\d+)\.(\d+)\.(\d+)(?:[-+][0-9A-Za-z.-]+)?(?:\s|$)`)

// NewDefaultRunner constructs the production system diagnostic runner.
func NewDefaultRunner() Runner {
	sandboxes := sbx.NewClient(defaultCommandTimeout)
	return NewRunner(
		sandboxes,
		compatibilityCheck{},
		diagnosticsCheck{},
		networkPolicyCheck{},
		kitSourcesCheck{},
	)
}

type compatibilityCheck struct{}

func (compatibilityCheck) ID() string       { return "sbx-compatibility" }
func (compatibilityCheck) Grade() Grade     { return GradeRequired }
func (compatibilityCheck) Requires() []Fact { return nil }
func (compatibilityCheck) Run(ctx context.Context, environment Environment) Result {
	executable, err := environment.Sandboxes.Locate()
	if err != nil {
		return Result{
			Status:   StatusFail,
			Summary:  "sbx is not installed or is not executable",
			Guidance: "Install Docker Sandboxes and ensure `sbx` is available on PATH.",
		}
	}

	output := environment.Sandboxes.Version(ctx, executable)
	if output.Err != nil {
		return Result{
			Status:   StatusFail,
			Summary:  "sbx version could not be determined",
			Guidance: "Run `sbx version` and resolve the reported error.",
		}
	}

	version, ok := parseSbxVersion(string(output.Stdout))
	if !ok {
		return Result{
			Status:   StatusFail,
			Summary:  "sbx returned an unrecognized version",
			Guidance: fmt.Sprintf("Install a supported sbx version (%s to <%s).", SupportedSbxMinimum, SupportedSbxMaximumExclusive),
		}
	}

	if compareVersions(version, supportedMinimum()) < 0 || compareVersions(version, supportedMaximum()) >= 0 {
		return Result{
			Status: StatusFail,
			Summary: fmt.Sprintf(
				"sbx %s is unsupported; supported range is %s to <%s",
				version,
				SupportedSbxMinimum,
				SupportedSbxMaximumExclusive,
			),
			Guidance: "Install a supported Docker Sandboxes release.",
		}
	}

	return Result{
		Status:  StatusPass,
		Summary: fmt.Sprintf("sbx %s is compatible", version),
		Provides: map[Fact]string{
			FactSbxExecutable: executable,
			FactSbxCompatible: version.String(),
		},
	}
}

type diagnosticsCheck struct{}

func (diagnosticsCheck) ID() string   { return "docker-diagnostics" }
func (diagnosticsCheck) Grade() Grade { return GradeRequired }
func (diagnosticsCheck) Requires() []Fact {
	return []Fact{FactSbxExecutable, FactSbxCompatible}
}
func (diagnosticsCheck) Run(ctx context.Context, environment Environment) Result {
	output := environment.Sandboxes.Diagnose(ctx, environment.Facts[FactSbxExecutable])

	summary, err := decodeDiagnosticSummary(output.Stdout)
	if err != nil {
		return Result{
			Status:   StatusFail,
			Summary:  "Docker diagnostics could not be summarized",
			Guidance: "Run `sbx diagnose` for detailed results.",
		}
	}

	result := Result{
		Status: StatusPass,
		Summary: fmt.Sprintf(
			"%d passed, %d warned, %d failed, %d skipped",
			summary.Pass,
			summary.Warn,
			summary.Fail,
			summary.Skip,
		),
		Guidance: "Run `sbx diagnose` for detailed results.",
	}
	if summary.Fail > 0 {
		result.Status = StatusFail
	} else if summary.Warn > 0 {
		result.Status = StatusWarn
	}
	return result
}

type diagnosticSummary struct {
	Pass int
	Warn int
	Fail int
	Skip int
}

func decodeDiagnosticSummary(data []byte) (diagnosticSummary, error) {
	var envelope struct {
		Version string `json:"version"`
		Summary struct {
			Pass *int `json:"pass"`
			Warn *int `json:"warn"`
			Fail *int `json:"fail"`
			Skip *int `json:"skip"`
		} `json:"summary"`
	}
	if err := json.Unmarshal(data, &envelope); err != nil {
		return diagnosticSummary{}, err
	}
	if envelope.Version != "1.0" || envelope.Summary.Pass == nil || envelope.Summary.Warn == nil ||
		envelope.Summary.Fail == nil || envelope.Summary.Skip == nil {
		return diagnosticSummary{}, fmt.Errorf("unsupported diagnostic envelope")
	}
	summary := diagnosticSummary{
		Pass: *envelope.Summary.Pass,
		Warn: *envelope.Summary.Warn,
		Fail: *envelope.Summary.Fail,
		Skip: *envelope.Summary.Skip,
	}
	if summary.Pass < 0 || summary.Warn < 0 || summary.Fail < 0 || summary.Skip < 0 {
		return diagnosticSummary{}, fmt.Errorf("negative diagnostic total")
	}
	return summary, nil
}

type networkPolicyCheck struct{}

func (networkPolicyCheck) ID() string   { return "network-policy" }
func (networkPolicyCheck) Grade() Grade { return GradeAdvisory }
func (networkPolicyCheck) Requires() []Fact {
	return []Fact{FactSbxExecutable, FactSbxCompatible}
}
func (networkPolicyCheck) Run(ctx context.Context, environment Environment) Result {
	output := environment.Sandboxes.ListPolicies(ctx, environment.Facts[FactSbxExecutable])
	if output.Err != nil {
		return policyInspectionWarning()
	}

	rules, err := decodePolicyRules(output.Stdout)
	if err != nil {
		return policyInspectionWarning()
	}

	for _, rule := range rules {
		if rule.Status == "active" && rule.Scope == "global" && rule.Origin == "org" {
			return Result{Status: StatusPass, Summary: "global network policy is organisation-managed"}
		}
	}
	for _, rule := range rules {
		if rule.Status == "active" && rule.Scope == "global" && rule.Origin == "local" {
			return Result{Status: StatusPass, Summary: "global network policy is initialized"}
		}
	}

	return Result{
		Status:   StatusWarn,
		Summary:  "global network policy is not initialized",
		Guidance: "Review the available presets, then initialize one with `sbx policy init`.",
	}
}

type policyRule struct {
	PolicyID     string `json:"policy_id"`
	Scope        string `json:"scope"`
	ResourceType string `json:"resource_type"`
	Origin       string `json:"origin"`
	Status       string `json:"status"`
}

func decodePolicyRules(data []byte) ([]policyRule, error) {
	var envelope struct {
		Rules json.RawMessage `json:"rules"`
	}
	if err := json.Unmarshal(data, &envelope); err != nil {
		return nil, err
	}
	if len(envelope.Rules) == 0 || string(envelope.Rules) == "null" {
		return nil, fmt.Errorf("missing policy rules")
	}
	var rules []policyRule
	if err := json.Unmarshal(envelope.Rules, &rules); err != nil {
		return nil, err
	}
	for _, rule := range rules {
		if rule.PolicyID == "" || rule.Scope == "" || rule.ResourceType == "" || rule.Origin == "" || rule.Status == "" {
			return nil, fmt.Errorf("incomplete policy rule")
		}
	}
	return rules, nil
}

func policyInspectionWarning() Result {
	return Result{
		Status:   StatusWarn,
		Summary:  "global network policy could not be inspected",
		Guidance: "Run `sbx policy ls --type network` to inspect policy state.",
	}
}

type kitSourcesCheck struct{}

func (kitSourcesCheck) ID() string   { return "kit-sources" }
func (kitSourcesCheck) Grade() Grade { return GradeAdvisory }
func (kitSourcesCheck) Requires() []Fact {
	return []Fact{FactSbxExecutable, FactSbxCompatible}
}
func (kitSourcesCheck) Run(ctx context.Context, environment Environment) Result {
	output := environment.Sandboxes.GetKitAllowedSources(ctx, environment.Facts[FactSbxExecutable])
	if output.Err != nil {
		return kitSourcesInspectionWarning()
	}

	setting, err := decodeKitSources(output.Stdout)
	if err != nil {
		return kitSourcesInspectionWarning()
	}

	source := describeSettingSource(setting.Source)
	for _, value := range setting.Value {
		if value == "*" {
			return Result{
				Status:   StatusWarn,
				Summary:  fmt.Sprintf("remote kit sources are unrestricted (source: %s)", source),
				Guidance: kitSourcesGuidance(setting.Source),
			}
		}
	}

	return Result{
		Status:  StatusPass,
		Summary: fmt.Sprintf("remote kit sources are restricted (source: %s)", source),
	}
}

type kitSourcesSetting struct {
	Value  []string
	Source string
}

func decodeKitSources(data []byte) (kitSourcesSetting, error) {
	var envelope struct {
		Key    string          `json:"key"`
		Value  json.RawMessage `json:"value"`
		Source string          `json:"source"`
	}
	if err := json.Unmarshal(data, &envelope); err != nil {
		return kitSourcesSetting{}, err
	}
	if envelope.Key != "kit.allowedSources" || len(envelope.Value) == 0 || envelope.Source == "" {
		return kitSourcesSetting{}, fmt.Errorf("unsupported kit source setting")
	}
	var value []string
	if err := json.Unmarshal(envelope.Value, &value); err != nil {
		return kitSourcesSetting{}, err
	}
	if value == nil {
		return kitSourcesSetting{}, fmt.Errorf("missing kit source value")
	}
	return kitSourcesSetting{Value: value, Source: envelope.Source}, nil
}

func kitSourcesInspectionWarning() Result {
	return Result{
		Status:   StatusWarn,
		Summary:  "remote kit source restrictions could not be inspected",
		Guidance: "Run `sbx settings get --json kit.allowedSources` to inspect the effective value.",
	}
}

func describeSettingSource(source string) string {
	switch strings.ToLower(source) {
	case "override", "local":
		return "local override"
	case "env", "environment":
		return "environment"
	case "org", "organization", "organisation", "remote", "managed":
		return "organisation-managed"
	case "default":
		return "default"
	default:
		return source
	}
}

func kitSourcesGuidance(source string) string {
	switch describeSettingSource(source) {
	case "local override":
		return "Narrow or unset the local `kit.allowedSources` override."
	case "environment":
		return "Narrow or unset `DOCKER_SANDBOXES_KIT_ALLOWED_SOURCES`."
	case "organisation-managed":
		return "Ask your Docker Sandboxes administrator to restrict the allowed sources."
	default:
		return "Restrict `kit.allowedSources` to trusted publisher prefixes."
	}
}

type semanticVersion struct {
	major int
	minor int
	patch int
}

func (v semanticVersion) String() string {
	return fmt.Sprintf("v%d.%d.%d", v.major, v.minor, v.patch)
}

func parseSbxVersion(output string) (semanticVersion, bool) {
	matches := sbxVersionPattern.FindStringSubmatch(output)
	if len(matches) != 4 {
		return semanticVersion{}, false
	}
	components := make([]int, 3)
	for index, match := range matches[1:] {
		component, err := strconv.Atoi(match)
		if err != nil {
			return semanticVersion{}, false
		}
		components[index] = component
	}
	return semanticVersion{major: components[0], minor: components[1], patch: components[2]}, true
}

func compareVersions(left, right semanticVersion) int {
	leftComponents := []int{left.major, left.minor, left.patch}
	rightComponents := []int{right.major, right.minor, right.patch}
	for index := range leftComponents {
		if leftComponents[index] < rightComponents[index] {
			return -1
		}
		if leftComponents[index] > rightComponents[index] {
			return 1
		}
	}
	return 0
}

func supportedMinimum() semanticVersion { return semanticVersion{major: 0, minor: 35, patch: 0} }
func supportedMaximum() semanticVersion { return semanticVersion{major: 0, minor: 38, patch: 0} }
