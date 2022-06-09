package helper

import (
	"encoding/binary"
	"os"
)

const BlockSize = 4096

//  Based on the below cal
const MaxLeafSize = 30

//  DiskBlock -- size 4096
type DiskBlock struct {
	Id                  uint64   // 8
	CurenLeafSize       uint64   // 8
	CurrentChildrenSize uint64   // 8
	ChildrenBlocksIds   []uint64 // its in range (8id * 30 Sub-id)
	DataSet             []*Pairs // 3810 = 127*30
	// 3810+8+8+8 = 3834
	// 4096-3834 = 262
	// 262-(8*30) = 22
}

//  22 bytes wasted

// SetData takes
func (block *DiskBlock) SetData(data []*Pairs) {
	block.DataSet = data
	block.CurenLeafSize = uint64(len(data))
}

func (block *DiskBlock) SetChildren(childrenBlockIds []uint64) {
	block.ChildrenBlocksIds = childrenBlockIds
	block.CurrentChildrenSize = uint64(len(childrenBlockIds))
}

func Uint64ToBytes(index uint64) []byte {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(index))
	return b
}

func Uint64FromBytes(b []byte) uint64 {
	return uint64(binary.LittleEndian.Uint64(b))
}

type BlockService struct {
	file *os.File
}

func (bs BlockService) GetLatestBlockID() (int64, error) {
	fi, err := bs.file.Stat()
	if err != nil {
		return -1, err
	}

	length := fi.Size()
	if length == 0 {
		return -1, nil
	}

	// Calculate page number required to  be fetched from disk
	return (int64(fi.Size()) / int64(BlockSize)) - 1, nil
}

func (bs *BlockService) RootBlockExists() bool {
	lastestBlockID, err := bs.GetLatestBlockID()
	// fmt.Println(lastestBlockID)
	if err != nil {
		// Need to write a new block, Apparently no fileor path error
		return false
	} else if lastestBlockID == -1 {
		return false
	} else {
		return true
	}
}

func (bs *BlockService) GetBufferFromBlock(block *DiskBlock) []byte {
	blockBufer := make([]byte, BlockSize)
	blockOffset := 0

	// Write Block Index
	copy(blockBufer[blockOffset:], Uint64ToBytes(block.Id))
	blockOffset += 8
	copy(blockBufer[blockOffset:], Uint64ToBytes(block.CurenLeafSize))
	blockOffset += 8
	copy(blockBufer[blockOffset:], Uint64ToBytes(block.CurrentChildrenSize))
	blockOffset += 8

	// Write actual pair now
	for i := 0; i < int(block.CurenLeafSize); i++ {
		copy(blockBufer[blockOffset:], ConvertPairsToBytes(block.DataSet[i]))
		blockOffset += PairSize
	}

	// Read childrenBlock Indexes
	for i := 0; i < int(block.CurrentChildrenSize); i++ {
		copy(blockBufer[blockOffset:], Uint64ToBytes(block.ChildrenBlocksIds[i]))
		blockOffset += 8
	}
	return blockBufer

}

func (bs *BlockService) WriteBlockToDisk(block *DiskBlock) error {
	seekOffset := BlockSize * block.Id
	blockBuffer := bs.GetBufferFromBlock(block)
	_, err := bs.file.Seek(int64(seekOffset), 0)
	if err != nil {
		return err
	}
	_, err = bs.file.Write(blockBuffer)
	if err != nil {
		return err
	}
	return nil
}

func (bs *BlockService) NewBlock() (*DiskBlock, error) {
	latestBlockID, err := bs.GetLatestBlockID()
	block := &DiskBlock{}
	if err != nil {
		// This means file doesnt exist
		block.Id = 0
	} else {
		block.Id = uint64(latestBlockID) + 1
	}
	block.CurenLeafSize = 0
	err = bs.WriteBlockToDisk(block)
	if err != nil {
		return nil, err
	}
	return block, nil

}

