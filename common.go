// Copyright (c) 2014-2014 PPCD developers.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package ppcutil

import (
	"math/big"
)

func TargetToDifficulty(bits uint32) (diff float64) {
	nShift := (bits >> 24) & 0xff
	diff = float64(0x0000ffff) / float64(bits&0x00ffffff)
	for ; nShift < 29; nShift++ {
		diff *= 256.0
	}
	for ; nShift > 29; nShift-- {
		diff /= 256.0
	}
	return
}

func DifficultyToTarget(diff float64) (target *big.Int) {
	mantissa := 0x0000ffff / diff
	exp := 1
	tmp := mantissa
	for tmp >= 256.0 {
		tmp /= 256.0
		exp++
	}
	for i := 0; i < exp; i++ {
		mantissa *= 256.0
	}
	target = new(big.Int).Lsh(big.NewInt(int64(mantissa)), uint(26-exp)*8)
	return
}
