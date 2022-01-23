package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

func main() {
	// Program Name is always the first (implicit) argument
	if len(os.Args) == 1 {
		fmt.Println("tufin cli use :")
		fmt.Println("     cluster        to deploy k3s cluster")
		fmt.Println("     deploy         to deploy Wordpress App")
		fmt.Println("     status         get status of pods in default namespace")
		os.Exit(1)
	} else {
		//cmd := os.Args[1]
		switch os.Args[1] {
		case "cluster":
			fmt.Printf("Deploy k3s cluster \n")
			app := "bash"
			args := []string{"-c", "curl -sfL https://get.k3s.io | sh -"}
			execCmd(app, args)
		case "deploy":
			fmt.Printf("Deploy wordpress App \n")
			app := "kubectl"
			args := []string{"apply", "-f", "wordpressApp.yaml"}
			execCmd(app, args)
		case "status":
			fmt.Printf("Status \n")
			app := "kubectl"
			args := []string{"get", "pods"}
			execCmd(app, args)
		case "--help":
			fmt.Println("tufin cli use :")
			fmt.Println("     cluster        to deploy k3s cluster")
			fmt.Println("     deploy         to deploy Wordpress App")
			fmt.Println("     status         get status of pods in default namespace")
		default:
			fmt.Println("Command not found use tufin --help to see disponible options")
			os.Exit(1)
		}
	}

}

func execCmd(app string, args []string) {
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	exPath := filepath.Dir(ex)
	cmd := exec.Command(app, args...)
	cmd.Dir = exPath

	var stdBuffer bytes.Buffer
	mw := io.MultiWriter(os.Stdout, &stdBuffer)

	cmd.Stdout = mw
	cmd.Stderr = mw

	// Execute the command
	if err := cmd.Run(); err != nil {
		log.Panic(err)
	}
}
