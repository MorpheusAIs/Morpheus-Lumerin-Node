//SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

struct AppStorage {
  //
  // OTHER
  //
  // Number of seconds to delay the stake return when a user closes out a session using a user signed receipt
  int256 stakeDelay;
}
