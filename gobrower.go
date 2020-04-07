// http_server project main.go
package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"gitee.com/assem/safe_device_blockchain/blockchain"
)

var (
	port    = "8889"     //默认端口
	wwwroot = "./static" //网站根目录
)

func main() {
	blockchain.Connect()
	// blockchain.GetIndexData()
	startHTTP()
}

func startHTTP() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", IndexHandler)
	log.Println("HTTP Server Listening to 0.0.0.0: " + port)
	server := &http.Server{Addr: "localhost:" + port, Handler: mux}
	err := server.ListenAndServe()
	if err != nil {
		log.Fatal("ListenAndServe: ", err.Error())
	}
}

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	// 解析form
	err := r.ParseForm()
	if err != nil {
		w.WriteHeader(400)
		return
	}
	// 检查是否POST请求
	if r.Method == "GET" {
		err := r.ParseForm()
		if err != nil {
			w.WriteHeader(400)
			return
		}
		switch r.URL.Path {
		case "/query_info/index": //# 获取主页的数据，包括预览，交易量图，最近的区块和交易
			result := blockchain.GetIndexData()
			json.NewEncoder(w).Encode(result)
		case "/query_info/transaction_list": //根据交易哈希，获取交易详情和交易回执
			log.Println("transactionHash: ", r.Form.Get("transactionHash"))
			log.Println("page: ", r.Form.Get("page"))
			log.Println("blockNumber: ", r.Form.Get("blockNumber"))
			result := blockchain.GetTransactionListData(r.Form.Get("blockNumber"), r.Form.Get("transactionHash"), r.Form.Get("page"))
			json.NewEncoder(w).Encode(result)
		case "/query_info/transaction_detail": //根据交易哈希，获取交易详情和交易回执
			log.Println("transactionHash: ", r.Form.Get("transactionHash"))
			result := blockchain.GetTransactionDetailData(r.Form.Get("transactionHash"))
			json.NewEncoder(w).Encode(result)
			// transactionHash = request.args.get('transactionHash', '')
			// result = get_transaction_detail_data(transactionHash)
		default:
			if strings.HasPrefix(r.URL.String(), "/") {
				had := http.StripPrefix("/", http.FileServer(http.Dir(wwwroot)))
				had.ServeHTTP(w, r)
			} else {
				http.Error(w, "404 not found", 404)
			}
		}

	}

}

func getHtmlFile(path string) (fileHtml string) {
	dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	realPath := dir + "/" + wwwroot + "/" + path
	if PathExists(realPath) {
		file, err := os.Open(realPath)
		if err != nil {
			panic(err)
		}
		defer file.Close()
		fileContent, _ := ioutil.ReadAll(file)
		fileHtml = string(fileContent)

	} else {
		fileHtml = "404 page not found"
	}

	return fileHtml
}

func PathExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return false
}
