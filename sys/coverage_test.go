// the coverage anchor for the sys layer NWG1000.
//
// symbols.txt is the authoritative export set of the full libnwep, checked in and
// regenerated from the built .so (see its header). this test diffs the nwep_*
// symbols the sys package actually calls (the C.nwep_* references in its source)
// against that set. it fails on a phantom (a call to a symbol the library does not
// export, a typo or a removed symbol that would fail at link) and on a hole (an
// exported symbol the sys layer never reaches), so coverage stays total and
// mechanical, not promised.

package sys

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

// a C.nwep_* reference in the go source, the function or type names sys names.
var cRef = regexp.MustCompile(`C\.(nwep_[a-z0-9_]+)`)

// a function prototype name in a header, an nwep_* token immediately before a (.
var protoName = regexp.MustCompile(`\b(nwep_[a-z0-9_]+)\s*\(`)

// a static inline helper in a header, callable from cgo but not an exported .so
// symbol, so it is allowed in the sys calls without appearing in symbols.txt.
var inlineHelper = regexp.MustCompile(`static\s+inline\s+\w[\w ]*\b(nwep_[a-z0-9_]+)\s*\(`)

// authoritative reads the checked-in export set from symbols.txt.
func authoritative(t *testing.T) map[string]bool {
	t.Helper()
	data, err := os.ReadFile("symbols.txt")
	if err != nil {
		t.Fatalf("read symbols.txt: %v", err)
	}
	set := map[string]bool{}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "#") {
			set[line] = true
		}
	}
	return set
}

// functionNames returns the header function names and, separately, the static
// inline helpers, so the sys references can be filtered to header functions and
// the inline helpers can be excused from the phantom check.
func functionNames(t *testing.T) (funcs, inline map[string]bool) {
	t.Helper()
	funcs, inline = map[string]bool{}, map[string]bool{}
	for _, h := range []string{"nwep.h", "nwep_trust.h"} {
		data, err := os.ReadFile(filepath.Join("..", "..", "..", "include", h))
		if err != nil {
			t.Fatalf("read %s: %v", h, err)
		}
		for _, m := range protoName.FindAllStringSubmatch(string(data), -1) {
			funcs[m[1]] = true
		}
		for _, m := range inlineHelper.FindAllStringSubmatch(string(data), -1) {
			inline[m[1]] = true
		}
	}
	return funcs, inline
}

// sysCalls collects the C.nwep_* references across the sys source, split into
// exported-function calls (the coverage set) and static-inline-helper calls (which
// are real but not in symbols.txt, so the phantom check excuses them).
func sysCalls(t *testing.T) (exported, inlineCalls map[string]bool) {
	t.Helper()
	funcs, inline := functionNames(t)
	exported, inlineCalls = map[string]bool{}, map[string]bool{}
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if !strings.HasSuffix(e.Name(), ".go") || strings.HasSuffix(e.Name(), "_test.go") {
			continue
		}
		data, err := os.ReadFile(e.Name())
		if err != nil {
			t.Fatal(err)
		}
		for _, m := range cRef.FindAllStringSubmatch(string(data), -1) {
			switch {
			case inline[m[1]]:
				inlineCalls[m[1]] = true
			case funcs[m[1]]: // keep only header functions, not c type names
				exported[m[1]] = true
			}
		}
	}
	return exported, inlineCalls
}

// TestNoPhantomCalls fails if sys calls a symbol the library does not export.
func TestNoPhantomCalls(t *testing.T) {
	auth := authoritative(t)
	exported, _ := sysCalls(t)
	var phantoms []string
	for name := range exported {
		if !auth[name] {
			phantoms = append(phantoms, name)
		}
	}
	sort.Strings(phantoms)
	if len(phantoms) != 0 {
		t.Fatalf("sys calls %v which are not exported by libnwep (typo or removed symbol)", phantoms)
	}
}

// TestCoverageIsTotal fails if any exported symbol is never reached by sys.
func TestCoverageIsTotal(t *testing.T) {
	auth := authoritative(t)
	calls, _ := sysCalls(t)
	var holes []string
	for name := range auth {
		if !calls[name] {
			holes = append(holes, name)
		}
	}
	sort.Strings(holes)
	if len(holes) != 0 {
		t.Fatalf("%d/%d symbols covered, the sys layer never reaches %v NWG1000", len(calls), len(auth), holes)
	}
	t.Logf("nwep/sys coverage: %d / %d symbols", len(calls), len(auth))
}

// TestAuthoritativeSize guards symbols.txt against accidental truncation.
func TestAuthoritativeSize(t *testing.T) {
	if n := len(authoritative(t)); n != 159 {
		t.Fatalf("symbols.txt has %d symbols, expected 159 (146 core + 13 trust)", n)
	}
}
