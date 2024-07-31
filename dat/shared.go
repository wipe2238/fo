package dat

import "io"

type falloutShared struct{}

func (falloutShared) seek(stream io.ReadSeeker, file FalloutFile) (err error) {
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

func (shared falloutShared) getBytesReal(stream io.ReadSeeker, file FalloutFile) (bytesReal []byte, err error) {
	if bytesReal, err = shared.getBytesPacked(stream, file); err != nil {
		return nil, err
	}

	if !file.GetPacked() {
		return bytesReal, nil
	}

	return file.GetBytesUnpacked(bytesReal)
}

func (shared falloutShared) getBytesPacked(stream io.ReadSeeker, file FalloutFile) (bytes []byte, err error) {
	if err = shared.seek(stream, file); err != nil {
		return nil, err
	}

	bytes = make([]byte, file.GetSizePacked())
	if _, err = stream.Read(bytes); err != nil {
		return nil, err
	}

	return bytes, nil
}
