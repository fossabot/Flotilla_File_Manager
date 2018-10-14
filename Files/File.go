/*
* @Author: Ximidar
* @Date:   2018-10-02 16:48:31
* @Last Modified by:   Ximidar
* @Last Modified time: 2018-10-14 11:19:12
 */

package Files

import (
	"fmt"
	"os"
	ospath "path"
	"time"
)

type File struct {
	Name     string           `json:"name"`
	Path     string           `json:"path"`
	FileType string           `json:"filetype"`
	Size     int64            `json:"size"`
	ModTime  time.Time        `json:"modtime"`
	IsDir    bool             `json:"isdir"`
	Contents map[string]*File `json:"contents,omitempty"`
}

func New_File(path string, filetype string) *File {
	file := new(File)

	file.Path = path
	file.FileType = filetype
	file.IsDir = true
	file.Contents = make(map[string]*File)
	return file
}

func (file *File) Update_Info() {
	stats, err := os.Stat(file.Path)
	if err != nil {
		fmt.Println(err)
		return
	}

	file.populate_file_info(stats)
}

func (file *File) Index_fs() {

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
			file_path := ospath.Clean(file.Path + "/" + fileinfo.Name())
			dir := New_File(file_path, "folder")
			dir.populate_file_info(fileinfo)
			// This should be the first time we enter a path into the structure so there should be no clashes
			file.Contents[dir.Name] = dir
			go dir.Index_fs()
		} else {
			file_path := ospath.Clean(file.Path + "/" + fileinfo.Name())
			pfile := New_File(file_path, "file")
			pfile.populate_file_info(fileinfo)
			file.Contents[pfile.Name] = pfile
		}

	}

}

func (file *File) populate_file_info(info os.FileInfo) {
	file.Name = info.Name()
	file.Size = info.Size()
	file.IsDir = info.IsDir()
	file.ModTime = info.ModTime()

}
