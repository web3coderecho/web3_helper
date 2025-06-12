package scan

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/web3coderecho/web3_helper/eth_helper"
)

type Scan struct {
	Address   []common.Address `json:"address,omitempty"` // 合约地址（可选）
	Topics    [][]common.Hash  `json:"topics,omitempty"`  // 日志主题数组（最多4个）
	ethHelper *eth_helper.EthHelper
}

func NewScanFilterQuery(address []common.Address, topics [][]common.Hash, eth *eth_helper.EthHelper) *Scan {
	return &Scan{
		Address:   address,
		Topics:    topics,
		ethHelper: eth,
	}
}

func (s *Scan) Scan(ctx context.Context, from, to uint64, blockHash string) ([]types.Log, error) {
	filterQuery := s.GetEthFilterQuery(from, to, blockHash)
	return s.ethHelper.FilterLogs(ctx, filterQuery)
}

func (s *Scan) GetEthFilterQuery(from, to uint64, blockHash string) ethereum.FilterQuery {
	if blockHash == "" {
		return ethereum.FilterQuery{
			FromBlock: big.NewInt(int64(from)),
			ToBlock:   big.NewInt(int64(to)),
			Addresses: s.Address,
			Topics:    s.Topics,
		}
	} else {
		bHash := common.HexToHash(blockHash)
		return ethereum.FilterQuery{
			BlockHash: &bHash,
			Addresses: s.Address,
			Topics:    s.Topics,
		}
	}
}
