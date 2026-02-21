package email

import (
	"bytes"
	"html/template"
	"time"
)

var (
	verifyPINTmpl = template.Must(template.New("verifyPIN").Parse(`<!doctype html><html><body style="font-family:Arial,Helvetica,sans-serif;background:#f6f9fc;padding:24px">
  <div style="max-width:560px;margin:0 auto;background:#fff;border:1px solid #e6ecf1;border-radius:12px;padding:24px">
    <h2 style="margin:0 0 8px 0;color:#111">MedikaOne</h2>
    <p style="color:#555">Halo {{.FirstName}}, terima kasih telah mendaftar!</p>
    <p style="color:#555">Gunakan PIN berikut untuk verifikasi email Anda (berlaku {{.TTLMinutes}} menit):</p>
    <div style="text-align:center;margin:16px 0">
      <span style="display:inline-block;font-size:28px;letter-spacing:6px;font-weight:700;background:#0ea5e9;color:#fff;padding:12px 16px;border-radius:10px">{{.PIN}}</span>
    </div>
    <p style="color:#667">Jika Anda tidak meminta verifikasi ini, abaikan email ini.</p>
    <p style="color:#999;font-size:12px;margin-top:16px;padding-top:16px;border-top:1px solid #e6ecf1">💡 Tips: Jika email ini masuk ke folder Spam, tandai sebagai "Not Spam" agar email berikutnya masuk ke Inbox.</p>
  </div>
  <div style="text-align:center;color:#99a;font-size:12px;margin-top:10px">© {{.Year}} MedikaOne. Semua hak dilindungi.</div>
</body></html>`))

	resetPINTmpl = template.Must(template.New("resetPIN").Parse(`<!doctype html><html><body style="font-family:Arial,Helvetica,sans-serif;background:#f6f9fc;padding:24px">
  <div style="max-width:560px;margin:0 auto;background:#fff;border:1px solid #e6ecf1;border-radius:12px;padding:24px">
    <h2 style="margin:0 0 8px 0;color:#111">MedikaOne</h2>
    <p style="color:#555">Halo {{.FirstName}}, berikut PIN reset password (berlaku {{.TTLMinutes}} menit):</p>
    <div style="text-align:center;margin:16px 0">
      <span style="display:inline-block;font-size:28px;letter-spacing:6px;font-weight:700;background:#0ea5e9;color:#fff;padding:12px 16px;border-radius:10px">{{.PIN}}</span>
    </div>
    <p style="color:#667">Jika kamu tidak meminta reset ini, abaikan email ini.</p>
  </div>
  <div style="text-align:center;color:#99a; font-size:12px;margin-top:10px">© {{.Year}} MedikaOne</div>
</body></html>`))
)

type emailData struct {
	FirstName  string
	PIN        string
	TTLMinutes int
	Year       int
}

// RenderVerifyPIN menghasilkan HTML untuk email verifikasi PIN.
func RenderVerifyPIN(firstName, pin string, ttlMinutes int) string {
	if firstName == "" {
		firstName = "Pengguna"
	}
	data := emailData{
		FirstName:  firstName,
		PIN:        pin,
		TTLMinutes: ttlMinutes,
		Year:       YearNow(),
	}
	var buf bytes.Buffer
	if err := verifyPINTmpl.Execute(&buf, data); err != nil {
		return "" // fallback empty or log error if needed
	}
	return buf.String()
}

// RenderResetPIN menghasilkan HTML untuk email reset password (PIN).
func RenderResetPIN(firstName, pin string, ttlMinutes int) string {
	if firstName == "" {
		firstName = "Pengguna"
	}
	data := emailData{
		FirstName:  firstName,
		PIN:        pin,
		TTLMinutes: ttlMinutes,
		Year:       YearNow(),
	}
	var buf bytes.Buffer
	if err := resetPINTmpl.Execute(&buf, data); err != nil {
		return ""
	}
	return buf.String()
}

// YearNow disediakan agar mudah ditest/mocking kalau perlu.
func YearNow() int { return time.Now().Year() }

// ====== Review Workflow Email Templates ======

