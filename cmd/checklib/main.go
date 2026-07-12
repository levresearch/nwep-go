// cmd/checklib reports whether the nwep native library is available for the
// Go binding and prints the exact build command to use.
//
// Usage:
//
//	go run nwep/cmd/checklib
package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	osLinux   = "linux"
	osWindows = "windows"
)

func main() {
	// NWEP_LIB_DIR: set by the Windows installer, or a manual override on any platform.
	if override := os.Getenv("NWEP_LIB_DIR"); override != "" {
		if found := probe(override); found != "" {
			fmt.Printf("found (NWEP_LIB_DIR): %s\n", found)
			printBuildCmd(override)
			return
		}
		fmt.Fprintf(os.Stderr, "checklib: NWEP_LIB_DIR=%q is set but no nwep library found there\n", override)
		os.Exit(1)
	}

	// Linux: try pkg-config first - the installer writes nwep.pc.
	if runtime.GOOS == osLinux {
		if dir, flags := pkgConfigResult(); dir != "" {
			fmt.Printf("found via pkg-config: %s\n", dir)
			fmt.Println()
			if strings.Contains(dir, ".local") {
				// user install needs PKG_CONFIG_PATH set explicitly.
				fmt.Printf("build with:\n\n  PKG_CONFIG_PATH=%s/pkgconfig \\\n  go build -tags nwep_pkgconfig ./...\n\n",
					dir)
			} else {
				fmt.Printf("build with:\n\n  go build -tags nwep_pkgconfig ./...\n\n")
			}
			_ = flags
			return
		}
	}

	// Probe known installer default directories.
	systemDirs, userDirs := installerDefaultDirs()
	for _, dir := range append(systemDirs, userDirs...) {
		if found := probe(dir); found != "" {
			fmt.Printf("found: %s\n", found)
			if isUserDir(dir, userDirs) {
				printBuildCmd(dir)
			}
			return
		}
	}

	printNotFound(systemDirs, userDirs)
	os.Exit(1)
}

func pkgConfigResult() (libdir, flags string) {
	out, err := exec.Command("pkg-config", "--variable=libdir", "github.com/levresearch/nwep-go").Output()
	if err != nil {
		return "", ""
	}
	libdir = strings.TrimSpace(string(out))
	if libdir == "" {
		return "", ""
	}
	flagsOut, err := exec.Command("pkg-config", "--libs", "github.com/levresearch/nwep-go").Output()
	if err != nil {
		return libdir, ""
	}
	return libdir, strings.TrimSpace(string(flagsOut))
}

func probe(dir string) string {
	for _, name := range libCandidates() {
		p := filepath.Join(dir, name)
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return ""
}

func libCandidates() []string {
	switch runtime.GOOS {
	case osWindows:
		return []string{"nwep.dll", "libnwep.a", "libnwep-full.a"}
	default:
		return []string{"libnwep.so", "libnwep.a", "libnwep-full.a"}
	}
}

func installerDefaultDirs() (system, user []string) {
	switch runtime.GOOS {
	case osLinux:
		system = []string{"/usr/local/lib"}
		if home, err := os.UserHomeDir(); err == nil {
			user = []string{filepath.Join(home, ".local", "lib")}
		}
	case osWindows:
		if lad := os.Getenv("LOCALAPPDATA"); lad != "" {
			user = []string{
				filepath.Join(lad, "Programs", "github.com/levresearch/nwep-go", "lib"),
				filepath.Join(lad, "Programs", "github.com/levresearch/nwep-go", "bin"),
			}
		}
		if pf := os.Getenv("ProgramFiles"); pf != "" {
			system = []string{
				filepath.Join(pf, "github.com/levresearch/nwep-go", "lib"),
				filepath.Join(pf, "github.com/levresearch/nwep-go", "bin"),
			}
		}
	}
	return
}

func isUserDir(dir string, userDirs []string) bool {
	for _, u := range userDirs {
		if dir == u {
			return true
		}
	}
	return false
}

func printBuildCmd(libDir string) {
	fmt.Println()
	switch runtime.GOOS {
	case osWindows:
		fmt.Printf("build with:\n\n  set CGO_LDFLAGS=-L%s\n  go build ./...\n\n", libDir)
	default:
		fmt.Printf("note: %s is not in the default linker search path.\n"+
			"build with:\n\n  export CGO_LDFLAGS=\"-L%s\"\n  go build ./...\n\n", libDir, libDir)
	}
}

func printNotFound(systemDirs, userDirs []string) {
	fmt.Fprintln(os.Stderr, "checklib: nwep library not found.")
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "Install nwep with the GUI installer (https://rebuildtheinter.net/install),")
	fmt.Fprintln(os.Stderr, "or build it yourself with `zig build` in the nwep repo root.")
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "Expected install locations:")
	for _, d := range append(userDirs, systemDirs...) {
		fmt.Fprintf(os.Stderr, "  %s\n", d)
	}
	if runtime.GOOS == osWindows {
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "The installer also sets NWEP_LIB_DIR automatically - if it's not set,")
		fmt.Fprintln(os.Stderr, "try opening a new terminal after installing.")
	}
	if runtime.GOOS == osLinux {
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "For a user install, also try:")
		fmt.Fprintln(os.Stderr, "  PKG_CONFIG_PATH=~/.local/lib/pkgconfig go run nwep/cmd/checklib")
	}
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "If the library is in a non-standard location, set NWEP_LIB_DIR.")
}
