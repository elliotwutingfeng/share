// Copyright 2022, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package os extend the standard os package to provide additional
// functionalities.
package os

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/shuLhan/share/lib/ascii"
)

// ConfirmYesNo display a question to standard output and read for answer
// from input Reader for simple yes "y" or no "n" answer.
// If input Reader is nil, it will set to standard input.
// If "defIsYes" is true and answer is empty (only new line), then it will
// return true.
func ConfirmYesNo(in io.Reader, msg string, defIsYes bool) bool {
	var (
		r         *bufio.Reader
		b, answer byte
		err       error
	)

	if in == nil {
		r = bufio.NewReader(os.Stdin)
	} else {
		r = bufio.NewReader(in)
	}

	yon := "[y/N]"

	if defIsYes {
		yon = "[Y/n]"
	}

	fmt.Printf("%s %s ", msg, yon)

	for {
		b, err = r.ReadByte()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			break
		}
		if b == ' ' || b == '\t' {
			continue
		}
		if b == '\n' {
			break
		}
		if answer == 0 {
			answer = b
		}
	}

	if answer == 'y' || answer == 'Y' {
		return true
	}
	if answer == 0 {
		return defIsYes
	}

	return false
}

// Copy file from in to out.
// If the output file is already exist, it will be truncated.
// If the file is not exist, it will created with permission set to user's
// read-write only.
func Copy(out, in string) (err error) {
	fin, err := os.Open(in)
	if err != nil {
		return fmt.Errorf(`Copy: failed to open input file: %s`, err)
	}

	fout, err := os.OpenFile(out, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf(`Copy: failed to open output file: %s`, err)
	}

	defer func() {
		err := fout.Close()
		if err != nil {
			log.Printf(`Copy: failed to close output file: %s`, err)
		}
	}()
	defer func() {
		err := fin.Close()
		if err != nil {
			log.Printf(`Copy: failed to close input file: %s`, err)
		}
	}()

	buf := make([]byte, 1024)
	for {
		n, err := fin.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		if n == 0 {
			break
		}
		_, err = fout.Write(buf[:n])
		if err != nil {
			return err
		}
	}

	return nil
}

// IsBinary will return true if content of file is binary.
// If file is not exist or there is an error when reading or closing the file,
// it will return false.
func IsBinary(file string) bool {
	var (
		total     int
		printable int
	)

	f, err := os.Open(file)
	if err != nil {
		return false
	}

	content := make([]byte, 768)

	for total < 512 {
		n, err := f.Read(content)
		if err != nil {
			break
		}

		content = content[:n]

		for x := 0; x < len(content); x++ {
			if ascii.IsSpace(content[x]) {
				continue
			}
			if content[x] >= 33 && content[x] <= 126 {
				printable++
			}
			total++
		}
	}

	err = f.Close()
	if err != nil {
		return false
	}

	ratio := float64(printable) / float64(total)

	return ratio <= float64(0.75)
}

// IsDirEmpty will return true if directory is not exist or empty; otherwise
// it will return false.
func IsDirEmpty(dir string) (ok bool) {
	d, err := os.Open(dir)
	if err != nil {
		ok = true
		return
	}

	_, err = d.Readdirnames(1)
	if err != nil {
		if err == io.EOF {
			ok = true
		}
	}

	_ = d.Close()

	return
}

// IsFileExist will return true if relative path is exist on parent directory;
// otherwise it will return false.
func IsFileExist(parent, relpath string) bool {
	path := filepath.Join(parent, relpath)

	fi, err := os.Stat(path)
	if err != nil {
		return false
	}
	if fi.IsDir() {
		return false
	}
	return true
}

// PathFold replace the path "in" with tilde "~" if its prefix match with
// user's home directory from [os.UserHomeDir].
func PathFold(in string) (out string, err error) {
	var (
		logp = `PathFold`

		userHomeDir string
	)

	in = filepath.Clean(in)

	userHomeDir, err = os.UserHomeDir()
	if err != nil {
		return ``, fmt.Errorf(`%s: %s: %w`, logp, in, err)
	}
	if strings.HasPrefix(in, userHomeDir) {
		out = filepath.Join(`~`, in[len(userHomeDir):])
	} else {
		out = in
	}
	return out, nil
}

// PathUnfold expand the tilde "~/" prefix into user's home directory using
// [os.UserHomeDir] and environment variables using [os.ExpandEnv] inside
// the string path "in".
func PathUnfold(in string) (out string, err error) {
	var (
		logp = `PathUnfold`

		userHomeDir string
	)

	if strings.HasPrefix(in, `~/`) {
		userHomeDir, err = os.UserHomeDir()
		if err != nil {
			return ``, fmt.Errorf(`%s: %s: %w`, logp, in, err)
		}
		out = filepath.Join(userHomeDir, in[2:])
	} else {
		out = in
	}

	out = os.ExpandEnv(out)
	out = filepath.Clean(out)

	return out, nil
}

// RmdirEmptyAll remove directory in path if it's empty until one of the
// parent is not empty.
func RmdirEmptyAll(path string) error {
	if len(path) == 0 {
		return nil
	}
	fi, err := os.Stat(path)
	if err != nil {
		return RmdirEmptyAll(filepath.Dir(path))
	}
	if !fi.IsDir() {
		return nil
	}
	if !IsDirEmpty(path) {
		return nil
	}
	err = os.Remove(path)
	if err != nil {
		return err
	}

	return RmdirEmptyAll(filepath.Dir(path))
}
