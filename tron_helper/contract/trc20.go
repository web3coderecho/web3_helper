package contract

import (
	"encoding/hex"
	"errors"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/shopspring/decimal"
	"github.com/web3coderecho/web3_helper/eth_helper/contract/erc20"
	tron "github.com/web3coderecho/web3_helper/tron_helper"
	"github.com/web3coderecho/web3_helper/utils"
)

type Trc20 struct {
	Chain           *tron.Tron
	ContractAddress string
	decimals        int64
	privateKey      string
}

func NewTrc20(tron *tron.Tron, contractAddress string) *Trc20 {
	return &Trc20{
		Chain:           tron,
		ContractAddress: contractAddress,
	}
}

func (t *Trc20) SetPrivateKey(privateKey string) {
	t.privateKey = privateKey
}

// 获取 TRC20 小数位数
func (t *Trc20) Decimals() (int64, error) {
	if t.decimals > 0 {
		return t.decimals, nil
	}
	grpcClient := t.Chain.GetGrpcClient()
	defer grpcClient.Stop()
	decimals, err := grpcClient.TRC20GetDecimals(t.ContractAddress)
	if err != nil {
		return 0, err
	}
	t.decimals = decimals.Int64()
	return t.decimals, nil
}

func (t *Trc20) DecodeTransfer(to string, amount decimal.Decimal) (string, error) {
	grpcClient := t.Chain.GetGrpcClient()
	defer grpcClient.Stop()
	parsedABI, err := abi.JSON(strings.NewReader(erc20.Erc20MetaData.ABI))
	if err != nil {
		return "", err
	}
	data, err := parsedABI.Pack("transfer", common.HexToAddress(utils.TronToEth(to)), amount.BigInt())
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(data), nil
}

// 获取 TRC20 余额
func (t *Trc20) BalanceOf(address string) (decimal.Decimal, error) {
	grpcClient := t.Chain.GetGrpcClient()
	balance, err := grpcClient.TRC20ContractBalance(address, t.ContractAddress)
	defer grpcClient.Stop()
	if err != nil {
		return decimal.Zero, err
	}
	decimals, err := t.Decimals()
	if err != nil {
		return decimal.Zero, err
	}
	if balance.Cmp(big.NewInt(0)) <= 0 {
		return decimal.Zero, nil
	}
	return decimal.NewFromBigInt(balance, int32(-decimals)), nil
}
func (t *Trc20) EstimateGas(from string, data string) (int64, error) {
	grpcClient := t.Chain.GetGrpcClient()
	defer grpcClient.Stop()
	tx, err := grpcClient.TRC20Call(from, t.ContractAddress, data, true, 0)
	if err != nil {
		return 0, err
	}
	return tx.EnergyUsed, nil
}

func (t *Trc20) CheckEnergy(from, to string, amount decimal.Decimal) (int64, error) {
	grpcClient := t.Chain.GetGrpcClient()
	defer grpcClient.Stop()
	decimals, err := t.Decimals()
	if err != nil {
		return 0, err
	}
	amount = amount.Mul(decimal.NewFromInt(10).Pow(decimal.NewFromInt(decimals)))
	callData, err := t.DecodeTransfer(to, amount)
	if err != nil {
		return 0, err
	}
	feeLimit, err := t.EstimateGas(from, callData)
	if err != nil {
		return 0, err
	}
	energyLimit, err := t.Chain.GetAccountResource(from)
	if err != nil {
		return 0, err
	}
	if energyLimit < feeLimit {
		return feeLimit - energyLimit, nil
	}
	return 0, nil
}

func (t *Trc20) Transfer(from string, to string, amount decimal.Decimal) (string, error) {
	if t.privateKey == "" {
		return "", errors.New("privateKey is empty")
	}
	grpcClient := t.Chain.GetGrpcClient()
	defer grpcClient.Stop()
	decimals, err := t.Decimals()
	if err != nil {
		return "", err
	}
	amount = amount.Mul(decimal.NewFromInt(10).Pow(decimal.NewFromInt(decimals)))
	callData, err := t.DecodeTransfer(to, amount)
	if err != nil {
		return "", err
	}
	feeLimit, err := t.EstimateGas(from, callData)
	if err != nil {
		return "", err
	}
	feeLimit = 1000000000
	transaction, err := grpcClient.TRC20Send(from, to, t.ContractAddress, amount.BigInt(), feeLimit)
	if err != nil {
		return "", err
	}
	signTransaction, err := t.Chain.SignTransaction(transaction, t.privateKey)
	if err != nil {
		return "", err
	}
	return t.Chain.SendRawTransaction(signTransaction)
}
