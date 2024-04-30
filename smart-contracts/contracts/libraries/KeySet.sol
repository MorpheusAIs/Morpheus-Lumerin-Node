// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.24;

// inspired by https://github.com/rob-Hitchens/UnorderedKeySet/blob/master/contracts/HitchensUnorderedKeySet.sol
library KeySet {
    struct Set {
        mapping(bytes32 => uint) keyPointers;
        bytes32[] keyList;
    }

    error ZeroKey();
    error KeyExists();
    error KeyNotFound();

    function insert(Set storage self, bytes32 key) internal {
        if (key == 0x0) {
            revert ZeroKey();
        }
        if (exists(self, key)) {
            revert KeyExists();
        }

        self.keyPointers[key] = self.keyList.length;
        self.keyList.push(key);
    }

    function remove(Set storage self, bytes32 key) internal {
        if (!exists(self, key)) {
            revert KeyNotFound();
        }
        
        bytes32 keyToMove = self.keyList[count(self)-1];
        uint rowToReplace = self.keyPointers[key];
        self.keyPointers[keyToMove] = rowToReplace;
        self.keyList[rowToReplace] = keyToMove;
        delete self.keyPointers[key];
        self.keyList.pop();
    }

    function count(Set storage self) internal view returns (uint) {
        return (self.keyList.length);
    }

    function exists(
        Set storage self,
        bytes32 key
    ) internal view returns (bool) {
        if (self.keyList.length == 0) return false;
        return self.keyList[self.keyPointers[key]] == key;
    }

    function keyAtIndex(
        Set storage self,
        uint index
    ) internal view returns (bytes32) {
        return self.keyList[index];
    }

    function keys(Set storage self) internal view returns (bytes32[] memory) {
        return self.keyList;
    }
}

library AddressSet {
    struct Set {
        mapping(address => uint) keyPointers;
        address[] keyList;
    }

    error ZeroKey();
    error KeyExists();
    error KeyNotFound();

    function insert(Set storage self, address key) internal {
        if (key == address(0)) {
            revert ZeroKey();
        }
        if (exists(self, key)) {
            revert KeyExists();
        }

        self.keyPointers[key] = self.keyList.length;
        self.keyList.push(key);
    }

    function remove(Set storage self, address key) internal {
        if (!exists(self, key)) {
            revert KeyNotFound();
        }

        address keyToMove = self.keyList[count(self)-1];
        uint rowToReplace = self.keyPointers[key];
        self.keyPointers[keyToMove] = rowToReplace;
        self.keyList[rowToReplace] = keyToMove;
        delete self.keyPointers[key];
        self.keyList.pop();
    }

    function count(Set storage self) internal view returns (uint) {
        return (self.keyList.length);
    }

    function exists(
        Set storage self,
        address key
    ) internal view returns (bool) {
        if (self.keyList.length == 0) return false;
        return self.keyList[self.keyPointers[key]] == key;
    }

    function keyAtIndex(
        Set storage self,
        uint index
    ) internal view returns (address) {
        return self.keyList[index];
    }

    function keys(Set storage self) internal view returns (address[] memory) {
        return self.keyList;
    }
}
