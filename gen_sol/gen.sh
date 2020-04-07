#! /bin/sh
string=$1
obj=${string%.sol}
echo $obj
# 安装solc https://github.com/ethereum/solidity/releases/tag/v0.4.25
solc --bin -o ./ $1 && solc --abi -o ./ $1

abigen --bin=$obj.bin --abi=$obj.abi --pkg=contracts --out=../contracts/$obj.go