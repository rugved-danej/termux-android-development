package cmd

import (
	"strings"
	"testing"
)

func TestCompletionScripts(t *testing.T) {
	// 1. Verify Bash completion script contains key commands and profiles
	if !strings.Contains(BashCompletionScript, "complete -F _tadev_completions tadev termux-android-dev") {
		t.Errorf("Bash completion missing registration line")
	}
	if !strings.Contains(BashCompletionScript, "25+21") || !strings.Contains(BashCompletionScript, "ndk") {
		t.Errorf("Bash completion missing profiles or ndk command")
	}

	// 2. Verify Zsh completion script
	if !strings.Contains(ZshCompletionScript, "#compdef tadev termux-android-dev") {
		t.Errorf("Zsh completion missing compdef header")
	}
	if !strings.Contains(ZshCompletionScript, "clean:Purge temporary build caches") {
		t.Errorf("Zsh completion missing clean description")
	}

	// 3. Verify Fish completion script
	if !strings.Contains(FishCompletionScript, "complete -c tadev") {
		t.Errorf("Fish completion missing complete commands")
	}
	if !strings.Contains(FishCompletionScript, "complete -c termux-android-dev") {
		t.Errorf("Fish completion missing alias complete commands")
	}
}
