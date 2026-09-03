package mapper

import (
	"testing"
	"time"

	"github.com/api-monolith-template/internal/constant"
	"github.com/api-monolith-template/internal/model/entity"
)

func TestToReviewerWorkspaceResponseSelectsLatestMainManuscriptFile(t *testing.T) {
	baseTime := time.Date(2026, time.September, 1, 10, 0, 0, 0, time.UTC)
	assignment := &entity.ReviewAssignment{ID: "assignment-id"}
	manuscript := &entity.Manuscript{
		ID:    "manuscript-id",
		Title: "A manuscript",
		Files: []entity.ManuscriptFile{
			{
				ID:         "cover-letter",
				FileType:   constant.ManuscriptFileTypeCoverLetter,
				FilePath:   "/cover-letter.pdf",
				Version:    1,
				UploadedAt: baseTime,
			},
			{
				ID:         "main-v1",
				FileType:   constant.ManuscriptFileTypeMain,
				FilePath:   "/manuscript-v1.pdf",
				Version:    1,
				UploadedAt: baseTime.Add(time.Minute),
			},
			{
				ID:         "figure",
				FileType:   constant.ManuscriptFileTypeFigure,
				FilePath:   "/figure.png",
				Version:    1,
				UploadedAt: baseTime.Add(2 * time.Minute),
			},
			{
				ID:         "main-v2",
				FileType:   constant.ManuscriptFileTypeMain,
				FilePath:   "/manuscript-v2.pdf",
				Version:    2,
				UploadedAt: baseTime.Add(3 * time.Minute),
			},
		},
	}

	got := ToReviewerWorkspaceResponse(assignment, manuscript, nil, nil, nil, false, nil, 0)
	if got.Manuscript.File == nil {
		t.Fatal("workspace manuscript file is nil")
	}
	if got.Manuscript.File.ID != "main-v2" {
		t.Fatalf("workspace manuscript file ID = %q, want latest MAIN file", got.Manuscript.File.ID)
	}
	if got.Manuscript.File.FilePath != "/manuscript-v2.pdf" {
		t.Fatalf("workspace manuscript file path = %q, want /manuscript-v2.pdf", got.Manuscript.File.FilePath)
	}
}

func TestToReviewerWorkspaceResponseOmitsFileWhenNoMainExists(t *testing.T) {
	assignment := &entity.ReviewAssignment{ID: "assignment-id"}
	manuscript := &entity.Manuscript{
		ID: "manuscript-id",
		Files: []entity.ManuscriptFile{
			{ID: "cover-letter", FileType: constant.ManuscriptFileTypeCoverLetter},
		},
	}

	got := ToReviewerWorkspaceResponse(assignment, manuscript, nil, nil, nil, false, nil, 0)
	if got.Manuscript.File != nil {
		t.Fatalf("workspace manuscript file = %+v, want nil when MAIN is absent", got.Manuscript.File)
	}
}
