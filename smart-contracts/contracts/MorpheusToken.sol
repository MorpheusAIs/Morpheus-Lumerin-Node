// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.0;

import { ERC20 } from "@openzeppelin/contracts/token/ERC20/ERC20.sol";

contract MorpheusToken is ERC20 {
  uint256 constant initialSupply = 1000000 * (10 ** 18);

  constructor() ERC20("Morpheus dev", "MOR") {
    _mint(msg.sender, initialSupply);
  }
}
