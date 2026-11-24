package registries

import (
	"errors"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
)

// ErrStage1NotDeployed is returned by the ProviderPayouts, DelegateStaking and
// DelegateStakingFunding wrappers when the diamond does not route their
// functions because it has not been cut to Stage 1.
var ErrStage1NotDeployed = errors.New("Stage 1 contracts are not deployed on this diamond")

func stage1Error(err error) error {
	if lib.IsSelectorNotRegistered(err) {
		return lib.WrapError(ErrStage1NotDeployed, err)
	}
	return lib.TryConvertGethError(err)
}
