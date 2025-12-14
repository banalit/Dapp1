// SPDX-License-Identifier: MIT
pragma solidity ^0.8;

contract MyCounter {

    uint256 public counter;
    event AddCounter(address sender, uint256 counter);

    function addCounter() public returns(uint256) {
        counter ++;
        emit AddCounter(msg.sender, counter);
        return counter;
    }

}