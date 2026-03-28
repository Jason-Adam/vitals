//go:build unix

package gather

import (
	"syscall"
	"unsafe"
)

// getTermWidth returns the terminal width for the given file descriptor
// via the TIOCGWINSZ ioctl, or 0 if the fd is not a terminal.
func getTermWidth(fd uintptr) int {
	var ws struct {
		Row, Col       uint16
		Xpixel, Ypixel uint16
	}
	_, _, err := syscall.Syscall(
		syscall.SYS_IOCTL, fd,
		syscall.TIOCGWINSZ,
		uintptr(unsafe.Pointer(&ws)),
	)
	if err != 0 || ws.Col == 0 {
		return 0
	}
	return int(ws.Col)
}
