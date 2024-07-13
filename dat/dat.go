package dat

import (
	"fmt"
	"io"

	"github.com/wipe2238/fo/x/dbg"
)

//

// FalloutDat represents single .dat file
//
// NOTE: Unstable interface
type FalloutDat interface {
	GetGame() uint8 // returns 1 or 2

	GetDirs() []FalloutDir

	GetDbg() dbg.Map
	FillDbg()

	// implementation details

	readDat(stream io.Reader) error
}

// FalloutDat represents single directory entry
//
// NOTE: Unstable interface
type FalloutDir interface {
	GetParentDat() FalloutDat

	GetName() string // returns base name (FILENAME.EXT)
	GetPath() string // returns full path (DIR//NAME/FILENAME.EXT)

	GetFiles() []FalloutFile

	GetDbg() dbg.Map
}

// FalloutFile represents a single file entry
//
// NOTE: Unstable interface
type FalloutFile interface {
	GetParentDir() FalloutDir

	GetName() string // returns base name (FILENAME.EXT)
	GetPath() string // returns full path (DIR/NAME/FILENAME.EXT)

	GetOffset() int64
	GetCompressMode() uint32
	GetSizeReal() int64
	GetSizePacked() int64

	GetBytesReal(io.ReadSeeker) ([]byte, error)
	GetBytesPacked(io.ReadSeeker) ([]byte, error)

	GetDbg() dbg.Map
}

//

func Fallout1(stream io.Reader) (fo1dat FalloutDat, err error) {
	fo1dat = new(falloutDatV1)

	if err = fo1dat.readDat(stream); err != nil {
		return nil, fmt.Errorf("dat.Fallout1() %w", err)
	}

	return fo1dat, nil
}

func Fallout2(_ io.Reader) (fo2dat FalloutDat, err error) {
	return nil, fmt.Errorf("dat.Fallout2() not implemented")

	/*
		fo2dat = new(falloutDat_2)

		   	if err = readOsFileInit( filename,fo2dat, closeAfterReading); err != nil {
		   		return nil, fmt.Errorf("dat.Fallout2(%s) %w", filename, err)
		   	}

		   return fo2dat, nil
	*/
}

//

// Should be used by DATx implementation before decompression
func seekFile(stream io.ReadSeeker, file FalloutFile) (err error) {
	// set stream to file end position
	// make sure stream won't EOF in a middle of reading
	if _, err = stream.Seek((file.GetOffset() + file.GetSizePacked()), io.SeekStart); err != nil {
		return err
	}

	// set stream to file start position
	if _, err = stream.Seek(file.GetOffset(), io.SeekStart); err != nil {
		return err
	}

	return nil
}
