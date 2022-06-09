package helper

import (
	"fmt"
	"os"
	"testing"
)

func initBlockService() *BlockService {
	dir_path := "/home/rozz/go/src/KeyValueStore/other/helper"
	path := fmt.Sprintf("%s/db/test.db", dir_path)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		db_path := fmt.Sprintf("%s/db", dir_path)
		os.Mkdir(db_path, os.ModePerm)
	}
	if _, err := os.Stat(path); err == nil {
		err := os.Remove(path)
		if err != nil {
			panic(err)
		}
	}
	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		panic(err)
	}
	return NewBlockService(file)
}

func TestShouldGetNegativeIfBlockNotPresent(t *testing.T) {
	blockService := initBlockService()
	latestBlockID, _ := blockService.GetLatestBlockID()
	if latestBlockID != -1 {
		t.Error("Shoud get negative block id")
	}
}

func TestShouldSuccessfullyInitializeNewBlock(t *testing.T) {
	blockService := initBlockService()
	block, err := blockService.GetRootBlock()
	if err != nil {
		t.Error(err)
	}
	if block.Id != 0 {
		t.Error("Root Block id should be zero")
	}
	if block.CurrentChildrenSize != 0 {
		t.Error("Block leeaf size should be zero")
	}
}

func TestShouldSaveNewBlockOnDisk(t *testing.T) {
	blockService := initBlockService()
	block, err := blockService.GetRootBlock()
	if err != nil {
		t.Error(err)
	}
	if block.Id != 0 {
		t.Error("Root Block id should be zero")
	}
	if block.CurrentChildrenSize != 0 {
		t.Error("Block leaf size should be zero")
	}
	elements := make([]*Pairs, 3)
	elements[0] = NewPair("hola", "amigos")
	elements[1] = NewPair("foo", "bar")
	elements[2] = NewPair("gooz", "bumps")
	block.SetData(elements)
	err = blockService.WriteBlockToDisk(block)
	if err != nil {
		t.Error(err)
	}
	block, err = blockService.GetRootBlock()
	if err != nil {
		t.Error(err)
	}
	if len(block.DataSet) == 0 {
		t.Error("Length of data field should not be zero")
	}
}
func TestShouldConvertPairToAndFromBytes(t *testing.T) {
	pair := &Pairs{}
	pair.SetKey("Hola")
	pair.SetValue("Amigos")
	pairBytes := ConvertPairsToBytes(pair)
	convertedPair := ConvertBytesToPairs(pairBytes)

	if pair.KeyLen != convertedPair.KeyLen || pair.ValueLen != convertedPair.ValueLen {
		t.Error("Lengths do not match")
	}

	if pair.Key != convertedPair.Key || pair.Value != convertedPair.Value {
		t.Error("Values do not match")
	}
}

func TestShouldConvertBlockToAndFromBytes(t *testing.T) {
	blockService := initBlockService()
	block := &DiskBlock{}
	block.SetChildren([]uint64{2, 3, 4, 6})

	elements := make([]*Pairs, 3)
	elements[0] = NewPair("hola", "amigos")
	elements[1] = NewPair("foo", "bar")
	elements[2] = NewPair("gooz", "bumps")
	block.SetData(elements)
	blockBuffer := blockService.GetBufferFromBlock(block)
	convertedBlock := blockService.GetBlockFromBuffer(blockBuffer)

	if convertedBlock.ChildrenBlocksIds[2] != 4 {
		t.Error("Should contain 4 at 2nd index")
	}

	if len(convertedBlock.DataSet) != len(block.DataSet) {
		t.Error("Length of blocks should be same")
	}

	if convertedBlock.DataSet[1].Key != block.DataSet[1].Key {
		t.Error("Keys dont match")
	}

	if convertedBlock.DataSet[2].Value != block.DataSet[2].Value {
		t.Error("Values dont match")
	}
}

func TestShouldConvertToAndFromDiskNode(t *testing.T) {
	bs := initBlockService()
	node := &DiskNode{}
	node.blockID = 55
	elements := make([]*Pairs, 3)
	elements[0] = NewPair("hola", "amigos")
	node.keys = elements
	node.childrenBlockIDs = []uint64{1000, 10001}
	block := bs.ConvertDiskNodeToBlock(node)

	if block.Id != 55 {
		t.Error("Should have same block ID as node block ID")
	}
	if block.ChildrenBlocksIds[1] != 10001 {
		t.Error("Block ids should match")
	}

	nodeFromBlock := bs.ConvertBlockToDiskNode(block)

	if nodeFromBlock.blockID != node.blockID {
		t.Error("Block ids should match")
	}

	if nodeFromBlock.childrenBlockIDs[0] != 1000 {
		t.Error("Child Block ids should match")
	}
	if nodeFromBlock.keys[0].Key != "hola" {
		t.Error("Data elements should match")
	}
}
