package main

// Every Go file belongs to a package. The special name "main" tells Go this is
// an executable program (not a reusable library). The entry point is main().

import (
	"bufio"      // bufio wraps a Reader/Writer with buffering for efficiency.
	             // Here we use bufio.Scanner to read stdin line by line.
	"fmt"        // fmt handles formatted I/O: printing to the terminal, building strings, etc.
	"io"         // io provides basic interfaces for reading and writing data streams.
	             // We use io.Copy to copy bytes from one file to another.
	"os"         // os gives access to operating system features: files, args, stderr, exit codes.
	"path/filepath" // filepath manipulates file paths in a way that works across operating systems
	                // (uses / on Unix, \ on Windows). We use it to split a path into dir and filename.
	"strconv"    // strconv converts between strings and other types.
	             // We use strconv.Atoi to turn a string like "3" into the integer 3.
	"strings"    // strings provides utilities for working with strings:
	             // trimming, splitting, joining, building, etc.
)

// ── Helper ───────────────────────────────────────────────────────────────────

// cornerOrLine decides which character to place at a given column position
// within a single row. Every quad row follows the same rule:
//   - column 0            → left border character
//   - column width-1      → right border character
//   - anything in between → middle (fill) character
//
// Each quad function passes its own set of characters, so the same logic
// produces different shapes depending on what left/mid/right are.
func cornerOrLine(col, width int, left, mid, right string) string {
	if col == 0 {
		return left
	} else if col == width-1 {
		return right
	}
	return mid
}

// ── Quad generator ───────────────────────────────────────────────────────────

// generateQuad builds the complete text output of the named quad for the given
// width and height, and returns it as a single string.
//
// This is the single source of truth for what each quad looks like:
//   - the drawer (runQuad) prints this string directly
//   - the checker (runChecker) generates it and compares it against stdin
//
// strings.Builder is an efficient way to build a string piece by piece without
// creating lots of intermediate strings in memory.
func generateQuad(name string, width, height int) string {
	var sb strings.Builder

	// Iterate row by row (outer loop) then character by character (inner loop).
	for row := 0; row < height; row++ {
		for col := 0; col < width; col++ {
			// The switch picks the correct border characters for this quad.
			// Each case defines what the top row, bottom row, and middle rows look like.
			switch name {

			case "quadA":
				// Top and bottom rows: corners are 'o', fill is '-'
				// Middle rows:         sides are '|', interior is ' '
				if row == 0 || row == height-1 {
					sb.WriteString(cornerOrLine(col, width, "o", "-", "o"))
				} else {
					sb.WriteString(cornerOrLine(col, width, "|", " ", "|"))
				}

			case "quadB":
				// Top row:    '/' on the left, '*' fill, '\' on the right
				// Bottom row: '\' on the left, '*' fill, '/' on the right (flipped)
				// Middle rows: '*' sides, ' ' interior
				if row == 0 {
					sb.WriteString(cornerOrLine(col, width, "/", "*", "\\"))
				} else if row == height-1 {
					sb.WriteString(cornerOrLine(col, width, "\\", "*", "/"))
				} else {
					sb.WriteString(cornerOrLine(col, width, "*", " ", "*"))
				}

			case "quadC":
				// Top row:    'A' corners, 'B' fill
				// Bottom row: 'C' corners, 'B' fill
				// Middle rows: 'B' sides, ' ' interior
				if row == 0 {
					sb.WriteString(cornerOrLine(col, width, "A", "B", "A"))
				} else if row == height-1 {
					sb.WriteString(cornerOrLine(col, width, "C", "B", "C"))
				} else {
					sb.WriteString(cornerOrLine(col, width, "B", " ", "B"))
				}

			case "quadD":
				// Top and bottom rows are identical: 'A' left, 'B' fill, 'C' right
				// Middle rows: 'B' sides, ' ' interior
				if row == 0 || row == height-1 {
					sb.WriteString(cornerOrLine(col, width, "A", "B", "C"))
				} else {
					sb.WriteString(cornerOrLine(col, width, "B", " ", "B"))
				}

			case "quadE":
				// Top row:    'A' left, 'B' fill, 'C' right
				// Bottom row: 'C' left, 'B' fill, 'A' right  (top row mirrored)
				// Middle rows: 'B' sides, ' ' interior
				if row == 0 {
					sb.WriteString(cornerOrLine(col, width, "A", "B", "C"))
				} else if row == height-1 {
					sb.WriteString(cornerOrLine(col, width, "C", "B", "A"))
				} else {
					sb.WriteString(cornerOrLine(col, width, "B", " ", "B"))
				}
			}
		}
		// Every row ends with a newline, just like fmt.Println() does.
		sb.WriteString("\n")
	}

	// Return the entire shape as one string.
	return sb.String()
}

