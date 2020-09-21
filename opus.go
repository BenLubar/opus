package opus

// #cgo CFLAGS: -isystem opusfile-0.12/include -I opusfile-0.12/include -isystem opus-1.3.1/include -I opus-1.3.1/include -I opus-1.3.1/celt -I opus-1.3.1/silk
// #cgo CFLAGS: -DOPUS_BUILD -DUSE_ALLOCA -DHAVE_LRINT -DHAVE_LRINTF
// #cgo LDFLAGS: -lm
// #include <opusfile.h>
import "C"
import (
	"runtime"
	"unsafe"

	"golang.org/x/xerrors"
)

func DecodeOpus(b []byte) ([]byte, int, error) {
	defer runtime.KeepAlive(b)

	var err C.int

	f := C.op_open_memory((*C.uchar)(&b[0]), C.size_t(len(b)), &err)
	if f == nil {
		return nil, 0, xerrors.Errorf("opus: native error number %d", int(err))
	}
	defer C.op_free(f)

	var (
		pcm    []byte
		pcmBuf [5760 * 2]C.opus_int16
		link   C.int
	)

	ret := C.op_read(f, &pcmBuf[0], C.int(len(pcmBuf)), &link)
	if ret < 0 {
		return nil, 0, xerrors.Errorf("opus: native error %d", int(ret))
	}

	numChannels := int(C.op_head(f, link).channel_count)

	for ret != 0 {
		unsafePCMBuf := (*[len(pcmBuf) * 2]byte)(unsafe.Pointer(&pcmBuf[0]))
		pcm = append(pcm, unsafePCMBuf[:int(ret)*numChannels*2]...)
		runtime.KeepAlive(pcmBuf)

		ret = C.op_read(f, &pcmBuf[0], C.int(len(pcmBuf)), &link)
	}

	if ret < 0 {
		return nil, 0, xerrors.Errorf("opus: native error %d", int(ret))
	}

	return pcm, numChannels, nil
}
