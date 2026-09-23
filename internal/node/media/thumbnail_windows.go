//go:build windows

package media

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"image"
	"image/jpeg"
	"runtime"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

const (
	coinitApartmentThreaded = 0x2
	rpcEChangedMode         = 0x80010106
	siigbfThumbnailOnly     = 0x8
	biRGB                   = 0
	dibRGBColors            = 0
)

var (
	ole32                    = windows.NewLazySystemDLL("ole32.dll")
	shell32                  = windows.NewLazySystemDLL("shell32.dll")
	gdi32                    = windows.NewLazySystemDLL("gdi32.dll")
	user32                   = windows.NewLazySystemDLL("user32.dll")
	procCoInitializeEx       = ole32.NewProc("CoInitializeEx")
	procCoUninitialize       = ole32.NewProc("CoUninitialize")
	procSHCreateItemFromPath = shell32.NewProc("SHCreateItemFromParsingName")
	procGetObjectW           = gdi32.NewProc("GetObjectW")
	procGetDIBits            = gdi32.NewProc("GetDIBits")
	procDeleteObject         = gdi32.NewProc("DeleteObject")
	procGetDC                = user32.NewProc("GetDC")
	procReleaseDC            = user32.NewProc("ReleaseDC")
	iidShellItemImageFactory = windows.GUID{Data1: 0xbcc18b79, Data2: 0xba16, Data3: 0x442f, Data4: [8]byte{0x80, 0xc4, 0x8a, 0x59, 0xc3, 0x0c, 0x46, 0x3b}}
)

type shellItemImageFactory struct {
	vtbl *shellItemImageFactoryVtbl
}

type shellItemImageFactoryVtbl struct {
	queryInterface uintptr
	addRef         uintptr
	release        uintptr
	getImage       uintptr
}

type bitmap struct {
	type_        int32
	width        int32
	height       int32
	widthBytes   int32
	planes       uint16
	bitsPerPixel uint16
	_            uint32
	bits         unsafe.Pointer
}

type bitmapInfoHeader struct {
	size          uint32
	width         int32
	height        int32
	planes        uint16
	bitCount      uint16
	compression   uint32
	sizeImage     uint32
	xPelsPerMeter int32
	yPelsPerMeter int32
	clrUsed       uint32
	clrImportant  uint32
}

type bitmapInfo struct {
	header bitmapInfoHeader
}

// GenerateThumbnail retrieves the thumbnail provided by the Windows Shell.
func GenerateThumbnail(ctx context.Context, videoPath string) (string, error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	hr, _, _ := procCoInitializeEx.Call(0, coinitApartmentThreaded)
	if hresultFailed(hr) && uint32(hr) != rpcEChangedMode {
		return "", fmt.Errorf("initialize COM: HRESULT %#x", uint32(hr))
	}
	if uint32(hr) != rpcEChangedMode {
		defer procCoUninitialize.Call()
	}

	path, err := windows.UTF16PtrFromString(videoPath)
	if err != nil {
		return "", err
	}

	var factory *shellItemImageFactory
	hr, _, _ = procSHCreateItemFromPath.Call(
		uintptr(unsafe.Pointer(path)),
		0,
		uintptr(unsafe.Pointer(&iidShellItemImageFactory)),
		uintptr(unsafe.Pointer(&factory)),
	)
	if hresultFailed(hr) {
		return "", fmt.Errorf("create Windows Shell item: HRESULT %#x", uint32(hr))
	}
	defer syscallN(factory.vtbl.release, uintptr(unsafe.Pointer(factory)))

	var handle windows.Handle
	requestedSize := uintptr(uint64(320) | uint64(320)<<32)
	hr = syscallN(
		factory.vtbl.getImage,
		uintptr(unsafe.Pointer(factory)),
		requestedSize,
		siigbfThumbnailOnly,
		uintptr(unsafe.Pointer(&handle)),
	)
	if hresultFailed(hr) {
		return "", fmt.Errorf("get Windows Shell thumbnail: HRESULT %#x", uint32(hr))
	}
	defer procDeleteObject.Call(uintptr(handle))

	thumbnail, err := bitmapToJPEG(handle)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(thumbnail), nil
}

func bitmapToJPEG(handle windows.Handle) ([]byte, error) {
	var bm bitmap
	result, _, callErr := procGetObjectW.Call(
		uintptr(handle),
		unsafe.Sizeof(bm),
		uintptr(unsafe.Pointer(&bm)),
	)
	if result == 0 {
		return nil, fmt.Errorf("read Windows thumbnail bitmap: %w", callErr)
	}
	if bm.width <= 0 || bm.height == 0 {
		return nil, fmt.Errorf("invalid Windows thumbnail dimensions: %dx%d", bm.width, bm.height)
	}

	height := bm.height
	if height < 0 {
		height = -height
	}
	pixels := make([]byte, int(bm.width)*int(height)*4)
	info := bitmapInfo{header: bitmapInfoHeader{
		size:        uint32(unsafe.Sizeof(bitmapInfoHeader{})),
		width:       bm.width,
		height:      -height,
		planes:      1,
		bitCount:    32,
		compression: biRGB,
	}}

	dc, _, callErr := procGetDC.Call(0)
	if dc == 0 {
		return nil, fmt.Errorf("acquire Windows device context: %w", callErr)
	}
	defer procReleaseDC.Call(0, dc)

	result, _, callErr = procGetDIBits.Call(
		dc,
		uintptr(handle),
		0,
		uintptr(height),
		uintptr(unsafe.Pointer(&pixels[0])),
		uintptr(unsafe.Pointer(&info)),
		dibRGBColors,
	)
	if result == 0 {
		return nil, fmt.Errorf("read Windows thumbnail pixels: %w", callErr)
	}

	img := image.NewRGBA(image.Rect(0, 0, int(bm.width), int(height)))
	for i := 0; i < len(pixels); i += 4 {
		img.Pix[i] = pixels[i+2]
		img.Pix[i+1] = pixels[i+1]
		img.Pix[i+2] = pixels[i]
		img.Pix[i+3] = 0xff
	}

	var output bytes.Buffer
	if err := jpeg.Encode(&output, img, &jpeg.Options{Quality: 80}); err != nil {
		return nil, err
	}
	return output.Bytes(), nil
}

func hresultFailed(result uintptr) bool {
	return int32(uint32(result)) < 0
}

func syscallN(trap uintptr, args ...uintptr) uintptr {
	result, _, _ := syscall.SyscallN(trap, args...)
	return result
}
