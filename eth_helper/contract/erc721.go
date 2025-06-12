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
	"github.com/web3coderecho/web3_helper/eth_helper/contract/erc721"
)

type ERC721 struct {
	ContractAddress common.Address
	Name            string
	Symbol          string
	eth             *eth_helper.EthHelper
}

func NewErc721(eth *eth_helper.EthHelper, address common.Address) *ERC721 {
	return &ERC721{
		ContractAddress: address,
		eth:             eth,
	}
}

func (erc *ERC721) GetErc721(ctx context.Context) (*erc721.Erc721, *ethclient.Client, error) {
	client, err := erc.eth.NewEthClient(ctx)
	if err != nil {
		return nil, nil, err
	}
	erc721Contract, err := erc721.NewErc721(erc.ContractAddress, client)
	if err != nil {
		client.Close()
		return nil, nil, err
	}
	return erc721Contract, client, nil
}

func (erc *ERC721) GetName(ctx context.Context) (string, error) {
	if erc.Name != "" {
		return erc.Name, nil
	}
	caller, client, err := erc.GetErc721(ctx)
	defer client.Close()
	if err != nil {
		return "", err
	}
	name, err := caller.Name(nil)
	if err != nil {
		return "", err
	}
	erc.Name = name
	return erc.Name, nil
}

func (erc *ERC721) GetSymbol(ctx context.Context) (string, error) {
	if erc.Symbol != "" {
		return erc.Symbol, nil
	}
	caller, client, err := erc.GetErc721(ctx)
	defer client.Close()
	if err != nil {
		return "", err
	}
	symbol, err := caller.Symbol(nil)
	if err != nil {
		return "", err
	}
	erc.Symbol = symbol
	return erc.Symbol, nil
}

func (erc *ERC721) OwnerOf(ctx context.Context, tokenId *big.Int) (common.Address, error) {
	caller, client, err := erc.GetErc721(ctx)
	defer client.Close()
	if err != nil {
		return common.Address{}, err
	}
	return caller.OwnerOf(nil, tokenId)
}

func (erc *ERC721) BalanceOf(ctx context.Context, owner common.Address) (int64, error) {
	caller, client, err := erc.GetErc721(ctx)
	defer client.Close()
	if err != nil {
		return 0, err
	}
	balance, err := caller.BalanceOf(nil, owner)
	if err != nil {
		return 0, err
	}
	return balance.Int64(), nil
}

func (erc *ERC721) Approve(ctx context.Context, from common.Address, to common.Address, tokenId *big.Int, privateKey *ecdsa.PrivateKey) (common.Hash, error) {
	abi, _ := erc721.Erc721MetaData.GetAbi()
	data, err := abi.Pack("approve", to, tokenId)
	if err != nil {
		return common.Hash{}, err
	}
	return erc.eth.Transaction(ctx, from, privateKey, erc.ContractAddress, decimal.Zero, 0, common.Big0, 0, data)
}

func (erc *ERC721) SafeTransferFrom(ctx context.Context, from common.Address, to common.Address, tokenId *big.Int, data []byte, privateKey *ecdsa.PrivateKey) (common.Hash, error) {
	abi, _ := erc721.Erc721MetaData.GetAbi()
	data, err := abi.Pack("safeTransferFrom", from, to, tokenId, "")
	if err != nil {
		return common.Hash{}, err
	}
	return erc.eth.Transaction(ctx, from, privateKey, erc.ContractAddress, decimal.Zero, 0, common.Big0, 0, data)
}

func (erc *ERC721) SetApprovalForAll(ctx context.Context, from common.Address, operator common.Address, approved bool, privateKey *ecdsa.PrivateKey) (common.Hash, error) {
	abi, _ := erc721.Erc721MetaData.GetAbi()
	data, err := abi.Pack("setApprovalForAll", operator, approved)
	if err != nil {
		return common.Hash{}, err
	}
	return erc.eth.Transaction(ctx, from, privateKey, erc.ContractAddress, decimal.Zero, 0, common.Big0, 0, data)
}

func (erc *ERC721) IsApprovedForAll(ctx context.Context, owner common.Address, operator common.Address) (bool, error) {
	caller, client, err := erc.GetErc721(ctx)
	defer client.Close()
	if err != nil {
		return false, err
	}
	return caller.IsApprovedForAll(nil, owner, operator)
}

func (erc *ERC721) TransferFrom(ctx context.Context, from common.Address, to common.Address, tokenId *big.Int, privateKey *ecdsa.PrivateKey) (common.Hash, error) {
	abi, _ := erc721.Erc721MetaData.GetAbi()
	data, err := abi.Pack("transferFrom", from, to, tokenId, "")
	if err != nil {
		return common.Hash{}, err
	}
	return erc.eth.Transaction(ctx, from, privateKey, erc.ContractAddress, decimal.Zero, 0, common.Big0, 0, data)
}

func (erc *ERC721) ParseTransfer(ctx context.Context, log types.Log) (*erc721.Erc721Transfer, error) {
	filterer, client, err := erc.GetErc721(ctx)
	defer client.Close()
	if err != nil {
		return nil, err
	}
	return filterer.ParseTransfer(log)
}
