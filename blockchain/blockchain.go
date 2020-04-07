package blockchain

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"gitee.com/assem/safe_device_blockchain/contracts"
	"github.com/KasperLiu/gobcos/accounts/abi"
	"github.com/KasperLiu/gobcos/client"
	"github.com/KasperLiu/gobcos/common"
	"github.com/tidwall/gjson"
)

var gc *client.Client

func Connect() {
	groupID := uint(1)
	c, err := client.Dial("http://127.0.0.1:8545", groupID) // change to your RPC URL and GroupID
	if err != nil {
		// handle err
		fmt.Println("connect error: ", err)
	}
	gc = c
}

// func ConnectTest() {
// 	groupID := uint(1)
// 	c, err := client.Dial("http://127.0.0.1:8545", groupID) // change to your RPC URL and GroupID
// 	if err != nil {
// 		// handle err
// 		fmt.Println("connect error: ", err)
// 	}
// 	bn, _ := client.GetBlockNumber(context.Background())
// 	fmt.Println(string(bn))
// 	// 创建私钥
// 	privateKey, err := crypto.GenerateKey()
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	// load the contract
// 	address := common.HexToAddress("0x1f494c56c3ad1e6738f3500d19499cd3541160ea") //contract addree in hex: 0x0626918C51A1F36c7ad4354BB1197460A533a2B9
// 	instance, err := contracts.NewContracts(address, c)
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	auth := bind.NewKeyedTransactor(privateKey)
// 	tx, err := instance.SetPassLog(auth, "123456")
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	fmt.Printf("tx sent: %s\n", tx.Hash().Hex())

// 	// wait for the mining
// 	receipt, err := bind.WaitMined(context.Background(), c, tx)
// 	if err != nil {
// 		log.Fatalf("tx mining error:%v\n", err)
// 	}
// 	fmt.Printf("transaction hash of receipt: %s\n", receipt.GetTransactionHash())

// 	// read the result
// 	opts := &bind.CallOpts{From: common.HexToAddress("0xFbb18d54e9Ee57529cda8c7c52242EFE879f064F")} // 0xFbb18d54e9Ee57529cda8c7c52242EFE879f064F
// 	result, err := instance.GetPassLog(opts)
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	fmt.Println(string(result[:])) // "bar"
// }

