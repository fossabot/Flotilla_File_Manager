/*
* @Author: Ximidar
* @Date:   2018-10-10 02:38:49
* @Last Modified by:   Ximidar
* @Last Modified time: 2018-10-17 11:22:04
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

	fm.Init_Paths()
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
		err := fm.Add_File_To_Structure(event.Name)
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
		file, err := fm.Get_File_By_Path(event.Name)
		if err != nil {
			fmt.Println(err)
		}
		file.Update_Info()
	}
}

// RemoveFileFromStructure Will remove a file object from the Structure Map
func (fm *FileManager) RemoveFileFromStructure(path string) error {
	file, err := fm.Get_File_By_Path(ospath.Dir(path))
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

func (fm *FileManager) Add_File_To_Structure(path string) error {
	file, err := fm.Get_File_By_Path(path)
	if err == nil {
		return errors.New("File already exists")
	}

	file, err = fm.Get_File_By_Path(ospath.Dir(path))
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
	file.Index_fs()

	return nil

}

// This function will be given a path and it will return the correct file
func (fm FileManager) Get_File_By_Path(path string) (*Files.File, error) {
	// if the path is the root file, then get the root file
	if path == fm.Root_Folder_Path {
		return fm.Structure["root"], nil
	}
	path_copy := path
	// cut off everything before the root file
	path_copy = strings.Replace(path_copy, fm.Root_Folder_Path+"/", "", 1)

	// split remaining names up by "/" char
	path_keys := strings.Split(path_copy, "/")
	fmt.Println(path_keys)

	// pull file out of structure

	pulled_file := new(Files.File)
	structure := fm.Structure["root"].Contents
	key_length := len(path_keys)
	for index, value := range path_keys {
		if index+1 == key_length {
			var ok bool
			pulled_file, ok = structure[value]
			if !ok {
				return nil, errors.New("File doesn't exist")
			}
			fmt.Println("Pulled File")
		} else {
			temp_structure, ok := structure[value]
			if !ok {
				return nil, errors.New("File doesn't exist")
			}
			structure = temp_structure.Contents
		}
	}

	return pulled_file, nil
}

// This function will check that our Root Folder Path Exists and is created, then it will add fsnotify
// watchers to every path
func (fm *FileManager) Init_Paths() {

	// Create root folder path
	fm.create_root_folder_path()

	// Index FS
	fm.index_fs()

	go fm.FSWatcher()

}

// walk the root directory and index filesystem while also adding directories to the
// fs watcher
func (fm *FileManager) index_fs() {
	filepath.Walk(fm.Root_Folder_Path, func(path string, info os.FileInfo, err error) error {

		if err != nil {
			fmt.Printf("prevent panic by handling failure accessing a path %q: %v\n", path, err)
			return err
		}
		if info.IsDir() {
			fm.Watcher.Add(path)
		}

		return nil

	})

	root_structure := Files.New_File(fm.Root_Folder_Path, "folder")
	root_structure.Index_fs()
	fm.Structure["root"] = root_structure
	fm.Print_Structure()
}

func (fm FileManager) Print_Structure() {
	marshed, err := fm.Get_JSON_Structure()
	if err != nil {
		fmt.Println("Couldn't get json structure:", err)
		return
	}
	fmt.Println(string(marshed))
}

func (fm FileManager) Get_JSON_Structure() ([]byte, error) {
	marshed, err := json.MarshalIndent(fm.Structure, "", "    ")
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	return marshed, nil
}

func (fm *FileManager) create_root_folder_path() {
	// Check and create basic data directory
	home := os.Getenv("HOME")
	fm.Root_Folder_Path = ospath.Clean(home + "/Flotilla_Data")

	if _, err := os.Stat(fm.Root_Folder_Path); os.IsNotExist(err) {
		err := os.Mkdir(fm.Root_Folder_Path, 0755)
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

func (fm FileManager) Add_File(src, dst string) error {

	if in_path := fm.is_in_root_path(dst); !in_path {
		return errors.New("Destination is not in Flotilla Root Path")
	}

	err := fm.cp(src, dst)

	if err != nil {
		return err
	}

	return nil
}

func (fm *FileManager) Move_File(src, dst string) (err error) {
	err = fm.Add_File(src, dst)
	if err != nil {
		return err
	}
	err = fm.Delete_File(src)
	return err
}

func (fm *FileManager) Delete_File(path string) error {

	if in_path := fm.is_in_root_path(path); !in_path {
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

func (fm FileManager) is_in_root_path(path string) (in_path bool) {
	defer func() {
		// recover from panic if one occured.
		if recover() != nil {
			in_path = false
		}
	}()
	in_path = false
	if path[:len(fm.Root_Folder_Path)] == fm.Root_Folder_Path {
		in_path = true
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
