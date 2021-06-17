// Package opus is a thin wrapper around the opusfile library.
package opus

// #cgo CFLAGS: -isystem opusfile-0.12/include -I opusfile-0.12/include -isystem opus-1.3.1/include -I opus-1.3.1/include -I opus-1.3.1/celt -I opus-1.3.1/silk
// #cgo CFLAGS: -DOPUS_BUILD -DUSE_ALLOCA -DHAVE_LRINT -DHAVE_LRINTF
// #cgo android arm CFLAGS: -Dfseeko=fseek -Dftello=ftell
// #cgo LDFLAGS: -lm
// #include <opusfile.h>
import "C"
import (
	"runtime"
	"strconv"
	"unsafe"
)

// DecodeOpus decodes an ogg/opus file in memory to 48kHz 16-bit PCM.
func DecodeOpus(b []byte) (pcm []byte, numChannels int, err error) {
	defer runtime.KeepAlive(b)

	var err C.int

	f := C.op_open_memory((*C.uchar)(&b[0]), C.size_t(len(b)), &err)
	if f == nil {
		return nil, 0, opusError(err)
	}
	defer C.op_free(f)

	var (
		pcmBuf [5760 * 2]C.opus_int16
		link   C.int
	)

	ret := C.op_read(f, &pcmBuf[0], C.int(len(pcmBuf)), &link)
	if ret < 0 {
		return nil, 0, opusError(err)
	}

	numChannels = int(C.op_head(f, link).channel_count)

	for ret != 0 {
		unsafePCMBuf := (*[unsafe.Sizeof(pcmBuf)]byte)(unsafe.Pointer(&pcmBuf[0]))
		pcm = append(pcm, unsafePCMBuf[:int(ret)*numChannels*2]...)
		runtime.KeepAlive(pcmBuf)

		ret = C.op_read(f, &pcmBuf[0], C.int(len(pcmBuf)), &link)
	}

	if ret < 0 {
		return nil, 0, opusError(err)
	}

	return pcm, numChannels, nil
}

// DecodeOpusFloat decodes an ogg/opus file in memory to 48kHz 32-bit float PCM.
func DecodeOpusFloat(b []byte) (pcm []byte, numChannels int, err error) {
	defer runtime.KeepAlive(b)

	var err C.int

	f := C.op_open_memory((*C.uchar)(&b[0]), C.size_t(len(b)), &err)
	if f == nil {
		return nil, 0, opusError(err)
	}
	defer C.op_free(f)

	var (
		pcmBuf [5760 * 2]C.float
		link   C.int
	)

	ret := C.op_read_float(f, &pcmBuf[0], C.int(len(pcmBuf)), &link)
	if ret < 0 {
		return nil, 0, opusError(err)
	}

	numChannels = int(C.op_head(f, link).channel_count)

	for ret != 0 {
		unsafePCMBuf := (*[unsafe.Sizeof(pcmBuf)]byte)(unsafe.Pointer(&pcmBuf[0]))
		pcm = append(pcm, unsafePCMBuf[:int(ret)*numChannels*2]...)
		runtime.KeepAlive(pcmBuf)

		ret = C.op_read_float(f, &pcmBuf[0], C.int(len(pcmBuf)), &link)
	}

	if ret < 0 {
		return nil, 0, opusError(err)
	}

	return pcm, numChannels, nil
}

type opusError C.int

func (err opusError) Error() string {
	switch C.int(err) {
	case C.OP_EREAD:
		return "opus: io error"
	case C.OP_EFAULT:
		return "opus: internal error"
	case C.OP_EIMPL:
		return "opus: stream uses unsupported features"
	case C.OP_EINVAL:
		return "opus: invalid parameter"
	case C.OP_ENOTFORMAT:
		return "opus: not a valid OGG file"
	case C.OP_EBADHEADER:
		return "opus: malformed header"
	case C.OP_EVERSION:
		return "opus: unrecognized version number"
	case C.OP_EBADPACKET:
		return "opus: bad packet"
	case C.OP_EBADLINK:
		return "opus: malformed bitstream"
	case C.OP_ENOSEEK:
		return "opus: unable to seek"
	case C.OP_EBADTIMESTAMP:
		return "opus: invalid timestamp"
	default:
		return "opus: unknown native error code " + strconv.Itoa(int(err))
	}
}
