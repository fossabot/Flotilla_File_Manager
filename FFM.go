/*
* @Author: Ximidar
* @Date:   2018-10-01 18:58:24
* @Last Modified by:   Ximidar
* @Last Modified time: 2018-10-17 11:39:48
 */
package main

import (
	"fmt"

	NI "github.com/ximidar/Flotilla/Flotilla_File_Manager/NatsFile"
)

func main() {
	fmt.Println("Creating File Manager")
	NatsIO, err := NI.NewNatsFile()
	if err != nil {
		panic(err)
	}
	fmt.Println(NatsIO.FileManager.RootFolderPath)
	Run()
}

// Run will keep the program alive
func Run() {
	for {
		select {
		default:
			continue
		}
	}
}
