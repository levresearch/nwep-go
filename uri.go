// the uri type, a parsed web://nodeid[:port]/path reference NW110900.

package nwep

import "github.com/levresearch/nwep-go/sys"

// URI is a parsed web/1 reference, the target node, an optional port, and a path.
type URI struct {
	// NodeID is the target node the reference names.
	NodeID NodeID
	// Port is the explicit port, or 0 when the reference omitted one.
	Port uint16
	// Path is the request path, beginning with a slash NW040400.
	Path string
}

// ParseURI parses a web://nodeid_base58[:port]/path reference (nwep_uri_parse).
//
// returns the parsed URI.
// errors with a protocol error when input is not a valid web/1 reference.
func ParseURI(input string) (URI, error) {
	id, port, path, rc := sys.URIParse(input)
	if err := check(rc); err != nil {
		return URI{}, err
	}
	return URI{NodeID: NodeID(id), Port: port, Path: path}, nil
}
