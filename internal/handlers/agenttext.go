package handlers

import "encoding/json"

// agentText builds a tool's MODEL-facing result text. MCP clients show the model
// the `content` (this text), NOT `structuredContent` (which drives the UI
// surfaces). So any data the agent needs to CHAIN the next call — compiled nodes,
// soul ids, validation findings, the slide id — has to live here, or the agent is
// blind. summary is the human-readable line; payload is JSON-serialized and
// appended so the agent can read/copy it.
func agentText(summary string, payload any) string {
	b, err := json.Marshal(payload)
	if err != nil || len(b) == 0 || string(b) == "null" {
		return summary
	}
	return summary + "\n" + string(b)
}
