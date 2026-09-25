package apispec

// templateFacts is what the chat-template scan contributes beyond reasoning
// bindings: capability facts with their sources (tools variable or
// <tool_call> -> tools, vision tokens -> modalities.input image) and the
// reasoning default ({%- set enable_thinking = true|false %}, set thinking =
// ...). They rank below explicit backend facts and above the registry, and
// apply only when the stack is known (unknown wire vocabulary otherwise).
type templateFacts struct {
	caps                  CapabilityFacts
	thinkingDefault       string
	thinkingDefaultSource string
}

// capsFromTemplate is the seam for the scan in family.go; until it lands the
// template contributes nothing.
func capsFromTemplate(_ string) templateFacts {
	return templateFacts{}
}
