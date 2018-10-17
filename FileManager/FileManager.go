/*
* @Author: Ximidar
* @Date:   2018-10-10 02:38:49
* @Last Modified by:   Ximidar
* @Last Modified time: 2018-10-17 11:55:52
 */

package FileManager

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	ospath "path"
	"path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/ximidar/Flotilla/Flotilla_File_Manager/Files"
)

// FileManager is an object for keeping track of a folder and
// it's contents.
type FileManager struct {
	RootFolderPath string
	Structure      map[string]*Files.File
	Watcher        *fsnotify.Watcher
}

// NewFileManager will contstruct a FileManager Object
func NewFileManager() *FileManager {
	fm := new(FileManager)
	fm.Structure = make(map[string]*Files.File)

	// Create notify watcher
	var err error
	fm.Watcher, err = fsnotify.NewWatcher()

	if err != nil {
		fmt.Printf("Err: %v\n", err)
		panic(err)
	}

	fm.InitPaths()
	return fm
}

// FSWatcher will watch for events and handle them as they come in
func (fm *FileManager) FSWatcher() {
	for {
		select {
		// watch for events
		case event := <-fm.Watcher.Events:
			fm.HandleEvent(event)
		// watch for errors
		case err := <-fm.Watcher.Errors:
			fmt.Println("ERROR", err)
			break
		}
	}
}

// HandleEvent will handle a File system Event
func (fm *FileManager) HandleEvent(event fsnotify.Event) {
	fmt.Printf("Event at %s\n", event.Name)

	op := event.Op
	fmt.Printf("Event Op: %v\n", op)
	switch op {

	case fsnotify.Create:
		fmt.Println("Create Event")
		err := fm.AddFileToStructure(event.Name)
		if err != nil {
			fmt.Println(err)
		}
	case fsnotify.Remove:
		fmt.Println("Remove Event")
		err := fm.RemoveFileFromStructure(event.Name)
		if err != nil {
			fmt.Println(err)
		}
	case fsnotify.Rename:
		fmt.Println("Rename Event")
		err := fm.RemoveFileFromStructure(event.Name)
		if err != nil {
			fmt.Println(err)
		}

	case fsnotify.Write:
		fmt.Println("Write Event")
		file, err := fm.GetFileByPath(event.Name)
		if err != nil {
			fmt.Println(err)
		}
		file.UpdateInfo()
	}
}

// RemoveFileFromStructure Will remove a file object from the Structure Map
func (fm *FileManager) RemoveFileFromStructure(path string) error {
	file, err := fm.GetFileByPath(ospath.Dir(path))
	basename := ospath.Base(path)
	if err != nil {
		return err
	}
	if tempfile, ok := file.Contents[basename]; ok {
		if tempfile.IsDir {
			fm.Watcher.Remove(tempfile.Path)
		}
		delete(file.Contents, basename)
	} else {
		return errors.New("Could not find file to delete")
	}
	return nil
}

// AddFileToStructure will add a file to the structure that was not previously there
func (fm *FileManager) AddFileToStructure(path string) error {
	file, err := fm.GetFileByPath(path)
	if err == nil {
		return errors.New("File already exists")
	}

	file, err = fm.GetFileByPath(ospath.Dir(path))
	if err != nil {
		fmt.Printf("Could not get %v\n", ospath.Dir(path))
		return err
	}

	// figure out if we need to add another watcher
	stats, fileerr := os.Stat(path)
	if fileerr != nil {
		fmt.Println("Could not stat path", path)
		return fileerr
	}
	if stats.IsDir() {
		fmt.Println("Adding path to fsnotify watcher", path)
		fm.Watcher.Add(path)
	}

	// tell the directory to re-index all it's files
	file.Indexfs()

	return nil

}

// GetFileByPath This function will be given a path and it will return the correct file
func (fm FileManager) GetFileByPath(path string) (*Files.File, error) {
	// if the path is the root file, then get the root file
	if path == fm.RootFolderPath {
		return fm.Structure["root"], nil
	}
	pathCopy := path
	// cut off everything before the root file
	pathCopy = strings.Replace(pathCopy, fm.RootFolderPath+"/", "", 1)

	// split remaining names up by "/" char
	pathKeys := strings.Split(pathCopy, "/")
	fmt.Println(pathKeys)

	// pull file out of structure

	pulledFile := new(Files.File)
	structure := fm.Structure["root"].Contents
	keyLength := len(pathKeys)
	for index, value := range pathKeys {
		if index+1 == keyLength {
			var ok bool
			pulledFile, ok = structure[value]
			if !ok {
				return nil, errors.New("File doesn't exist")
			}
			fmt.Println("Pulled File")
		} else {
			tempStructure, ok := structure[value]
			if !ok {
				return nil, errors.New("File doesn't exist")
			}
			structure = tempStructure.Contents
		}
	}

	return pulledFile, nil
}

