package email

import (
	"bytes"
	"fmt"
	"html/template"
	"strings"
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
    <p style="color:#999;font-size:12px;margin-top:16px;padding-top:16px;border-top:1px solid #e6ecf1">Jika email ini masuk folder Spam, tandai sebagai bukan spam agar pesan berikutnya masuk ke kotak masuk.</p>
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
    <p style="color:#667">Jika Anda tidak meminta reset ini, abaikan email ini.</p>
  </div>
  <div style="text-align:center;color:#99a;font-size:12px;margin-top:10px">© {{.Year}} MedikaOne. Semua hak dilindungi.</div>
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
		return ""
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

// YearNow returns the current calendar year for email footers.
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

	reviewerInvitationTmpl = template.Must(template.New("reviewerInvitation").Parse(`<!DOCTYPE html>
<html lang="id">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width,initial-scale=1.0">
<meta http-equiv="X-UA-Compatible" content="IE=edge">
<title>Undangan menjadi reviewer</title>
</head>
<body style="margin:0;padding:0;background:#e8eef4;font-family:'Segoe UI',Roboto,'Helvetica Neue',Arial,sans-serif;-webkit-font-smoothing:antialiased;">
<span style="display:none!important;visibility:hidden;mso-hide:all;font-size:1px;line-height:1px;max-height:0;max-width:0;opacity:0;overflow:hidden;">{{.Preheader}}</span>
<table role="presentation" width="100%" cellspacing="0" cellpadding="0" style="background:#e8eef4;padding:28px 12px;">
  <tr>
    <td align="center">
      <table role="presentation" width="100%" cellspacing="0" cellpadding="0" style="max-width:560px;background:#ffffff;border-radius:16px;overflow:hidden;box-shadow:0 8px 30px rgba(15,23,42,0.08);border:1px solid #dbe4ee;">
        <tr>
          <td style="background:#0369a1;padding:22px 26px;">
            <p style="margin:0;font-size:11px;letter-spacing:0.14em;text-transform:uppercase;color:rgba(255,255,255,0.88);font-weight:600;">MedikaOne Journal</p>
            <h1 style="margin:10px 0 0 0;font-size:22px;font-weight:700;color:#ffffff;line-height:1.35;">Undangan menjadi reviewer</h1>
          </td>
        </tr>
        <tr>
          <td style="padding:26px 26px 8px 26px;">
            <p style="margin:0 0 14px 0;font-size:16px;line-height:1.65;color:#334155;">Halo <strong style="color:#0f172a;">{{.ReviewerName}}</strong>,</p>
            <p style="margin:0 0 20px 0;font-size:15px;line-height:1.65;color:#475569;">Anda diundang oleh editor <strong style="color:#0f172a;">{{.EditorName}}</strong> untuk meninjau manuskrip berikut.</p>
            <table role="presentation" width="100%" cellspacing="0" cellpadding="0" style="background:#f0f7ff;border:1px solid #bfdbfe;border-radius:12px;margin:0 0 22px 0;">
              <tr>
                <td style="padding:18px 20px;">
                  <p style="margin:0;font-size:17px;font-weight:700;color:#0f172a;line-height:1.45;">{{.ManuscriptTitle}}</p>
                  <p style="margin:10px 0 0 0;font-size:14px;color:#475569;">Batas waktu review: <strong style="color:#0369a1;">{{.DueDate}}</strong></p>
                </td>
              </tr>
            </table>
            <p style="margin:0 0 14px 0;font-size:15px;line-height:1.65;color:#475569;">Undangan berlaku <strong>7 hari</strong>. Gunakan tautan langsung dari sistem berikut (salin ke peramban jika perlu):</p>
            <table role="presentation" cellspacing="0" cellpadding="0" style="margin:18px 0 18px 0;">
              <tr>
                <td align="center" style="border-radius:10px;" bgcolor="#0369a1">
                  <a href="{{.AcceptURL}}" style="display:inline-block;padding:12px 18px;font-size:14px;font-weight:700;color:#ffffff;text-decoration:none;border-radius:10px;background:#0369a1;">Terima undangan</a>
                </td>
                <td style="width:10px"></td>
                <td align="center" style="border-radius:10px;" bgcolor="#b91c1c">
                  <a href="{{.DeclineURL}}" style="display:inline-block;padding:12px 18px;font-size:14px;font-weight:700;color:#ffffff;text-decoration:none;border-radius:10px;background:#b91c1c;">Tolak undangan</a>
                </td>
              </tr>
            </table>
            <p style="margin:0 0 6px 0;font-size:12px;font-weight:600;color:#334155;">Tautan (salin ke peramban jika tombol tidak berfungsi)</p>
            <p style="margin:0 0 10px 0;font-size:13px;line-height:1.55;word-break:break-all;"><a href="{{.AcceptURL}}" style="color:#0369a1">{{.AcceptURL}} (Accept)</a></p>
            <p style="margin:0 0 0 0;font-size:13px;line-height:1.55;word-break:break-all;"><a href="{{.DeclineURL}}" style="color:#b91c1c">{{.DeclineURL}} (Decline)</a></p>
            <p style="margin:22px 0 0 0;padding-top:18px;border-top:1px solid #e2e8f0;font-size:12px;line-height:1.55;color:#94a3b8;">Jika Anda tidak merespons dalam 7 hari, undangan akan kedaluwarsa. Jangan bagikan tautan ini kepada orang lain.</p>
          </td>
        </tr>
        <tr>
          <td style="background:#f8fafc;padding:14px 26px;text-align:center;border-top:1px solid #e2e8f0;">
            <p style="margin:0;font-size:11px;color:#94a3b8;">© {{.Year}} MedikaOne. Semua hak dilindungi.</p>
          </td>
        </tr>
      </table>
    </td>
  </tr>
</table>
</body>
</html>`))

	registeredReviewerAssignmentTmpl = template.Must(template.New("registeredReviewerAssignment").Parse(`<!DOCTYPE html>
<html lang="id">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width,initial-scale=1.0">
<title>Penugasan review baru</title>
</head>
<body style="margin:0;padding:0;background:#e8eef4;font-family:'Segoe UI',Roboto,'Helvetica Neue',Arial,sans-serif;">
<table role="presentation" width="100%" cellspacing="0" cellpadding="0" style="background:#e8eef4;padding:28px 12px;">
  <tr>
    <td align="center">
      <table role="presentation" width="100%" cellspacing="0" cellpadding="0" style="max-width:560px;background:#ffffff;border-radius:16px;overflow:hidden;box-shadow:0 8px 30px rgba(15,23,42,0.08);border:1px solid #dbe4ee;">
        <tr>
          <td style="background:#0369a1;padding:22px 26px;">
            <p style="margin:0;font-size:11px;letter-spacing:0.14em;text-transform:uppercase;color:rgba(255,255,255,0.88);font-weight:600;">MedikaOne Journal</p>
            <h1 style="margin:10px 0 0 0;font-size:22px;font-weight:700;color:#ffffff;line-height:1.35;">Penugasan review baru</h1>
          </td>
        </tr>
        <tr>
          <td style="padding:26px 26px 8px 26px;">
            <p style="margin:0 0 14px 0;font-size:16px;line-height:1.65;color:#334155;">Halo <strong style="color:#0f172a;">{{.ReviewerName}}</strong>,</p>
            <p style="margin:0 0 20px 0;font-size:15px;line-height:1.65;color:#475569;"><strong style="color:#0f172a;">{{.EditorName}}</strong> menugaskan Anda sebagai reviewer untuk manuskrip berikut. Silakan login ke portal reviewer untuk memulai.</p>
            <table role="presentation" width="100%" cellspacing="0" cellpadding="0" style="background:#f0f7ff;border:1px solid #bfdbfe;border-radius:12px;margin:0 0 22px 0;">
              <tr>
                <td style="padding:18px 20px;">
                  <p style="margin:0;font-size:17px;font-weight:700;color:#0f172a;line-height:1.45;">{{.ManuscriptTitle}}</p>
                  <p style="margin:10px 0 0 0;font-size:14px;color:#475569;">Batas waktu review: <strong style="color:#0369a1;">{{.DueDate}}</strong></p>
                </td>
              </tr>
            </table>
            <table role="presentation" cellspacing="0" cellpadding="0" style="margin:18px 0 18px 0;">
              <tr>
                <td align="center" style="border-radius:10px;" bgcolor="#0369a1">
                  <a href="{{.PortalURL}}" style="display:inline-block;padding:12px 18px;font-size:14px;font-weight:700;color:#ffffff;text-decoration:none;border-radius:10px;background:#0369a1;">Buka portal reviewer</a>
                </td>
              </tr>
            </table>
            <p style="margin:0 0 6px 0;font-size:12px;font-weight:600;color:#334155;">Tautan</p>
            <p style="margin:0;font-size:13px;line-height:1.55;word-break:break-all;"><a href="{{.PortalURL}}" style="color:#0369a1">{{.PortalURL}}</a></p>
          </td>
        </tr>
        <tr>
          <td style="background:#f8fafc;padding:14px 26px;text-align:center;border-top:1px solid #e2e8f0;">
            <p style="margin:0;font-size:11px;color:#94a3b8;">© {{.Year}} MedikaOne. Semua hak dilindungi.</p>
          </td>
        </tr>
      </table>
    </td>
  </tr>
</table>
</body>
</html>`))

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
      <p style="margin:8px 0 0 0;color:#555;font-size:14px">Putaran review: <strong>{{.RoundNumber}}</strong></p>
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
	pre := "Undangan menjadi reviewer dari " + editorName
	if t := strings.TrimSpace(manuscriptTitle); t != "" {
		pre = "Undangan review manuskrip: " + t
	}
	if len(pre) > 140 {
		pre = pre[:137] + "..."
	}
	data := struct {
		ReviewerName    string
		ManuscriptTitle string
		EditorName      string
		DueDate         string
		AcceptURL       string
		DeclineURL      string
		Year            int
		Preheader       string
	}{reviewerName, manuscriptTitle, editorName, dueDate, acceptURL, declineURL, YearNow(), pre}
	var buf bytes.Buffer
	if err := reviewerInvitationTmpl.Execute(&buf, data); err != nil {
		return ""
	}
	return buf.String()
}

