/*
* @Author: Ximidar
* @Date:   2018-10-17 17:14:20
* @Last Modified by:   Ximidar
* @Last Modified time: 2018-10-23 00:03:48
 */

package FileStreamer

import (
	"bufio"
	"errors"
	"io"
	"os"

	"github.com/ximidar/Flotilla/Flotilla_File_Manager/Files"
)

// Adapter is an interface for other classes to be read files
// AKA This will read files to the nats interface
type Adapter interface {
	LineReader(line string)
	ProgressUpdate(file *Files.File, currentLine int, readBytes int)
}

// FileStreamer Takes a file and streams it to the Nats Comm Object
type FileStreamer struct {
	SelectedFile *Files.File
	Playing      bool
	DonePlaying  bool
	LineNumber   int
	Adapter      Adapter

	currentBytes   int
	playingChannel chan bool
}

// NewFileStreamer will construct a FileStreamer object
func NewFileStreamer(adapter Adapter) (*FileStreamer, error) {
	fs := new(FileStreamer)
	fs.Adapter = adapter
	return fs, nil
}

// SelectFile will select a file
func (fs *FileStreamer) SelectFile(file *Files.File) error {
	fs.SelectedFile = file

	return nil
}

// SetPlaying will set the playing Variable to true or False
func (fs *FileStreamer) SetPlaying(set bool) {
	fs.Playing = set
	if set {
		fs.playingChannel <- set
	}

	// TODO Publish to NATS server if we are playing
}

///////////////////
// File Streamer //
///////////////////

// StreamFile will "Play" the selected file to the nats interface
func (fs *FileStreamer) StreamFile() error {

	// Check if the selected file is there
	if !fs.checkSelectedFile() {
		return errors.New("current file cannot be played")
	}

	// open file
	file, err := fs.openSelectedFile()
	if err != nil {
		return err
	}
	defer file.Close()

	// Make reader
	reader := bufio.NewReader(file)
	if err != nil {
		return err
	}

	// Play the file
	for !fs.DonePlaying {

		if fs.Playing {
			line, lineErr := reader.ReadString('\n')
			if lineErr != nil {
				if lineErr == io.EOF {
					fs.DonePlaying = true
					break
				}
				return lineErr
			}
			fs.readLine(line)
			fs.Adapter.ProgressUpdate(fs.SelectedFile, fs.LineNumber, fs.currentBytes)
		} else {
			// Wait for Playing to be true again
			select {
			case <-fs.playingChannel:
				continue
			}
		}

	}

	return nil
}

func (fs *FileStreamer) checkSelectedFile() bool {
	if fs.SelectedFile == nil {
		return false
	}

	if _, err := os.Stat(fs.SelectedFile.Path); !os.IsNotExist(err) {
		return false
	}

	return true
}

func (fs *FileStreamer) openSelectedFile() (*os.File, error) {
	f, err := os.Open(fs.SelectedFile.Path)
	if err != nil {
		return nil, err
	}
	return f, nil
}

// this function will read the line to the adapter
func (fs *FileStreamer) readLine(line string) {
	fs.currentBytes += len(line)
	fs.Adapter.LineReader(line)
	fs.LineNumber++
}
