package main

import (
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// gremlinsConfigGlob matches the per-package Gremlins configs. The test reads
// every file: a new package adds a file, and a missing file is a missing gate.
const gremlinsConfigGlob = ".gremlins/*.yaml"

// gremlinsSharedBlock is the part of a config that every package repeats.
// Gremlins reads one config per run, so the shared values drift unless a test
// compares them.
type gremlinsSharedBlock struct {
	workers            string
	timeoutCoefficient string
	excludeFiles       []string
}

// readGremlinsSharedBlock returns the shared block of a Gremlins config. The
// values stay text: the test compares them across files and does not interpret
// them.
func readGremlinsSharedBlock(t *testing.T, path string) gremlinsSharedBlock {
	t.Helper()
	var block gremlinsSharedBlock
	inExcludeFiles := false
	for _, line := range strings.Split(readTestFile(t, path), "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		switch {
		case strings.HasPrefix(trimmed, "workers:"):
			block.workers = strings.TrimSpace(strings.TrimPrefix(trimmed, "workers:"))
			inExcludeFiles = false
		case strings.HasPrefix(trimmed, "timeout-coefficient:"):
			block.timeoutCoefficient = strings.TrimSpace(strings.TrimPrefix(trimmed, "timeout-coefficient:"))
			inExcludeFiles = false
		case strings.HasPrefix(trimmed, "exclude-files:"):
			inExcludeFiles = true
		case inExcludeFiles && strings.HasPrefix(trimmed, "-"):
			block.excludeFiles = append(block.excludeFiles, strings.TrimSpace(strings.TrimPrefix(trimmed, "-")))
		default:
			inExcludeFiles = false
		}
	}
	return block
}

// TestGremlinsConfigsShareTheSameBlock checks that every .gremlins/*.yaml file
// repeats the same workers, timeout-coefficient and exclude-files. The fifteen
// package configs must not diverge: a shared setting changed in one file only
// makes one package measure differently from the others.
func TestGremlinsConfigsShareTheSameBlock(t *testing.T) {
	paths, err := filepath.Glob(gremlinsConfigGlob)
	if err != nil {
		t.Fatalf("glob %s: %v", gremlinsConfigGlob, err)
	}
	if len(paths) != 15 {
		t.Fatalf("%s matches %d files, want 15", gremlinsConfigGlob, len(paths))
	}
	reference := readGremlinsSharedBlock(t, paths[0])
	for _, path := range paths[1:] {
		if got := readGremlinsSharedBlock(t, path); !reflect.DeepEqual(got, reference) {
			t.Errorf("%s shared block = %+v, want %+v (as in %s)", path, got, reference, paths[0])
		}
	}
}
