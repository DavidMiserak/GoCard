// File: internal/data/editor.go

package data

import (
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/DavidMiserak/GoCard/internal/model"
	tea "github.com/charmbracelet/bubbletea"
)

type EditorResponse struct {
	FileName string
	ExitCode error
	IsEdit   bool   // true = editing, false = adding
	CardID   string // original card ID (for edits)
}

func getShellEditor() (string, error) {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		// Default to vi
		editor = "vi"
	}

	return exec.LookPath(editor)
}

func LaunchEditor(file string, isEdit bool, cardID string) tea.Cmd {
	editor, err := getShellEditor()

	if err != nil {
		return nil // can't launch
	}

	cmdToRun := exec.Command(editor, file)

	return tea.ExecProcess(cmdToRun, func(result error) tea.Msg {
		return EditorResponse{
			FileName: file,
			ExitCode: result,
			IsEdit:   isEdit,
			CardID:   cardID,
		}
	})
}

func getCardTemplate() string {
	currentDate := time.Now().Format("2006-01-02")
	template := fmt.Sprintf(`---
tags: []
created: %s
review_interval: 0
---

# Title

## Question

## Answer

`, currentDate)

	return template

}

// Abstracted method to create temporary files with desired text
func createTmpFileWithText(text string) (string, error) {
	tmpFile, err := os.CreateTemp("", "GoCard-tmp-*.md")
	if err != nil {
		return "", err
	}

	_, err = tmpFile.WriteString(text)
	if err != nil {
		tmpFile.Close()
		return "", err
	}

	err = tmpFile.Close()
	if err != nil {
		return "", err
	}

	return tmpFile.Name(), nil
}

func CreateTmpFileWithCard(card model.Card) (string, error) {
	originalFileContents, err := os.ReadFile(card.ID)
	if err != nil {
		return "", err
	}

	return createTmpFileWithText(string(originalFileContents))
}

func CreateTmpFileWithTemplate() (string, error) {
	return createTmpFileWithText(getCardTemplate())
}
