/*
* @Author: Ximidar
* @Date:   2018-10-17 17:14:20
* @Last Modified by:   Ximidar
* @Last Modified time: 2018-10-17 17:36:23
 */

package FileStreamer

import (
	"github.com/ximidar/Flotilla/Flotilla_File_Manager/Files"
)

// FileStreamer Takes a file and streams it to the Nats Comm Object
type FileStreamer struct {
	SelectedFile *Files.File
	Playing      bool
}

// NewFileStreamer will construct a FileStreamer object
func NewFileStreamer() (*FileStreamer, error) {
	fs := new(FileStreamer)
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

	// TODO Publish to NATS server if we are playing
}
