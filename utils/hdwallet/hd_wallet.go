package hdwallet

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/btcsuite/btcd/btcutil/hdkeychain"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/tyler-smith/go-bip39"
	"github.com/web3coderecho/web3_helper/utils"
)

var MainNetParams = &chaincfg.MainNetParams

// HDWallet 结构体
type HDWallet struct {
	Mnemonic string
	Seed     []byte
	Master   *hdkeychain.ExtendedKey
}

// EncryptMnemonicToKeystore 使用密码加密助记词，并生成标准的 Keystore JSON 数据
func (w *HDWallet) EncryptMnemonicToKeystore(password string) ([]byte, error) {
	data := []byte(w.Mnemonic)
	// 使用 scrypt 参数加密任意数据（非私钥），返回 keystore 格式 CryptoJSON
	cryptoJson, err := keystore.EncryptDataV3(
		data,             // 助记词字节
		[]byte(password), // 密码
		keystore.StandardScryptN,
		keystore.StandardScryptP,
	)
	if err != nil {
		return nil, err
	}
	// 返回 JSON 序列化后的字节数组
	return json.Marshal(cryptoJson)
}

// GetRootKeyPair 获取根地址和根私钥
func (w *HDWallet) GetRootKeyPair() (*ETHAddressInfo, error) {
	// 获取根私钥
	privateKey, err := w.Master.ECPrivKey()
	if err != nil {
		return nil, fmt.Errorf("failed to get root private key: %v", err)
	}
	ecdsaPrivateKey := privateKey.ToECDSA()
	// 生成对应的以太坊地址
	ethAddr := crypto.PubkeyToAddress(ecdsaPrivateKey.PublicKey)
	return &ETHAddressInfo{
		Address:    ethAddr,
		PrivateKey: ecdsaPrivateKey,
	}, nil
}

func DecryptMnemonicFromKeystore(keystoreJSON []byte, password string) (*HDWallet, error) {
	var keyStore keystore.CryptoJSON
	if err := json.Unmarshal(keystoreJSON, &keyStore); err != nil {
		return nil, err
	}
	// 解密任意数据
	decryptedData, err := keystore.DecryptDataV3(keyStore, password)
	if err != nil {
		return nil, err
	}
	mnemonic := string(bytes.TrimRight(decryptedData, "\x00"))
	return RecoverHDWallet(mnemonic)
}

func NewHDWallet(entropyBits int) (*HDWallet, error) {
	entropy, err := bip39.NewEntropy(entropyBits)
	if err != nil {
		return nil, err
	}
	mnemonic, err := bip39.NewMnemonic(entropy)
	if err != nil {
		return nil, err
	}
	seed := bip39.NewSeed(mnemonic, "")
	master, err := hdkeychain.NewMaster(seed, MainNetParams)
	if err != nil {
		return nil, err
	}
	return &HDWallet{Mnemonic: mnemonic, Seed: seed, Master: master}, nil
}

// RecoverHDWallet 通过助记词恢复钱包
func RecoverHDWallet(mnemonic string) (*HDWallet, error) {
	if !bip39.IsMnemonicValid(mnemonic) {
		return nil, fmt.Errorf("助记词无效")
	}
	seed := bip39.NewSeed(mnemonic, "")
	master, err := hdkeychain.NewMaster(seed, MainNetParams)
	if err != nil {
		return nil, err
	}
	return &HDWallet{Mnemonic: mnemonic, Seed: seed, Master: master}, nil
}

// ValidateMnemonic 校验助记词
func (w *HDWallet) ValidateMnemonic() error {
	if !bip39.IsMnemonicValid(w.Mnemonic) {
		return fmt.Errorf("助记词无效")
	}
	return nil
}

