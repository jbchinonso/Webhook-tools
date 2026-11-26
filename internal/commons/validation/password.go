package validation

import (
	"unicode"
)

// passwordRules holds the composition rules. Kept as a struct so it can be
// changed in one place and, if you ever want, loaded from config.
type passwordRules struct {
	MinLength    int
	MaxLength    int
	NeedsUpper   bool
	NeedsLower   bool
	NeedsDigit   bool
	NeedsSpecial bool
}

// defaultPasswordRules is the composition policy for Webhook. Length floor
// is deliberately higher than the naive 8 — length dominates entropy.
var defaultPasswordRules = passwordRules{
	MinLength:    10,
	MaxLength:    72, // bcrypt's hard limit; longer is silently truncated
	NeedsUpper:   true,
	NeedsLower:   true,
	NeedsDigit:   true,
	NeedsSpecial: true,
}

// checkPassword validates a plaintext password against the policy.
// Returns (ok, reason). Reason is a short human-readable string suitable
// for surfacing in the API's error.details.
func checkPassword(pw string) (bool, string) {
	r := defaultPasswordRules

	if len(pw) < r.MinLength {
		return false, "must be at least 10 characters"
	}
	if len(pw) > r.MaxLength {
		return false, "must be at most 72 characters"
	}

	var hasUpper, hasLower, hasDigit, hasSpecial bool
	for _, c := range pw {
		switch {
		case unicode.IsUpper(c):
			hasUpper = true
		case unicode.IsLower(c):
			hasLower = true
		case unicode.IsDigit(c):
			hasDigit = true
		case unicode.IsPunct(c) || unicode.IsSymbol(c) || c == ' ':
			hasSpecial = true
		}
	}

	if r.NeedsUpper && !hasUpper {
		return false, "must contain at least one uppercase letter"
	}
	if r.NeedsLower && !hasLower {
		return false, "must contain at least one lowercase letter"
	}
	if r.NeedsDigit && !hasDigit {
		return false, "must contain at least one number"
	}
	if r.NeedsSpecial && !hasSpecial {
		return false, "must contain at least one special character"
	}

	return true, ""
}