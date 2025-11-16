// Package certerr provides typed errors and helpers for certificate-related
// operations across the repository. It standardizes error construction and
// matching so callers can reliably branch on error source/kind using the
// Go 1.13+ `errors.Is` and `errors.As` helpers.
//
// Guidelines
//   - Always wrap underlying causes using the helper constructors or with
//     fmt.Errorf("context: %w", err).
//   - Do not include sensitive data (keys, passwords, tokens) in error
//     messages; add only non-sensitive, actionable context.
//   - Prefer programmatic checks via errors.Is (for sentinel errors) and
//     errors.As (to retrieve *certerr.Error) rather than relying on error
//     string contents.
//
// Typical usage
//
//	if err := doParse(); err != nil {
//	    return certerr.ParsingError(certerr.ErrorSourceCertificate, err)
//	}
//
// Callers may branch on error kinds and sources:
//
//	var e *certerr.Error
//	if errors.As(err, &e) {
//	    switch e.Kind {
//	    case certerr.KindParse:
//	        // handle parse error
//	    }
//	}
//
// Sentinel errors are provided for common conditions like
// `certerr.ErrEncryptedPrivateKey` and can be matched with `errors.Is`.
package certerr
