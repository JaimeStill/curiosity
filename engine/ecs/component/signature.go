package component

// MaxCID is the upper bound on the number of component types registered
// simultaneously within one engine instance. It sizes both the archetype
// storage's per-CID column index (a flat array, per D-030 §6) and the
// Signature bitset below; per-Signature memory cost at MaxCID = 2048 is
// 256 bytes (32 uint64 words).
const MaxCID = 2048

// Signature is a bitset over component IDs: bit (cid - 1) is set iff
// the corresponding component is present. Stored as [MaxCID/64]uint64
// so the full ID space (1..MaxCID) fits in 32 native-width words.
//
// The (cid - 1) offset is the consequence of InvalidID = 0: valid CIDs
// (1..MaxCID) map to bit positions 0..MaxCID-1 within the bitset.
type Signature [MaxCID / 64]uint64

/*
	bit = cid - 1: cid is index + 1 since 0 is the invalid ID.
	bit / 64 = word: result is in the range 0..31 (32 possible words)
	bit % 64 = position: result is in the range 0..63 (64 possible positions)

	s[bit / 64] (which word) |= 1 << (bit % 64) (which bit IN that word)

	1 << k: left-shift the value 1 by k positions.
					The result has one bit set, at position k
					counting from the least-significant bit (LSB = posiiton 0).

	1 << (bit % 64): which position is the bit set in 0..63

	position: 63 62 61 ... 3 2 1 0
  bits:      0  0  0 ... 0 0 0 1

  position: 63 62 61 60  59 58 57 56  ...  3 2 1 0
  bits:      1  0  0  0   0  0  0  0  ...  0 0 0 0
  hex:         \_8_/        \_0_/     ...   \_0_/
              digit 15     digit 14        digit 0

	hex digit:    0    1    2    3    4    5    6    7    8    9    A    B    C    D    E    F
  binary:    0000 0001 0010 0011 0100 0101 0110 0111 1000 1001 1010 1011 1100 1101 1110 1111

	A single hex digit is called a nibble and covers 4 bits. A byte is 8 bits = 2 hex digits.
	bit position in nibble: 3 2 1 0
  contributes value:      8 4 2 1


	A single bit set -> nibble value is exactly one of {1, 2, 4, 8}
	two bits set     -> nibble value is the sum (e.g., 0011 = 3, 0101 = 5, 0111 = 7)

	1 <<  0 = 0x0001    1 <<  8 = 0x0100
  1 <<  1 = 0x0002    1 <<  9 = 0x0200
  1 <<  2 = 0x0004    1 << 10 = 0x0400
  1 <<  3 = 0x0008    1 << 11 = 0x0800
  1 <<  4 = 0x0010    1 << 12 = 0x1000
  1 <<  5 = 0x0020    1 << 13 = 0x2000
  1 <<  6 = 0x0040    1 << 14 = 0x4000
  1 <<  7 = 0x0080    1 << 15 = 0x8000

	|  cid | bit = cid - 1 | bit / 64 (word) | bit % 64 (pos) |             1 << (bit % 64)               |
	|------|---------------|-----------------|----------------|-------------------------------------------|
	|    1 |             0 |               0 |              0 | 0x...0001             (bit 0 of word 0)   |
	|   64 |            63 |               0 |             63 | 0x8000_0000_0000_0000 (top bit of word 0) |
  |   65 |            64 |               1 |              0 | 0x...0001             (bit 0 of word 1)   |
  |  128 |           127 |               1 |             63 | top bit of word 1                         |
  | 2048 |          2047 |              31 |             63 | top bit of word 31                        |

	s.Set(1) (bit 0 of word 0):
	s[0] = 0x0000_0000_0000_0001

	s.Set(2) (bit 1 of word 0):
	s[0] = 0x0000_0000_0000_0002

	s.Set(65) (bit 0 of word 1):
	s[1] = 0x0000_0000_0000_0001
*/

// Set marks cid as present in s. The caller must pass a valid CID
// (cid != InvalidID and cid <= MaxCID); the precondition is upheld
// structurally by sourcing CIDs from IDFor, which never returns
// InvalidID and never returns values above MaxCID under normal
// registration patterns.
func (s *Signature) Set(cid ID) {
	bit := cid - 1
	s[bit/64] |= 1 << (bit % 64)
}

// Has reports whether cid is present in s. Same valid-CID precondition
// as Set.
func (s Signature) Has(cid ID) bool {
	bit := cid - 1
	return s[bit/64]&(1<<(bit%64)) != 0
}

// Contains reports whether s is a superset of other: every CID set in
// other is also set in s. The intended use is query iteration —
// testing whether an archetype's signature satisfies the required set
// of a query.
func (s Signature) Contains(other Signature) bool {
	for i := range s {
		if s[i]&other[i] != other[i] {
			return false
		}
	}
	return true
}

// SignatureOf returns a Signature with each CID in cids set. The slice
// is read-only — SignatureOf neither retains nor mutates it.
func SignatureOf(cids []ID) Signature {
	var sig Signature
	for _, cid := range cids {
		sig.Set(cid)
	}
	return sig
}
