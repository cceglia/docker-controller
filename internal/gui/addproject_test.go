package gui

import (
	"errors"
	"path/filepath"
	"testing"

	"dockercontroller/internal/compose"
	"dockercontroller/internal/gui/guitest"
	"dockercontroller/internal/gui/view"
)

// formWith builds a form with the named files already chosen, so a test says
// only what it is about.
func formWith(files ...string) AddProjectForm {
	form := AddProjectForm{}
	for _, file := range files {
		form.AddFile(file)
	}
	return form
}

func TestAddProjectForm_AddFileKeepsTheMainFileFirst(t *testing.T) {
	for _, tc := range []struct {
		name  string
		paths []string
		want  []string
	}{
		{"no file", nil, nil},
		{"one file", []string{"/home/me/shop/docker-compose.yml"}, []string{"/home/me/shop/docker-compose.yml"}},
		{
			"the second file comes after the main one",
			[]string{"/home/me/shop/docker-compose.yml", "/home/me/shop/compose.override.yml"},
			[]string{"/home/me/shop/docker-compose.yml", "/home/me/shop/compose.override.yml"},
		},
		{
			"a file chosen twice is kept once",
			[]string{"/home/me/shop/docker-compose.yml", "/home/me/shop/docker-compose.yml"},
			[]string{"/home/me/shop/docker-compose.yml"},
		},
		{
			"a blank path is ignored",
			[]string{"", "/home/me/shop/docker-compose.yml"},
			[]string{"/home/me/shop/docker-compose.yml"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			form := formWith(tc.paths...)
			guitest.AssertStrings(t, "Files", form.Files, tc.want)
		})
	}
}

func TestAddProjectForm_RemoveFileMovesTheMainFile(t *testing.T) {
	main := "/home/me/shop/docker-compose.yml"
	override := "/home/me/shop/compose.override.yml"

	t.Run("removing the main file picks the next one", func(t *testing.T) {
		form := formWith(main, override)
		form.RemoveFile(0)
		guitest.AssertStrings(t, "Files", form.Files, []string{override})
		if got := form.MainFile(); got != override {
			t.Errorf("MainFile = %q, want %q", got, override)
		}
	})

	t.Run("removing an additional file keeps the main one", func(t *testing.T) {
		form := formWith(main, override)
		form.RemoveFile(1)
		guitest.AssertStrings(t, "Files", form.Files, []string{main})
	})

	t.Run("removing outside the list changes nothing", func(t *testing.T) {
		form := formWith(main)
		form.RemoveFile(3)
		form.RemoveFile(-1)
		guitest.AssertStrings(t, "Files", form.Files, []string{main})
	})
}

func TestAddProjectForm_PickOpensOnlyAfterARealChoice(t *testing.T) {
	main := "/home/me/shop/docker-compose.yml"
	override := "/home/me/shop/compose.override.yml"
	wiki := "/home/me/wiki/docker-compose.yml"

	for _, tc := range []struct {
		name      string
		prep      func(*AddProjectForm)
		paths     []string
		fail      bool
		wantAdded bool
		wantFiles []string
		wantName  string
		wantAsked []string
	}{
		{
			name: "a closed picker opens nothing",
		},
		{
			name:      "a closed picker keeps the chosen files and name",
			prep:      func(f *AddProjectForm) { f.AddFile(main); f.Propose("shop") },
			wantFiles: []string{main},
			wantName:  "shop",
		},
		{
			name:      "the first pick brings the files and proposes the name",
			paths:     []string{main, override},
			wantAdded: true,
			wantFiles: []string{main, override},
			wantName:  "shop",
			wantAsked: []string{main},
		},
		{
			name:      "an extra pick keeps the name of the main file",
			prep:      func(f *AddProjectForm) { f.AddFile(main); f.Propose("shop") },
			paths:     []string{override},
			wantAdded: true,
			wantFiles: []string{main, override},
			wantName:  "shop",
		},
		{
			name:      "a pick after the main file left proposes the new name",
			prep:      func(f *AddProjectForm) { f.AddFile(main); f.Propose("shop"); f.RemoveFile(0) },
			paths:     []string{wiki},
			wantAdded: true,
			wantFiles: []string{wiki},
			wantName:  "wiki",
			wantAsked: []string{wiki},
		},
		{
			name:      "a typed name survives the pick",
			prep:      func(f *AddProjectForm) { f.Rename("mine") },
			paths:     []string{main},
			wantAdded: true,
			wantFiles: []string{main},
			wantName:  "mine",
			wantAsked: []string{main},
		},
		{
			name:      "a file already chosen adds nothing",
			prep:      func(f *AddProjectForm) { f.AddFile(main); f.Propose("shop") },
			paths:     []string{main},
			wantFiles: []string{main},
			wantName:  "shop",
		},
		{
			name:  "a blank path adds nothing",
			paths: []string{"", "   "},
		},
		{
			name:      "a failed proposal keeps the files",
			paths:     []string{main},
			fail:      true,
			wantAdded: true,
			wantFiles: []string{main},
			wantAsked: []string{main},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var asked []string
			propose := func(path string) (string, error) {
				asked = append(asked, path)
				if tc.fail {
					return "", compose.ErrInvalidProject
				}
				return filepath.Base(filepath.Dir(path)), nil
			}

			var form AddProjectForm
			if tc.prep != nil {
				tc.prep(&form)
			}
			added, err := form.Pick(tc.paths, propose)

			if added != tc.wantAdded {
				t.Errorf("Pick added = %t, want %t", added, tc.wantAdded)
			}
			if got := err != nil; got != tc.fail {
				t.Errorf("Pick error = %v, want a failure: %t", err, tc.fail)
			}
			guitest.AssertStrings(t, "Files", form.Files, tc.wantFiles)
			if form.Name != tc.wantName {
				t.Errorf("Name = %q, want %q", form.Name, tc.wantName)
			}
			guitest.AssertStrings(t, "propose", asked, tc.wantAsked)
		})
	}
}

