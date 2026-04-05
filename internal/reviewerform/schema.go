package reviewerform

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// PublicSchema is the slim JSON shape for GET /v1/reviewer/review-form-schema (id, label, options only per field).
type PublicSchema struct {
	SchemaVersion int             `json:"schema_version"`
	Sections      []PublicSection `json:"sections"`
}

type PublicSection struct {
	ID     string        `json:"id"`
	Title  string        `json:"title"`
	Fields []PublicField `json:"fields"`
}

// PublicField is one answerable key (radio_group parents are flattened into their items).
type PublicField struct {
	ID      string   `json:"id"`
	Label   string   `json:"label"`
	Options []string `json:"options"`
}

//go:embed schema.json
var embeddedSchemaJSON []byte

// Schema is the parsed independent-review form (copy + field definitions).
type Schema struct {
	SchemaVersion int          `json:"schema_version"`
	Page          PageCopy     `json:"page"`
	Sections      []SectionDef `json:"sections"`
	fieldIndex    map[string]fieldMeta
}

type PageCopy struct {
	Title             string `json:"title"`
	Description       string `json:"description,omitempty"`
	InstructionBanner string `json:"instruction_banner,omitempty"`
	RejectionNote     string `json:"rejection_note,omitempty"`
	SubmitLabel       string `json:"submit_label,omitempty"`
	SaveDraftLabel    string `json:"save_draft_label,omitempty"`
}

type SectionDef struct {
	ID          string     `json:"id"`
	Title       string     `json:"title"`
	Description string     `json:"description,omitempty"`
	Expandable  bool       `json:"expandable"`
	Fields      []FieldDef `json:"fields"`
}

type FieldDef struct {
	ID                 string      `json:"id"`
	Type               string      `json:"type"`
	Label              string      `json:"label"`
	Help               string      `json:"help,omitempty"`
	Description        string      `json:"description,omitempty"`
	Required           bool        `json:"required"`
	AllowPDFAttachment bool        `json:"allow_pdf_attachment,omitempty"`
	Min                int         `json:"min,omitempty"`
	Max                int         `json:"max,omitempty"`
	LowLabel           string      `json:"low_label,omitempty"`
	HighLabel          string      `json:"high_label,omitempty"`
	Items              []RadioItem `json:"items,omitempty"`
}

type RadioItem struct {
	ID      string   `json:"id"`
	Label   string   `json:"label"`
	Options []string `json:"options"`
}

type fieldMeta struct {
	typ      string
	min, max int
	options  map[string]struct{}
	label    string
}

// ValidationIssue describes one problem with answers against the form schema (for API error detail).
type ValidationIssue struct {
	FieldID    string `json:"field_id"`
	Label      string `json:"label,omitempty"`
	Code       string `json:"code"`
	MessageEng string `json:"message_eng"`
	MessageIdn string `json:"message_idn"`
}

// Load parses the embedded JSON schema (single source of truth for copy + questions).
func Load() (*Schema, error) {
	var s Schema
	if err := json.Unmarshal(embeddedSchemaJSON, &s); err != nil {
		return nil, fmt.Errorf("reviewerform: parse schema: %w", err)
	}
	s.buildIndex()
	return &s, nil
}

// RawJSON returns the embedded file bytes for API responses.
func RawJSON() json.RawMessage { return json.RawMessage(embeddedSchemaJSON) }

// ToPublic builds a minimal schema payload for clients (no page, no expandable, no type/required/min/max; checklist flattened).
func (s *Schema) ToPublic() PublicSchema {
	out := PublicSchema{SchemaVersion: s.SchemaVersion, Sections: make([]PublicSection, 0, len(s.Sections))}
	for _, sec := range s.Sections {
		ps := PublicSection{ID: sec.ID, Title: sec.Title, Fields: make([]PublicField, 0, len(sec.Fields))}
		for _, f := range sec.Fields {
			switch f.Type {
			case "long_text":
				ps.Fields = append(ps.Fields, PublicField{ID: f.ID, Label: f.Label, Options: []string{}})
			case "radio_group":
				for _, it := range f.Items {
					opts := append([]string(nil), it.Options...)
					ps.Fields = append(ps.Fields, PublicField{ID: it.ID, Label: it.Label, Options: opts})
				}
			case "scale":
				opts := scaleOptionStrings(f.Min, f.Max)
				ps.Fields = append(ps.Fields, PublicField{ID: f.ID, Label: f.Label, Options: opts})
			}
		}
		out.Sections = append(out.Sections, ps)
	}
	return out
}