// 获取主页展示信息
func GetIndexData() interface{} {
	transCount, _ := gc.GetTotalTransactionCount(context.Background())
	transCountJSON := gjson.ParseBytes(transCount)
	blockNumber := transCountJSON.Get("blockNumber").String()
	txSumHex := transCountJSON.Get("txSum").String()
	txSum, _ := strconv.ParseUint(txSumHex[2:], 16, 32)
	failedTxSum := 0
	pbftView, _ := gc.GetPBFTView(context.Background())
	// # print("首页数据:", blockNumber, txSum, failedTxSum, pbftView)
	consensusStatus, _ := gc.GetConsensusStatus(context.Background())
	obj := gjson.ParseBytes(consensusStatus)
	node := obj.Array()
	highestblockNumber := node[0].Get("highestblockNumber")
	nodeList := node[1].Array()
	// fmt.Println("nodeList: ", len(nodeList))
	var newNodeList []map[string]string // 节点信息
	for _, v := range nodeList {
		// fmt.Println(k, v, v.Get("nodeId").String())
		var m = make(map[string]string)
		m["nodeId"] = v.Get("nodeId").String()
		m["view"] = v.Get("view").String()
		m["highestblockNumber"] = highestblockNumber.String()
		m["static"] = "正常"
		newNodeList = append(newNodeList, m)
	}

	highBlockNum, _ := strconv.ParseUint(blockNumber[2:len(blockNumber)], 16, 32)
	fmt.Println("highBlockNum: ", highBlockNum)
	var newTransactionList []map[string]interface{}
	var newBlockList []map[string]interface{} //区块列表
	var trsCountList [15]int                  // 统计日交易数量
	trsCountListNum := 15                     // 统计15天的交易数量
	// 获取最近15天日期
	nowTime := time.Now().Unix()
	var dataList []string
	for i := 0; i < 15; i++ {
		_, m, d := time.Unix(nowTime, 0).Date()
		dm := strconv.Itoa(int(m)) + "-" + strconv.Itoa(d)
		dataList = append(dataList, dm)
		nowTime -= 86400
	}
	for i, j := 0, len(dataList)-1; i < j; i, j = i+1, j-1 {
		dataList[i], dataList[j] = dataList[j], dataList[i]
	}

	for i := highBlockNum; i > 0; i-- {
		// # if get_block_Number >= 0:
		num := "0x" + strconv.Itoa(int(i))
		res, _ := gc.GetBlockByNumber(context.Background(), num, true)
		bkJSON := gjson.ParseBytes(res)
		//获取区块时间
		timestamps := bkJSON.Get("timestamp").String()
		timestamp, _ := strconv.ParseInt(timestamps[2:], 16, 64)
		blockTime := time.Unix(timestamp/1000, 0).Format("2006-01-02 15:04:05")

		// 交易信息
		transactions := bkJSON.Get("transactions").Array()

		// 根据日期存储交易数量
		_, mth, day := time.Unix(timestamp/1000, 0).Date()
		dm := strconv.Itoa(int(mth)) + "-" + strconv.Itoa(day)
		fmt.Println("dm: ", dm, "  dataList[trsCountListNum-1]", dataList[trsCountListNum-1])
		for j := 0; j < trsCountListNum; j++ {
			if dm == dataList[j] {
				fmt.Println("len(transactions): ", len(transactions))
				trsCountList[j] = len(transactions)
			}
		}
		//
		var m = make(map[string]interface{})
		for _, v := range transactions {
			// fmt.Println(k, v)
			for k1, v1 := range v.Map() {
				m[k1] = v1.String()
			}

			bn := v.Get("blockNumber").String()
			m["blockNumber"], _ = strconv.ParseUint(bn[2:], 16, 32)
			tix := v.Get("transactionIndex").String()
			m["transactionIndex"], _ = strconv.ParseUint(string(tix)[2:], 16, 32)

			// loc, _ := time.LoadLocation("Asia/Shanghai") //设置时区
		}
		m["time"] = blockTime
		newTransactionList = append(newTransactionList, m)

		// 区块列表
		oneBlockMsg := map[string]interface{}{
			"block_num":       i,
			"block_hash":      bkJSON.Get("hash").String(),
			"time":            blockTime,
			"transaction_num": len(transactions),
		}
		newBlockList = append(newBlockList, oneBlockMsg)
	}
	pbftViewNum, _ := strconv.ParseUint(string(pbftView)[3:len(pbftView)-1], 16, 32)
	result := map[string]interface{}{
		"blockNumber":      highBlockNum,
		"txSum":            txSum,
		"failedTxSum":      failedTxSum,
		"pbftView":         pbftViewNum,
		"nodeList":         newNodeList,
		"block_list":       newBlockList,
		"transaction_list": newTransactionList,
		"data_list":        dataList,
		"count_list":       trsCountList,
	}
	// fmt.Println(result)
	return result
}