func TestAddProjectForm_ProposeStopsAtTheNameTheUserTyped(t *testing.T) {
	main := "/home/me/shop/docker-compose.yml"
	override := "/home/me/shop/compose.override.yml"

	for _, tc := range []struct {
		name   string
		steps  []func(*AddProjectForm)
		want   string
		edited bool
	}{
		{
			name:  "no file, no name",
			want:  "",
			steps: nil,
		},
		{
			name:  "the proposal of the main file is shown",
			steps: []func(*AddProjectForm){func(f *AddProjectForm) { f.AddFile(main); f.Propose("shop") }},
			want:  "shop",
		},
		{
			name: "another main file replaces an untouched proposal",
			steps: []func(*AddProjectForm){
				func(f *AddProjectForm) { f.AddFile(main); f.Propose("shop") },
				func(f *AddProjectForm) { f.RemoveFile(0); f.AddFile(override); f.Propose("override") },
			},
			want: "override",
		},
		{
			name: "a typed name survives a later main file",
			steps: []func(*AddProjectForm){
				func(f *AddProjectForm) { f.AddFile(main); f.Propose("shop") },
				func(f *AddProjectForm) { f.Rename("mine"); f.Propose("other") },
			},
			want:   "mine",
			edited: true,
		},
		{
			name: "a name the user cleared stays cleared",
			steps: []func(*AddProjectForm){
				func(f *AddProjectForm) { f.AddFile(main); f.Propose("shop") },
				func(f *AddProjectForm) { f.Rename(""); f.Propose("shop") },
			},
			want:   "",
			edited: true,
		},
		{
			name:  "a proposal is not a typed name",
			steps: []func(*AddProjectForm){func(f *AddProjectForm) { f.AddFile(main); f.Propose("shop") }},
			want:  "shop",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var form AddProjectForm
			for _, step := range tc.steps {
				step(&form)
			}
			if form.Name != tc.want {
				t.Errorf("Name = %q, want %q", form.Name, tc.want)
			}
			if form.NameEdited() != tc.edited {
				t.Errorf("NameEdited = %t, want %t", form.NameEdited(), tc.edited)
			}
		})
	}
}

func TestAddProjectForm_CanAddNeedsAFileAndAName(t *testing.T) {
	main := "/home/me/shop/docker-compose.yml"
	for _, tc := range []struct {
		name string
		form AddProjectForm
		want bool
	}{
		{"nothing chosen", AddProjectForm{}, false},
		{"a file without a name", formWith(main), false},
		{"a name without a file", AddProjectForm{Name: "shop"}, false},
		{"a blank name", AddProjectForm{Files: []string{main}, Name: "   "}, false},
		{"a file and a name", AddProjectForm{Files: []string{main}, Name: "shop"}, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.form.CanAdd(); got != tc.want {
				t.Errorf("CanAdd = %t, want %t", got, tc.want)
			}
		})
	}
}

