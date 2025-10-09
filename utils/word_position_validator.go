package utils

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/techmaster-vietnam/dd_goshare/pkg/models"
)

// ValidateAndFixWordPositions kiểm tra và sửa lại vị trí bắt đầu của các từ trong dialog.fill_in_words
// Trả về true nếu có thay đổi, false nếu không có thay đổi
func ValidateAndFixWordPositions(dialog *models.Dialog, fillInWords *models.FillInWords, fillInBlanks *models.FillInBlank) (bool, error) {
	if fillInWords == nil || len(fillInWords.Words) == 0 {
		return false, nil // Không có từ nào để kiểm tra
	}

	// Parse timestamp data từ dialog.Result để lấy vị trí thực tế
	timestampPositions, err := extractWordPositionsFromTimestamp(dialog.Result)
	if err != nil {
		return false, fmt.Errorf("failed to extract word positions from timestamp: %v", err)
	}

	hasChanges := false

	// Kiểm tra từng từ trong fill_in_words
	for i := range fillInWords.Words {
		word := &fillInWords.Words[i]

		// Tìm vị trí thực tế từ timestamp
		actualPosition, found := timestampPositions[strings.ToLower(word.Word)]

		if found && actualPosition != word.Start {
			fmt.Printf("Fixing word '%s': old position %d -> new position %d\n",
				word.Word, word.Start, actualPosition)
			word.Start = actualPosition
			hasChanges = true
		} else if !found {
			fmt.Printf("Warning: Could not find word '%s' in timestamp data\n", word.Word)
		}
	}

	// Nếu có thay đổi, cập nhật lại Words_to_fill
	if hasChanges {
		updatedJSON, err := json.Marshal(fillInWords)
		if err != nil {
			return false, fmt.Errorf("failed to marshal updated fill_in_words: %v", err)
		}
		fillInBlanks.WordsIndex = json.RawMessage(updatedJSON)
	}

	return hasChanges, nil
}

// extractWordPositionsFromTimestamp trích xuất vị trí từ timestamp data
func extractWordPositionsFromTimestamp(resultData json.RawMessage) (map[string]int, error) {
	var timestampResult models.TimestampResult
	if err := json.Unmarshal(resultData, &timestampResult); err != nil {
		return nil, fmt.Errorf("failed to unmarshal timestamp result: %v", err)
	}

	positions := make(map[string]int)

	// Extract positions từ words array
	for _, wordData := range timestampResult.Words {
		if len(wordData) >= 3 {
			// Format: [timestamp, word, position]
			if word, ok := wordData[1].(string); ok {
				if position, ok := wordData[2].(float64); ok {
					positions[strings.ToLower(word)] = int(position)
				}
			}
		}
	}

	return positions, nil
}

// ValidateAndFixWordPositionsWithUpdate kiểm tra, sửa và lưu lại vào database
func ValidateAndFixWordPositionsWithUpdate(dialog *models.Dialog, fillInWords *models.FillInWords, fillInBlanks *models.FillInBlank, updateFunc func(*models.FillInBlank) error) error {
	hasChanges, err := ValidateAndFixWordPositions(dialog, fillInWords, fillInBlanks)
	if err != nil {
		return fmt.Errorf("failed to validate word positions: %v", err)
	}

	if hasChanges {
		if err := updateFunc(fillInBlanks); err != nil {
			return fmt.Errorf("failed to update dialog in database: %v", err)
		}
		fmt.Printf("Successfully updated word positions for dialog %s\n", dialog.ID)
	} else {
		fmt.Printf("No position corrections needed for dialog %s\n", dialog.ID)
	}

	return nil
}

// GetWordPositionReport tạo báo cáo về vị trí các từ trong fill_in_words
func GetWordPositionReport(dialog *models.Dialog, fillInBlanks *models.FillInBlank) (*models.WordPositionReport, error) {
	report := &models.WordPositionReport{
		DialogID: dialog.ID,
		Words:    []models.WordPositionStatus{},
	}

	if len(fillInBlanks.WordsIndex) == 0 {
		return report, nil
	}

	var fillInWords models.FillInWords
	if err := json.Unmarshal(fillInBlanks.WordsIndex, &fillInWords); err != nil {
		return nil, fmt.Errorf("failed to unmarshal fill_in_words: %v", err)
	}

	if len(fillInWords.Words) == 0 {
		return report, nil
	}

	// Parse timestamp data để lấy vị trí thực tế
	timestampPositions, err := extractWordPositionsFromTimestamp(dialog.Result)
	if err != nil {
		return nil, fmt.Errorf("failed to extract word positions from timestamp: %v", err)
	}

	report.TotalWords = len(fillInWords.Words)

	for i, word := range fillInWords.Words {
		// Tìm vị trí thực tế từ timestamp
		actualPosition, found := timestampPositions[strings.ToLower(word.Word)]

		wordStatus := models.WordPositionStatus{
			Index:           i + 1,
			Word:            word.Word,
			CurrentPosition: word.Start,
		}

		if !found {
			wordStatus.Status = "✗ NOT FOUND"
			wordStatus.StatusCode = "not_found"
			wordStatus.Message = "Word not found in timestamp data"
			report.NotFoundWords++
		} else if actualPosition != word.Start {
			wordStatus.Status = "⚠ INCORRECT"
			wordStatus.StatusCode = "incorrect"
			wordStatus.CorrectPosition = &actualPosition
			wordStatus.Message = fmt.Sprintf("Should be at position %d", actualPosition)
			report.IncorrectWords++
		} else {
			wordStatus.Status = "✓ CORRECT"
			wordStatus.StatusCode = "correct"
			wordStatus.CorrectPosition = &actualPosition
			report.CorrectWords++
		}

		report.Words = append(report.Words, wordStatus)
	}

	return report, nil
}
