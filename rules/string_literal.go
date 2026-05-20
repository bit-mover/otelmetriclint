package rules

// StringLiteral fires when the recognizer could not resolve the call's
// name argument to a string literal — meaning dynamic concatenation,
// an identifier, or some other non-literal produced no usable name. Every
// other rule depends on the literal text, so a non-literal name aborts
// further analysis of this call.
type StringLiteral struct{}

// ID returns the rule's stable identifier used in config and diagnostics.
func (StringLiteral) ID() string { return "string_literal" }

// Check returns one diagnostic if call.Name is empty, signalling the
// recognizer found a metric-creation call site but couldn't resolve its
// name to a string literal.
func (s StringLiteral) Check(call MetricCall) []Diagnostic {
	if call.Name != "" {
		return nil
	}
	return []Diagnostic{{
		Pos:     call.Pos,
		Message: "metric name must be a string literal — dynamic names cannot be validated",
		RuleID:  s.ID(),
	}}
}
