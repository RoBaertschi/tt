//go:build unix
// +build unix

package term

import (
	"errors"

	"golang.org/x/sys/unix"
)

var restore unix.Termios
var rawModeEnabled = false

func EnterRawMode() error {
	termios, err := unix.IoctlGetTermios(unix.Stdin, unix.TCGETS)

	if err != nil {
		return err
	}

	restore = *termios
	termios.Lflag = termios.Lflag &^ (unix.ECHO | unix.ICANON | unix.ISIG | unix.IEXTEN)
	termios.Iflag = termios.Iflag &^ (unix.IXON | unix.ICRNL | unix.BRKINT | unix.INPCK | unix.ISTRIP)
	termios.Cflag = termios.Cflag | unix.CS8

	if err := unix.IoctlSetTermios(unix.Stdin, unix.TCSETSF, termios); err != nil {
		return err
	}

	rawModeEnabled = true

	return nil
}

func LeaveRawMode() error {
	if !rawModeEnabled {
		return errors.New("raw mode is not enabled")
	}

	err := unix.IoctlSetTermios(unix.Stdin, unix.TCSETSF, &restore)
	if err != nil {
		return err
	}
	rawModeEnabled = false
	return nil
}
