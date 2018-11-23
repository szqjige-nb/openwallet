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

package tron

//AddressDecoderStruct for Interface AddressDecoder
type AddressDecoderStruct struct {
	wm *WalletManager //钱包管理者
}

//NewAddressDecoder 地址解析器
func NewAddressDecoder(wm *WalletManager) *AddressDecoderStruct {
	decoder := AddressDecoderStruct{}
	decoder.wm = wm
	return &decoder
}

//PrivateKeyToWIF 私钥转WIF
func (decoder *AddressDecoderStruct) PrivateKeyToWIF(priv []byte, isTestnet bool) (string, error) {
	return "", nil
}

//PublicKeyToAddress 公钥转地址
func (decoder *AddressDecoderStruct) PublicKeyToAddress(pub []byte, isTestnet bool) (string, error) {

	address, err := decoder.wm.CreateAddressRef(pub, false) // isPrivate == false
	if err != nil {
		return "", err
	}

	// pkHash := owcrypt.Hash(pub, 0, owcrypt.HASH_ALG_HASH160)
	// address := addressEncoder.AddressEncode(pkHash, cfg)

	return address, nil
}

//RedeemScriptToAddress 多重签名赎回脚本转地址
func (decoder *AddressDecoderStruct) RedeemScriptToAddress(pubs [][]byte, required uint64, isTestnet bool) (string, error) {
	return "", nil
}

//WIFToPrivateKey WIF转私钥
func (decoder *AddressDecoderStruct) WIFToPrivateKey(wif string, isTestnet bool) ([]byte, error) {

	return nil, nil

}