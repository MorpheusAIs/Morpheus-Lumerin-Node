/**
 * Placeholder replies reported from the field: models answered with their chat
 * template's tool syntax (a `<tool_calls>` block naming list_files, and
 * plan-update markup) instead of an answer. Shared by the detector, runner and
 * renderer regression tests so every layer is held to the same inputs.
 */
export const LEAKED_TOOL_CALLS_BLOCK = `<tool_calls>
[{"name": "list_files", "arguments": {"path": "."}}]
</tool_calls>`

export const LEAKED_PLAN_UPDATE = `<update_plan_step>
{"id": "step-1", "status": "in_progress", "note": "Listing the project files"}
</update_plan_step>`

export const LEAKED_PLACEHOLDER_REPLY = `I'll start by looking at the project.

${LEAKED_TOOL_CALLS_BLOCK}

${LEAKED_PLAN_UPDATE}`

export const LEAKED_SINGLE_TOOL_CALL = `<tool_call>
{"name": "list_files", "arguments": {"path": "."}}
</tool_call>`

export const LEAKED_BARE_JSON_CALL = `{"name": "list_files", "arguments": {"path": "."}}`
