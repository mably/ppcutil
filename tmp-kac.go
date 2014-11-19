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

	"compress/bzip2"
	"compress/gzip"
	"encoding/csv"
	"encoding/hex"
	"io"
	"os"
	"strconv"
	"strings"
)

var (
	zeroSha               = btcwire.ShaHash{}
	initialHashTargetBits = uint32(0x1c00ffff)
	stakeTargetSpacing    = int64(10 * 60) // 10 minutes
	targetSpacingWorkMax  = int64(stakeTargetSpacing * 12)
	targetTimespan        = int64(7 * 24 * 60 * 60)
)

func minInt(a int64, b int64) int64 {
	if a < b {
		return a
	}
	return b
}

// https://github.com/ppcoin/ppcoin/blob/v0.4.0ppc/src/main.cpp#L894
// ppcoin: find last block index up to pindex
func getLastBlockIndex(db btcdb.Db, last *btcutil.Block, proofOfStake bool) (block *btcutil.Block) {
	block = last
	for true {
		if block == nil {
			break
		}
		//TODO dirty workaround, ppcoin doesn't point to genesis block
		if block.Height() == 0 {
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

// GetNextTargetRequired TODO(kac-) golint
// https://github.com/ppcoin/ppcoin/blob/v0.4.0ppc/src/main.cpp#L902
func GetNextTargetRequired(params btcnet.Params, db btcdb.Db, last *btcutil.Block, proofOfStake bool) (compact uint32) {
	if last == nil {
		return params.PowLimitBits // genesis block
	}
	prev := getLastBlockIndex(db, last, proofOfStake)
	if prev == nil {
		return initialHashTargetBits // first block
	}
	block, _ := db.FetchBlockBySha(&prev.MsgBlock().Header.PrevBlock)
	prevPrev := getLastBlockIndex(db, block, proofOfStake)
	if prevPrev == nil {
		return initialHashTargetBits // second block
	}
	actualSpacing := prev.MsgBlock().Header.Timestamp.Unix() - prevPrev.MsgBlock().Header.Timestamp.Unix()
	newTarget := btcchain.CompactToBig(prev.MsgBlock().Header.Bits)
	var targetSpacing int64
	if proofOfStake {
		targetSpacing = stakeTargetSpacing
	} else {
		targetSpacing = minInt(targetSpacingWorkMax, stakeTargetSpacing*(1+last.Height()-prev.Height()))
	}
	interval := targetTimespan / targetSpacing
	tmp := new(big.Int)
	newTarget.Mul(newTarget,
		tmp.SetInt64(interval-1).Mul(tmp, big.NewInt(targetSpacing)).Add(tmp, big.NewInt(actualSpacing+actualSpacing)))
	newTarget.Div(newTarget, tmp.SetInt64(interval+1).Mul(tmp, big.NewInt(targetSpacing)))
	if newTarget.Cmp(params.PowLimit) > 0 {
		newTarget = params.PowLimit
	}
	return btcchain.BigToCompact(newTarget)
}

// CBlkIdx TODO(kac-) golint
type CBlkIdx struct {
	Prev                  *CBlkIdx
	Next                  *CBlkIdx
	Height                uint32
	Mint                  uint64
	Supply                uint64
	GeneratedModifier     bool
	EntropyBit            bool
	ProofOfStake          bool
	StakeModifier         []byte
	StakeModifierChecksum []byte
	HashProofOfStake      []byte
	PrevOutHash           []byte
	PrevOutN              uint32
	StakeTime             uint32
	HashMerkleRoot        []byte
	BlockHash             []byte
	BlockTrust            []byte
	ChainTrust            []byte
}

// ReadCBlockIndex TODO(kac-) golint
func ReadCBlockIndex(blockIndexFile string) (rootIndex *CBlkIdx) {

	fi, _ := os.Open(blockIndexFile)
	defer fi.Close()
	var r io.Reader = fi

	if strings.HasSuffix(blockIndexFile, ".bz2") {
		r = bzip2.NewReader(r)
	} else if strings.HasSuffix(blockIndexFile, ".gz") {
		r, _ = gzip.NewReader(r)
	}
	ci := csv.NewReader(r)
	ci.Read() // header
	var blk, root, prev *CBlkIdx
	for rec, err := ci.Read(); err == nil; rec, err = ci.Read() {
		blk = new(CBlkIdx)
		i, _ := strconv.Atoi(rec[1])
		blk.Height = uint32(i)
		i64, _ := strconv.Atoi(rec[2])
		blk.Mint = uint64(i64)
		i64, _ = strconv.Atoi(rec[3])
		blk.Supply = uint64(i64)
		i, _ = strconv.Atoi(rec[4])
		blk.GeneratedModifier = i == 1
		i, _ = strconv.Atoi(rec[5])
		blk.EntropyBit = i == 1
		i, _ = strconv.Atoi(rec[6])
		blk.ProofOfStake = i == 1
		by, _ := hex.DecodeString(rec[7])
		blk.StakeModifier = by
		by, _ = hex.DecodeString(rec[8])
		blk.StakeModifierChecksum = by
		by, _ = hex.DecodeString(rec[9])
		blk.HashProofOfStake = by
		sa := strings.Split(rec[10], ":")
		by, _ = hex.DecodeString(sa[0])
		blk.PrevOutHash = by
		i, _ = strconv.Atoi(sa[1])
		blk.PrevOutN = uint32(i)
		i, _ = strconv.Atoi(rec[11])
		blk.StakeTime = uint32(i)
		by, _ = hex.DecodeString(rec[12])
		blk.HashMerkleRoot = by
		by, _ = hex.DecodeString(rec[13])
		blk.BlockHash = by
		by, _ = hex.DecodeString(rec[14])
		blk.BlockTrust = by
		by, _ = hex.DecodeString(rec[15])
		blk.ChainTrust = by

		if prev == nil {
			root = blk
		} else {
			blk.Prev = prev
			prev.Next = blk
		}
		prev = blk
	}
	return root
}
