package dat

import (
	"fmt"
	"io"
	"os"

	"github.com/wipe2238/fo/x/dbg"
)

//

const errPackage = "fo/dat:"

// FalloutDat represents single .dat file
//
// NOTE: Unstable interface
type FalloutDat interface {
	GetGame() uint8 // returns 1 or 2

	GetDirs() []FalloutDir

	GetDbg() dbg.Map

	// Stream position must be set to start of DAT before calling this function
	SetDbg(io.Seeker) error

	// Deprecated: use `SetDbg()` insead
	FillDbg()

	// implementation details

	readDat(stream io.ReadSeeker) error
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
	SetDbg(io.Seeker) error
}

// FalloutFile represents a single file entry
//
// NOTE: Unstable interface
type FalloutFile interface {
	GetParentDat() FalloutDat
	GetParentDir() FalloutDir

	GetName() string // returns base name (FILENAME.EXT)
	GetPath() string // returns full path (DIR/NAME/FILENAME.EXT)

	GetOffset() int64
	GetSizeReal() int64
	GetSizePacked() int64
	GetPacked() bool
	GetPackedMode() uint32

	GetBytesReal(io.ReadSeeker) ([]byte, error)
	GetBytesPacked(io.ReadSeeker) ([]byte, error)

	GetDbg() dbg.Map
	SetDbg(io.Seeker) error
}

//

// Open attempts to read .dat file with unspecified version as DAT2 or DAT1 (in that order)
//
// On success, returns opened stream (required for extracting files data) and `FalloutDat` object.
func Open(filename string) (osFile *os.File, dat FalloutDat, err error) {
	if osFile, err = os.Open(filename); err != nil {
		return nil, nil, err
	}

	// TODO: add storing/restoring stream position

	for _, reader := range [2]func(io.ReadSeeker) (FalloutDat, error){Fallout1, Fallout2} {
		if dat, err = reader(osFile); err != nil {
			osFile.Seek(0, io.SeekStart)
			continue
		}

		return osFile, dat, nil
	}

	osFile.Close()
	return nil, nil, fmt.Errorf("%s Open(%s) cannot guess DAT file version", errPackage, filename)
}

// Fallout1 reads already opened stream as DAT1
func Fallout1(stream io.ReadSeeker) (dat1 FalloutDat, err error) {
	dat1 = new(falloutDatV1)

	if err = dat1.readDat(stream); err != nil {
		return nil, fmt.Errorf("%s Fallout1() %w", errPackage, err)
	}

	return dat1, nil
}

// Fallout2 reads already opened stream as DAT2
func Fallout2(stream io.ReadSeeker) (dat2 FalloutDat, err error) {
	dat2 = new(falloutDatV2)

	if err = dat2.readDat(stream); err != nil {
		return nil, fmt.Errorf("%s Fallout2() %w", errPackage, err)
	}

	return dat2, nil
}
