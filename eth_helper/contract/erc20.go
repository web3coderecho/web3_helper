package contract

import (
	"context"
	"crypto/ecdsa"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/shopspring/decimal"
	"github.com/web3coderecho/web3_helper/eth_helper"
	"github.com/web3coderecho/web3_helper/eth_helper/contract/erc20"
)

type ERC20 struct {
	ContractAddress common.Address
	Name            string
	Symbol          string
	Decimals        int
	eth             *eth_helper.EthHelper
}

func NewErc20(eth eth_helper.EthHelper, address common.Address) *ERC20 {
	return &ERC20{
		ContractAddress: address,
		eth:             &eth,
	}
}

func (erc *ERC20) GetErc20Caller(ctx context.Context) (*erc20.Erc20Caller, *ethclient.Client, error) {
	client, err := erc.eth.NewEthClient(ctx)
	if err != nil {
		return nil, nil, err
	}
	erc20caller, err := erc20.NewErc20Caller(erc.ContractAddress, client)
	if err != nil {
		client.Close()
		return nil, nil, err
	}
	return erc20caller, client, nil
}
func (erc *ERC20) GetErc20Transactor(ctx context.Context) (*erc20.Erc20Transactor, *ethclient.Client, error) {
	client, err := erc.eth.NewEthClient(ctx)
	if err != nil {
		return nil, nil, err
	}
	transactor, err := erc20.NewErc20Transactor(erc.ContractAddress, client)
	if err != nil {
		client.Close()
		return nil, nil, err
	}
	return transactor, client, nil
}

func (erc *ERC20) GetErc20Filterer(ctx context.Context) (*erc20.Erc20Filterer, *ethclient.Client, error) {
	client, err := erc.eth.NewEthClient(ctx)
	if err != nil {
		return nil, nil, err
	}
	filterer, err := erc20.NewErc20Filterer(erc.ContractAddress, client)
	if err != nil {
		client.Close()
		return nil, nil, err
	}
	return filterer, client, nil
}

func (erc *ERC20) GetDecimals(ctx context.Context) (int, error) {
	if erc.Decimals != 0 {
		return erc.Decimals, nil
	}
	caller, client, err := erc.GetErc20Caller(ctx)
	if err != nil {
		return 0, err
	}
	defer client.Close()
	decimals, err := caller.Decimals(nil)
	if err != nil {
		return 0, err
	}
	erc.Decimals = int(decimals)
	return erc.Decimals, err
}

func (erc *ERC20) BalanceOf(ctx context.Context, address common.Address) (decimal.Decimal, error) {
	caller, client, err := erc.GetErc20Caller(ctx)
	if err != nil {
		return decimal.Zero, err
	}
	defer client.Close()
	balance, err := caller.BalanceOf(nil, address)
	if err != nil {
		return decimal.Zero, err
	}
	if balance.Cmp(common.Big0) == 0 {
		return decimal.Zero, nil
	}
	decimals, err := erc.GetDecimals(ctx)
	if err != nil {
		return decimal.Zero, err
	}
	return eth_helper.FromWeiWithDecimals(balance, decimals), err
}

func (erc *ERC20) GetName(ctx context.Context) (string, error) {
	if erc.Name != "" {
		return erc.Name, nil
	}
	caller, client, err := erc.GetErc20Caller(ctx)
	if err != nil {
		return "", err
	}
	defer client.Close()
	name, err := caller.Name(nil)
	if err != nil {
		return "", err
	}
	erc.Name = name
	return erc.Name, err
}
func (erc *ERC20) GetSymbol(ctx context.Context) (string, error) {
	if erc.Symbol != "" {
		return erc.Symbol, nil
	}
	caller, client, err := erc.GetErc20Caller(ctx)
	if err != nil {
		return "", err
	}
	defer client.Close()
	symbol, err := caller.Symbol(nil)
	if err != nil {
		return "", err
	}
	erc.Symbol = symbol
	return erc.Symbol, err
}

func (erc *ERC20) TotalSupply(ctx context.Context) (decimal.Decimal, error) {
	caller, client, err := erc.GetErc20Caller(ctx)
	if err != nil {
		return decimal.Zero, err
	}
	defer client.Close()
	totalSupply, err := caller.TotalSupply(nil)
	if err != nil {
		return decimal.Zero, err
	}
	decimals, err := erc.GetDecimals(ctx)
	if err != nil {
		return decimal.Zero, err
	}
	return eth_helper.FromWeiWithDecimals(totalSupply, decimals), err
}
func (erc *ERC20) Allowance(ctx context.Context, owner, spender common.Address) (decimal.Decimal, error) {
	caller, client, err := erc.GetErc20Caller(ctx)
	if err != nil {
		return decimal.Zero, err
	}
	defer client.Close()
	allowance, err := caller.Allowance(nil, owner, spender)
	if err != nil {
		return decimal.Zero, err
	}
	decimals, err := erc.GetDecimals(ctx)
	if err != nil {
		return decimal.Zero, err
	}
	return eth_helper.FromWeiWithDecimals(allowance, decimals), err
}

func (erc *ERC20) Transfer(ctx context.Context, from, to common.Address, amount decimal.Decimal, privateKey *ecdsa.PrivateKey) (common.Hash, error) {
	transactor, client, err := erc.GetErc20Transactor(ctx)
	defer client.Close()
	if err != nil {
		return common.Hash{}, err
	}
	balance, err := erc.BalanceOf(ctx, from)
	if err != nil {
		return common.Hash{}, err
	}
	if balance.LessThan(amount) {
		return common.Hash{}, fmt.Errorf("erc20 balance is not enough")
	}
	transaction, err := transactor.Transfer(nil, to, eth_helper.ToWeiWithDecimals(amount, erc.Decimals))
	if err != nil {
		return common.Hash{}, err
	}
	return erc.eth.Transaction(ctx, from, privateKey, erc.ContractAddress, decimal.Zero, 0, common.Big0, 0, transaction.Data())
}

func (erc *ERC20) Approve(ctx context.Context, from, spender common.Address, amount decimal.Decimal, privateKey *ecdsa.PrivateKey) (common.Hash, error) {
	transactor, client, err := erc.GetErc20Transactor(ctx)
	defer client.Close()
	if err != nil {
		return common.Hash{}, err
	}
	transaction, err := transactor.Approve(nil, spender, eth_helper.ToWeiWithDecimals(amount, erc.Decimals))
	if err != nil {
		return common.Hash{}, err
	}
	return erc.eth.Transaction(ctx, from, privateKey, erc.ContractAddress, decimal.Zero, 0, common.Big0, 0, transaction.Data())
}

func (erc *ERC20) ParseTransfer(ctx context.Context, log types.Log) (*erc20.Erc20Transfer, error) {
	filterer, client, err := erc.GetErc20Filterer(ctx)
	defer client.Close()
	if err != nil {
		return nil, err
	}
	return filterer.ParseTransfer(log)
}
