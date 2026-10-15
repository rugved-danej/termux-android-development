package cmd

import (
	"testing"

	"tadev/pkg/env"
)

func TestJavaProfilesExist(t *testing.T) {
	expected := []string{"25+21", "25+17", "21+17", "21", "25", "17"}
	for _, id := range expected {
		p, ok := env.GetJavaProfile(id)
		if !ok {
			t.Errorf("expected Java profile %s to exist", id)
		}
		if p.ID != id {
			t.Errorf("expected profile ID %s, got %s", id, p.ID)
		}
	}
}

func TestJavaProfileHostDirs(t *testing.T) {
	p21, _ := env.GetJavaProfile("21")
	if p21.HostJavaDir != "java-21-openjdk" {
		t.Errorf("expected java-21-openjdk, got %s", p21.HostJavaDir)
	}

	p25, _ := env.GetJavaProfile("25")
	if p25.HostJavaDir != "java-25-openjdk" {
		t.Errorf("expected java-25-openjdk, got %s", p25.HostJavaDir)
	}

	p17, _ := env.GetJavaProfile("17")
	if p17.HostJavaDir != "java-17-openjdk" {
		t.Errorf("expected java-17-openjdk, got %s", p17.HostJavaDir)
	}
}
