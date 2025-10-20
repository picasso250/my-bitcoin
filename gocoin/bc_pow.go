package gocoin

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"log"
	"math"
	"math/big"
)

// difficulty 定义了挖矿的难度。
// 在比特币中，这个值会定期调整，但在这里我们使用一个固定的值。
// 值越高，找到有效哈希所需的前导零就越多，难度越大。
const difficulty = 12

// ProofOfWork 结构体包含了指向区块的指针和挖矿目标
type ProofOfWork struct {
	block  *Block
	target *big.Int
}

// NewProofOfWork 构建并返回一个新的 ProofOfWork 实例
func NewProofOfWork(b *Block) *ProofOfWork {
	target := big.NewInt(1)
	// target = 1 * 2^(256 - difficulty)
	// 难度值越大，target越小，计算出小于target的哈希就越难
	target.Lsh(target, uint(256-difficulty))

	pow := &ProofOfWork{b, target}

	return pow
}

// prepareData 准备用于哈希计算的数据
// 它将区块的各个字段（除了Hash本身）与 nonce 组合在一起
func (pow *ProofOfWork) prepareData(nonce int) []byte {
	// 将 int64 转换为 []byte
	timestampBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(timestampBytes, uint64(pow.block.Timestamp))

	// 将 int 转换为 []byte
	nonceBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(nonceBytes, uint64(nonce))

	data := bytes.Join(
		[][]byte{
			pow.block.PrevBlockHash,
			pow.block.HashTransactions(),
			timestampBytes,
			nonceBytes,
		},
		[]byte{},
	)

	return data
}

// Run 执行工作量证明（挖矿）
// 它会不断尝试不同的 nonce，直到找到一个哈希值小于 target
func (pow *ProofOfWork) Run() (int, []byte) {
	var hashInt big.Int
	var hash [32]byte
	nonce := 0

	fmt.Printf("Mining a new block with target: %x\n", pow.target.Bytes())
	for nonce < math.MaxInt64 {
		data := pow.prepareData(nonce)
		hash = sha256.Sum256(data)
		
		if math.Remainder(float64(nonce), 100000) == 0 {
			fmt.Printf("\r%x", hash)
		}

		hashInt.SetBytes(hash[:])

		// Cmp compares x and y and returns:
		// -1 if x <  y
		//  0 if x == y
		// +1 if x >  y
		if hashInt.Cmp(pow.target) == -1 {
			fmt.Printf("\r%x\n\n", hash)
			break
		} else {
			nonce++
		}
	}

	return nonce, hash[:]
}

// Validate 验证一个区块的工作量证明是否有效
func (pow *ProofOfWork) Validate() bool {
	var hashInt big.Int

	data := pow.prepareData(pow.block.Nonce)
	hash := sha256.Sum256(data)
	hashInt.SetBytes(hash[:])

	isValid := hashInt.Cmp(pow.target) == -1

	return isValid
}

// IntToHex converts an int64 to a byte array
func IntToHex(num int64) []byte {
	buff := new(bytes.Buffer)
	err := binary.Write(buff, binary.BigEndian, num)
	if err != nil {
		log.Panic(err)
	}

	return buff.Bytes()
}