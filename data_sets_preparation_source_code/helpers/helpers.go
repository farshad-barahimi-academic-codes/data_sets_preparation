/*
	Copyright (c) 2022 Farshad Barahimi. Licensed under the MIT license.

	This file (this code) is written by Farshad Barahimi.

	The purpose of writing this code is academic.
*/

package helpers

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

type DataSetPreparationInformation struct {
	PrefixOfInputDownloadURLs string
	InputDownloadURLs         []string
	Parameters                []string
	Preparation               interface {
		Prepare(dataSetPreparationInformation *DataSetPreparationInformation, outputDirectory string)
	}
	OnlySpecificPreparation bool
}

func (dataSetPreparationInformation *DataSetPreparationInformation) Prepare(outputDirectory string) {

	if !dataSetPreparationInformation.OnlySpecificPreparation {
		_, err := os.Stat(outputDirectory)
		if !os.IsNotExist(err) {
			panic("Not finished successfully. Output directory already exists.")
		}

		os.MkdirAll(outputDirectory, 600)
		os.MkdirAll(filepath.Join(outputDirectory, "Downloaded_files"), 600)
		os.MkdirAll(filepath.Join(outputDirectory, "Uncompressed_downloaded_files"), 600)
		os.MkdirAll(filepath.Join(outputDirectory, "Temporary_files"), 600)
		os.MkdirAll(filepath.Join(outputDirectory, "Final_files"), 600)

		for _, inputDownloadURL := range dataSetPreparationInformation.InputDownloadURLs {
			url := dataSetPreparationInformation.PrefixOfInputDownloadURLs + inputDownloadURL
			relativePath := inputDownloadURL
			if strings.HasPrefix(inputDownloadURL, "https://") ||
				strings.HasPrefix(inputDownloadURL, "http://") ||
				strings.HasPrefix(inputDownloadURL, "ftp://") {
				url = inputDownloadURL
				relativePath = path.Base(url)
			}

			filePath := filepath.Join(outputDirectory, "Downloaded_files", relativePath)
			os.MkdirAll(filepath.Dir(filePath), 600)
			downloadFile(url, filePath)

			if strings.HasSuffix(filePath, ".tar.gz") {
				fmt.Println("Uncompressing...", filepath.Base(filePath))
				uncompressedDirectory := filepath.Dir(filepath.Join(outputDirectory, "Uncompressed_downloaded_files", relativePath))
				uncompressedDirectory = filepath.Join(uncompressedDirectory, strings.TrimSuffix(filepath.Base(filePath), ".tar.gz"))
				_, err := os.Stat(uncompressedDirectory)
				if !os.IsNotExist(err) {
					panic("Not finished successfully.")
				}
				uncompressTarGz(filePath, uncompressedDirectory)
				fmt.Println("| Uncompressed", filepath.Base(filePath))
				fmt.Println()
			}

			if strings.HasSuffix(filePath, ".zip") {
				fmt.Println("Uncompressing...", filepath.Base(filePath))
				uncompressedDirectory := filepath.Dir(filepath.Join(outputDirectory, "Uncompressed_downloaded_files", relativePath))
				uncompressedDirectory = filepath.Join(uncompressedDirectory, strings.TrimSuffix(filepath.Base(filePath), ".zip"))
				_, err := os.Stat(uncompressedDirectory)
				if !os.IsNotExist(err) {
					panic("Not finished successfully.")
				}
				uncompressZip(filePath, uncompressedDirectory)
				fmt.Println("| Uncompressed", filepath.Base(filePath))
				fmt.Println()
			}
		}
	}

	if dataSetPreparationInformation.Preparation != nil {
		fmt.Println("Processing further ...")
		dataSetPreparationInformation.Preparation.Prepare(dataSetPreparationInformation, outputDirectory)
		fmt.Println("| Processed further ...")
	}
}

func downloadFile(url string, filePath string) {
	headResponse, err := http.Head(url)
	defer headResponse.Body.Close()
	if err != nil {
		panic("Not finished successfully.")
	}

	urlFileSize := headResponse.ContentLength
	fmt.Println(time.Now().Format(time.UnixDate))
	fmt.Println("Downloading...", url, "|Size ~=", urlFileSize/1024, "KB")

	file, err := os.Create(filePath)
	defer file.Close()
	if err != nil {
		panic("Not finished successfully.")
	}

	response, err := http.Get(url)
	defer response.Body.Close()
	if err != nil {
		panic("Not finished successfully.")
	}

	fileSize, err := io.Copy(file, response.Body)

	if fileSize != urlFileSize {
		panic("Not finished successfully.")
	}

	fmt.Println("| Downloaded", url, "| Size~= ", fileSize/1024, "KB")
	fmt.Println()
}

func uncompressTarGz(compressedFile string, uncompressedDirectory string) {
	file, err := os.Open(compressedFile)
	if err != nil {
		panic("Not finished successfully.")
	}

	gzipReader, err := gzip.NewReader(file)
	if err != nil {
		panic("Not finished successfully.")
	}

	tarReader := tar.NewReader(gzipReader)

	for true {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			panic("Not finished successfully.")
		}

		if header.Typeflag == tar.TypeDir {
			directory := filepath.Join(uncompressedDirectory, header.Name)
			if !strings.HasPrefix(directory, uncompressedDirectory) {
				panic("Not finished successfully.")
			}
			os.MkdirAll(directory, 600)
		} else if header.Typeflag == tar.TypeReg {
			filePath := filepath.Join(uncompressedDirectory, header.Name)
			if !strings.HasPrefix(filePath, uncompressedDirectory) {
				panic("Not finished successfully.")
			}

			uncompressedFile, err := os.Create(filePath)
			if err != nil {
				panic("Not finished successfully.")
			}
			fileSize, err := io.Copy(uncompressedFile, tarReader)
			if err != nil {
				panic("Not finished successfully.")
			}
			if fileSize != header.FileInfo().Size() {
				panic("Not finished successfully.")
			}

			uncompressedFile.Close()
		} else {
			panic("Not finished successfully.")
		}
	}

}

func uncompressZip(compressedFile string, uncompressedDirectory string) {
	zipReader, err := zip.OpenReader(compressedFile)
	if err != nil {
		panic("Not finished successfully.")
	}

	for _, file := range zipReader.File {
		if file.FileInfo().IsDir() {
			directory := filepath.Join(uncompressedDirectory, file.Name)
			if !strings.HasPrefix(directory, uncompressedDirectory) {
				panic("Not finished successfully.")
			}
			os.MkdirAll(directory, 600)
		} else {
			filePath := filepath.Join(uncompressedDirectory, file.Name)
			if !strings.HasPrefix(filePath, uncompressedDirectory) {
				panic("Not finished successfully.")
			}
			_, err := os.Stat(filepath.Dir(filePath))
			if os.IsNotExist(err) {
				os.MkdirAll(filepath.Dir(filePath), 600)
			}

			uncompressedFile, err := os.Create(filePath)
			if err != nil {
				panic("Not finished successfully.")
			}

			opennedFile, err := file.Open()
			if err != nil {
				panic("Not finished successfully.")
			}

			fileSize, err := io.Copy(uncompressedFile, opennedFile)
			if err != nil {
				panic("Not finished successfully.")
			}
			if fileSize != file.FileInfo().Size() {
				panic("Not finished successfully.")
			}

			uncompressedFile.Close()
		}
	}

}