// InitPaths This function will check that our Root Folder Path Exists and is created, then it will add fsnotify
// watchers to every path
func (fm *FileManager) InitPaths() {

	// Create root folder path
	fm.createRootFolderPath()

	// Index FS
	fm.indexfs()

	go fm.FSWatcher()

}

// walk the root directory and index filesystem while also adding directories to the
// fs watcher
func (fm *FileManager) indexfs() {
	filepath.Walk(fm.RootFolderPath, func(path string, info os.FileInfo, err error) error {

		if err != nil {
			fmt.Printf("prevent panic by handling failure accessing a path %q: %v\n", path, err)
			return err
		}
		if info.IsDir() {
			fm.Watcher.Add(path)
		}

		return nil

	})

	rootStructure := Files.NewFile(fm.RootFolderPath, "folder")
	rootStructure.Indexfs()
	fm.Structure["root"] = rootStructure
	fm.PrintStructure()
}

// PrintStructure will dump the Structure object to the console
func (fm FileManager) PrintStructure() {
	marshed, err := fm.GetJSONStructure()
	if err != nil {
		fmt.Println("Couldn't get json structure:", err)
		return
	}
	fmt.Println(string(marshed))
}

// GetJSONStructure will create a JSON object of the Structure Object
func (fm FileManager) GetJSONStructure() ([]byte, error) {
	marshed, err := json.MarshalIndent(fm.Structure, "", "    ")
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	return marshed, nil
}

func (fm *FileManager) createRootFolderPath() {
	// Check and create basic data directory
	home := os.Getenv("HOME")
	fm.RootFolderPath = ospath.Clean(home + "/Flotilla_Data")

	if _, err := os.Stat(fm.RootFolderPath); os.IsNotExist(err) {
		err := os.Mkdir(fm.RootFolderPath, 0750)
		if err != nil {
			panic(err)
		}
	}
}

////////// File Operations
/////////////////////////////////////////////////////////////////////////////
// reminder: You do not need to mess with Structure here at all.           //
// fsnotify functions will take care of maintaining the correct structure  //
/////////////////////////////////////////////////////////////////////////////

// AddFile will copy a file from src to dst
func (fm FileManager) AddFile(src, dst string) error {

	if inPath := fm.isInRootPath(dst); !inPath {
		return errors.New("Destination is not in Flotilla Root Path")
	}

	err := fm.cp(src, dst)

	if err != nil {
		return err
	}

	return nil
}

// MoveFile will move a file from src to dst
func (fm *FileManager) MoveFile(src, dst string) (err error) {
	err = fm.AddFile(src, dst)
	if err != nil {
		return err
	}
	err = fm.DeleteFile(src)
	return err
}

// DeleteFile will delete a file at path
func (fm *FileManager) DeleteFile(path string) error {

	if inPath := fm.isInRootPath(path); !inPath {
		return errors.New("Cannot delete file that is not in Flotilla Root Path")
	}
	file, err := os.Stat(path)
	if err != nil {
		return err
	}
	if !file.Mode().IsRegular() {
		// cannot copy non-regular files (e.g., directories,
		// symlinks, devices, etc.)
		return fmt.Errorf("CopyFile: non-regular source file %s (%q)", file.Name(), file.Mode().String())
	}

	// delete the file
	os.Remove(path)

	return nil
}

func (fm FileManager) isInRootPath(path string) (inPath bool) {
	defer func() {
		// recover from panic if one occurred.
		if recover() != nil {
			inPath = false
		}
	}()
	inPath = false
	if path[:len(fm.RootFolderPath)] == fm.RootFolderPath {
		inPath = true
	}

	return

}

// CopyFile copies a file from src to dst. If src and dst files exist, and are
// the same, then return success. Otherise copy the file contents from src to dst.
func (fm FileManager) cp(src, dst string) (err error) {
	sfi, err := os.Stat(src)
	if err != nil {
		return
	}
	if !sfi.Mode().IsRegular() {
		// cannot copy non-regular files (e.g., directories,
		// symlinks, devices, etc.)
		return fmt.Errorf("CopyFile: non-regular source file %s (%q)", sfi.Name(), sfi.Mode().String())
	}
	dfi, err := os.Stat(dst)
	if err != nil {
		if !os.IsNotExist(err) {
			return
		}
	} else {
		if !(dfi.Mode().IsRegular()) {
			return fmt.Errorf("CopyFile: non-regular destination file %s (%q)", dfi.Name(), dfi.Mode().String())
		}
		if os.SameFile(sfi, dfi) {
			return
		}
	}
	err = fm.copyFileContents(src, dst)
	return
}

// copyFileContents copies the contents of the file named src to the file named
// by dst. The file will be created if it does not already exist. If the
// destination file exists, all it's contents will be replaced by the contents
// of the source file.
func (fm FileManager) copyFileContents(src, dst string) (err error) {
	in, err := os.Open(src)
	if err != nil {
		return
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return
	}
	defer func() {
		cerr := out.Close()
		if err == nil {
			err = cerr
		}
	}()
	if _, err = io.Copy(out, in); err != nil {
		return
	}
	err = out.Sync()
	return
}
