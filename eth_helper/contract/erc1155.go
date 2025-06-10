package contract

import (
	"context"
	"crypto/ecdsa"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/shopspring/decimal"
	"github.com/web3coderecho/web3_helper/eth_helper"
	"github.com/web3coderecho/web3_helper/eth_helper/contract/erc1155"
)

type ERC1155 struct {
	ContractAddress common.Address
	Name            string
	Symbol          string
	eth             *eth_helper.EthHelper
}

func NewErc1155(eth *eth_helper.EthHelper, address common.Address) *ERC1155 {
	return &ERC1155{
		ContractAddress: address,
		eth:             eth,
	}
}

func (erc *ERC1155) GetErc1155(ctx context.Context) (*erc1155.Erc1155, *ethclient.Client, error) {
	client, err := erc.eth.NewEthClient(ctx)
	if err != nil {
		return nil, nil, err
	}
	erc1155Contract, err := erc1155.NewErc1155(erc.ContractAddress, client)
	if err != nil {
		client.Close()
		return nil, nil, err
	}
	return erc1155Contract, client, nil
}

func (erc *ERC1155) BalanceOf(ctx context.Context, account common.Address, id *big.Int) (decimal.Decimal, error) {
	caller, client, err := erc.GetErc1155(ctx)
	defer client.Close()
	if err != nil {
		return decimal.Zero, err
	}
	balance, err := caller.BalanceOf(nil, account, id)
	if err != nil {
		return decimal.Zero, err
	}
	return decimal.NewFromBigInt(balance, 0), nil
}

func (erc *ERC1155) BalanceOfBatch(ctx context.Context, accounts []common.Address, ids []*big.Int) ([]*big.Int, error) {
	caller, client, err := erc.GetErc1155(ctx)
	defer client.Close()
	if err != nil {
		return nil, err
	}
	return caller.BalanceOfBatch(nil, accounts, ids)
}

func (erc *ERC1155) IsApprovedForAll(ctx context.Context, account common.Address, operator common.Address) (bool, error) {
	caller, client, err := erc.GetErc1155(ctx)
	defer client.Close()
	if err != nil {
		return false, err
	}
	return caller.IsApprovedForAll(nil, account, operator)
}

func (erc *ERC1155) SafeTransferFrom(ctx context.Context, from common.Address, to common.Address, id *big.Int, value *big.Int, data []byte, privateKey *ecdsa.PrivateKey) (common.Hash, error) {
	transactor, client, err := erc.GetErc1155(ctx)
	defer client.Close()
	if err != nil {
		return common.Hash{}, err
	}
	tx, err := transactor.SafeTransferFrom(nil, from, to, id, value, data)
	if err != nil {
		return common.Hash{}, err
	}
	return erc.eth.Transaction(ctx, from, privateKey, erc.ContractAddress, decimal.Zero, 0, common.Big0, 0, tx.Data())
}

func (erc *ERC1155) SafeBatchTransferFrom(ctx context.Context, from common.Address, to common.Address, ids []*big.Int, values []*big.Int, data []byte, privateKey *ecdsa.PrivateKey) (common.Hash, error) {
	transactor, client, err := erc.GetErc1155(ctx)
	defer client.Close()
	if err != nil {
		return common.Hash{}, err
	}
	tx, err := transactor.SafeBatchTransferFrom(nil, from, to, ids, values, data)
	if err != nil {
		return common.Hash{}, err
	}
	return erc.eth.Transaction(ctx, from, privateKey, erc.ContractAddress, decimal.Zero, 0, common.Big0, 0, tx.Data())
}

func (erc *ERC1155) SetApprovalForAll(ctx context.Context, from, operator common.Address, approved bool, privateKey *ecdsa.PrivateKey) (common.Hash, error) {
	transactor, client, err := erc.GetErc1155(ctx)
	defer client.Close()
	if err != nil {
		return common.Hash{}, err
	}
	tx, err := transactor.SetApprovalForAll(nil, operator, approved)
	if err != nil {
		return common.Hash{}, err
	}
	return erc.eth.Transaction(ctx, from, privateKey, erc.ContractAddress, decimal.Zero, 0, common.Big0, 0, tx.Data())
}

func (erc *ERC1155) Uri(ctx context.Context, id *big.Int) (string, error) {
	caller, client, err := erc.GetErc1155(ctx)
	defer client.Close()
	if err != nil {
		return "", err
	}
	return caller.Uri(nil, id)
}

func (erc *ERC1155) ParseTransferSingle(log types.Log) (*erc1155.Erc1155TransferSingle, error) {
	filterer, client, err := erc.GetErc1155(context.Background())
	defer client.Close()
	if err != nil {
		return nil, err
	}
	return filterer.ParseTransferSingle(log)
}
func (erc *ERC1155) ParseTransferBatch(log types.Log) (*erc1155.Erc1155TransferBatch, error) {
	filterer, client, err := erc.GetErc1155(context.Background())
	defer client.Close()
	if err != nil {
		return nil, err
	}
	return filterer.ParseTransferBatch(log)
}
