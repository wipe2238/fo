package dat

import (
	"fmt"
	"os"
	"strings"
)

type falloutFileLoader struct {
	FalloutFile

	DataReal   []byte
	DataPacked []byte
}

func newLoader(version uint8) (file *falloutFileLoader) {
	switch version {
	case 1:
		file = &falloutFileLoader{FalloutFile: new(falloutFileV1)}
	case 2:
		file = &falloutFileLoader{FalloutFile: new(falloutFileV2)}
	default:
		panic(fmt.Sprintf("%s invalid version '%d", errPackage, version))
	}

	return file
}

// If loading `filenamePacked` results in error,
func (loader *falloutFileLoader) loadFiles(filenameReal string, filenamePacked string) (err error) {
	if err = loader.loadFileType(filenameReal, "real"); err != nil {
		return err
	}

	if err = loader.loadFileType(filenamePacked, "packed"); err != nil {
		return err
	}

	return nil
}

// Once file is loaded, inner `FalloutFile`
func (loader *falloutFileLoader) loadFileType(filename string, dataType string) (err error) {
	switch strings.ToLower(dataType) {
	case "real":
		if loader.DataReal, err = os.ReadFile(filename); err != nil {
			return err
		}

		loader.setSize(int64(len(loader.DataReal)), -1)

	case "packed":
		if loader.DataPacked, err = os.ReadFile(filename); err != nil {
			return err
		}

		loader.setSize(-1, int64(len(loader.DataPacked)))

	default:
		return fmt.Errorf("%s unknown dataType '%s'", errPackage, dataType)
	}

	return nil
}

// If `size*` < 0, related fields won't be changed
func (loader *falloutFileLoader) setSize(sizeReal int64, sizePacked int64) {
	if sizeReal < 0 && sizePacked < 0 {
		return
	}

	if file, ok := loader.FalloutFile.(*falloutFileV1); file != nil && ok {
		if sizeReal >= 0 {
			file.SizeReal = uint32(sizeReal)
		}
		if sizePacked >= 0 {
			file.SizePacked = uint32(sizePacked)
		}
	} else if file, ok := loader.FalloutFile.(*falloutFileV2); file != nil && ok {
		if sizeReal >= 0 {
			file.SizeReal = uint32(sizeReal)
		}
		if sizePacked >= 0 {
			file.SizePacked = uint32(sizePacked)
		}
	} else {
		panic(fmt.Sprintf("%s loader: cannot set size[%d, %d]: unknown FalloutFile implementation", errPackage, sizeReal, sizePacked))
	}
}