func scaleOptionStrings(min, max int) []string {
	if max < min || min < 1 {
		return []string{}
	}
	n := max - min + 1
	opts := make([]string, 0, n)
	for i := min; i <= max; i++ {
		opts = append(opts, strconv.Itoa(i))
	}
	return opts
}

func (s *Schema) buildIndex() {
	s.fieldIndex = make(map[string]fieldMeta)
	for _, sec := range s.Sections {
		for _, f := range sec.Fields {
			switch f.Type {
			case "long_text":
				s.fieldIndex[f.ID] = fieldMeta{typ: "long_text", label: f.Label}
			case "scale":
				s.fieldIndex[f.ID] = fieldMeta{typ: "scale", min: f.Min, max: f.Max, label: f.Label}
			case "radio_group":
				for _, it := range f.Items {
					optSet := make(map[string]struct{})
					for _, o := range it.Options {
						optSet[strings.TrimSpace(o)] = struct{}{}
					}
					s.fieldIndex[it.ID] = fieldMeta{typ: "radio", options: optSet, label: it.Label}
				}
			}
		}
	}
}

// CollectAnswerValidationIssues lists every validation problem in schema order (sections → fields).
func (s *Schema) CollectAnswerValidationIssues(schemaVersion int, answers map[string]any, complete bool) []ValidationIssue {
	var issues []ValidationIssue
	if schemaVersion == 0 {
		schemaVersion = s.SchemaVersion
	}
	if schemaVersion != s.SchemaVersion {
		return []ValidationIssue{{
			FieldID:    "",
			Code:       "schema_version_mismatch",
			MessageEng: fmt.Sprintf("schema_version must be %d (got %d)", s.SchemaVersion, schemaVersion),
			MessageIdn: fmt.Sprintf("schema_version harus %d (diterima %d)", s.SchemaVersion, schemaVersion),
		}}
	}
	if answers == nil {
		answers = map[string]any{}
	}

	for _, sec := range s.Sections {
		for _, f := range sec.Fields {
			switch f.Type {
			case "long_text":
				issues = append(issues, s.validateLongTextField(f.ID, f.Label, answers[f.ID], complete)...)
			case "radio_group":
				for _, it := range f.Items {
					meta := s.fieldIndex[it.ID]
					issues = append(issues, s.validateRadioField(it.ID, meta.label, meta.options, answers[it.ID], complete)...)
				}
			case "scale":
				meta := s.fieldIndex[f.ID]
				issues = append(issues, s.validateScaleField(f.ID, meta.label, meta.min, meta.max, answers[f.ID], complete)...)
			}
		}
	}

	if complete {
		var extras []string
		for k := range answers {
			if _, ok := s.fieldIndex[k]; !ok {
				extras = append(extras, k)
			}
		}
		sort.Strings(extras)
		for _, k := range extras {
			issues = append(issues, ValidationIssue{
				FieldID:    k,
				Code:       "unknown_field",
				MessageEng: fmt.Sprintf("Unknown answer key %q (not in form schema)", k),
				MessageIdn: fmt.Sprintf("Kunci jawaban %q tidak dikenal (bukan bagian form)", k),
			})
		}
	}
	return issues
}

