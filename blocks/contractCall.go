package blocks

import (
	"fmt"
	"log"
	"math/big"

	"context"
	"crypto/ecdsa"

	counter "eth-demo/contract"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

func ContractCall() {
	client, err := ethclient.Dial("https://eth-sepolia.g.alchemy.com/v2/vlr5uknz65163GXJIeUoE")
	if err != nil {
		log.Fatal(err)
	}
	contractAddr := "0xDd0f22bA35281CEacc64dD20B948F6070409ed3F"
	tokenAddress := common.HexToAddress(contractAddr)
	instance, err := counter.NewContract(tokenAddress, client)
	if err != nil {
		log.Fatal(err)
	}
	counterNum, err := instance.Counter(&bind.CallOpts{})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("counterNum:", counterNum)

	//======write data
	privateKey, err := crypto.HexToECDSA("privatekey")
	if err != nil {
		log.Fatal(err)
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Fatal("cannot assert type: publicKey is not of type *ecdsa.PublicKey")
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		log.Fatal(err)
	}

	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	// 手动设置 GasLimit（避免 EstimateGas 触发栈溢出，或正确构造 EstimateGas 参数）
	gasLimit := uint64(300000) // 足够调用 AddCounter() 的 Gas 上限

	// 4.3 正确构造合约调用的 Data 字段（核心修复！）
	// 方式1：使用生成的绑定方法（推荐，避免手动组装错误）
	// 生成交易签名器
	chainId, err := client.ChainID(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("chainId:", chainId)
	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, chainId) // Sepolia 链 ID=11155111
	if err != nil {
		log.Fatal("创建签名器失败：", err)
	}
	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0) // 转账金额（无 ETH 转账则为 0）
	auth.GasLimit = gasLimit   // Gas 上限
	auth.GasPrice = gasPrice   // Gas 价格

	// 调用合约的 AddCounter 方法（使用绑定的方法，自动组装 Data）
	tx, err := instance.AddCounter(auth)
	if err != nil {
		log.Fatal("调用 AddCounter 失败：", err)
	}
	fmt.Printf("交易已发送，哈希：%s\n", tx.Hash().Hex())

	// 等待交易上链（可选，验证结果）
	receipt, err := bind.WaitMined(context.Background(), client, tx)
	if err != nil {
		log.Fatal("等待交易上链失败：", err)
	}
	fmt.Printf("交易已上链，区块高度：%d，状态：%d\n", receipt.BlockNumber, receipt.Status)

	counterNum, err = instance.Counter(&bind.CallOpts{})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("counterNum:", counterNum)
}
