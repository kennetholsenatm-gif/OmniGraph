package netbox

import (
	"encoding/json"
	"fmt"
	"net"
	"strings"
)

// WebhookV1 is the versioned JSON body for a custom NetBox receiver or OmniGraph webhook.
const WebhookAPIVersion = "omnigraph/netbox-sync/v1"

type WebhookV1 struct {
	APIVersion     string `json:"apiVersion"`
	Action         string `json:"action"`
	IP             string `json:"ip,omitempty"`
	CIDR           string `json:"cidr,omitempty"`
	Role           string `json:"role,omitempty"`
	SiteID         int    `json:"siteId,omitempty"`
	DeviceID       int    `json:"deviceId,omitempty"`
	Environment    string `json:"environment,omitempty"`
	IdempotencyKey string `json:"idempotencyKey,omitempty"`
}

// Validate checks required fields and address shapes.
func (w *WebhookV1) Validate() error {
	if w == nil {
		return fmt.Errorf("netbox: nil webhook payload")
	}
	if strings.TrimSpace(w.Action) == "" {
		return fmt.Errorf("netbox: action is required")
	}
	hasIP := strings.TrimSpace(w.IP) != ""
	hasCIDR := strings.TrimSpace(w.CIDR) != ""
	if !hasIP && !hasCIDR {
		return fmt.Errorf("netbox: ip or cidr is required")
	}
	if hasIP {
		if net.ParseIP(strings.TrimSpace(w.IP)) == nil {
			return fmt.Errorf("netbox: invalid ip %q", w.IP)
		}
	}
	if hasCIDR {
		_, _, err := net.ParseCIDR(strings.TrimSpace(w.CIDR))
		if err != nil {
			return fmt.Errorf("netbox: cidr: %w", err)
		}
	}
	return nil
}

// MarshalJSON returns the canonical wire form with apiVersion set.
func (w *WebhookV1) MarshalJSON() ([]byte, error) {
	if err := w.Validate(); err != nil {
		return nil, err
	}
	type alias WebhookV1
	a := alias(*w)
	if strings.TrimSpace(a.APIVersion) == "" {
		a.APIVersion = WebhookAPIVersion
	}
	return json.Marshal(&a)
}