// RenderRegisteredReviewerAssignment notifies an existing reviewer account (no token flow).
func RenderRegisteredReviewerAssignment(reviewerName, manuscriptTitle, editorName, dueDate, portalURL string) string {
	data := struct {
		ReviewerName    string
		ManuscriptTitle string
		EditorName      string
		DueDate         string
		PortalURL       string
		Year            int
	}{reviewerName, manuscriptTitle, editorName, dueDate, portalURL, YearNow()}
	var buf bytes.Buffer
	if err := registeredReviewerAssignmentTmpl.Execute(&buf, data); err != nil {
		return ""
	}
	return buf.String()
}

// RenderRegisteredReviewerAssignmentPlain is the plain-text part for transactional email APIs.
func RenderRegisteredReviewerAssignmentPlain(reviewerName, manuscriptTitle, editorName, dueDate, portalURL string) string {
	var b strings.Builder
	b.WriteString("MedikaOne Journal - Penugasan review\n\n")
	b.WriteString("Halo ")
	b.WriteString(reviewerName)
	b.WriteString(",\n\n")
	b.WriteString(editorName)
	b.WriteString(" menugaskan Anda sebagai reviewer untuk manuskrip:\n\n")
	if strings.TrimSpace(manuscriptTitle) != "" {
		b.WriteString(manuscriptTitle)
	} else {
		b.WriteString("(lihat portal)")
	}
	b.WriteString("\nBatas waktu review: ")
	b.WriteString(dueDate)
	b.WriteString("\n\nBuka portal reviewer:\n")
	b.WriteString(portalURL)
	b.WriteString("\n\n---\n© ")
	b.WriteString(fmt.Sprintf("%d", YearNow()))
	b.WriteString(" MedikaOne.\n")
	return b.String()
}

