package eth_helper

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"log"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/shopspring/decimal"
	"github.com/web3coderecho/web3_helper/eth_helper/eth_interface"
	"github.com/web3coderecho/web3_helper/utils"
)

type EthHelper struct {
	rpcURL   string
	chainId  *big.Int
	gasPrice eth_interface.GasPriceInterface
}

func NewEthHelper(rpcURL string) *EthHelper {
	return &EthHelper{
		rpcURL: rpcURL,
		chainId: big.NewInt(0),
	}
}

func (e *EthHelper) SetGasPrice(gasPrice eth_interface.GasPriceInterface) {
	e.gasPrice = gasPrice
}

func (e *EthHelper) GetGasPrice(ctx context.Context) (*big.Int, error) {
	if e.gasPrice == nil {
		return e.FeeHistory(ctx, 20, []float64{25, 75})
	}
	gasPrice, err := e.gasPrice.GetGasPrice()
	if err != nil {
		return nil, err
	}
	return utils.ToWeiWithDecimals(gasPrice, 9), nil
}

// NewEthClient 初始化并连接到以太坊节点
func (e *EthHelper) NewEthClient(ctx context.Context) (*ethclient.Client, error) {
	client, err := ethclient.DialContext(ctx, e.rpcURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Ethereum node: %v", err)
	}
	return client, nil
}

func (e *EthHelper) GetBlockNumber(ctx context.Context) (uint64, error) {
	client, err := e.NewEthClient(ctx)
	if err != nil {
		return 0, err
	}
	defer client.Close()
	blockNumber, err := client.BlockNumber(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to get block number: %v", err)
	}
	return blockNumber, nil
}
func (e *EthHelper) GetBlockByNumber(ctx context.Context, blockNumber int64) (*types.Block, error) {
	client, err := e.NewEthClient(ctx)
	if err != nil {
		return nil, err
	}
	defer client.Close()
	return client.BlockByNumber(ctx, big.NewInt(blockNumber))
}

func (e *EthHelper) EstimateGas(ctx context.Context, from, to common.Address, data []byte, value decimal.Decimal) (uint64, error) {
	client, err := e.NewEthClient(ctx)
	if err != nil {
		return 0, err
	}
	defer client.Close()

	return client.EstimateGas(ctx, ethereum.CallMsg{
		From:  from,
		To:    &to,
		Data:  data,
		Value: utils.ToEther(value),
	})
}

func (e *EthHelper) GetBalance(ctx context.Context, address common.Address) (decimal.Decimal, error) {
	client, err := e.NewEthClient(ctx)
	if err != nil {
		return decimal.Zero, err
	}
	defer client.Close()
	balance, err := client.BalanceAt(ctx, address, nil)
	if err != nil {
		return decimal.Zero, err
	}
	return utils.FromEther(balance), nil
}

func (e *EthHelper) GetChainId(ctx context.Context) (*big.Int, error) {
	client, err := e.NewEthClient(ctx)
	if err != nil {
		return e.chainId, err
	}
	defer client.Close()
	if e.chainId == nil || e.chainId.Cmp(big.NewInt(0)) <= 0 {
		chainId, err := client.ChainID(ctx)
		if err != nil {
			return e.chainId, err
		}
		e.chainId = chainId
	}
	return e.chainId, nil
}

func (e *EthHelper) GetTransactionCount(ctx context.Context, address common.Address) (uint64, error) {
	client, err := e.NewEthClient(ctx)
	if err != nil {
		return 0, err
	}
	defer client.Close()
	return client.PendingNonceAt(ctx, address)
}

func (e *EthHelper) GetTransactionReceipt(ctx context.Context, txHash common.Hash) (*types.Receipt, error) {
	client, err := e.NewEthClient(ctx)
	if err != nil {
		return nil, err
	}
	defer client.Close()
	return client.TransactionReceipt(ctx, txHash)
}

func (e *EthHelper) GetTransactionByHash(ctx context.Context, txHash common.Hash) (*types.Transaction, bool, error) {
	client, err := e.NewEthClient(ctx)
	if err != nil {
		return nil, false, err
	}
	defer client.Close()
	return client.TransactionByHash(ctx, txHash)
}

