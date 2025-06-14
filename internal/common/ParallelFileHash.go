package common

import (
	"fmt"
	"github.com/zeebo/blake3"
	"io"
	"os"
)

func ParallelFileHash(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	const chunkSize = 1 << 20 // 1MB
	hasher := blake3.New()

	buf := make([]byte, chunkSize)
	for {
		n, err := file.Read(buf)
		if err != nil && err != io.EOF {
			return "", err
		}
		if n == 0 {
			break
		}
		hasher.Write(buf[:n])
	}

	return fmt.Sprintf("%x", hasher.Sum(nil)), nil
}
