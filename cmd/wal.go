package main

import "os"

type WAL struct {
	file *os.File
}

func NewWal() (*WAL, error) {
	file, err := os.OpenFile(WalName, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}

	return &WAL{file: file}, nil
}

func (w *WAL) Writeln(line string) error {
	return w.Write(line + "\n")
}

func (w *WAL) Write(str string) error {
	_, err := w.file.WriteString(str)
	if err != nil {
		return err
	}
	return nil
}

func (w *WAL) Close() {
	w.file.Close()
}