var (
	editorAssignedTmpl = template.Must(template.New("editorAssigned").Parse(`<!doctype html><html><body style="font-family:Arial,Helvetica,sans-serif;background:#f6f9fc;padding:24px">
  <div style="max-width:560px;margin:0 auto;background:#fff;border:1px solid #e6ecf1;border-radius:12px;padding:24px">
    <h2 style="margin:0 0 8px 0;color:#111">MedikaOne Journal</h2>
    <p style="color:#555">Halo {{.EditorName}},</p>
    <p style="color:#555">Anda telah ditugaskan oleh <strong>{{.ChiefEditorName}}</strong> sebagai editor untuk manuskrip berikut:</p>
    <div style="background:#f0f7ff;border:1px solid #cce0ff;border-radius:8px;padding:16px;margin:16px 0">
      <p style="margin:0;font-weight:700;color:#333">{{.ManuscriptTitle}}</p>
    </div>
    <p style="color:#555">Silakan login ke sistem untuk meninjau dan mengelola proses review.</p>
    <p style="color:#667">Jika Anda memiliki pertanyaan, hubungi Chief Editor.</p>
  </div>
  <div style="text-align:center;color:#99a;font-size:12px;margin-top:10px">© {{.Year}} MedikaOne. Semua hak dilindungi.</div>
</body></html>`))

	reviewerInvitationTmpl = template.Must(template.New("reviewerInvitation").Parse(`<!doctype html><html><body style="font-family:Arial,Helvetica,sans-serif;background:#f6f9fc;padding:24px">
  <div style="max-width:560px;margin:0 auto;background:#fff;border:1px solid #e6ecf1;border-radius:12px;padding:24px">
    <h2 style="margin:0 0 8px 0;color:#111">MedikaOne Journal</h2>
    <p style="color:#555">Halo {{.ReviewerName}},</p>
    <p style="color:#555">Anda diundang oleh editor <strong>{{.EditorName}}</strong> untuk me-review manuskrip berikut:</p>
    <div style="background:#f0f7ff;border:1px solid #cce0ff;border-radius:8px;padding:16px;margin:16px 0">
      <p style="margin:0;font-weight:700;color:#333">{{.ManuscriptTitle}}</p>
      <p style="margin:8px 0 0 0;color:#555;font-size:14px">Batas waktu review: <strong>{{.DueDate}}</strong></p>
    </div>
    <p style="color:#555">Undangan ini berlaku selama <strong>7 hari</strong>. Silakan terima atau tolak undangan melalui link berikut:</p>
    <div style="text-align:center;margin:20px 0">
      <a href="{{.AcceptURL}}" style="display:inline-block;background:#0ea5e9;color:#fff;padding:12px 24px;border-radius:8px;text-decoration:none;font-weight:600;margin-right:8px">Terima Undangan</a>
      <a href="{{.DeclineURL}}" style="display:inline-block;background:#ef4444;color:#fff;padding:12px 24px;border-radius:8px;text-decoration:none;font-weight:600">Tolak Undangan</a>
    </div>
    <p style="color:#999;font-size:12px;margin-top:16px;padding-top:16px;border-top:1px solid #e6ecf1">Jika Anda tidak me-respond undangan ini dalam 7 hari, undangan akan otomatis kedaluwarsa.</p>
  </div>
  <div style="text-align:center;color:#99a;font-size:12px;margin-top:10px">© {{.Year}} MedikaOne. Semua hak dilindungi.</div>
</body></html>`))

	submissionDecisionTmpl = template.Must(template.New("submissionDecision").Parse(`<!doctype html><html><body style="font-family:Arial,Helvetica,sans-serif;background:#f6f9fc;padding:24px">
  <div style="max-width:560px;margin:0 auto;background:#fff;border:1px solid #e6ecf1;border-radius:12px;padding:24px">
    <h2 style="margin:0 0 8px 0;color:#111">MedikaOne Journal</h2>
    <p style="color:#555">Halo {{.AuthorName}},</p>
    <p style="color:#555">Kami ingin menginformasikan keputusan editorial untuk manuskrip Anda:</p>
    <div style="background:#f0f7ff;border:1px solid #cce0ff;border-radius:8px;padding:16px;margin:16px 0">
      <p style="margin:0;font-weight:700;color:#333">{{.ManuscriptTitle}}</p>
      <p style="margin:8px 0 0 0;font-size:16px;font-weight:700;color:{{.DecisionColor}}">Keputusan: {{.Decision}}</p>
    </div>
    {{if .Comments}}<div style="background:#fafafa;border:1px solid #e6ecf1;border-radius:8px;padding:16px;margin:16px 0">
      <p style="margin:0;color:#555;font-weight:600">Komentar Editor:</p>
      <p style="margin:8px 0 0 0;color:#555">{{.Comments}}</p>
    </div>{{end}}
    <p style="color:#555">Silakan login ke sistem untuk detail lebih lanjut.</p>
  </div>
  <div style="text-align:center;color:#99a;font-size:12px;margin-top:10px">© {{.Year}} MedikaOne. Semua hak dilindungi.</div>
</body></html>`))

	reviewResultToAuthorTmpl = template.Must(template.New("reviewResultToAuthor").Parse(`<!doctype html><html><body style="font-family:Arial,Helvetica,sans-serif;background:#f6f9fc;padding:24px">
  <div style="max-width:560px;margin:0 auto;background:#fff;border:1px solid #e6ecf1;border-radius:12px;padding:24px">
    <h2 style="margin:0 0 8px 0;color:#111">MedikaOne Journal</h2>
    <p style="color:#555">Halo {{.AuthorName}},</p>
    <p style="color:#555">Hasil review untuk manuskrip Anda telah tersedia:</p>
    <div style="background:#f0f7ff;border:1px solid #cce0ff;border-radius:8px;padding:16px;margin:16px 0">
      <p style="margin:0;font-weight:700;color:#333">{{.ManuscriptTitle}}</p>
      <p style="margin:8px 0 0 0;color:#555;font-size:14px">Review Round: <strong>{{.RoundNumber}}</strong></p>
    </div>
    {{if .ReviewSummary}}<div style="background:#fafafa;border:1px solid #e6ecf1;border-radius:8px;padding:16px;margin:16px 0">
      <p style="margin:0;color:#555;font-weight:600">Ringkasan Review:</p>
      <p style="margin:8px 0 0 0;color:#555">{{.ReviewSummary}}</p>
    </div>{{end}}
    <p style="color:#555">Silakan login ke sistem untuk melihat detail review lengkap dan mengambil langkah selanjutnya.</p>
  </div>
  <div style="text-align:center;color:#99a;font-size:12px;margin-top:10px">© {{.Year}} MedikaOne. Semua hak dilindungi.</div>
</body></html>`))
)

