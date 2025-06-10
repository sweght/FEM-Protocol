package protocol

import (
	"encoding/json"
	"fmt"
)

// GenericEnvelope provides a unified interface for handling any FEP envelope type
type GenericEnvelope struct {
	BaseEnvelope
	Body json.RawMessage `json:"body"`
}

// ParseEnvelope parses a generic envelope from JSON bytes
func ParseEnvelope(data []byte) (*GenericEnvelope, error) {
	var envelope GenericEnvelope
	if err := json.Unmarshal(data, &envelope); err != nil {
		return nil, fmt.Errorf("failed to parse envelope: %w", err)
	}
	return &envelope, nil
}

// ParseTypedEnvelope parses a generic envelope into a specific typed envelope
func (g *GenericEnvelope) ParseTypedEnvelope() (interface{}, error) {
	switch g.Type {
	case EnvelopeRegisterAgent:
		var envelope RegisterAgentEnvelope
		envelope.BaseEnvelope = g.BaseEnvelope
		if err := json.Unmarshal(g.Body, &envelope.Body); err != nil {
			return nil, err
		}
		return &envelope, nil

	case EnvelopeRegisterBroker:
		var envelope RegisterBrokerEnvelope
		envelope.BaseEnvelope = g.BaseEnvelope
		if err := json.Unmarshal(g.Body, &envelope.Body); err != nil {
			return nil, err
		}
		return &envelope, nil

	case EnvelopeEmitEvent:
		var envelope EmitEventEnvelope
		envelope.BaseEnvelope = g.BaseEnvelope
		if err := json.Unmarshal(g.Body, &envelope.Body); err != nil {
			return nil, err
		}
		return &envelope, nil

	case EnvelopeRenderInstruction:
		var envelope RenderInstructionEnvelope
		envelope.BaseEnvelope = g.BaseEnvelope
		if err := json.Unmarshal(g.Body, &envelope.Body); err != nil {
			return nil, err
		}
		return &envelope, nil

	case EnvelopeToolCall:
		var envelope ToolCallEnvelope
		envelope.BaseEnvelope = g.BaseEnvelope
		if err := json.Unmarshal(g.Body, &envelope.Body); err != nil {
			return nil, err
		}
		return &envelope, nil

	case EnvelopeToolResult:
		var envelope ToolResultEnvelope
		envelope.BaseEnvelope = g.BaseEnvelope
		if err := json.Unmarshal(g.Body, &envelope.Body); err != nil {
			return nil, err
		}
		return &envelope, nil

	case EnvelopeRevoke:
		var envelope RevokeEnvelope
		envelope.BaseEnvelope = g.BaseEnvelope
		if err := json.Unmarshal(g.Body, &envelope.Body); err != nil {
			return nil, err
		}
		return &envelope, nil

	default:
		return nil, fmt.Errorf("unknown envelope type: %s", g.Type)
	}
}

// GetBodyAs unmarshals the envelope body into the provided struct
func (g *GenericEnvelope) GetBodyAs(v interface{}) error {
	return json.Unmarshal(g.Body, v)
}