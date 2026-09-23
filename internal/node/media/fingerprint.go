package media

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"

	"github.com/zeebo/blake3"
)

const fingerprintChunkSize int64 = 1 << 20

// FileFingerprint returns a fast fingerprint based on the file size and up to
// 1 MiB sampled from the beginning, middle, and end of the file. It is not a
// full-content hash: changes outside the sampled ranges are not detected.
func FileFingerprint(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return "", err
	}

	fileSize := info.Size()
	hasher := blake3.New()
	var encoded [8]byte
	binary.LittleEndian.PutUint64(encoded[:], uint64(fileSize))
	_, _ = hasher.Write(encoded[:])

	sampleSize := min(fileSize, fingerprintChunkSize)
	offsets := []int64{
		0,
		max((fileSize-sampleSize)/2, 0),
		max(fileSize-sampleSize, 0),
	}
	buf := make([]byte, sampleSize)
	var previousOffset int64 = -1
	for _, offset := range offsets {
		if offset == previousOffset {
			continue
		}
		previousOffset = offset

		binary.LittleEndian.PutUint64(encoded[:], uint64(offset))
		_, _ = hasher.Write(encoded[:])
		n, err := file.ReadAt(buf, offset)
		if err != nil && err != io.EOF {
			return "", err
		}
		_, _ = hasher.Write(buf[:n])
	}

	return fmt.Sprintf("%x", hasher.Sum(nil)), nil
}
