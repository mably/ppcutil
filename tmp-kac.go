// Copyright (c) 2014-2014 PPCD developers.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package ppcutil

import (
	"github.com/mably/btcchain"
	"github.com/mably/btcdb"
	"github.com/mably/btcnet"
	"github.com/mably/btcutil"
	"github.com/mably/btcwire"
	"math/big"
)

var (
	ZeroSha                      = btcwire.ShaHash{}
	InitialHashTargetBits uint32 = 0x1c00ffff
	StakeTargetSpacing    int64  = 10 * 60 // 10 minutes
	TargetSpacingWorkMax  int64  = StakeTargetSpacing * 12
	TargetTimespan        int64  = 7 * 24 * 60 * 60
)

func MinInt(a int64, b int64) int64 {
	if a < b {
		return a
	}
	return b
}

// https://github.com/ppcoin/ppcoin/blob/v0.4.0ppc/src/main.cpp#L894
// ppcoin: find last block index up to pindex
func GetLastBlockIndex(db btcdb.Db, last *btcutil.Block, proofOfStake bool) (block *btcutil.Block) {
	block = last
	for true {
		if block == nil {
			break
		}
		//TODO dirty workaround, ppcoin doesn't point to genesis block
		if block.Height()==0{
			return nil
		}
		prevExists, err := db.ExistsSha(&block.MsgBlock().Header.PrevBlock)
		if err != nil || !prevExists {
			break
		}
		if block.MsgBlock().IsProofOfStake() == proofOfStake {
			break
		}
		block, _ = db.FetchBlockBySha(&block.MsgBlock().Header.PrevBlock)
	}
	return block
}

// https://github.com/ppcoin/ppcoin/blob/v0.4.0ppc/src/main.cpp#L902
func GetNextTargetRequired(params btcnet.Params, db btcdb.Db, last *btcutil.Block, proofOfStake bool) (compact uint32) {
	if last == nil {
		return params.PowLimitBits // genesis block
	}
	prev := GetLastBlockIndex(db, last, proofOfStake)
	if prev == nil {
		return InitialHashTargetBits // first block
	}
	block, _ := db.FetchBlockBySha(&prev.MsgBlock().Header.PrevBlock)
	prevPrev := GetLastBlockIndex(db, block, proofOfStake)
	if prevPrev == nil {
		return InitialHashTargetBits // second block
	}
	actualSpacing := prev.MsgBlock().Header.Timestamp.Unix() - prevPrev.MsgBlock().Header.Timestamp.Unix()
	newTarget := btcchain.CompactToBig(prev.MsgBlock().Header.Bits)
	var targetSpacing int64
	if proofOfStake {
		targetSpacing = StakeTargetSpacing
	} else {
		targetSpacing = MinInt(TargetSpacingWorkMax, StakeTargetSpacing*(1+last.Height()-prev.Height()))
	}
	interval := TargetTimespan / targetSpacing
	tmp := new(big.Int)
	newTarget.Mul(newTarget,
		tmp.SetInt64(interval-1).Mul(tmp, big.NewInt(targetSpacing)).Add(tmp, big.NewInt(actualSpacing+actualSpacing)))
	newTarget.Div(newTarget, tmp.SetInt64(interval+1).Mul(tmp, big.NewInt(targetSpacing)))
	if newTarget.Cmp(params.PowLimit) > 0 {
		newTarget = params.PowLimit
	}
	return btcchain.BigToCompact(newTarget)
}
