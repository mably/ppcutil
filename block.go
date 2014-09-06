// Copyright (c) 2014-2014 PPCD developers.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package ppcutil

import (
	"github.com/mably/btcutil"
)

func BlockUnixTime(block *btcutil.Block) int64 {
	return block.MsgBlock().Header.Timestamp.Unix()
}

func IsBlockProofOfStake(block *btcutil.Block) bool {
	return IsMsgBlockProofOfStake(block.MsgBlock())
}