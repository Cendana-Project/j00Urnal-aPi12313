package seeder

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"time"

	"gorm.io/gorm"

	"github.com/api-monolith-template/internal/constant"
	"github.com/api-monolith-template/internal/model/entity"
	"github.com/api-monolith-template/internal/service/storage"
)

func SeedSampleManuscripts(db *gorm.DB) error {
	// 1. Get user IDs
	chiefEditorID := "550e8400-e29b-41d4-a716-446655440030"
	editorID := "550e8400-e29b-41d4-a716-446655440020"
	authorID := "550e8400-e29b-41d4-a716-446655440040"
	// reviewerID := "550e8400-e29b-41d4-a716-446655440050"

	// 2. Create Journal
	journalID := "990e8400-e29b-41d4-a716-446655440001"
	journal := entity.Journal{
		ID:          journalID,
		Name:        "Journal of Advanced Technology",
		Description: "A sample journal for testing",
		Status:      constant.PublicationStatusPublished,
		CreatedBy:   chiefEditorID,
		CreatedAt:   time.Now(),
	}
	if err := db.FirstOrCreate(&journal, entity.Journal{ID: journalID}).Error; err != nil {
		return fmt.Errorf("seed journal: %w", err)
	}

	// 3. Create Volume
	volID := "990e8400-e29b-41d4-a716-446655440002"
	volume := entity.Volume{
		ID:        volID,
		JournalID: journalID,
		Year:      time.Now().Year(),
		Number:    1,
		Status:    constant.PublicationStatusPublished,
		CreatedAt: time.Now(),
	}
	if err := db.FirstOrCreate(&volume, entity.Volume{ID: volID}).Error; err != nil {
		return fmt.Errorf("seed volume: %w", err)
	}

	// 4. Create Issue
	issueID := "990e8400-e29b-41d4-a716-446655440003"
	issue := entity.Issue{
		ID:              issueID,
		VolumeID:        volID,
		Number:          1,
		PublicationDate: time.Now(),
		Status:          constant.PublicationStatusPublished,
		CreatedAt:       time.Now(),
	}
	if err := db.FirstOrCreate(&issue, entity.Issue{ID: issueID}).Error; err != nil {
		return fmt.Errorf("seed issue: %w", err)
	}

	// 5. Create Manuscripts
	manuscripts := []entity.Manuscript{
		{
			ID:            "880e8400-e29b-41d4-a716-446655440001",
			JournalID:     journalID,
			Title:         "The Future of Artificial Intelligence",
			Abstract:      "This paper discusses the future impact of AI on various industries.",
			Status:        constant.ManuscriptStatusSubmitted,
			IsTncAccepted: true,
			CreatedAt:     time.Now(),
			PublishedAt:   time.Now(), // Need non-null default
		},
		{
			ID:               "880e8400-e29b-41d4-a716-446655440002",
			JournalID:        journalID,
			Title:            "Quantum Computing Revisited",
			Abstract:         "An overview of recent advancements in quantum computing hardware.",
			Status:           constant.ManuscriptStatusUnderReview,
			AssignedEditorID: &editorID,
			IsTncAccepted:    true,
			CreatedAt:        time.Now(),
			PublishedAt:      time.Now(),
		},
		{
			ID:               "880e8400-e29b-41d4-a716-446655440003",
			JournalID:        journalID,
			IssueID:          &issueID,
			Title:            "Machine Learning in Healthcare",
			Abstract:         "A study on predictive models for patient readmission.",
			Status:           constant.ManuscriptStatusPublished,
			AssignedEditorID: &editorID,
			IsTncAccepted:    true,
			CreatedAt:        time.Now(),
			PublishedAt:      time.Now(),
		},
	}

	for i, m := range manuscripts {
		aid := authorID
		m.MainAuthorID = &aid
		m.SubmittedByUserID = &aid
		if err := db.FirstOrCreate(&m, entity.Manuscript{ID: m.ID}).Error; err != nil {
			return fmt.Errorf("seed manuscript idx %d: %w", i, err)
		}
	}

	// Seed sample manuscript files (PDFs)
	if err := SeedSampleFiles(db); err != nil {
		return fmt.Errorf("seed sample files: %w", err)
	}

	return nil
}