func (s *Schema) validateLongTextField(id, label string, v any, complete bool) []ValidationIssue {
	if !complete {
		if v == nil {
			return nil
		}
		if _, ok := v.(string); !ok {
			return []ValidationIssue{{
				FieldID:    id,
				Label:      label,
				Code:       "invalid_type",
				MessageEng: "Value must be a string",
				MessageIdn: "Nilai harus berupa teks (string)",
			}}
		}
		return nil
	}
	// complete: long_text optional — only validate type if present
	if v == nil {
		return nil
	}
	if _, ok := v.(string); !ok {
		return []ValidationIssue{{
			FieldID:    id,
			Label:      label,
			Code:       "invalid_type",
			MessageEng: "Value must be a string",
			MessageIdn: "Nilai harus berupa teks (string)",
		}}
	}
	return nil
}

func (s *Schema) validateRadioField(id, label string, options map[string]struct{}, v any, complete bool) []ValidationIssue {
	if !okPresent(v) {
		if !complete {
			return nil
		}
		return []ValidationIssue{{
			FieldID:    id,
			Label:      label,
			Code:       "required",
			MessageEng: "This question must be answered before submit",
			MessageIdn: "Pertanyaan ini wajib dijawab sebelum submit",
		}}
	}
	str, ok := v.(string)
	if !ok {
		return []ValidationIssue{{
			FieldID:    id,
			Label:      label,
			Code:       "invalid_type",
			MessageEng: "Value must be a string (e.g. Yes / No)",
			MessageIdn: "Nilai harus teks (mis. Yes / No)",
		}}
	}
	str = strings.TrimSpace(str)
	if _, ok := options[str]; !ok {
		return []ValidationIssue{{
			FieldID:    id,
			Label:      label,
			Code:       "invalid_option",
			MessageEng: fmt.Sprintf("Invalid choice %q; use one of the allowed options", str),
			MessageIdn: fmt.Sprintf("Pilihan %q tidak valid; gunakan salah satu opsi yang diizinkan", str),
		}}
	}
	return nil
}

func (s *Schema) validateScaleField(id, label string, min, max int, v any, complete bool) []ValidationIssue {
	if !okPresent(v) {
		if !complete {
			return nil
		}
		return []ValidationIssue{{
			FieldID:    id,
			Label:      label,
			Code:       "required",
			MessageEng: fmt.Sprintf("Rating required (%d–%d)", min, max),
			MessageIdn: fmt.Sprintf("Nilai skala wajib diisi (%d–%d)", min, max),
		}}
	}
	n, err := toInt(v)
	if err != nil {
		return []ValidationIssue{{
			FieldID:    id,
			Label:      label,
			Code:       "invalid_type",
			MessageEng: "Value must be a number",
			MessageIdn: "Nilai harus berupa angka",
		}}
	}
	if n < min || n > max {
		return []ValidationIssue{{
			FieldID:    id,
			Label:      label,
			Code:       "out_of_range",
			MessageEng: fmt.Sprintf("Must be between %d and %d (inclusive)", min, max),
			MessageIdn: fmt.Sprintf("Harus antara %d dan %d (inklusif)", min, max),
		}}
	}
	return nil
}

func okPresent(v any) bool {
	if v == nil {
		return false
	}
	if s, ok := v.(string); ok && strings.TrimSpace(s) == "" {
		return false
	}
	return true
}

// ValidateAnswers checks keys against schema. If complete=true, every indexed field must be present and valid.
func (s *Schema) ValidateAnswers(schemaVersion int, answers map[string]any, complete bool) error {
	issues := s.CollectAnswerValidationIssues(schemaVersion, answers, complete)
	if len(issues) == 0 {
		return nil
	}
	return fmt.Errorf("%s", issues[0].MessageEng)
}

func toInt(v any) (int, error) {
	switch x := v.(type) {
	case float64:
		return int(x), nil
	case int:
		return x, nil
	case int64:
		return int(x), nil
	case json.Number:
		i, err := x.Int64()
		return int(i), err
	case string:
		return strconv.Atoi(strings.TrimSpace(x))
	default:
		return 0, fmt.Errorf("expected number")
	}
}
