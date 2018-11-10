/*
* @Author: Ximidar
* @Date:   2018-10-17 17:14:20
* @Last Modified by:   Ximidar
* @Last Modified time: 2018-11-02 20:39:07
 */

package FileStreamer

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/ximidar/Flotilla/Flotilla_File_Manager/Files"
)

// Adapter is an interface for other classes to be read files
// AKA This will read files to the NATS interface
type Adapter interface {
	LineReader(line string)
	ProgressUpdate(file *Files.File, currentLine int, readBytes int)
}

// FileStreamer Takes a file and streams it to the NATS Comm Object
type FileStreamer struct {
	SelectedFile *Files.File
	Playing      bool
	DonePlaying  bool
	LineNumber   int
	Adapter      Adapter

	currentBytes      int
	playingChannel    chan bool
	continueStreaming chan bool
}

// NewFileStreamer will construct a FileStreamer object
func NewFileStreamer(adapter Adapter) (*FileStreamer, error) {
	fs := new(FileStreamer)
	fs.Adapter = adapter

	// Setup Channels
	fs.playingChannel = make(chan bool, 10)
	fs.continueStreaming = make(chan bool)

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

// MonitorFeedback will watch for ok or wait and only continue to the next line when it's okay to do so
func (fs *FileStreamer) MonitorFeedback(line string) {
	monitorFor := []string{"ok", "wait"}
	for _, key := range monitorFor {
		if strings.Contains(line, key) {
			fmt.Println("Continuing!")
			fs.continueStreaming <- true
		}
	}
}

///////////////////
// File Streamer //
///////////////////

// StreamFile will "Play" the selected file to the NATS interface
func (fs *FileStreamer) StreamFile() error {

	fmt.Println("Checking if file can be played")
	// Check if the selected file is there
	if !fs.checkSelectedFile() {
		return errors.New("current file cannot be played")
	}
	fmt.Println("Opening File")
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
	fmt.Println("Playing file")
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

			// Check line to update progress and send the line to the buffer
			fs.checkLine(line)

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

// checkLine will filter out any comments not to be sent
func (fs *FileStreamer) checkLine(line string) {
	// Update the progress
	fs.updateProgress(line)

	// Check if the line is a comment
	trimline := strings.Replace(line, " ", "", -1)
	if trimline[:1] == ";" {
		return
	}

	// Read the line
	go fs.readLine(line)

	// After Reading a line wait for the buffer to ask for another
	select {
	case <-fs.continueStreaming:
		return
		// TODO add cases for pausing and playing and canceling and such
	}
}

func (fs *FileStreamer) checkSelectedFile() bool {
	if fs.SelectedFile == nil {
		return false
	}

	if _, err := os.Stat(fs.SelectedFile.Path); os.IsNotExist(err) {
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

// this function will read the line to the adapter. This is also where we should modify the line for "printing mode" ie G1 X10 turns into n350 G1 X10!9
func (fs *FileStreamer) readLine(line string) {
	fs.Adapter.LineReader(line)
}

// updateProgress will update the progression through the file
func (fs *FileStreamer) updateProgress(line string) {
	fs.currentBytes += len(line)
	fs.LineNumber++
	fs.Adapter.ProgressUpdate(fs.SelectedFile, fs.LineNumber, fs.currentBytes)
}
