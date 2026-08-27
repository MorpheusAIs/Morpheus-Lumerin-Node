package structs

import "github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"

const DefaultUserStakeIterations uint8 = 255

// QueryUserStakeIterations limits the number of on-hold entries inspected by
// the contract. A pointer lets the API distinguish an omitted value (use the
// default) from an explicit zero (reject it).
type QueryUserStakeIterations struct {
	Iterations *uint8 `form:"iterations" binding:"omitempty,gte=1" example:"255"`
}

type UserStakeWithdrawalRequest struct {
	Iterations *uint8 `json:"iterations" binding:"omitempty,gte=1" example:"255"`
}

type UserStakesOnHoldRes struct {
	Available *lib.BigInt `json:"available" example:"100000000" swaggertype:"string"`
	Hold      *lib.BigInt `json:"hold" example:"200000000" swaggertype:"string"`
}

func UserStakeIterationsOrDefault(iterations *uint8) uint8 {
	if iterations == nil {
		return DefaultUserStakeIterations
	}

	return *iterations
}
