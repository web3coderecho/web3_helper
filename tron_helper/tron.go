package tron

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/fbsobreira/gotron-sdk/pkg/client"
	"github.com/fbsobreira/gotron-sdk/pkg/common"
	"github.com/fbsobreira/gotron-sdk/pkg/proto/api"
	"github.com/fbsobreira/gotron-sdk/pkg/proto/core"
	"github.com/golang/protobuf/proto"
	"github.com/shopspring/decimal"
	"google.golang.org/grpc"
)

type Tron struct {
	TronJsonRpc   string
	TronApi       string
	TronProApiKey string
	apiKey        int
}

func NewTron(tronJsonRpc, tronApi, tronProApiKey string) *Tron {
	return &Tron{
		TronJsonRpc:   tronJsonRpc,
		TronApi:       tronApi,
		TronProApiKey: tronProApiKey,
	}
}

func (t *Tron) NewTronJsonRpcClient(ctx context.Context) *rpc.Client {
	rpcClient, err := rpc.DialOptions(ctx, t.TronJsonRpc, rpc.WithHTTPClient(&http.Client{
		Timeout: 120 * time.Second,
	}))
	if err != nil {
		panic(err)
	}
	rpcClient.SetHeader("Content-Type", "application/json")
	rpcClient.SetHeader("TRON-PRO-API-KEY", t.TronProApiKey)
	return rpcClient
}

func (t *Tron) GetLogs(ctx context.Context, params map[string]interface{}) []types.Log {
	var result []types.Log
	grpcClient := t.NewTronJsonRpcClient(ctx)
	defer grpcClient.Close()
	if err := grpcClient.CallContext(ctx, &result, "eth_getLogs", params); err != nil {
		log.Fatalf(fmt.Sprintf("failed to get logs: %v", err))
	}
	if len(result) > 0 {
		return result
	}
	return nil
}

func (t *Tron) GetBlockNumber(ctx context.Context) (int64, error) {
	grpcClient := t.GetGrpcClient()
	defer grpcClient.Stop()
	block, err := grpcClient.Client.GetBlock(ctx, nil)
	if block == nil {
		return 0, err
	}
	return block.BlockHeader.RawData.Number, err
}

func (t *Tron) CheckTransactionStatus(txHash string) error {
	resultCh := make(chan bool, 1)
	var wg sync.WaitGroup
	wg.Add(1)
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	go func() {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			default:
				tx, err := t.GetTransactionStatus(txHash)
				if err != nil {
					time.Sleep(3 * time.Second)
					continue
				}
				if tx {
					resultCh <- tx
					return
				} else {
					continue
				}
			}
		}
	}()
	wg.Wait()
	select {
	case result := <-resultCh:
		if !result {
			return errors.New("activation failed")
		}
		return nil
	case <-ctx.Done():
		return errors.New("timeout waiting for transaction status")
	}
}

func (t *Tron) GetTransactionStatus(txHash string) (bool, error) {
	grpcClient := t.GetGrpcClient()
	defer grpcClient.Stop()
	res, err := grpcClient.GetTransactionInfoByID(txHash)
	if res == nil {
		return false, err
	}
	return res.Receipt.Result == core.Transaction_Result_SUCCESS, nil
}

func (t *Tron) GetGrpcClient() *client.GrpcClient {
	grpcClient := client.NewGrpcClient("")
	_ = grpcClient.SetAPIKey(t.TronProApiKey)
	err := grpcClient.Start(grpc.WithInsecure())
	if err != nil {
		log.Fatal(fmt.Sprintf("failed to get grpc client: %v", err))
	}
	return grpcClient
}

// 获取用户能量
func (t *Tron) GetAccountResource(address string) (int64, error) {
	grpcClient := t.GetGrpcClient()
	defer grpcClient.Stop()
	res, err := grpcClient.GetAccountResource(address)
	if res == nil {
		return 0, err
	}
	return res.EnergyLimit, nil
}
func (t *Tron) Balance(address string) (decimal.Decimal, error) {
	grpcClient := t.GetGrpcClient()
	defer grpcClient.Stop()
	res, err := grpcClient.GetAccount(address)

	if err != nil {
		errMsg := "account not found"
		if err.Error() == errMsg {
			return decimal.Zero, nil
		}
		return decimal.Zero, err
	}
	if res == nil {
		return decimal.Zero, err
	}
	return decimal.NewFromBigInt(big.NewInt(res.Balance), -6), nil
}

func (t *Tron) TransferTrx(from, to string, amount decimal.Decimal, privateKey string) (string, error) {
	grpcClient := t.GetGrpcClient()
	defer grpcClient.Stop()
	amount = amount.Mul(decimal.NewFromInt(10).Pow(decimal.NewFromInt(6)))
	transaction, err := grpcClient.Transfer(from, to, amount.IntPart())
	if err != nil {
		return "", err
	}
	signTransaction, err := t.SignTransaction(transaction, privateKey)
	if err != nil {
		return "", err
	}
	return t.SendRawTransaction(signTransaction)
}

func (t *Tron) SendRawTransaction(transaction *api.TransactionExtention) (string, error) {
	grpcClient := t.GetGrpcClient()
	defer grpcClient.Stop()
	response, err := grpcClient.Broadcast(transaction.Transaction)
	if response == nil {
		return "", err
	}
	if response.Result {
		return strings.TrimPrefix(common.BytesToHexString(transaction.GetTxid()), "0x"), nil
	} else {
		return "", errors.New(string(response.Message))
	}
}
func (t *Tron) SignTransaction(transaction *api.TransactionExtention, privateKey string) (*api.TransactionExtention, error) {
	// 获取交易的 RawData
	rawData, err := proto.Marshal(transaction.Transaction.GetRawData())
	if err != nil {
		return nil, fmt.Errorf("failed to marshal raw data: %v", err)
	}
	h256h := sha256.New()
	h256h.Write(rawData)
	hash := h256h.Sum(nil)
	// 使用私钥签名
	privateKeyECDSA, err := crypto.HexToECDSA(privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to convert private key: %v", err)
	}
	signature, err := crypto.Sign(hash, privateKeyECDSA)
	if err != nil {
		return nil, fmt.Errorf("failed to sign transaction: %v", err)
	}
	// 将签名添加到交易中
	transaction.Transaction.Signature = append(transaction.Transaction.Signature, signature)
	return transaction, nil
}
