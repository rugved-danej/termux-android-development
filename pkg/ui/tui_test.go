package ui

import (
	"testing"
	"time"

	"atomicgo.dev/keyboard"
	"atomicgo.dev/keyboard/keys"
)

func TestRunSelectMenuKeyboard(t *testing.T) {
	items := []MenuItem{
		{Key: "1", Title: "First Action", Description: "Description of first action"},
		{Key: "2", Title: "Second Action", Description: "Description of second action"},
	}

	go func() {
		time.Sleep(100 * time.Millisecond)
		_ = keyboard.SimulateKeyPress(keys.CtrlD) // Toggle details
		time.Sleep(50 * time.Millisecond)
		_ = keyboard.SimulateKeyPress(keys.Down) // Move down
		time.Sleep(50 * time.Millisecond)
		_ = keyboard.SimulateKeyPress(keys.Enter) // Select second item
	}()

	idx, confirmed, err := RunSelectMenu("Test Menu", "Test Device", items)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !confirmed {
		t.Fatalf("expected confirmed true")
	}
	if idx != 1 {
		t.Fatalf("expected idx 1, got %d", idx)
	}
}

func TestRunCheckboxMenuKeyboard(t *testing.T) {
	items := []*CheckboxOption{
		{ID: "opt1", Label: "Option 1", Description: "Desc 1", Checked: false},
		{ID: "opt2", Label: "Option 2", Description: "Desc 2", Checked: true},
	}

	go func() {
		time.Sleep(100 * time.Millisecond)
		_ = keyboard.SimulateKeyPress(keys.Space) // Toggle opt1 to true
		time.Sleep(50 * time.Millisecond)
		_ = keyboard.SimulateKeyPress(keys.Enter) // Confirm
	}()

	res, confirmed, err := RunCheckboxMenu("Test Chooser", "Subtitle", items)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !confirmed {
		t.Fatalf("expected confirmed true")
	}
	if !res[0].Checked {
		t.Fatalf("expected opt1 to be checked")
	}
}

func TestRunJavaProfileMenuKeyboard(t *testing.T) {
	items := []MenuItem{
		{Key: "25+21", Title: "[25 + 21] OpenJDK 25 (Host) + OpenJDK 21 (Gradle Toolchain)"},
		{Key: "21", Title: "[21] OpenJDK 21 LTS Only"},
	}

	go func() {
		time.Sleep(100 * time.Millisecond)
		_ = keyboard.SimulateKeyPress(keys.Tab) // Trigger help URL
		time.Sleep(50 * time.Millisecond)
		_ = keyboard.SimulateKeyPress(keys.Down) // Move down
		time.Sleep(50 * time.Millisecond)
		_ = keyboard.SimulateKeyPress(keys.Enter) // Select second item
	}()

	idx, confirmed, err := RunJavaProfileMenu("SELECT JAVA PROFILE", "Test Device", items, "https://example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !confirmed {
		t.Fatalf("expected confirmed true")
	}
	if idx != 1 {
		t.Fatalf("expected idx 1, got %d", idx)
	}
}
