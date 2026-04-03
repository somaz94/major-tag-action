package tagger

import "errors"

// Sentinel errors for common failure modes.
var (
	// ErrInvalidTag indicates the tag does not match semver format (vX.Y.Z).
	ErrInvalidTag = errors.New("tag does not match semver format (expected vX.Y.Z)")

	// ErrAuthFailed indicates authentication configuration failure.
	ErrAuthFailed = errors.New("failed to configure authentication")

	// ErrTagUpdate indicates a tag update operation failure.
	ErrTagUpdate = errors.New("failed to update tag")

	// ErrInvalidSHA indicates an invalid commit SHA format.
	ErrInvalidSHA = errors.New("invalid commit SHA format")
)
