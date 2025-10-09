package services

import (
	repo "github.com/techmaster-vietnam/dd_goshare/pkg/repositories"
	"github.com/techmaster-vietnam/dd_goshare/pkg/models"
)

// DialogService handles business logic for dialogs
type DialogService struct {
	dialogRepo *repo.DialogRepository
}

// NewDialogService creates a new DialogService instance

func NewDialogService(dialogRepo *repo.DialogRepository, 
	) *DialogService {
	return &DialogService{
		dialogRepo: dialogRepo,
	}
}

// GetAllDialogs retrieves all dialogs
func (s *DialogService) GetAllDialogs() ([]models.Dialog, error) {
	return s.dialogRepo.GetAllDialogs()
}

// GetDialog retrieves a dialog by ID
func (s *DialogService) GetDialog(dialogID string) (*models.Dialog, error) {
	return s.dialogRepo.GetDialog(dialogID)
}