func (bs *BlockService) GetBlockFromBuffer(blockBuffer []byte) *DiskBlock {
	blockOffset := 0
	block := &DiskBlock{}

	// Read Block Index
	block.Id = Uint64FromBytes(blockBuffer[blockOffset:])
	blockOffset += 8
	block.CurenLeafSize = Uint64FromBytes(blockBuffer[blockOffset:])
	blockOffset += 8
	block.CurrentChildrenSize = Uint64FromBytes(blockBuffer[blockOffset:])
	blockOffset += 8

	// Read actual pairs now
	block.DataSet = make([]*Pairs, block.CurenLeafSize)
	for i := 0; i < int(block.CurenLeafSize); i++ {
		block.DataSet[i] = ConvertBytesToPairs(blockBuffer[blockOffset:])
		blockOffset += PairSize
	}
	// Read children block indexes
	block.ChildrenBlocksIds = make([]uint64, block.CurrentChildrenSize)
	for i := 0; i < int(block.CurrentChildrenSize); i++ {
		block.ChildrenBlocksIds[i] = Uint64FromBytes(blockBuffer[blockOffset:])
		blockOffset += 8
	}
	return block
}

func (bs *BlockService) GetBlockFromDiskByBlockNumber(index int64) (*DiskBlock, error) {
	if index < 0 {
		panic("Index less than 0 asked")
	}

	offset := index * BlockSize
	_, err := bs.file.Seek(offset, 0)
	if err != nil {
		return nil, err
	}

	blockBuffer := make([]byte, BlockSize)
	_, err = bs.file.Read(blockBuffer)
	if err != nil {
		return nil, err
	}
	block := bs.GetBlockFromBuffer(blockBuffer)
	return block, nil
}

func (bs *BlockService) GetRootBlock() (*DiskBlock, error) {

	/*
		1. Check if root block exist
		2. If exists, fetch it, else initialize a new block
	*/
	if !bs.RootBlockExists() {
		// Need to write a new block
		return bs.NewBlock()
	}

	return bs.GetBlockFromDiskByBlockNumber(0)
}

func (bs *BlockService) ConvertDiskNodeToBlock(node *DiskNode) *DiskBlock {
	block := &DiskBlock{Id: node.blockID}
	tempElements := make([]*Pairs, len(node.getElements()))
	for index, element := range node.getElements() {
		tempElements[index] = element
	}
	block.SetData(tempElements)
	tempBlockIDs := make([]uint64, len(node.GetChildBlockIDs()))
	for index, childBlockID := range node.GetChildBlockIDs() {
		tempBlockIDs[index] = childBlockID
	}
	block.SetChildren(tempBlockIDs)
	return block

}

func (bs *BlockService) ConvertBlockToDiskNode(block *DiskBlock) *DiskNode {
	node := &DiskNode{
		blockID:      block.Id,
		blockService: bs,
		keys:         make([]*Pairs, block.CurenLeafSize),
	}

	for index := range node.keys {
		node.keys[index] = block.DataSet[index]
	}
	node.childrenBlockIDs = make([]uint64, block.CurrentChildrenSize)
	for index := range node.childrenBlockIDs {
		node.childrenBlockIDs[index] = block.ChildrenBlocksIds[index]
	}
	return node
}

func (bs *BlockService) GetNodeAtBlockID(blockID uint64) (*DiskNode, error) {
	block, err := bs.GetBlockFromDiskByBlockNumber(int64(blockID))
	if err != nil {
		return nil, err
	}
	return bs.ConvertBlockToDiskNode(block), nil
}

func (bs *BlockService) SaveNewNodeToDisk(n *DiskNode) error {
	// Get block ID to be assigned to this block
	lastestBlockID, err := bs.GetLatestBlockID()
	if err != nil {
		return err
	}
	n.blockID = uint64(lastestBlockID) + 1
	block := bs.ConvertDiskNodeToBlock(n)
	return bs.WriteBlockToDisk(block)
}

func (bs *BlockService) UpdateNodeToDisk(n *DiskNode) error {
	block := bs.ConvertDiskNodeToBlock(n)
	return bs.WriteBlockToDisk(block)
}

func (bs *BlockService) UpdateRootNode(n *DiskNode) error {
	n.blockID = 0
	return bs.UpdateNodeToDisk(n)
}

func NewBlockService(file *os.File) *BlockService {
	return &BlockService{file}
}

/**
. Dynamicaly calculate blockSize
2. Then based on the blocksize, calculate the maxLeafSize
*/

func (bs *BlockService) GetMaxLeafSize() int {
	return MaxLeafSize
}
