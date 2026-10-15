package env

import (
	"testing"
)

func TestAvailableJavaProfiles(t *testing.T) {
	if len(AvailableJavaProfiles) == 0 {
		t.Fatal("expected at least one available Java profile")
	}

	foundDefault := false
	for _, p := range AvailableJavaProfiles {
		if p.ID == "" {
			t.Errorf("empty ID for profile: %+v", p)
		}
		if p.Title == "" {
			t.Errorf("empty Title for profile %s", p.ID)
		}
		if p.Description == "" {
			t.Errorf("empty Description for profile %s", p.ID)
		}
		if len(p.Packages) == 0 {
			t.Errorf("expected packages for profile %s", p.ID)
		}
		if p.HostJavaDir == "" {
			t.Errorf("expected HostJavaDir for profile %s", p.ID)
		}
		if len(p.ToolchainDirs) == 0 {
			t.Errorf("expected ToolchainDirs for profile %s", p.ID)
		}
		if p.IsDefault {
			foundDefault = true
		}
	}

	if !foundDefault {
		t.Error("expected at least one default Java profile")
	}
}

func TestNormalizeJavaProfileID(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"25+21", "25+21"},
		{"25 + 21", "25+21"},
		{"25, 21", "25+21"},
		{"25 & 21", "25+21"},
		{"25 and 21", "25+21"},
		{"25_21", "25+21"},
		{"25", "25"},
		{"default", "21"},
		{"21", "21"},
		{"jdk21", "21"},
		{"17", "17"},
		{"jdk17", "17"},
		{"25+17", "25+17"},
		{"25 + 17", "25+17"},
		{"21+17", "21+17"},
	}

	for _, tt := range tests {
		got := NormalizeJavaProfileID(tt.input)
		if got != tt.expected {
			t.Errorf("NormalizeJavaProfileID(%q) = %q, expected %q", tt.input, got, tt.expected)
		}
	}
}

func TestGetJavaProfile(t *testing.T) {
	p21, ok := GetJavaProfile("21")
	if !ok {
		t.Fatal("expected profile 21 to be found")
	}
	if p21.HostJavaDir != "java-21-openjdk" {
		t.Errorf("unexpected host dir for 21: %s", p21.HostJavaDir)
	}

	p25, ok := GetJavaProfile("25")
	if !ok {
		t.Fatal("expected profile 25 to be found")
	}
	if p25.HostJavaDir != "java-25-openjdk" {
		t.Errorf("unexpected host dir for 25: %s", p25.HostJavaDir)
	}

	p25_21, ok := GetJavaProfile("25+21")
	if !ok {
		t.Fatal("expected profile 25+21 to be found")
	}
	if p25_21.RequiresTur {
		t.Errorf("expected official stable profile 25+21 not to require tur-repo")
	}

	p17, ok := GetJavaProfile("17")
	if !ok {
		t.Fatal("expected profile 17 to be found")
	}
	if p17.HostJavaDir != "java-17-openjdk" {
		t.Errorf("unexpected host dir for 17: %s", p17.HostJavaDir)
	}

	_, ok = GetJavaProfile("nonexistent_profile_999")
	if ok {
		t.Error("expected nonexistent profile to return false")
	}
}

func TestGetPackagesForProfile(t *testing.T) {
	p21, _ := GetJavaProfile("21")
	pkgs := GetPackagesForProfile(p21)

	found21 := false
	for _, pkg := range pkgs {
		if pkg == "openjdk-21" {
			found21 = true
		}
		if pkg == "openjdk-25" {
			t.Error("profile 21 should not include openjdk-25")
		}
	}
	if !found21 {
		t.Error("expected openjdk-21 in packages for profile 21")
	}
}
