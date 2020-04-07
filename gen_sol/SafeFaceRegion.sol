pragma solidity ^0.4.24;

contract SafeFaceRegion{
    string passLog;

    constructor() public{
    //    passLog = "Hello, World!";
    }

    function getPassLog() constant public returns(string){
        return passLog;
    }

    function setPassLog(string n) public{
    	passLog = n;
    }
}
