package main

import (
	"archive/zip"
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	
	"github.com/ledongthuc/pdf"
)

func main() {
	argvSearch()
	inputLoop()
}

func recovery() {  
    if r := recover(); r != nil {
        fmt.Println("recovered:", r)
    }
}

func argvSearch() {
	defer recovery()
		
	if len(os.Args) <= 2 {
		fmt.Println("please enter root directory path and string to search")
	}
	if len(os.Args[2]) < 2 {
		fmt.Println("minimum search string should contain at least 2 characters")
	}
	if exists, err := isDirectoryExists(os.Args[1]); err != nil {
		fmt.Println(err)
	} else if !exists {
		fmt.Printf("%s directory does not exists\n", os.Args[1])
	} else {
		searchRecursively(os.Args[1], (os.Args[2]))
	}
}

func inputLoop() {
	// SET HERE YOUR FOLDER PATH:
	// by enter path into this ` ` quotation mark
	f := `WRITE-YOUR-PATH-FOLDER-HERE`
	
	// print info
	fmt.Println("\n=== Program loop start ===")
	fmt.Println("folder path: ", f)
	fmt.Println("Enter input to search: ")
	scanner := bufio.NewScanner(os.Stdin)
	
	// loop forever
	for scanner.Scan() {
		line := scanner.Text()
		if line == "q" {
			fmt.Println("\nExit!")
            break
		} else if len(line) > 2 {
			searchRecursively(f, (scanner.Text()))
		} else {
			fmt.Println("minimum search string should contain at least 2 characters")
		}
		fmt.Println("\nEnter input to search: ")
	}
}

func searchRecursively(path string, s string) {
	files := getAllFiles(path)
	i := 0
	if files != nil {
		for path, ext := range getAllFiles(path) {
			//fmt.Printf("searching %s...\n", path)
			switch ext {
			case ".pdf":
				if found, _ := searchPDF(path, s); found {
					fmt.Printf("FOUND! %s\n", path)
					i++
				}
				break

			case ".docx":
				if found, _ := searchDocx(path, s); found {
					fmt.Printf("FOUND! %s\n", path)
					i++
				}
				break

			case ".doc":
				break
			}
		}
	} 
	if i == 0 {
		fmt.Printf("\n'%s' Not FOUND :(\n", s)
	} else {
		fmt.Printf("\n'%s' FOUND in %d file(s) :)\n", s, i)
	}
}

func getAllFiles(path string) map[string]string {
	fileList := make(map[string]string, 0)
	if err := filepath.Walk(path, func(path string, f os.FileInfo, err error) error {
		ext := filepath.Ext(path)
		if strings.EqualFold(ext, ".pdf") || strings.EqualFold(ext, ".doc") || strings.EqualFold(ext, ".docx") {
			fileList[path] = strings.ToLower(ext)
		}
		return nil
	}); err != nil {
		fmt.Println(err)
		return nil
	}
	return fileList
}

func isDirectoryExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func printError() {
	err := recover()
	if err != nil && fmt.Sprint(err) != "malformed PDF: reading at offset 0: stream not present" {
		fmt.Println(err)
	}
}

func searchPDF(path string, s string) (bool, error) {
	defer printError()
	_, r, err := pdf.Open(path)
	if err != nil {
		return false, err
	}
	totalPage := r.NumPage()
	for pageIndex := 1; pageIndex <= totalPage; pageIndex++ {
		p := r.Page(pageIndex)
		if p.V.IsNull() {
			continue
		}
		content := p.Content()
		var textBuilder bytes.Buffer
		defer textBuilder.Reset()
		if content.Text != nil {
			for _, t := range content.Text {
				textBuilder.WriteString(t.S)
			}
		}
		str := textBuilder.String()
		if strings.Contains(str, s) {
			return true, nil
		}

	}
	return false, nil
}

func readZipFile(zf *zip.File) ([]byte, error) {
	f, err := zf.Open()
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return ioutil.ReadAll(f)
}
func searchDocx(path string, s string) (bool, error) {
	file, err := os.Open(path)
	if err != nil {
		fmt.Println(err)
	}
	defer file.Close()
	body, err := ioutil.ReadAll(file)
	if err != nil {
		fmt.Println(err)
	}
	zipReader, err := zip.NewReader(bytes.NewReader(body), int64(len(body)))
	if err != nil {
		fmt.Println(err)
	}
	for _, zipFile := range zipReader.File {
		if zipFile.Name == "word/document.xml" {
			unzippedFileBytes, err := readZipFile(zipFile)

			if err != nil {
				log.Println(err)

			}
			if bytes.Contains(unzippedFileBytes, []byte(s)) {
				return true, nil
			}
			continue
		}
	}
	return false, nil
}
