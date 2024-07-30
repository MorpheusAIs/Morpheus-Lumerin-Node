// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.0;

import { ERC20 } from "@openzeppelin/contracts/token/ERC20/ERC20.sol";

contract LumerinToken is ERC20 {
  uint256 constant initialSupply = 1_000_000_000 * (10 ** 8);

  constructor() ERC20("Lumerin dev", "LMR") {
    _mint(msg.sender, initialSupply);
  }

  function decimals() public pure override returns (uint8) {
    return 8;
  }
}
