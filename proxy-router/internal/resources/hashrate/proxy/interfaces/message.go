package interfaces

// Placed in the separate package to avoid circular dependencies
// when using compile-time checks in stratumv1_message package

type MiningMessageGeneric interface {
	Serialize() []byte
}

type MiningMessageWithID interface {
	MiningMessageGeneric
	GetID() int
	SetID(int)
}