// RenderReviewerInvitationPlain is the plain-text alternative for transactional APIs and accessibility.
func RenderReviewerInvitationPlain(reviewerName, manuscriptTitle, editorName, dueDate, acceptURL, declineURL string) string {
	var b strings.Builder
	b.WriteString("MedikaOne Journal - Undangan menjadi reviewer\n\n")
	b.WriteString("Halo ")
	b.WriteString(reviewerName)
	b.WriteString(",\n\n")
	b.WriteString("Anda diundang oleh editor ")
	b.WriteString(editorName)
	b.WriteString(" untuk meninjau manuskrip berikut:\n\n")
	b.WriteString("Judul: ")
	if strings.TrimSpace(manuscriptTitle) != "" {
		b.WriteString(manuscriptTitle)
	} else {
		b.WriteString("(lihat portal)")
	}
	b.WriteString("\nBatas waktu review: ")
	b.WriteString(dueDate)
	b.WriteString("\n\nUndangan berlaku 7 hari.\n\n")
	b.WriteString("Terima undangan:\n")
	b.WriteString(acceptURL)
	b.WriteString("\n\nTolak undangan:\n")
	b.WriteString(declineURL)
	b.WriteString("\n\n---\n© ")
	b.WriteString(fmt.Sprintf("%d", YearNow()))
	b.WriteString(" MedikaOne.\n")
	return b.String()
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

// ====== Generic Workflow Notification (shared card layout) ======

var simpleNotificationTmpl = template.Must(template.New("simpleNotification").Parse(`<!doctype html><html><body style="font-family:Arial,Helvetica,sans-serif;background:#f6f9fc;padding:24px">
  <div style="max-width:560px;margin:0 auto;background:#fff;border:1px solid #e6ecf1;border-radius:12px;padding:24px">
    <h2 style="margin:0 0 8px 0;color:#111">MedikaOne Journal</h2>
    <p style="color:#555">Halo {{.RecipientName}},</p>
    <p style="color:#555">{{.IntroLine}}</p>
    <div style="background:#f0f7ff;border:1px solid #cce0ff;border-radius:8px;padding:16px;margin:16px 0">
      <p style="margin:0;font-weight:700;color:#333">{{.ManuscriptTitle}}</p>
      {{if .DetailLine}}<p style="margin:8px 0 0 0;color:#555;font-size:14px">{{.DetailLine}}</p>{{end}}
    </div>
    <p style="color:#555">Silakan login ke sistem untuk detail lebih lanjut.</p>
  </div>
  <div style="text-align:center;color:#99a;font-size:12px;margin-top:10px">© {{.Year}} MedikaOne. Semua hak dilindungi.</div>
</body></html>`))

func renderSimpleNotification(recipientName, introLine, manuscriptTitle, detailLine string) string {
	if strings.TrimSpace(recipientName) == "" {
		recipientName = "Pengguna"
	}
	data := struct {
		RecipientName   string
		IntroLine       string
		ManuscriptTitle string
		DetailLine      string
		Year            int
	}{recipientName, introLine, manuscriptTitle, detailLine, YearNow()}
	var buf bytes.Buffer
	if err := simpleNotificationTmpl.Execute(&buf, data); err != nil {
		return ""
	}
	return buf.String()
}

// RenderManuscriptSubmitted confirms receipt of a new submission to the author.
func RenderManuscriptSubmitted(authorName, manuscriptTitle string) string {
	return renderSimpleNotification(authorName,
		"Manuskrip Anda telah berhasil disubmit dan akan segera diperiksa oleh tim editorial kami.",
		manuscriptTitle, "")
}

// RenderManuscriptInProduction notifies the author that their accepted manuscript entered production.
func RenderManuscriptInProduction(authorName, manuscriptTitle string) string {
	return renderSimpleNotification(authorName,
		"Manuskrip Anda telah memasuki tahap produksi (copyediting & layout).",
		manuscriptTitle, "")
}

// RenderManuscriptPublished notifies the author that their manuscript has been published.
func RenderManuscriptPublished(authorName, manuscriptTitle, issueLabel string) string {
	return renderSimpleNotification(authorName,
		"Selamat! Manuskrip Anda telah dipublikasikan.",
		manuscriptTitle, issueLabel)
}

// RenderNewSubmissionAlert notifies Chief Editor/Super Admin that a new manuscript needs editor assignment.
func RenderNewSubmissionAlert(recipientName, manuscriptTitle, authorName string) string {
	return renderSimpleNotification(recipientName,
		"Manuskrip baru telah disubmit dan menunggu penugasan editor.",
		manuscriptTitle, "Penulis: "+authorName)
}

// RenderReviewerAccepted notifies the editor that a reviewer accepted their invitation.
func RenderReviewerAccepted(editorName, reviewerName, manuscriptTitle string) string {
	return renderSimpleNotification(editorName,
		reviewerName+" telah menerima undangan review untuk manuskrip berikut.",
		manuscriptTitle, "")
}

// RenderReviewerDeclined notifies the editor that a reviewer declined their invitation.
func RenderReviewerDeclined(editorName, reviewerName, manuscriptTitle, reason string) string {
	detail := ""
	if strings.TrimSpace(reason) != "" {
		detail = "Alasan: " + reason
	}
	return renderSimpleNotification(editorName,
		reviewerName+" menolak undangan review untuk manuskrip berikut.",
		manuscriptTitle, detail)
}

// RenderReviewerReportSubmitted notifies the editor that a reviewer submitted their report.
func RenderReviewerReportSubmitted(editorName, reviewerName, manuscriptTitle string) string {
	return renderSimpleNotification(editorName,
		reviewerName+" telah menyerahkan laporan review untuk manuskrip berikut. Silakan tinjau dan buat keputusan editorial.",
		manuscriptTitle, "")
}

// RenderReviewerWithdrawn notifies the editor that a reviewer withdrew from an assignment.
func RenderReviewerWithdrawn(editorName, reviewerName, manuscriptTitle string) string {
	return renderSimpleNotification(editorName,
		reviewerName+" mengundurkan diri dari penugasan review untuk manuskrip berikut.",
		manuscriptTitle, "")
}

// RenderExtensionRequested notifies the editor that a reviewer requested a due-date extension.
func RenderExtensionRequested(editorName, reviewerName, manuscriptTitle, requestedDue string) string {
	return renderSimpleNotification(editorName,
		reviewerName+" mengajukan perpanjangan batas waktu review untuk manuskrip berikut.",
		manuscriptTitle, "Batas waktu yang diminta: "+requestedDue)
}

// RenderExtensionDecision notifies the reviewer of the editor's decision on their extension request.
func RenderExtensionDecision(reviewerName, manuscriptTitle string, approved bool, dueDate string) string {
	intro := "Permintaan perpanjangan batas waktu review Anda telah ditolak."
	detail := ""
	if approved {
		intro = "Permintaan perpanjangan batas waktu review Anda telah disetujui."
		detail = "Batas waktu baru: " + dueDate
	}
	return renderSimpleNotification(reviewerName, intro, manuscriptTitle, detail)
}
