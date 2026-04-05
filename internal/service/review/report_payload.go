package review

import (
	"encoding/json"
	"fmt"

	"github.com/api-monolith-template/internal/constant"
	"github.com/api-monolith-template/internal/reviewerform"
)

func invalidReviewerReportWithIssues(issues []reviewerform.ValidationIssue) error {
	e := constant.ErrInvalidReviewerReport
	e.Data = map[string]any{"issues": issues}
	return e
}

const maxReviewerAnswerRunes = 100000

type reviewerStoredReport struct {
	SchemaVersion int            `json:"schema_version"`
	Answers       map[string]any `json:"answers"`
	Flags         map[string]any `json:"flags,omitempty"`
}

func parseReviewerStoredPayload(b []byte) (reviewerStoredReport, error) {
	var p reviewerStoredReport
	if len(b) == 0 {
		return reviewerStoredReport{Answers: map[string]any{}, Flags: map[string]any{}}, nil
	}
	if err := json.Unmarshal(b, &p); err != nil {
		return reviewerStoredReport{}, err
	}
	if p.Answers == nil {
		p.Answers = map[string]any{}
	}
	if p.Flags == nil {
		p.Flags = map[string]any{}
	}
	return p, nil
}

func marshalReviewerPayload(p reviewerStoredReport) ([]byte, error) {
	return json.Marshal(p)
}

func mergeReviewerPayload(existing []byte, schemaVersion int, patchAnswers, patchFlags map[string]any, defaultSchema int) (reviewerStoredReport, error) {
	base, err := parseReviewerStoredPayload(existing)
	if err != nil {
		return reviewerStoredReport{}, err
	}
	sv := base.SchemaVersion
	if sv == 0 {
		sv = defaultSchema
	}
	if schemaVersion != 0 {
		sv = schemaVersion
	}
	if sv == 0 {
		sv = defaultSchema
	}
	out := reviewerStoredReport{
		SchemaVersion: sv,
		Answers:       cloneAnyMap(base.Answers),
		Flags:         cloneAnyMap(base.Flags),
	}
	if patchAnswers != nil {
		for k, v := range patchAnswers {
			out.Answers[k] = v
		}
	}
	if patchFlags != nil {
		for k, v := range patchFlags {
			out.Flags[k] = v
		}
	}
	return out, nil
}

func cloneAnyMap(in map[string]any) map[string]any {
	out := make(map[string]any, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

func validateReviewerAnswerSizes(answers map[string]any) error {
	for k, v := range answers {
		s, ok := v.(string)
		if !ok {
			continue
		}
		if len([]rune(s)) > maxReviewerAnswerRunes {
			return fmt.Errorf("answer %q exceeds maximum length", k)
		}
	}
	return nil
}
