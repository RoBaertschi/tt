package term

import (
	"errors"
	"fmt"
	"os"
)

const ESC = "\x1b"
const CSI = ESC + "["
const Reset = CSI + "0m"

// Colors

const (
	BlackFg   = "30"
	BlackBg   = "40"
	RedFg     = "31"
	RedBg     = "41"
	GreenFg   = "32"
	GreenBg   = "42"
	YellowFg  = "33"
	YellowBg  = "43"
	BlueFg    = "34"
	BlueBg    = "44"
	MagentaFg = "35"
	MagentaBg = "35"
	CyanFg    = "36"
	CyanBg    = "46"
	WhiteFg   = "37"
	WhiteBg   = "47"
)

func Color(col string) string {
	return CSI + col + "m"
}

// Other CSI

func CursorUp(amount int) string {
	return fmt.Sprintf("%s%dA", CSI, amount)
}

func CursorDown(amount int) string {
	return fmt.Sprintf("%s%dB", CSI, amount)
}

func CursorForward(amount int) string {
	return fmt.Sprintf("%s%dC", CSI, amount)
}

func CursorBack(amount int) string {
	return fmt.Sprintf("%s%dD", CSI, amount)
}

// Not ANSI.SYS
func CursorNextLine(amount int) string {
	return fmt.Sprintf("%s%dE", CSI, amount)
}

// Not ANSI.SYS
func CursorPreviousLine(amount int) string {
	return fmt.Sprintf("%s%dF", CSI, amount)
}

// Not ANSI.SYS
// Moves cursor to column amount
func CursorHorizontalAbsolute(amount int) string {
	return fmt.Sprintf("%s%dG", CSI, amount)
}

// Moves the cursor on the 1-based grid in the terminal
func CursorPostition(x, y int) string {
	return fmt.Sprintf("%s%d;%dH", CSI, x, y)
}

type EraseInDisplayMode int

const (
	ClearFromCursor EraseInDisplayMode = iota // clear from cursor to the end of the screen
	ClearToCursor                             // clear to cursor from the begining of the screen
	ClearEntireScreen
	ClearEntireScreenAndScrollbackBuffer // xterm extension
)

func EraseInDisplay(mode EraseInDisplayMode) string {
	return fmt.Sprintf("%s%dJ", CSI, mode)
}

type EraseInLineMode int

const (
	FromCursorToEnd EraseInLineMode = iota
	FromCursorToBegin
	EntireLine
)

func EraseInLine(mode EraseInLineMode) string {
	return fmt.Sprintf("%s%dK", CSI, mode)
}

// Not ANSI.SYS
func ScrollUp(amount int) string {
	return fmt.Sprintf("%s%dS", CSI, amount)
}

// Not ANSI.SYS
func ScrollDown(amount int) string {
	return fmt.Sprintf("%s%dT", CSI, amount)
}

// Moves the cursor on the 1-based grid in the terminal
// Same as Cursor Position but a format effector, not a editor function
// see https://en.wikipedia.org/wiki/ANSI_escape_code#Control_Sequence_Introducer_commands
func HorizontalVerticalPosition(x, y int) string {
	return fmt.Sprintf("%s%d;%dH", CSI, x, y)
}

// Gets the cursor position by transmitting CSIn;mR  n = row, m = column
// see https://en.wikipedia.org/wiki/ANSI_escape_code#Control_Sequence_Introducer_commands
// Use GetCursorPosition to get x and y
func DeviceStatusReport() string {
	return CSI + "6n"
}

var DidNotGetCsi = errors.New("could not get csi from Device Status Report sequence")

func GetCursorPosition() (row, column int, err error) {
	// _, err = os.Stdin.Seek(0, 2)
	// if err != nil {
	// 	return
	// }
	os.Stdout.Write([]byte(DeviceStatusReport()))
	// CSI is ESC and [
	csiBuffer := [2]byte{}
	os.Stdin.Read(csiBuffer[:])

	if string(csiBuffer[:]) != CSI {
		err = DidNotGetCsi
		return
	}

	miniBuff := [1]byte{}
	for {
		_, err = os.Stdin.Read(miniBuff[:])

		if err != nil {
			return
		}

		if miniBuff[0] == ';' {
			break
		}

		if '0' <= miniBuff[0] && miniBuff[0] <= '9' {
			row *= 10
			row += int(miniBuff[0] - '0')
		} else {
			err = fmt.Errorf("invalid byte for number %b", miniBuff[0])
			return
		}
	}

	for {
		_, err = os.Stdin.Read(miniBuff[:])

		if err != nil {
			return
		}

		if miniBuff[0] == 'R' {
			break
		}

		if '0' <= miniBuff[0] && miniBuff[0] <= '9' {
			column *= 10
			column += int(miniBuff[0] - '0')
		} else {
			err = fmt.Errorf("invalid byte for number %b", miniBuff[0])
			return
		}
	}

	return
}

func Exit(val int) {
	// LeaveRawMode()
	os.Exit(val)
}
