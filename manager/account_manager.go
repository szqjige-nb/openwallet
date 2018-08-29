/*
 * Copyright 2018 The OpenWallet Authors
 * This file is part of the OpenWallet library.
 *
 * The OpenWallet library is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The OpenWallet library is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 * GNU Lesser General Public License for more details.
 */

package manager

import (
	"github.com/blocktree/OpenWallet/openwallet"
	"fmt"
	"github.com/blocktree/OpenWallet/assets"
	"github.com/asdine/storm"
	"github.com/blocktree/go-OWCBasedFuncs/owkeychain"
	"time"
)

func (wm *WalletManager) CreateAssetsAccount(appID, walletID string, account *openwallet.AssetsAccount, otherOwnerKeys []string) (*openwallet.AssetsAccount, error) {

	wallet, err := wm.GetWalletInfo(appID, walletID)
	if err != nil {
		return nil, err
	}

	if len(account.Alias) == 0 {
		return nil, fmt.Errorf("account alias is empty")
	}

	if len(account.Symbol) == 0 {
		return nil, fmt.Errorf("account symbol is empty")
	}

	if account.Required == 0 {
		account.Required = 1
	}

	symbolInfo, err := assets.GetSymbolInfo(account.Symbol)
	if err != nil {
		return nil, err
	}

	if wallet.IsTrust {

		//使用私钥创建子账户
		key, err := wallet.HDKey()
		if err != nil {
			return nil, err
		}

		newAccIndex := wallet.AccountIndex + 1 + owkeychain.HardenedKeyStart

		account.HDPath = fmt.Sprintf("%s/%d",wallet.RootPath, newAccIndex)

		childKey, err := key.DerivedKeyWithPath(account.HDPath, symbolInfo.CurveType())
		if err != nil {
			return nil, err
		}

		account.PublicKey = childKey.OWEncode()
		account.Index = uint64(newAccIndex)
		account.AccountID = account.GetAccountID()

		wallet.AccountIndex = newAccIndex
	}

	account.AddressIndex = -1

	//组合拥有者
	account.OwnerKeys = []string{
		account.PublicKey,
	}

	account.OwnerKeys = append(account.OwnerKeys, otherOwnerKeys...)

	if len(account.PublicKey) == 0 {
		return nil, fmt.Errorf("account publicKey is empty")
	}

	//保存钱包到本地应用数据库
	db, err := wm.OpenDB(appID)
	if err != nil {
		return nil, err
	}

	tx, err := db.Begin(true)
	if err != nil {
		return nil, err
	}

	defer tx.Rollback()

	err = tx.Save(wallet)
	if err != nil {

		return nil, err
	}

	err = tx.Save(account)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	//TODO:创建新地址
	_, err = wm.CreateAddress(appID, walletID, account.GetAccountID(), 1)
	if err != nil {
		return nil, err
	}

	return account, nil
}

func (wm *WalletManager) GetAssetsAccountInfo(appID, walletID, accountID string) (*openwallet.AssetsAccount, error) {

	//打开数据库
	db, err := wm.OpenDB(appID)
	if err != nil {
		return nil, err
	}

	var account openwallet.AssetsAccount
	err = db.One("AccountID", accountID, &account)
	if err != nil {
		return nil, fmt.Errorf("can not find account: %s", accountID)
	}

	return &account, nil
}

func (wm *WalletManager) GetAssetsAccountList(appID, walletID string, offset, limit int) ([]*openwallet.AssetsAccount, error) {

	//打开数据库
	db, err := wm.OpenDB(appID)
	if err != nil {
		return nil, err
	}

	var accounts []*openwallet.AssetsAccount
	err = db.Find("WalletID", walletID, &accounts, storm.Limit(limit), storm.Skip(offset))
	if err != nil {
		return nil, err
	}

	return accounts, nil

}


func (wm *WalletManager) CreateAddress(appID, walletID string, accountID string, count uint64) ([]*openwallet.Address, error) {

	var (
		newKeys = make([][]byte, 0)
		address string
		addrs = make([]*openwallet.Address, 0)
	)

	account, err := wm.GetAssetsAccountInfo(appID, walletID, accountID)
	if err != nil {
		return nil, err
	}

	if count == 0 {
		return nil, fmt.Errorf("create address count is zero")
	}

	symbolInfo, err := assets.GetSymbolInfo(account.Symbol)
	if err != nil {
		return nil, err
	}

	//保存钱包到本地应用数据库
	db, err := wm.OpenDB(appID)
	if err != nil {
		return nil, err
	}

	tx, err := db.Begin(true)
	if err != nil {
		return nil, err
	}

	defer tx.Rollback()

	for i := uint64(0); i<count;i++{

		address = ""

		newIndex := account.AddressIndex + 1

		derivedPath := fmt.Sprintf("%s/%d", account.HDPath, newIndex)

		//通过多个拥有者公钥生成地址
		for _, pub := range account.OwnerKeys {

			pubkey, err := owkeychain.OWDecode(pub)
			if err != nil {
				return nil, err
			}

			newKey, err := pubkey.GenPublicChild(uint32(newIndex))
			newKeys = append(newKeys, newKey.GetPublicKeyBytes())
		}


		if len(account.OwnerKeys) > 1 {
			address, err = symbolInfo.AddressDecode().RedeemScriptToAddress(newKeys, account.Required, wm.cfg.isTestnet)
			if err != nil {
				return nil, err
			}
		} else {
			address, err = symbolInfo.AddressDecode().PublicKeyToAddress(newKeys[0], wm.cfg.isTestnet)
			if err != nil {
				return nil, err
			}
		}

		addr := &openwallet.Address{
			Address:   address,
			AccountID: accountID,
			HDPath:    derivedPath,
			CreatedAt: time.Now(),
			Symbol:    account.Symbol,
			Index:     uint64(newIndex),
			WatchOnly: false,
		}


		account.AddressIndex = newIndex

		err = tx.Save(account)
		if err != nil {

			return nil, err
		}

		err = tx.Save(addr)
		if err != nil {
			return nil, err
		}

		addrs = append(addrs, addr)

	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return addrs, nil
}

func (wm *WalletManager) GetAddressList(appID, walletID, accountID string, offset, limit int, watchOnly bool) ([]*openwallet.Address, error) {
	//TODO:待实现
	return nil, nil
}


func (wm *WalletManager) ImportWatchOnlyAddress(appID, walletID, accountID string, addresses []*openwallet.Address) error {
	//TODO:待实现
	return nil
}

