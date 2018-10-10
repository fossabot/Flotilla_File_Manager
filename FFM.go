/*
* @Author: Ximidar
* @Date:   2018-10-01 18:58:24
* @Last Modified by:   Ximidar
* @Last Modified time: 2018-10-10 02:47:47
 */
package main

import (
	"fmt"
	FM "github.com/ximidar/Flotilla/Flotilla_File_Manager/File_Manager"
)

func main() {
	fmt.Println("Creating File Manager")
	File_Manager := FM.New_File_Manager()
	fmt.Println(File_Manager.Root_Folder_Path)
	Run()
}

func Run() {
	for {
		select {
		default:
			continue
		}
	}
}
