package dat

import (
	"fmt"
	"io"
)

//

type consts struct {
	HeaderLen uint8
}

func Const() consts {
	return consts{
		HeaderLen: 3,
	}
}

//

type FalloutDatOptions struct {
}

func DefaultOptions() *FalloutDatOptions {
	return &FalloutDatOptions{}
}

//

// FalloutFile represents a single file stored in .dat
type FalloutFile interface {
	GetParentDir() FalloutDir

	GetName() string // returns base name (FILENAME.EXT)
	GetPath() string // returns full path (DIR/NAME/FILENAME.EXT)

	GetPacked() bool
	GetSizeReal() int32
	GetSizePacked() int32
	GetOffset() int32

	GetBytesReal(io.ReadSeeker) ([]byte, error)
	GetBytesPacked(io.ReadSeeker) ([]byte, error)
}

type FalloutDir interface {
	GetParentDat() FalloutDat

	GetName() string // returns base name (NAME)
	GetPath() string // returns full path (DIR/NAME)

	GetFiles() ([]FalloutFile, error)
}

type FalloutDat interface {
	readStream(stream io.Reader, options *FalloutDatOptions) error

	Reset()

	GetGame() byte // returns 1 or 2
	GetDirs() ([]FalloutDir, error)
	// GetDirsPaths() ([]FalloutDir, error)
}

//

func Fallout1(stream io.Reader, options *FalloutDatOptions) (fo1dat FalloutDat, err error) {
	fo1dat = new(falloutDat_1)

	if err = fo1dat.readStream(stream, options); err != nil {
		return nil, fmt.Errorf("dat.Fallout1() %w", err)
	}

	return fo1dat, nil
}

func Fallout2(filename string, options *FalloutDatOptions) (fo2dat FalloutDat, err error) {
	return nil, fmt.Errorf("dat.Fallout2(%s) not implemented", filename)

	/*
		fo2dat = new(falloutDat_2)

		   	if err = readOsFileInit( filename,fo2dat, closeAfterReading); err != nil {
		   		return nil, fmt.Errorf("dat.Fallout2(%s) %w", filename, err)
		   	}

		   return fo2dat, nil
	*/
}

//

// Should be used by DATx implementation before extraction
func seekFile(stream io.ReadSeeker, file FalloutFile) (err error) {
	var size = file.GetSizeReal()
	if file.GetPacked() {
		size = file.GetSizePacked()
	}

	// set stream to file end position
	// make sure stream won't EOF in a middle of reading
	if _, err = stream.Seek(int64((file.GetOffset() + size)), io.SeekStart); err != nil {
		return err
	}

	// set stream to file start position
	if _, err = stream.Seek(int64(file.GetOffset()), io.SeekStart); err != nil {
		return err
	}

	return nil
}

func Witness(...any) {}