//
func GetTransactionListData(blockNumber, transactionHash, page string) interface{} {
	// var transaction_list []map[string]interface{}
	fmt.Println(len(transactionHash))
	if len(blockNumber) > 0 {
		res, _ := gc.GetBlockByNumber(context.Background(), blockNumber, true)
		bkJSON := gjson.ParseBytes(res) // 交易信息

		//获取区块时间
		timestamps := bkJSON.Get("timestamp").String()
		timestamp, _ := strconv.ParseInt(timestamps[2:], 16, 64)
		blockTime := time.Unix(timestamp/1000, 0).Format("2006-01-02 15:04:05")
		// 获取区块交易信息
		transactions := bkJSON.Get("transactions").Array()
		var newTransactionList []map[string]interface{}
		for _, v := range transactions {
			var m = make(map[string]interface{})
			bn := v.Get("blockNumber").String()
			m["blockNumber"], _ = strconv.ParseUint(bn[2:len(bn)], 16, 32)
			tix := v.Get("transactionIndex").String()
			m["transactionIndex"], _ = strconv.ParseUint(string(tix)[2:], 16, 32)
			m["time"] = blockTime

			newTransactionList = append(newTransactionList, m)
		}
		result := map[string]interface{}{
			"transaction_num":  1,
			"transaction_list": newTransactionList,
		}
		return result
	} else if len(transactionHash) > 0 {
		res, _ := gc.GetTransactionByHash(context.Background(), transactionHash)
		trJSON := gjson.ParseBytes(res) // 交易信息
		// transactions := trJSON.Get("transactions").Array()
		bn := trJSON.Get("blockNumber").String()
		bni, _ := strconv.ParseUint(bn[2:len(bn)], 16, 32)
		tix := trJSON.Get("transactionIndex").String()
		transactionIndex, _ := strconv.ParseUint(tix[2:], 16, 32)
		timestamps := trJSON.Get("timestamp").String()
		timestamp, _ := strconv.ParseInt(timestamps[2:], 16, 64)
		blockTime := time.Unix(timestamp/1000, 0).Format("2006-01-02 15:04:05")
		result := map[string]interface{}{
			"transaction_num": 1,
			"transaction_list": map[string]interface{}{
				"hash":             trJSON.Get("hash").String(),
				"from":             trJSON.Get("from").String(),
				"to":               trJSON.Get("to").String(),
				"blockNumber":      bni,
				"transactionIndex": transactionIndex,
				"time":             blockTime,
			},
		}
		fmt.Println(result)
		return result
	} else {
		transCount, _ := gc.GetTotalTransactionCount(context.Background())
		transCountJSON := gjson.ParseBytes(transCount)
		blockNumber := transCountJSON.Get("blockNumber").String()
		highBlockNum, _ := strconv.ParseUint(blockNumber[2:], 16, 32)
		txSum := transCountJSON.Get("txSum").String()
		transactionNumber, _ := strconv.ParseUint(txSum[2:], 16, 32)
		page, _ := strconv.Atoi(page)
		// # get_block_Number = blockNumber - 10 * (int(page) - 1)
		var newTransactionList []map[string]interface{}
		for x := highBlockNum; x > 0; x-- {
			// # if get_block_Number >= 0:
			num := "0x" + strconv.FormatInt(int64(x), 16)
			fmt.Println("GetBlockByNumber:", num)
			res, _ := gc.GetBlockByNumber(context.Background(), num, true)
			bkJSON := gjson.ParseBytes(res)
			transactions := bkJSON.Get("transactions").Array()

			timestamps := bkJSON.Get("timestamp").String()
			timestamp, _ := strconv.ParseInt(timestamps[2:len(timestamps)], 16, 64)
			blockTime := time.Unix(timestamp/1000, 0).Format("2006-01-02 15:04:05")
			//
			for _, v := range transactions {
				// fmt.Println(k, v)
				var m = make(map[string]interface{})
				for k1, v1 := range v.Map() {
					m[k1] = v1.String()
				}

				bn := v.Get("blockNumber").String()
				m["blockNumber"], _ = strconv.ParseUint(bn[2:len(bn)], 16, 32)
				tix := v.Get("transactionIndex").String()
				m["transactionIndex"], _ = strconv.ParseUint(string(tix)[2:len(tix)], 16, 32)
				m["time"] = blockTime
				newTransactionList = append(newTransactionList, m)

				// loc, _ := time.LoadLocation("Asia/Shanghai") //设置时区
			}
		}
		result := map[string]interface{}{
			"high_block_num":   highBlockNum,
			"transaction_num":  transactionNumber,
			"transaction_list": newTransactionList[10*(page-1) : len(newTransactionList)%10*page],
		}
		// fmt.Println(result)
		return result
	}
}

func GetTransactionDetailData(hash string) interface{} {
	res, _ := gc.GetTransactionByHash(context.Background(), hash)
	trJSON := gjson.ParseBytes(res) // 交易信息
	inputHex := trJSON.Get("input").String()
	data := inputParser(contracts.ContractsABI, inputHex)

	transactionReceipt, _ := gc.GetTransactionReceipt(context.Background(), hash)
	inputHexRci := transactionReceipt.Input

	receiptData := inputParser(contracts.ContractsABI, inputHexRci)
	result := map[string]interface{}{
		"txresponse":         trJSON.Value(),
		"transactionReceipt": transactionReceipt,
		"chain_data":         data,
		"receipt_data":       receiptData,
	}
	return result
}

func inputParser(abiString, inputString string) []interface{} {
	cd := common.Hex2Bytes(inputString[2:])

	abispec, err := abi.JSON(strings.NewReader(abiString))
	if err != nil {
		log.Fatal(err)
	}
	sigdata, argdata := cd[:4], cd[4:]
	method, err := abispec.MethodById(sigdata)
	if err != nil {
		log.Fatal(err)
	}
	data, err := method.Inputs.UnpackValues(argdata)
	if err != nil {
		log.Fatal(err)
	}
	// fmt.Println(data)
	return data
}
