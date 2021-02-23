package main

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/cheggaaa/pb/v3"
)

var (
	ErrOffsetExceedsFileSize = errors.New("offset exceeds file size")
	ErrLimitIsNegative       = errors.New("limit number must be not negative")
)

func Copy(fromPath, toPath string, offset, limit int64) error {
	if limit < 0 {
		return ErrLimitIsNegative
	}

	// open fromPath file
	src, err := os.Open(fromPath)
	if err != nil {
		return fmt.Errorf("error during opening %s file: %w", fromPath, err)
	}
	defer src.Close()

	srcFileInfo, err := src.Stat()
	if err != nil {
		return fmt.Errorf("error during getting os.Stat info for %s file: %w", fromPath, err)
	}

	// check offset bigger than size here instead of creation of toPath file if we know that error will occur anyway
	if srcFileInfo.Size() < offset {
		return ErrOffsetExceedsFileSize
	}

	// create toPath file
	dst, err := os.Create(toPath)
	if err != nil {
		return fmt.Errorf("error during creation of %s file: %w", toPath, err)
	}
	defer dst.Close()

	if _, err := src.Seek(offset, io.SeekStart); err != nil {
		return fmt.Errorf("error during making an offset in %s file: %w", fromPath, err)
	}

	// here we count num of bytes to copy, this calculations mostly done for correct progress bar size values
	srcFileSize := srcFileInfo.Size()
	var bytesToCopy int64
	if limit == 0 || offset+limit > srcFileSize {
		bytesToCopy = srcFileSize - offset
	} else {
		bytesToCopy = limit
	}

	// creation and setup of a progress bar
	progressBar := pb.Full.Start64(bytesToCopy)
	progressBar.Set(pb.Bytes, true)
	progressBar.Set(pb.SIBytesPrefix, true)
	proxySrcReader := progressBar.NewProxyReader(src)
	defer progressBar.Finish()

	if _, err = io.CopyN(dst, proxySrcReader, bytesToCopy); err != nil {
		if !errors.Is(err, io.EOF) {
			return fmt.Errorf("error during bytes copying: %w", err)
		}
	}
	return nil
}