func TestAddProjectForm_CleanNameTrimsTheName(t *testing.T) {
	for _, tc := range []struct {
		name string
		in   string
		want string
	}{
		{"already clean", "shop", "shop"},
		{"surrounded by blanks", "  shop  ", "shop"},
		{"blank", "   ", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			form := AddProjectForm{Name: tc.in}
			if got := form.CleanName(); got != tc.want {
				t.Errorf("CleanName = %q, want %q", got, tc.want)
			}
		})
	}
}

// TestAddProjectForm_PickKeepsTheOrderOfThePicker checks that the form keeps
// the File compose in the order the picker returned them, across more picks and
// a removal: the first one stays the main one, because compose reads -f in that
// order.
func TestAddProjectForm_PickKeepsTheOrderOfThePicker(t *testing.T) {
	main := "/home/me/shop/docker-compose.yml"
	override := "/home/me/shop/compose.override.yml"
	extra := "/home/me/shop/compose.extra.yml"

	propose := func(path string) (string, error) {
		return filepath.Base(filepath.Dir(path)), nil
	}

	t.Run("one pick keeps the order of the picker", func(t *testing.T) {
		var form AddProjectForm
		if _, err := form.Pick([]string{override, main}, propose); err != nil {
			t.Fatalf("Pick error = %v", err)
		}
		guitest.AssertStrings(t, "Files", form.Files, []string{override, main})
		if got := form.MainFile(); got != override {
			t.Errorf("MainFile = %q, want the first file of the picker: %q", got, override)
		}
	})

	t.Run("a later pick appends the files after the chosen ones", func(t *testing.T) {
		var form AddProjectForm
		form.Pick([]string{main, override}, propose)
		form.Pick([]string{extra}, propose)
		guitest.AssertStrings(t, "Files", form.Files, []string{main, override, extra})
	})

	t.Run("removing the main file moves the first place down the order", func(t *testing.T) {
		var form AddProjectForm
		form.Pick([]string{main, override, extra}, propose)
		form.RemoveFile(0)
		guitest.AssertStrings(t, "Files", form.Files, []string{override, extra})
		if got := form.MainFile(); got != override {
			t.Errorf("MainFile = %q, want %q", got, override)
		}
	})
}

// TestAddRefusal_ExplainsTheRefusalBeforeTheSave checks the three refusals the
// window can tell before the Manager runs, with the text the user reads.
func TestAddRefusal_ExplainsTheRefusalBeforeTheSave(t *testing.T) {
	main := "/home/me/shop/docker-compose.yml"
	projects := []compose.Project{{Name: "shop"}}

	for _, tc := range []struct {
		name  string
		input string
		files []string
		err   error
		text  string
	}{
		{
			name:  "no File compose",
			input: "shop",
			err:   compose.ErrNoComposeFile,
			text:  "Serve almeno un File compose.",
		},
		{
			name:  "no name",
			files: []string{main},
			err:   compose.ErrInvalidName,
			text:  "Nome non valido: usa lettere minuscole, cifre, - e _.",
		},
		{
			name:  "a name with blank space only",
			input: "   ",
			files: []string{main},
			err:   compose.ErrInvalidName,
			text:  "Nome non valido: usa lettere minuscole, cifre, - e _.",
		},
		{
			name:  "a name already saved",
			input: "  shop  ",
			files: []string{main},
			err:   compose.ErrDuplicateProject,
			text:  "Esiste già un Progetto con questo nome.",
		},
		{
			name:  "a request the Manager may judge",
			input: "blog",
			files: []string{main},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := AddRefusal(tc.input, tc.files, projects)
			if tc.err == nil {
				if err != nil {
					t.Fatalf("AddRefusal error = %v, want none", err)
				}
				return
			}
			if !errors.Is(err, tc.err) {
				t.Fatalf("AddRefusal error = %v, want %v", err, tc.err)
			}
			if got := view.UserError(err).Error(); got != tc.text {
				t.Errorf("the refusal reads %q, want %q", got, tc.text)
			}
		})
	}
}

func TestNameTaken_RefusesASavedName(t *testing.T) {
	projects := []compose.Project{{Name: "shop"}, {Name: "wiki"}}
	for _, tc := range []struct {
		name  string
		input string
		want  bool
	}{
		{"no name", "", false},
		{"a saved name", "shop", true},
		{"another saved name", "wiki", true},
		{"a saved name with blanks", "  shop  ", true},
		{"an unknown name", "blog", false},
		{"a name of another case is not the same", "Shop", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := NameTaken(tc.input, projects); got != tc.want {
				t.Errorf("NameTaken(%q) = %t, want %t", tc.input, got, tc.want)
			}
		})
	}
}
