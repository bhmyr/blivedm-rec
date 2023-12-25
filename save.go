// Package main provides a CsvWriter type for writing data to CSV files.
package main

import (
	"encoding/csv"
	"os"
	"path"
)

// fp represents a file pointer.
type fp struct {
	path string   // path is the file path.
	ptr  *os.File // ptr is the file pointer.
}

// Files represents a map of file pointers.
type Files map[string]fp

// CsvWriter represents a CSV writer.
type CsvWriter struct {
	files Files  // files is a map of file pointers.
	name  string // name is the user name.
	t     string // time is the current day.
}

// NewCsvWriter creates a new CsvWriter instance.
// It initializes the file pointers and backup file paths.
func NewCsvWriter(name string) (*CsvWriter, error) {
	var t = getTime()
	var files = Files{}
	for _, v := range []string{"free", "paid"} {
		path := config.Dir + t + "_" + v + "_" + name + ".csv"
		file, _ := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
		files[v] = fp{path, file}

	}
	return &CsvWriter{
		files: files,
		name:  name,
		t:     t,
	}, nil
}

// NewFile creates a new file for writing data.
// It updates the file pointers and sets the current day.
func (w *CsvWriter) NewFile() {
	t := getTime()
	for _, v := range []string{"free", "paid"} {
		path := config.Dir + t + "_" + v + "_" + w.name + ".csv"
		file, _ := os.OpenFile(config.Dir+t+"_"+v+"_"+w.name+".csv", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
		w.files[v] = fp{path, file}
	}
	w.t = t
}

// stop closes all the file pointers.
func (w *CsvWriter) Stop() {
	for _, v := range []string{"free", "paid"} {
		w.files[v].ptr.Close()
	}
}

// Close closes all the file pointers, backs up the data, and removes the files.
func (w *CsvWriter) Backup() {
	for _, v := range []string{"free", "paid"} {
		w.files[v].ptr.Close()
		os.Rename(w.files[v].path, config.Backdir+path.Base(w.files[v].path))
	}
}

// Write writes a record to the appropriate file.
// It checks if the current day has changed and creates a new file if necessary.
func (w *CsvWriter) Write(paid bool, record []string) error {
	if w.t != getTime() {
		w.Backup()
		w.NewFile()
	}
	writer := func() *csv.Writer {
		if paid {
			return csv.NewWriter(w.files["paid"].ptr)
		}
		return csv.NewWriter(w.files["free"].ptr)
	}()
	err := writer.Write(record)
	if err != nil {
		return err
	}
	writer.Flush()
	return nil
}