// TransferETH 发送 ETH 转账交易
func (e *EthHelper) TransferETH(
	ctx context.Context,
	from common.Address,
	privateKey *ecdsa.PrivateKey,
	to common.Address,
	amount decimal.Decimal,
) (common.Hash, error) {
	return e.Transaction(ctx, from, privateKey, to, amount, 0, common.Big0, 0, nil)
}
func (e *EthHelper) CheckTransactionStatus(txHash common.Hash) error {
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
				tx, err := e.GetTransactionReceipt(ctx, txHash)
				if err != nil {
					log.Printf("Error getting transaction status: %v", err)
					time.Sleep(3 * time.Second)
					continue
				}
				if tx.Status == types.ReceiptStatusSuccessful {
					resultCh <- true
				} else {
					resultCh <- false
				}
				return
			}
		}
	}()
	wg.Wait()
	select {
	case result := <-resultCh:
		if !result {
			return fmt.Errorf("transaction failed %s", txHash.String())
		}
		return nil
	case <-ctx.Done():
		return fmt.Errorf("timeout waiting for transaction status %s", txHash.String())
	}
}
func (e *EthHelper) Transaction(
	ctx context.Context,
	from common.Address,
	privateKey *ecdsa.PrivateKey,
	to common.Address,
	amount decimal.Decimal,
	gasLimit uint64,
	gasPrice *big.Int,
	nonce uint64,
	data []byte,
) (common.Hash, error) {
	limit, price, _, err := e.Check(ctx, from, to, data, amount)
	if err != nil {
		return common.Hash{}, err
	}
	newGasPrice := big.NewInt(0)
	if gasPrice.Cmp(big.NewInt(0)) <= 0 || gasPrice.Sign() == 0 {
		newGasPrice = price
	}else{
		newGasPrice = gasPrice
	}
	if gasLimit <= limit {
		gasLimit = limit
	}
	if nonce == 0 {
		client, err := e.NewEthClient(ctx)
		if err != nil {
			return common.Hash{}, err
		}
		defer client.Close()
		nonce, err = client.PendingNonceAt(ctx, from)
		if err != nil {
			return common.Hash{}, fmt.Errorf("failed to get nonce: %v", err)
		}
	}
	// 3. 将 decimal.Decimal 转换为 *big.Int（Wei）
	value := utils.ToEther(amount)
	// 5. 获取 chainID
	chainID, err := e.GetChainId(ctx)
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to get chain ID: %v", err)
	}
	// 4. 创建交易对象
	tx := types.NewTx(&types.LegacyTx{
		Nonce:    nonce,
		To:       &to,
		Value:    value,
		Gas:      gasLimit,
		GasPrice: newGasPrice,
		Data:     data,
	})
	// 6. 使用私钥签名交易
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to sign transaction: %v", err)
	}
	return e.SendTransaction(ctx, signedTx)
}

func (e *EthHelper) Check(ctx context.Context, from, to common.Address, data []byte, amount decimal.Decimal) (gasLimit uint64, gasPrice *big.Int, gas decimal.Decimal, err error) {
	gasLimit, err = e.EstimateGas(ctx, from, to, data, amount)
	if err != nil {
		return 0, nil, decimal.Zero, fmt.Errorf("failed to estimate gas: %v", err)
	}
	gasPrice, err = e.GetGasPrice(ctx)
	gasFee := big.NewInt(0)
	gasFee.Mul(gasPrice, big.NewInt(int64(gasLimit)))
	fromBalance, err := e.GetBalance(ctx, from)
	gas = utils.FromEther(gasFee).Add(amount)
	if gas.Cmp(fromBalance) >= 0 {
		gas.Sub(fromBalance)
		err = fmt.Errorf("insufficient balance")
	}
	return gasLimit, gasPrice, gas, err
}

func (e *EthHelper) SendTransaction(ctx context.Context, tx *types.Transaction) (common.Hash, error) {
	client, err := e.NewEthClient(ctx)
	if err != nil {
		return common.Hash{}, err
	}
	defer client.Close()
	err = client.SendTransaction(ctx, tx)
	return tx.Hash(), err
}

func (e *EthHelper) FilterLogs(ctx context.Context, filterQuery ethereum.FilterQuery) ([]types.Log, error) {
	client, err := e.NewEthClient(ctx)
	if err != nil {
		return nil, err
	}
	defer client.Close()
	return client.FilterLogs(ctx, filterQuery)
}

func (e *EthHelper) FeeHistory(ctx context.Context, blockCount uint64, rewardPercentiles []float64) (*big.Int, error) {
	client, err := e.NewEthClient(ctx)
	if err != nil {
		return nil, err
	}
	defer client.Close()
	res, err := client.FeeHistory(ctx, blockCount, nil, rewardPercentiles)
	if err != nil {
		return nil, err
	}
	maxPriorityFee := e.estimateGasPrice(res)
	return maxPriorityFee, nil
}

func (e *EthHelper) estimateGasPrice(feeHistory *ethereum.FeeHistory) (maxPriorityFee *big.Int) {
	if len(feeHistory.Reward) == 0 {
		return maxPriorityFee
	}
	var totalTips big.Int
	for _, rewards := range feeHistory.Reward {
		if len(rewards) > 1 {
			totalTips.Add(&totalTips, rewards[1]) // 使用 75% 百分位 tip
		}
	}
	count := big.NewInt(int64(len(feeHistory.Reward)))
	avgTip := new(big.Int).Div(&totalTips, count)
	// 构造 gas 参数
	return new(big.Int).Set(avgTip)
}
