// shamir secret sharing, splitting a recovery key into shares NW150400.

package nwep

import "github.com/levresearch/nwep-go/sys"

// SplitSecret splits secret into shares pieces, any threshold of which recombine it.
//
// the intended use is splitting an offline recovery private key NW150400.
// returns the concatenated share bytes, hand them to CombineShares with the same
// share length to recover the secret.
// errors when threshold exceeds shares or the parameters are out of range.
func SplitSecret(secret []byte, threshold, shares int) ([]byte, error) {
	out, rc := sys.ShamirSplit(secret, threshold, shares)
	if err := check(rc); err != nil {
		return nil, err
	}
	return out, nil
}

// CombineShares recombines count concatenated shares of shareLen bytes each.
//
// returns the recovered secret. at least the original threshold of shares must be
// present for a correct result.
// errors when the shares are inconsistent or too few.
func CombineShares(shares []byte, count, shareLen int) ([]byte, error) {
	out, rc := sys.ShamirCombine(shares, count, shareLen)
	if err := check(rc); err != nil {
		return nil, err
	}
	return out, nil
}
