package helper

// importing ttthe neccessary packages
import (
	"encoding/binary"
	"fmt"
)

// initializing constants
// pairSize represent the length of byte store on the disk Sector
// Sector takes a byte of 512 we we approximately 4 pairs to get store with 12bytes remaining
// 2 bytes for keylength
// 2 bytes for valuelength
// 4 bytes for 4 pairs = 16
// each pairSize on sector 124 out 512 bytes
// 124 pairSize for 4 pairs = 496
//  496 + 16 = 512 bytes on block Size
const PairSize = 124

// maxlength of a key
const maxKeyLength = 30

//maxlength of a value
const maxValueLength = 93

// A pair struct
type Pairs struct {
	KeyLen   uint16 //2
	ValueLen uint16 //2
	Key      string //30
	Value    string //93
}

//  A  setKey method to put a key  and generate keylen
//  Takes a key as parameter to set a key and keylen for the pair
func (p *Pairs) SetKey(key string) {
	p.Key = key
	p.KeyLen = uint16(len(key))
}

// A setValue method sets a value and value length
//Takes value to be set as a string
func (p *Pairs) SetValue(value string) {
	p.Value = value
	p.ValueLen = uint16(len(value))
}

// newPair creates a new Pair object.
// Takes a key and a value
func NewPair(key string, value string) *Pairs {
	pair := new(Pairs)
	pair.SetKey(key)
	pair.SetValue(value)

	return pair
}

func (p *Pairs) Validate() error {
	if len(p.Key) > maxKeyLength {
		return fmt.Errorf("key length should not be more than 30, currently it is %d ", len(p.Key))
	}
	if len(p.Value) > maxValueLength {
		return fmt.Errorf("value length should not be more than 93, currently it is %d", len(p.Value))
	}
	return nil
}

// Before writting our pair object on disk
// we need to convert the individual attributes to bytes

// The function convert value uint16 to a a byte and return []byte
func uint16ToBytes(value uint16) []byte {
	byteVal := make([]byte, 2)
	binary.LittleEndian.PutUint16(byteVal, value)
	return byteVal
}

// Takes a slice of bytes and convert to uint
func uint16FromBytes(b []byte) uint16 {
	unitVal := uint16(binary.LittleEndian.Uint64(b))
	return unitVal
}

// takes a object pair an converts to a byte array [0000000]
func ConvertPairsToBytes(pair *Pairs) []byte {
	// initialize sliice of byte
	bytePair := make([]byte, PairSize)
	var pairOffset uint16
	pairOffset = 0
	copy(bytePair[pairOffset:], uint16ToBytes(pair.KeyLen))
	pairOffset += 2
	copy(bytePair[pairOffset:], uint16ToBytes(pair.ValueLen))
	pairOffset += 2
	keyByte := []byte(pair.Key)
	copy(bytePair[pairOffset:], keyByte[:pair.KeyLen])
	pairOffset += pair.KeyLen
	valByte := []byte(pair.Value)
	copy(bytePair[pairOffset:], valByte[:pair.ValueLen])
	return bytePair
}

// Convert byte to Pair object
func ConvertBytesToPairs(pairByte []byte) *Pairs {
	pair := new(Pairs)
	var pairOffset uint16
	pairOffset = 0
	//Read key length
	pair.KeyLen = uint16FromBytes(pairByte[pairOffset:])
	pairOffset += 2
	//Read Vaue length
	pair.ValueLen = uint16FromBytes(pairByte[pairOffset:])
	pairOffset += 2
	pair.Key = string(pairByte[pairOffset : pairOffset+pair.KeyLen])
	pairOffset += pair.KeyLen
	pair.Value = string(pairByte[pairOffset : pairOffset+pair.ValueLen])
	return pair
}