// ── Quad runner ───────────────────────────────────────────────────────────────

// runQuad is called when the binary is invoked as quadA, quadB, etc.
// It reads two command-line arguments (width and height), validates them,
// then prints the corresponding quad shape.
func runQuad(name string) {
	// os.Args is a slice of strings holding the command-line arguments.
	// os.Args[0] is always the program name itself, so the user's arguments
	// start at os.Args[1]. We expect exactly two: width and height.
	if len(os.Args) != 3 {
		fmt.Printf("Usage: %s <width> <height>\n", name)
		return
	}

	// strconv.Atoi converts a string to an int. It returns two values:
	// the integer result and an error (nil if conversion succeeded).
	width, errW := strconv.Atoi(os.Args[1])
	height, errH := strconv.Atoi(os.Args[2])

	// If either conversion failed (e.g. the user typed "abc"), the error is not nil.
	if errW != nil || errH != nil {
		fmt.Println("Width and height must be valid integers.")
		return
	}

	if width <= 0 || height <= 0 {
		fmt.Println("Width and height must be positive.")
		return
	}

	// fmt.Print (without "ln") prints without adding an extra newline.
	// generateQuad already ends every row (including the last) with '\n'.
	fmt.Print(generateQuad(name, width, height))
}

// ── Checker ───────────────────────────────────────────────────────────────────

// runChecker reads from stdin, works out the dimensions of the shape,
// generates the expected output for every quad at those dimensions, and
// reports whichever ones match.
func runChecker() {
	// bufio.NewScanner wraps os.Stdin in a buffered reader.
	// By default it reads one line at a time (splitting on '\n').
	scanner := bufio.NewScanner(os.Stdin)

	// The default buffer is 64 KB. We expand it to 1 MB so that very large
	// quad shapes (many columns/rows) don't cause a "token too long" error.
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

	// Read every line from stdin into a slice of strings.
	// scanner.Scan() returns true as long as there is more input.
	// scanner.Text() returns the current line without the trailing '\n'.
	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	// An empty input can't be a quad.
	if len(lines) == 0 {
		fmt.Println("Not a quad function")
		return
	}

	// For a valid quad, every row must have the same length.
	// We use the first line's length as the expected width.
	width := len(lines[0])
	height := len(lines)

	// A quad always has at least one character per row.
	if width == 0 {
		fmt.Println("Not a quad function")
		return
	}

	// If any line has a different length the shape is irregular → not a quad.
	// The blank identifier '_' discards the index since we only need the value.
	for _, line := range lines {
		if len(line) != width {
			fmt.Println("Not a quad function")
			return
		}
	}

	// Reconstruct the original input as a single string so we can compare it
	// against generateQuad's output. strings.Join puts '\n' between lines,
	// and we add a final '\n' because every quad row ends with one.
	input := strings.Join(lines, "\n") + "\n"

	// Try each quad in alphabetical order (A → E).
	// We store the names in a slice so the order is guaranteed.
	quads := []string{"quadA", "quadB", "quadC", "quadD", "quadE"}

	// matches collects the formatted result strings for every quad that fits.
	var matches []string
	for _, quad := range quads {
		if generateQuad(quad, width, height) == input {
			// fmt.Sprintf works like fmt.Printf but returns the string instead of printing it.
			matches = append(matches, fmt.Sprintf("[%s] [%d] [%d]", quad, width, height))
		}
	}

	if len(matches) == 0 {
		fmt.Println("Not a quad function")
		return
	}

	// strings.Join glues all match strings together with " || " between them.
	fmt.Println(strings.Join(matches, " || "))
}