// SeedSampleFiles uploads sample PDF files to storage and creates ManuscriptFile records.
func SeedSampleFiles(db *gorm.DB) error {
	journalID := "990e8400-e29b-41d4-a716-446655440001"
	issueID := "990e8400-e29b-41d4-a716-446655440003"

	now := time.Now()

	manuscripts := []struct {
		ID        string
		Title     string
		IssueID   *string
		FileTypes []constant.ManuscriptFileType
	}{
		{
			ID:        "880e8400-e29b-41d4-a716-446655440001",
			Title:     "The Future of Artificial Intelligence",
			IssueID:   nil,
			FileTypes: []constant.ManuscriptFileType{constant.ManuscriptFileTypeMain, constant.ManuscriptFileTypeCoverLetter},
		},
		{
			ID:        "880e8400-e29b-41d4-a716-446655440002",
			Title:     "Quantum Computing Revisited",
			IssueID:   nil,
			FileTypes: []constant.ManuscriptFileType{constant.ManuscriptFileTypeMain, constant.ManuscriptFileTypeCoverLetter},
		},
		{
			ID:        "880e8400-e29b-41d4-a716-446655440003",
			Title:     "Machine Learning in Healthcare",
			IssueID:   &issueID,
			FileTypes: []constant.ManuscriptFileType{constant.ManuscriptFileTypeMain, constant.ManuscriptFileTypeCopyedited},
		},
	}

	for _, m := range manuscripts {
		for _, ft := range m.FileTypes {
			var count int64
			db.Model(&entity.ManuscriptFile{}).
				Where("manuscript_id = ? AND file_type = ?", m.ID, ft).
				Count(&count)
			if count > 0 {
				continue
			}

			pdfBytes := generateSamplePDF(m.Title, string(ft))
			issueSegment := "submissions"
			if m.IssueID != nil {
				issueSegment = *m.IssueID
			}
			ext := ".pdf"
			storagePath := fmt.Sprintf("manuscripts/%s/%s/%s/%s/v1%s", journalID, issueSegment, m.ID, ft, ext)
			publicURL := fmt.Sprintf("https://placeholder-seed/%s", storagePath)

			svc := storage.NewService()
			uploadedURL, err := svc.Upload(context.Background(), pdfBytes, storagePath, "application/pdf")
			if err == nil {
				publicURL = uploadedURL
			} else {
				log.Printf("[seeder] storage upload failed for %s/%s (%s), using placeholder URL: %v", m.ID, ft, m.Title, err)
			}

			filename := fmt.Sprintf("%s_%s%s", m.Title, ft, ext)
			f := &entity.ManuscriptFile{
				ManuscriptID: m.ID,
				FileType:     ft,
				FilePath:     publicURL,
				Filename:     filename,
				MimeType:     "application/pdf",
				SizeBytes:    int64(len(pdfBytes)),
				Version:      1,
				UploadedAt:   now,
			}

			if err := db.Create(f).Error; err != nil {
				log.Printf("[seeder] failed to create file record for %s/%s: %v", m.ID, ft, err)
			}
		}
	}

	return nil
}

// generateSamplePDF creates a minimal valid PDF with the given title text.
func generateSamplePDF(title, label string) []byte {
	var b bytes.Buffer
	w := func(s string) { b.WriteString(s) }
	wf := func(format string, a ...any) { fmt.Fprintf(&b, format, a...) }

	w("%PDF-1.4\n")

	offsets := make(map[int]int)

	offsets[1] = b.Len()
	w("1 0 obj\n<< /Type /Catalog /Pages 2 0 R >>\nendobj\n")

	offsets[2] = b.Len()
	w("2 0 obj\n<< /Type /Pages /Kids [3 0 R] /Count 1 >>\nendobj\n")

	content := fmt.Sprintf("BT /F1 24 Tf 100 700 Td (%s - %s) Tj ET", title, label)
	offsets[3] = b.Len()
	wf("3 0 obj\n<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792]\n/Contents 4 0 R /Resources << /Font << /F1 5 0 R >> >> >>\nendobj\n")

	offsets[4] = b.Len()
	wf("4 0 obj\n<< /Length %d >>\nstream\n%s\nendstream\nendobj\n", len(content), content)

	offsets[5] = b.Len()
	w("5 0 obj\n<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>\nendobj\n")

	xrefOffset := b.Len()
	w("xref\n")
	wf("0 %d\n", len(offsets)+1)
	wf("%010d %05d %c \r\n", 0, 65535, 'f')
	for i := 1; i <= len(offsets); i++ {
		wf("%010d %05d %c \r\n", offsets[i], 0, 'n')
	}

	w("trailer\n")
	wf("<< /Size %d /Root 1 0 R >>\n", len(offsets)+1)
	w("startxref\n")
	wf("%d\n", xrefOffset)
	w("%%EOF\n")

	return b.Bytes()
}

func FlushManuscripts(db *gorm.DB) error {
	journalIDs := []string{"990e8400-e29b-41d4-a716-446655440001"}
	volIDs := []string{"990e8400-e29b-41d4-a716-446655440002"}
	issueIDs := []string{"990e8400-e29b-41d4-a716-446655440003"}
	manuIDs := []string{
		"880e8400-e29b-41d4-a716-446655440001",
		"880e8400-e29b-41d4-a716-446655440002",
		"880e8400-e29b-41d4-a716-446655440003",
	}

	if err := db.Exec(`DELETE FROM manuscript_files WHERE manuscript_id IN (?)`, manuIDs).Error; err != nil {
		return err
	}
	if err := db.Exec(`DELETE FROM manuscripts WHERE id IN (?)`, manuIDs).Error; err != nil {
		return err
	}
	if err := db.Exec(`DELETE FROM issues WHERE id IN (?)`, issueIDs).Error; err != nil {
		return err
	}
	if err := db.Exec(`DELETE FROM volumes WHERE id IN (?)`, volIDs).Error; err != nil {
		return err
	}
	if err := db.Exec(`DELETE FROM journals WHERE id IN (?)`, journalIDs).Error; err != nil {
		return err
	}

	return nil
}