// RenderEditorAssigned generates HTML for editor assignment notification.
func RenderEditorAssigned(editorName, manuscriptTitle, chiefEditorName string) string {
	data := struct {
		EditorName      string
		ManuscriptTitle string
		ChiefEditorName string
		Year            int
	}{editorName, manuscriptTitle, chiefEditorName, YearNow()}
	var buf bytes.Buffer
	if err := editorAssignedTmpl.Execute(&buf, data); err != nil {
		return ""
	}
	return buf.String()
}

// RenderReviewerInvitation generates HTML for reviewer invitation email.
func RenderReviewerInvitation(reviewerName, manuscriptTitle, editorName, dueDate, acceptURL, declineURL string) string {
	data := struct {
		ReviewerName    string
		ManuscriptTitle string
		EditorName      string
		DueDate         string
		AcceptURL       string
		DeclineURL      string
		Year            int
	}{reviewerName, manuscriptTitle, editorName, dueDate, acceptURL, declineURL, YearNow()}
	var buf bytes.Buffer
	if err := reviewerInvitationTmpl.Execute(&buf, data); err != nil {
		return ""
	}
	return buf.String()
}

// RenderSubmissionDecision generates HTML for manuscript decision notification.
func RenderSubmissionDecision(authorName, manuscriptTitle, decision, comments string) string {
	decisionColor := "#333"
	switch decision {
	case "ACCEPTED":
		decisionColor = "#16a34a"
	case "REJECTED":
		decisionColor = "#dc2626"
	case "REVISION_REQUIRED":
		decisionColor = "#d97706"
	}
	data := struct {
		AuthorName      string
		ManuscriptTitle string
		Decision        string
		DecisionColor   string
		Comments        string
		Year            int
	}{authorName, manuscriptTitle, decision, decisionColor, comments, YearNow()}
	var buf bytes.Buffer
	if err := submissionDecisionTmpl.Execute(&buf, data); err != nil {
		return ""
	}
	return buf.String()
}

// RenderReviewResultToAuthor generates HTML for forwarding review results to author.
func RenderReviewResultToAuthor(authorName, manuscriptTitle string, roundNumber int, reviewSummary string) string {
	data := struct {
		AuthorName      string
		ManuscriptTitle string
		RoundNumber     int
		ReviewSummary   string
		Year            int
	}{authorName, manuscriptTitle, roundNumber, reviewSummary, YearNow()}
	var buf bytes.Buffer
	if err := reviewResultToAuthorTmpl.Execute(&buf, data); err != nil {
		return ""
	}
	return buf.String()
}