// ── Installer ─────────────────────────────────────────────────────────────────

// install creates hard links named quadA–quadE pointing to this same binary,
// placed in the same directory as the binary itself.
//
// A hard link is a second directory entry for the same file on disk — it costs
// almost no space and any name resolves to the same executable bytes.
// If hard linking fails (e.g. the filesystem doesn't support it), we fall back
// to a plain file copy.
func install() {
	// os.Executable returns the absolute path of the running binary.
	self, err := os.Executable()
	if err != nil {
		// fmt.Fprintln lets us print to a specific writer; os.Stderr is the
		// standard error stream, separate from the normal output (os.Stdout).
		fmt.Fprintln(os.Stderr, "error resolving executable path:", err)
		os.Exit(1) // non-zero exit code signals failure to the shell
	}

	// filepath.Dir strips the filename and returns just the directory.
	dir := filepath.Dir(self)

	// filepath.Ext returns the file extension including the dot, e.g. ".exe".
	// On Unix this will be an empty string "".
	ext := filepath.Ext(self)

	for _, name := range []string{"quadA", "quadB", "quadC", "quadD", "quadE"} {
		// filepath.Join builds a correct path from parts, handling separators.
		target := filepath.Join(dir, name+ext)

		// Remove any existing file at the target path so we can replace it cleanly.
		os.Remove(target) // we intentionally ignore the error here (file may not exist)

		// Try to create a hard link first.
		if err := os.Link(self, target); err != nil {
			// Hard link failed — fall back to copying the binary.
			if err2 := copyExecutable(self, target); err2 != nil {
				fmt.Fprintf(os.Stderr, "failed to install %s: %v\n", name, err2)
				continue // skip to the next quad name
			}
		}
		fmt.Printf("installed %s\n", target)
	}
}

// copyExecutable copies the file at src to dst, preserving its permission bits
// (e.g. the executable bit on Unix so the file can be run with ./quadA).
func copyExecutable(src, dst string) error {
	// os.Open opens a file for reading. It returns a *File and an error.
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	// defer schedules in.Close() to run when this function returns,
	// no matter how it returns (normal, early return, or panic).
	// This guarantees the file handle is always released.
	defer in.Close()

	// Stat returns metadata about the file (size, permissions, etc.).
	info, err := in.Stat()
	if err != nil {
		return err
	}

	// os.OpenFile opens (or creates) a file with explicit flags and permissions.
	// O_CREATE: create the file if it doesn't exist.
	// O_WRONLY: open for writing only.
	// O_TRUNC:  if the file already exists, empty it first.
	// info.Mode() reuses the source file's permission bits (e.g. rwxr-xr-x).
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, info.Mode())
	if err != nil {
		return err
	}
	defer out.Close()

	// io.Copy streams all bytes from in to out efficiently.
	// The blank identifier '_' discards the byte count we don't need.
	_, err = io.Copy(out, in)
	return err
}

// ── Entry point ───────────────────────────────────────────────────────────────

// main is where Go starts execution. Its job here is purely to dispatch:
// it looks at the name the binary was called as and routes to the right function.
//
// This is the "multi-call binary" pattern — one compiled file that behaves
// differently depending on what name it is invoked under. The same trick is
// used by tools like busybox, which acts as ls, cat, echo, and many others
// all from a single binary.
func main() {
	// os.Args[0] is the path used to invoke the binary, e.g. "./quadA" or
	// "/usr/local/bin/quadchecker". filepath.Base strips the directory part,
	// leaving just the filename. strings.TrimSuffix removes ".exe" on Windows.
	name := strings.TrimSuffix(filepath.Base(os.Args[0]), ".exe")

	switch name {
	case "quadA", "quadB", "quadC", "quadD", "quadE":
		// The binary was called as one of the quad names → draw that quad.
		runQuad(name)
	default:
		// Any other name (including "quadchecker") → act as the checker.
		// But first check if the user passed "install" as an argument.
		if len(os.Args) > 1 && os.Args[1] == "install" {
			install()
			return
		}
		runChecker()
	}
}
