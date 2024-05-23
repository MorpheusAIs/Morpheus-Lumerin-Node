// SPDX-License-Identifier: GPL-3.0

pragma solidity >=0.8.2 <0.9.0;

import {Initializable} from "@openzeppelin/contracts-upgradeable/proxy/utils/Initializable.sol";
import {SafeERC20} from "@openzeppelin/contracts/token/ERC20/utils/SafeERC20.sol";
import {IERC20} from "@openzeppelin/contracts/token/ERC20/IERC20.sol";

contract LumerinMorpheusVault is Initializable {

    modifier onlyOwner()
    {
        require(msg.sender == owner, "you are not authorized");
        _;
    }

    using SafeERC20 for IERC20;
    address owner;
    IERC20 lumerin;
    IERC20 morpheus;

    mapping(address => uint) public staked;
    mapping(address => bool) public member;
    address[] public stakers;
    mapping(address => uint) public claimable;
    uint public total_claimable;
    uint public total_staked;

    function initialize(address arblmr, address arbmor)
        public initializer
    {
        owner = msg.sender;
        lumerin = IERC20(arblmr);
        morpheus = IERC20(arbmor);
    }

    function stake(uint amount)
        public
    {
        lumerin.safeTransferFrom(msg.sender, address(this), amount);
        if (!member[msg.sender]) {
            stakers.push(msg.sender);
            member[msg.sender] = true;
        }
        staked[msg.sender] += amount;
        total_staked += amount;
    }

    function unstake(uint amount)
        public
    {
        require(staked[msg.sender] >= amount, "Not enough staked");
        lumerin.safeTransferFrom(address(this), msg.sender, amount);
        staked[msg.sender] -= amount;
        total_staked -= amount;
    }

    function staked_amount(address staker)
        public view returns (uint)
    {
        if (staker != address(0)) {
            return staked[staker];
        }
        return total_staked;
    }

    function claimable_amount(address staker)
        public view returns (uint)
    {
        if (staker != address(0)) {
            return claimable[staker];
        }
        return total_claimable;
    }

    function claim()
        public
    {
        uint amount = claimable[msg.sender];
        require(amount > 0, "No claimable reward");
        morpheus.safeTransferFrom(address(this), msg.sender, amount);
        claimable[msg.sender] = 0;
        total_claimable -= amount;
    }

    function distribute(uint amount)
        public onlyOwner
    {
        for (uint i=0; i<stakers.length; i++) {
            uint reward = staked[stakers[i]]/total_staked*amount;
            claimable[stakers[i]] += reward;
            total_claimable += reward;
        }
        require(morpheus.balanceOf(address(this)) >= total_claimable,
            "Funds required");
    }
}

