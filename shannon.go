/* Copyright (c) 2017 Jeffrey Massung
 *
 * This software is provided 'as-is', without any express or implied
 * warranty.  In no event will the authors be held liable for any damages
 * arising from the use of this software.
 *
 * Permission is granted to anyone to use this software for any purpose,
 * including commercial applications, and to alter it and redistribute it
 * freely, subject to the following restrictions:
 *
 * 1. The origin of this software must not be misrepresented; you must not
 *    claim that you wrote the original software. If you use this software
 *    in a product, an acknowledgment in the product documentation would be
 *    appreciated but is not required.
 *
 * 2. Altered source versions must be plainly marked as such, and must not be
 *    misrepresented as being the original software.
 *
 * 3. This notice may not be removed or altered from any source distribution.
 */

package shannon

import (
	"bytes"
	"errors"
	"fmt"
	"math"
	"sort"
)

// Code is a Shannon-Fano code point.
type Code struct {
	Char rune
	Prob float64
	Bits uint32
	Size int
}

// Table is a simple lookup-map for encoding and decoding.
type Table map[rune]Code

// BuildTable returns Shannon-Fano table for encoding an decoding.
func BuildTable(freq map[rune]float64) (table Table) {
	var divide func([]Code)

	// initialize an empty list capable of holding all codes
	codes := make([]Code, 0, len(freq))

	// build up the list of codes
	for r, prob := range freq {
		codes = append(codes, Code{
			Char: r,
			Prob: prob,
		})
	}

	// sort the codes by probability
	sort.Slice(codes, func(a, b int) bool {
		return codes[a].Prob > codes[b].Prob
	})

	// recursively divide the codes, building the table
	divide = func(codes []Code) {
		var p int

		if len(codes) < 2 {
			return
		}

		// sum the total probability for this slice
		prob := 0.0
		for _, code := range codes {
			prob += code.Prob
		}

		// probability of the left half
		left := codes[0].Prob
		best := 1.0

		// find the optimal pivot
		for p = 1; p < len(codes)-1; p++ {
			if diff := math.Abs((prob - left) - left); diff < best {
				best = diff
			} else {
				break
			}

			// tally the probability on the left
			left += codes[p].Prob
		}

		// update the left half with 0's and right half with 1's
		for i := 0; i < len(codes); i++ {
			codes[i].Bits <<= 1
			codes[i].Size++

			if i >= p {
				codes[i].Bits |= 1
			}
		}

		// subdivide each branch
		divide(codes[:p])
		divide(codes[p:])
	}

	// perform the subdivision
	divide(codes)

	// create the resulting table
	table = make(Table)

	// construct the table from all the built codes
	for _, code := range codes {
		table[code.Char] = code
	}

	return
}

// BuildTableFromString builds a Shannon-Fano table from a string.
func BuildTableFromString(s string) Table {
	freq := make(map[rune]int)
	prob := make(map[rune]float64)

	// build a frequency table of all the runes
	for _, r := range s {
		freq[r] = freq[r] + 1
	}

	// build a probability table from the frequencies
	for r, n := range freq {
		prob[r] = float64(n) / float64(len(s))
	}

	return BuildTable(prob)
}

// BuildTableFromOrderedRunes builds a Shannon-Fano table from a slice of
// runes that are in order of most-to-least frequent. If the same rune
// occurs more than once, their probability is summed.
func BuildTableFromOrderedRunes(runes []rune) Table {
	prob := make(map[rune]float64)

	// summation of 0..len
	n := len(runes)
	m := float64(n*(n-1)/2 + n)

	// build the probability map
	for i, r := range runes {
		prob[r] += float64(n-i) / m
	}

	return BuildTable(prob)
}

// Encode a string using a Shannon-Fano table.
func (t Table) Encode(s string) ([]uint32, int, error) {
	if len(t) == 0 {
		return nil, 0, errors.New("empty shannon-fano table")
	}

	// add the first set of bits
	bitVec := make([]uint32, 1)
	size := 0

	// encode each rune in the string
	for _, r := range s {
		code, found := t[r]
		if !found {
			return nil, 0, fmt.Errorf("rune '%c' not found in shannon-fano table", r)
		}

		// pack if it fits completely
		if n := size & 0x1F; n+code.Size < 0x20 {
			bitVec[len(bitVec)-1] |= code.Bits << uint(0x20-n-code.Size)
		} else {
			n = code.Size - (0x20 - n)

			// append the last few bits
			bitVec[len(bitVec)-1] |= code.Bits >> uint(n)

			// create a new entry with the remaining bits
			bitVec = append(bitVec, code.Bits<<uint(0x20-n))
		}

		// tally the total size
		size += code.Size
	}

	return bitVec, size, nil
}

// Decode a bit vector into a string using a Shannon-Fano table.
func (t Table) Decode(bitVec []uint32, size int) (string, error) {
	var b bytes.Buffer
	var v uint32

	// ensure there are enough bits to decode
	if len(bitVec) <= size/32 {
		return "", errors.New("invalid bit vector")
	}

	// current bits/size being tested
	bits, n := uint32(0), 0

	// pop the first set of bits in the vector
	v, bitVec = bitVec[0], bitVec[1:]

	// pop bits until vector is completely consumed
	for i := 0; i < size; {
		bits, v = bits<<1|(v>>31), v<<1

		// tally bits, test for failure
		if n++; n > 32 {
			return "", errors.New("invalid bit vector or missing shannon code")
		}

		// pop the next set of bits from the bit vector
		if i++; i&0x1F == 0 {
			v, bitVec = bitVec[0], bitVec[1:]
		}

		// find a matching code
		for r, code := range t {
			if code.Size != n || code.Bits != bits {
				continue
			}

			// found matching code
			b.WriteRune(r)

			// reset the bits/size being tested
			bits, n = 0, 0
			break
		}
	}

	// ensure all bits were used
	if n != 0 {
		return b.String(), errors.New("encoded bits remaining")
	}

	return b.String(), nil
}
