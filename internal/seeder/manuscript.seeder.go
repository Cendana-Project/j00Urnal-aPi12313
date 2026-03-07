package seeder

import (
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/api-monolith-template/internal/constant"
	"github.com/api-monolith-template/internal/model/entity"
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
			MainAuthorID:  authorID,
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
			MainAuthorID:     authorID,
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
			MainAuthorID:     authorID,
			AssignedEditorID: &editorID,
			IsTncAccepted:    true,
			CreatedAt:        time.Now(),
			PublishedAt:      time.Now(),
		},
	}

	for i, m := range manuscripts {
		if err := db.FirstOrCreate(&m, entity.Manuscript{ID: m.ID}).Error; err != nil {
			return fmt.Errorf("seed manuscript idx %d: %w", i, err)
		}
	}

	return nil
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
