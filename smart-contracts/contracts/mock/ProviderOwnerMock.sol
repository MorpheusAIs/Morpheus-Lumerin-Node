// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import {Address} from "@openzeppelin/contracts/utils/Address.sol";

contract ProviderOwnerMock {
    using Address for address;

    enum OwnerMode {
        Valid,
        Revert,
        ShortReturn,
        DirtyReturn,
        GasBurn
    }

    address public immutable ownerAddress;
    OwnerMode public mode;
    uint256 public burnGas;

    constructor(address owner_) {
        ownerAddress = owner_;
    }

    function setMode(OwnerMode mode_, uint256 burnGas_) external {
        mode = mode_;
        burnGas = burnGas_;
    }

    function execute(address target_, bytes calldata data_) external returns (bytes memory) {
        return target_.functionCall(data_);
    }

    function owner() external view returns (address) {
        OwnerMode mode_ = mode;
        address owner_ = ownerAddress;

        if (mode_ == OwnerMode.Revert) {
            revert("ProviderOwnerMock: owner");
        }
        if (mode_ == OwnerMode.ShortReturn) {
            assembly {
                mstore(0, owner_)
                return(0, 31)
            }
        }
        if (mode_ == OwnerMode.DirtyReturn) {
            assembly {
                mstore(0, or(owner_, shl(160, 1)))
                return(0, 32)
            }
        }
        if (mode_ == OwnerMode.GasBurn) {
            uint256 start_ = gasleft();
            uint256 burnGas_ = burnGas;
            while (start_ - gasleft() < burnGas_) {}
        }

        return owner_;
    }
}
