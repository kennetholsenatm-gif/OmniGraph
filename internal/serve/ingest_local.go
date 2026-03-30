package serve

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/kennetholsenatm-gif/omnigraph/internal/omnistate"
)

// postIngestLocal accepts JSON from the web workspace after the browser reads local files
// (e.g. File System Access API). Each file is normalized into omnistate and merged.
//
// Request JSON:
//
//	{
//	  "clientSessionId": "optional opaque id",
//	  "files": [
//	    {
//	      "name": "terraform.tfstate",
//	      "contentType": "application/json",
//	      "encoding": "utf8" | "base64",
//	      "data": "..."
//	    }
//	  ]
//	}
//
// Response JSON: { "state": OmniGraphState, "errors": [...] }.
// HTTP 200 when the body is valid JSON and at least one file is present; partial failures
// appear in state.partialErrors and errors. HTTP 400 for empty files list or invalid envelope.
// Requires Authorization: Bearer (same experimental gate as other privileged APIs).
func (s *server) postIngestLocal(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	maxBytes := s.maxIngestBodyBytes
	if maxBytes <= 0 {
		maxBytes = 64 << 20
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
	defer r.Body.Close()

	var req ingestLocalRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		writeAPIErrorJSON(w, "INGEST_LOCAL_INVALID_JSON", "ingest/local: invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}
	if len(req.Files) == 0 {
		writeAPIErrorJSON(w, "INGEST_LOCAL_FILES_REQUIRED", "ingest/local: at least one file is required", http.StatusBadRequest)
		return
	}

	var frags []omnistate.OmniGraphStateFragment
	var topLevel []ingestResponseError

	for _, f := range req.Files {
		if err := validateIngestFileName(f.Name); err != nil {
			topLevel = append(topLevel, ingestResponseError{Path: f.Name, Code: "INGEST_NAME_INVALID", Message: "ingest/local: " + err.Error()})
			continue
		}
		raw, err := decodeIngestPayload(f.Encoding, f.Data)
		if err != nil {
			topLevel = append(topLevel, ingestResponseError{Path: f.Name, Code: "INGEST_DECODE_FAILED", Message: "ingest/local: " + err.Error()})
			continue
		}
		ref := omnistate.SourceRef{
			Type:     omnistate.DetectNormalizer(f.ContentType, f.Name).Kind(),
			Name:     f.Name,
			PathHint: f.Name,
		}
		n := omnistate.DetectNormalizer(f.ContentType, f.Name)
		frag, err := n.Normalize(r.Context(), omnistate.NormalizerInput{
			Data:        raw,
			ContentType: f.ContentType,
			Name:        f.Name,
			Ref:         ref,
		})
		if err != nil {
			topLevel = append(topLevel, ingestResponseError{Path: f.Name, Code: "INGEST_NORMALIZE_FAILED", Message: "ingest/local: " + err.Error()})
			continue
		}
		frags = append(frags, frag)
	}

	merged := omnistate.MergeFragments(req.ClientSessionID, frags...)
	merged.Revision = 1
	if s.syncHub != nil {
		// Side effect boundary: this replaces the hub's authoritative in-memory state.
		s.syncHub.replaceState(merged)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(ingestLocalResponse{
		State:  merged,
		Errors: topLevel,
	})
}

type ingestLocalRequest struct {
	ClientSessionID string           `json:"clientSessionId,omitempty"`
	Files           []ingestFileItem `json:"files"`
}

type ingestFileItem struct {
	Name        string `json:"name"`
	ContentType string `json:"contentType"`
	Encoding    string `json:"encoding"`
	Data        string `json:"data"`
	// ClientPathHint is optional display-only provenance from the File System Access API (not trusted as a server path).
	ClientPathHint string `json:"clientPathHint,omitempty"`
	// LastModifiedRFC3339 is optional metadata from the browser File object.
	LastModifiedRFC3339 string `json:"lastModified,omitempty"`
}

type ingestResponseError struct {
	Path    string `json:"path,omitempty"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

type ingestLocalResponse struct {
	State  omnistate.OmniGraphState `json:"state"`
	Errors []ingestResponseError    `json:"errors,omitempty"`
}

func validateIngestFileName(name string) error {
	s := strings.TrimSpace(name)
	if s == "" {
		return fmt.Errorf("empty name")
	}
	if strings.Contains(s, "..") {
		return fmt.Errorf("invalid name")
	}
	return nil
}

func decodeIngestPayload(encoding, data string) ([]byte, error) {
	switch strings.ToLower(strings.TrimSpace(encoding)) {
	case "utf8", "utf-8", "":
		return []byte(data), nil
	case "base64":
		return base64.StdEncoding.DecodeString(data)
	default:
		return nil, fmt.Errorf("unsupported encoding %q", encoding)
	}
}