// BatchGenETHAddresses 批量生成以太坊地址信息。
// 该方法根据给定的账户、变更、起始索引和数量参数，派生出一系列以太坊地址及其对应的私钥。
// 参数:
//
//	account - 钱包账户编号。
//	change - 变更编号，用于区分主要账户和变更账户。
//	startIdx - 开始索引，用于指定派生地址的起始位置。
//	count - 生成地址的数量。
//
// 返回值:
//
//	[]ETHAddressInfo - 包含以太坊地址信息的切片，每个信息包括地址和私钥。
//
// BatchGenETHAddresses 批量生成以太坊地址信息。
func (w *HDWallet) BatchGenETHAddresses(account, change, startIdx, count uint32) []*ETHAddressInfo {
	var result []*ETHAddressInfo
	for i := uint32(0); i < count; i++ {
		address, err := w.GenETHByIndex(account, change, startIdx+i)
		if err != nil {
			continue
		}
		result = append(result, address)
	}
	return result
}

// deriveKeyAtPath 派生指定 account、change、index 的私钥
func (w *HDWallet) deriveKeyAtPath(account, change, index uint32) (*ecdsa.PrivateKey, error) {
	purpose, err := w.Master.Derive(hardened(44))
	if err != nil {
		return nil, err
	}
	coinType, err := purpose.Derive(hardened(60))
	if err != nil {
		return nil, err
	}
	acc, err := coinType.Derive(hardened(account))
	if err != nil {
		return nil, err
	}
	chg, err := acc.Derive(change)
	if err != nil {
		return nil, err
	}
	addrKey, err := chg.Derive(index)
	if err != nil {
		return nil, err
	}
	privateKey, err := addrKey.ECPrivKey()
	if err != nil {
		return nil, err
	}
	return privateKey.ToECDSA(), nil
}

// GenETHByIndex 生成指定 account、change、index 的 ETH 地址和私钥
func (w *HDWallet) GenETHByIndex(account, change, index uint32) (*ETHAddressInfo, error) {
	privateKey, err := w.deriveKeyAtPath(account, change, index)
	if err != nil {
		return nil, err
	}
	ethAddr := crypto.PubkeyToAddress(privateKey.PublicKey)
	return NewEthAddressInfo(ethAddr, privateKey), nil
}

// ExportETHKeystore 导出ETH keystore V3 JSON
func (info *ETHAddressInfo) ExportETHKeystore(password string) ([]byte, error) {
	return keystore.EncryptKey(&keystore.Key{Address: crypto.PubkeyToAddress(info.PrivateKey.PublicKey), PrivateKey: info.PrivateKey}, password, keystore.StandardScryptN, keystore.StandardScryptP)
}

// ETHAddressInfo 保存ETH地址和私钥
type ETHAddressInfo struct {
	Address    common.Address
	PrivateKey *ecdsa.PrivateKey
}

func (info *ETHAddressInfo) ToTronAddress() string {
	return utils.EthToTron(info.Address)
}

// EncryptPrivateKey 使用指定密码加密私钥，返回 keystore JSON 数据
func (info *ETHAddressInfo) EncryptPrivateKey(password string) ([]byte, error) {
	key := &keystore.Key{
		Address:    info.Address,
		PrivateKey: info.PrivateKey,
	}
	return keystore.EncryptKey(key, password, keystore.StandardScryptN, keystore.StandardScryptP)
}

func DecryptPrivateKeyFromKeyStore(keystoreJSON []byte, password string) (*ETHAddressInfo, error) {
	key, err := keystore.DecryptKey(keystoreJSON, password)
	if err != nil {
		return nil, err
	}
	return &ETHAddressInfo{
		Address:    key.Address,
		PrivateKey: key.PrivateKey,
	}, nil
}

func NewEthAddressInfo(address common.Address, privateKey *ecdsa.PrivateKey) *ETHAddressInfo {
	return &ETHAddressInfo{Address: address, PrivateKey: privateKey}
}

func (info *ETHAddressInfo) PrivateKey2String() string {
	privateKeyBytes := crypto.FromECDSA(info.PrivateKey)
	return hex.EncodeToString(privateKeyBytes)
}

// hardened 返回硬化索引
func hardened(i uint32) uint32 {
	return i + hdkeychain.HardenedKeyStart
}
