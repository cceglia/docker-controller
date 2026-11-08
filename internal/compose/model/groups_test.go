package model

import "testing"

func TestNormalizeGroupName_TrimsTheEndsAndKeepsTheInside(t *testing.T) {
	for _, tc := range []struct {
		name string
		in   string
		want string
	}{
		{"no space", "Lavoro", "Lavoro"},
		{"spaces around", "  Lavoro  ", "Lavoro"},
		{"a tabulator and a newline", "\tLavoro\n", "Lavoro"},
		{"a non-breaking space around", "\u00a0Lavoro\u00a0", "Lavoro"},
		{"an inner space stays", "  Area 51 ", "Area 51"},
		{"spaces only", "   ", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := NormalizeGroupName(tc.in); got != tc.want {
				t.Errorf("NormalizeGroupName(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestSameGroupName_FoldsTheCaseAndKeepsTheUnicodeFoldingSimple(t *testing.T) {
	for _, tc := range []struct {
		name string
		a    string
		b    string
		want bool
	}{
		{"the same name", "Lavoro", "Lavoro", true},
		{"ASCII capitals", "lavoro", "LAVORO", true},
		{"accented capitals", "Città", "CITTÀ", true},
		{"spaces around do not count", " Casa ", "casa", true},
		{"an inner space counts", "Area 51", "Area51", false},
		{"simple folding keeps the eszett apart", "Straße", "STRASSE", false},
		{"different names", "Casa", "Lavoro", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := SameGroupName(tc.a, tc.b); got != tc.want {
				t.Errorf("SameGroupName(%q, %q) = %t, want %t", tc.a, tc.b, got, tc.want)
			}
		})
	}
}

func TestLessGroupName_OrdersWithoutCaseThenBySpelling(t *testing.T) {
	for _, tc := range []struct {
		name string
		a    string
		b    string
		want bool
	}{
		{"alphabetical without case", "alfa", "Beta", true},
		{"alphabetical without case reversed", "Beta", "alfa", false},
		{"the same name ignoring case orders by the exact spelling", "Casa", "casa", true},
		{"the same name ignoring case reversed", "casa", "Casa", false},
		{"the same name is not before itself", "Casa", "Casa", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := LessGroupName(tc.a, tc.b); got != tc.want {
				t.Errorf("LessGroupName(%q, %q) = %t, want %t", tc.a, tc.b, got, tc.want)
			}
		})
	}
}

func TestValidateGroupName_AppliesTheRuleOfTheMode(t *testing.T) {
	existing := []string{"Casa", "Lavoro"}
	for _, tc := range []struct {
		name         string
		given        string
		mode         GroupNameMode
		from         string
		status       GroupNameStatus
		normal       string
		existingName string
	}{
		{"a new name in the form", "  Nuovo ", GroupNameAssign, "", GroupNameValid, "Nuovo", ""},
		{"the empty name clears the form", "   ", GroupNameAssign, "", GroupNameValid, "", ""},
		{"a reserved label in the form", "Senza Gruppo", GroupNameAssign, "", GroupNameInvalid, "Senza Gruppo", ""},
		{"a reserved label with capitals", "nessun gruppo", GroupNameAssign, "", GroupNameInvalid, "nessun gruppo", ""},
		{"a saved name in the form is the Group to assign", "casa", GroupNameAssign, "", GroupNameDuplicate, "casa", "Casa"},
		{"a new Group", "Nuovo", GroupNameCreate, "", GroupNameValid, "Nuovo", ""},
		{"an empty new Group", "  ", GroupNameCreate, "", GroupNameInvalid, "", ""},
		{"a duplicate new Group", "LAVORO", GroupNameCreate, "", GroupNameDuplicate, "LAVORO", "Lavoro"},
		{"a reserved new Group", "Nessun Gruppo", GroupNameCreate, "", GroupNameInvalid, "Nessun Gruppo", ""},
		{"a rename to a new name", "Ufficio", GroupNameRename, "Casa", GroupNameValid, "Ufficio", ""},
		{"a rename that only changes the spelling", "CASA", GroupNameRename, "Casa", GroupNameValid, "CASA", ""},
		{"a rename onto another Group", "LAVORO", GroupNameRename, "Casa", GroupNameDuplicate, "LAVORO", "Lavoro"},
		{"an empty rename", "", GroupNameRename, "Casa", GroupNameInvalid, "", ""},
		{"a reserved rename", "Senza Gruppo", GroupNameRename, "Casa", GroupNameInvalid, "Senza Gruppo", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := ValidateGroupName(tc.given, existing, tc.mode, tc.from)
			if got.Status != tc.status || got.Name != tc.normal || got.Existing != tc.existingName {
				t.Errorf("ValidateGroupName(%q) = %+v, want status %d, name %q, existing %q",
					tc.given, got, tc.status, tc.normal, tc.existingName)
			}
		})
	}
}
