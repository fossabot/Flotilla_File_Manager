/*
* @Author: Ximidar
* @Date:   2018-10-01 18:58:24
* @Last Modified by:   Ximidar
* @Last Modified time: 2018-10-05 18:36:21
 */
package main

import (
	"encoding/json"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/ximidar/Flotilla/Flotilla_File_Manager/Files"
	"os"
	ospath "path"
	"path/filepath"
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
	case fsnotify.Write:
		fmt.Println("Write Event")
	case fsnotify.Remove:
		fmt.Println("Remove Event")
	case fsnotify.Rename:
		fmt.Println("Rename Event")
	case fsnotify.Chmod:
		fmt.Println("Permission Event")

	}
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
