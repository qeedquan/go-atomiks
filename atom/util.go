package atom

import (
	"io"
	"os"

	"github.com/qeedquan/go-media/sdl"
)

func ck(err error) {
	if err != nil {
		sdl.LogCritical(sdl.LOG_CATEGORY_APPLICATION, "%v", err)
		sdl.ShowSimpleMessageBox(sdl.MESSAGEBOX_ERROR, "Error", err.Error(), nil)
		os.Exit(1)
	}
}

func ek(err error) bool {
	if err != nil {
		sdl.LogError(sdl.LOG_CATEGORY_APPLICATION, "%v", err)
		return true
	}
	return false
}

func readByte(r io.ByteReader) int {
	b, _ := r.ReadByte()
	return int(b)
}

func readShort(r io.ByteReader) int {
	hi := readByte(r)
	lo := readByte(r)
	return hi<<8 | lo
}
