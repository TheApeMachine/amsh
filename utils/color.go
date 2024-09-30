package utils

import "syscall"

/*
colors is a list of ANSI escape codes for colored output.
*/
var colors = []string{
	"\033[31m", // Red
	"\033[32m", // Green
	"\033[33m", // Yellow
	"\033[34m", // Blue
	"\033[35m", // Magenta
	"\033[36m", // Cyan
	"\033[37m", // White
	"\033[90m", // Bright Black
	"\033[91m", // Bright Red
	"\033[40m", // Bright Green
	"\033[41m", // Bright Yellow
	"\033[42m", // Bright Blue
	"\033[43m", // Bright Magenta
	"\033[44m", // Bright Cyan
}

/* reset is an ANSI escape code to reset the color of the text. */
var reset = "\033[0m"

/* PrintRed writes a string in red color to the console, using syscall */
func PrintRed(text string) { syscall.Write(1, []byte(colors[0]+text+reset)) }

/* PrintGreen writes a string in green color to the console, using syscall */
func PrintGreen(text string) { syscall.Write(1, []byte(colors[1]+text+reset)) }

/* PrintYellow writes a string in yellow color to the console, using syscall */
func PrintYellow(text string) { syscall.Write(1, []byte(colors[2]+text+reset)) }

/* PrintBlue writes a string in blue color to the console, using syscall */
func PrintBlue(text string) { syscall.Write(1, []byte(colors[3]+text+reset)) }

/* PrintMagenta writes a string in magenta color to the console, using syscall */
func PrintMagenta(text string) { syscall.Write(1, []byte(colors[4]+text+reset)) }

/* PrintCyan writes a string in cyan color to the console, using syscall */
func PrintCyan(text string) { syscall.Write(1, []byte(colors[5]+text+reset)) }

/* PrintBrightBlack writes a string in bright black color to the console, using syscall */
func PrintBrightBlack(text string) { syscall.Write(1, []byte(colors[6]+text+reset)) }

/* PrintBrightRed writes a string in bright red color to the console, using syscall */
func PrintBrightRed(text string) { syscall.Write(1, []byte(colors[7]+text+reset)) }

/* PrintBrightGreen writes a string in bright green color to the console, using syscall */
func PrintBrightGreen(text string) { syscall.Write(1, []byte(colors[8]+text+reset)) }

/* PrintBrightYellow writes a string in bright yellow color to the console, using syscall */
func PrintBrightYellow(text string) { syscall.Write(1, []byte(colors[9]+text+reset)) }

/* PrintBrightBlue writes a string in bright blue color to the console, using syscall */
func PrintBrightBlue(text string) { syscall.Write(1, []byte(colors[10]+text+reset)) }

/* PrintBrightMagenta writes a string in bright magenta color to the console, using syscall */
func PrintBrightMagenta(text string) { syscall.Write(1, []byte(colors[11]+text+reset)) }

/* PrintBrightCyan writes a string in bright cyan color to the console, using syscall */
func PrintBrightCyan(text string) { syscall.Write(1, []byte(colors[12]+text+reset)) }

/* PrintColor writes a string in color to the console, using syscall */
func PrintColor(text string, color int) { syscall.Write(1, []byte(colors[color]+text+reset)) }
