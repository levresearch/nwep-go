// the library version and low-level code helpers.

package nwep

import "nwep/sys"

// Version returns the libnwep version string, for example 0.1.0 (nwep_version).
func Version() string { return sys.Version() }

// MethodString returns the human name of a method (nwep_method_str).
func MethodString(m Method) string { return sys.MethodStr(int(m)) }

// StatusString returns the human name of a status code (nwep_status_str).
func StatusString(status int) string { return sys.StatusStr(status) }
