/*
* @Author: Ximidar
* @Date:   2018-10-02 16:48:31
* @Last Modified by:   Ximidar
* @Last Modified time: 2018-12-07 15:19:41
 */

package Files

import (
	"fmt"
	"os"
	ospath "path"
	"time"
)

// File is a representation of a file or folder on the system
type File struct {
	Name     string           `json:"name"`
	Path     string           `json:"path"`
	FileType string           `json:"filetype"`
	Size     uint64           `json:"size"`
	ModTime  time.Time        `json:"modtime"`
	IsDir    bool             `json:"isdir"`
	Contents map[string]*File `json:"contents,omitempty"`
}

// NewFile is a constructor for File
func NewFile(path string, filetype string) *File {
	file := new(File)

	file.Path = path
	file.FileType = filetype
	file.IsDir = true
	file.Contents = make(map[string]*File)
	return file
}

// UpdateInfo will be called to update the meta data for the file
func (file *File) UpdateInfo() {

	// sometimes file gets dereferenced before we can update it. Recover from calling a null pointer
	defer func() {
		if recover() != nil {
			fmt.Println("UpdateInfo Failed")
			return
		}
	}()
	stats, err := os.Stat(file.Path)
	if err != nil {
		fmt.Println(err)
		return
	}

	file.populateFileInfo(stats)
}

// Indexfs will index all subdirectories if the file is a folder
func (file *File) Indexfs() {

	path, err := os.Open(file.Path)

	if err != nil {
		fmt.Printf("Error Accessing path: %v\nErr: %v", file.Path, err)
		return
	}

	// read dir
	files, err := path.Readdir(-1)
	path.Close()

	if err != nil {
		fmt.Printf("Error Reading Directory: %v\nErr: %v", file.Path, err)
		return
	}

	for _, fileinfo := range files {
		if fileinfo.IsDir() {
			filePath := ospath.Clean(file.Path + "/" + fileinfo.Name())
			dir := NewFile(filePath, "folder")
			dir.populateFileInfo(fileinfo)
			// This should be the first time we enter a path into the structure so there should be no clashes
			file.Contents[dir.Name] = dir
			go dir.Indexfs()
		} else {
			filePath := ospath.Clean(file.Path + "/" + fileinfo.Name())
			pfile := NewFile(filePath, "file")
			pfile.populateFileInfo(fileinfo)
			file.Contents[pfile.Name] = pfile
		}

	}

}

func (file *File) populateFileInfo(info os.FileInfo) {
	file.Name = info.Name()
	file.Size = uint64(info.Size())
	file.IsDir = info.IsDir()
	file.ModTime = info.ModTime()

}
