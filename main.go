package main

import (
	"io"
	"log"
	"os"
)

const (
	MAX_FILE_SIZE = 10_000_000
	ORIGINAL      = "original.bin"
	MODDED        = "modded.bin"
	TARGET        = "target.bin"
)

type Differences struct {
	Offset int
	Length int
	Bytes  []byte
}

func findDifferences(src, dst string) ([]Differences, error) {
	sourceFile, err := os.Open(src)
	if err != nil {
		return nil, err
	}
	defer sourceFile.Close()

	destinationFile, err := os.Open(dst)
	if err != nil {
		return nil, err
	}
	defer destinationFile.Close()

	var diffs []Differences
	bufferSrc := make([]byte, MAX_FILE_SIZE)
	bufferDst := make([]byte, MAX_FILE_SIZE)

	nSrc, errSrc := sourceFile.Read(bufferSrc)
	if errSrc != nil && errSrc != io.EOF {
		return nil, errSrc
	}

	nDst, errDst := destinationFile.Read(bufferDst)
	if errDst != nil && errDst != io.EOF {
		return nil, errDst
	}

	minLen := min(nSrc, nDst)
	differencesFound := false
	var bytes []byte

	for i := 0; i < minLen; i++ {
		if bufferSrc[i] == bufferDst[i] {
			if differencesFound {
				// save patch
				diffs = append(diffs, Differences{
					Offset: i - len(bytes),
					Length: len(bytes),
					Bytes:  bytes,
				})
				// reset
				bytes = []byte{}
			}
			differencesFound = false
		} else {
			bytes = append(bytes, bufferDst[i])
			differencesFound = true
		}
	}

	return diffs, nil
}

func applyDifferences(target string, diffs []Differences) error {
	file, err := os.OpenFile(target, os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	for _, diff := range diffs {
		if _, err := file.Seek(int64(diff.Offset), io.SeekStart); err != nil {
			return err
		}
		if _, err := file.Write(diff.Bytes); err != nil {
			return err
		}
	}

	return nil
}

func main() {
	differences, err := findDifferences(ORIGINAL, MODDED)
	if err != nil {
		log.Fatalf("Error finding differences: %v", err)
	}

	for _, diff := range differences {
		log.Printf("Difference at offset %d, length %d, bytes: %v", diff.Offset, diff.Length, diff.Bytes)
	}

	if err := applyDifferences(TARGET, differences); err != nil {
		log.Fatalf("Error applying differences: %v", err)
	}

	log.Println("Differences applied successfully!")
}
