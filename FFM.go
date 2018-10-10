/*
* @Author: Ximidar
* @Date:   2018-10-01 18:58:24
* @Last Modified by:   Ximidar
* @Last Modified time: 2018-10-10 01:10:10
 */
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/ximidar/Flotilla/Flotilla_File_Manager/Files"
	"os"
	ospath "path"
	"path/filepath"
	"strings"
)

type File_Manager struct {
	Root_Folder_Path      string
	Current_Selected_File *Files.File
	Structure             map[string]*Files.File
	Watcher               *fsnotify.Watcher
}

func New_File_Manager() *File_Manager {
	fm := new(File_Manager)
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

func (fm *File_Manager) Run() {
	for {
		select {
		default:
			continue
		}
	}
}

func (fm *File_Manager) FS_Watcher() {
	for {
		select {
		// watch for events
		case event := <-fm.Watcher.Events:
			fm.Handle_Event(event)
		// watch for errors
		case err := <-fm.Watcher.Errors:
			fmt.Println("ERROR", err)
			break
		}
	}
}

func (fm *File_Manager) Handle_Event(event fsnotify.Event) {
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
		err := fm.Remove_File_From_Structure(event.Name)
		if err != nil {
			fmt.Println(err)
		}
	case fsnotify.Rename:
		fmt.Println("Rename Event")
		err := fm.Remove_File_From_Structure(event.Name)
		if err != nil {
			fmt.Println(err)
		}
	}
}

func (fm *File_Manager) Remove_File_From_Structure(path string) error {
	file, err := fm.Get_File_By_Path(ospath.Dir(path))
	basename := ospath.Base(path)
	if err != nil {
		return err
	}

	if _, ok := file.Contents[basename]; ok {
		delete(file.Contents, basename)
	} else {
		return errors.New("Could not find file to delete")
	}

	return nil

}

func (fm *File_Manager) Add_File_To_Structure(path string) error {
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
func (fm File_Manager) Get_File_By_Path(path string) (*Files.File, error) {
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
func (fm *File_Manager) Init_Paths() {

	// Create root folder path
	fm.create_root_folder_path()

	// Index FS
	fm.index_fs()

	go fm.FS_Watcher()

}

// walk the root directory and index filesystem while also adding directories to the
// fs watcher
func (fm *File_Manager) index_fs() {
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

func (fm *File_Manager) Print_Structure() {
	marshed, err := json.MarshalIndent(fm.Structure, "", "    ")
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(string(marshed))
}

func (fm *File_Manager) create_root_folder_path() {
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

func main() {
	fmt.Println("Creating File Manager")
	FM := New_File_Manager()
	FM.Run()
}
