package settings

// DefaultFeatureFlags is the authoritative seed list of feature flags
// inserted on first run. Flags controlling PII mutation MUST keep
// Enabled: false per CLAUDE.md Hard Rule #20 — PII features are opt-in,
// never on by default.
var DefaultFeatureFlags = []FeatureFlag{
	{Name: "PII_REDACTION_ENABLED", Enabled: false, DefaultValue: "false"},
	{Name: "PII_RESPONSE_SANITIZATION", Enabled: false, DefaultValue: "false"},
	{Name: "RATE_LIMIT_ENABLED", Enabled: false, DefaultValue: "false"},
	{Name: "CACHE_ENABLED", Enabled: true, DefaultValue: "true"},
	{Name: "AUDIT_LOG_ENABLED", Enabled: true, DefaultValue: "true"},
}
