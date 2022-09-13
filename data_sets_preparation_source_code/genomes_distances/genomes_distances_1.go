/*
	Copyright (c) 2022 Farshad Barahimi. Licensed under the MIT license.

	This file (this code) is written by Farshad Barahimi.

	The purpose of writing this code is academic.
*/

package genomes_distances

import (
	"bufio"
	"fmt"
	"github.com/farshad-barahimi-academic-codes/data_sets_preparation/helpers"
	"io/ioutil"
	"math"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type GenomesDistancesPreparation1 struct {
}

func (genomesDistancesPreparation1 GenomesDistancesPreparation1) Prepare(dataSetPreparationInformation *helpers.DataSetPreparationInformation, outputDirectory string) {
	plinkPath := filepath.Join(outputDirectory, "Uncompressed_downloaded_files", "plink2_win64_20220503", "plink2.exe")

	fmt.Println(time.Now().Format(time.UnixDate))
	fmt.Println("Running PLINK 2 ...")

	var debug bool = false

	cmd := exec.Command(plinkPath, "--pfile", "genomes", "--make-rel", "square", "--out", "matrix", "--memory", "30000")

	cmd.Dir = filepath.Join(outputDirectory, "Temporary_files")

	if debug {
		stdoutPipe, err := cmd.StdoutPipe()
		if err != nil {
			panic("Not finished successfully.")
		}
		stdoutPipeScanner := bufio.NewScanner(stdoutPipe)
		go func() {
			for stdoutPipeScanner.Scan() {
				fmt.Println(stdoutPipeScanner.Text())
			}
		}()

		stderrPipe, err := cmd.StderrPipe()
		if err != nil {
			panic("Not finished successfully.")
		}
		stderrPipeScanner := bufio.NewScanner(stderrPipe)
		go func() {
			for stderrPipeScanner.Scan() {
				fmt.Println(stderrPipeScanner.Text())
			}
		}()
	}

	err := cmd.Start()
	if err != nil {
		panic("Not finished successfully.")
	}

	err = cmd.Wait()

	if err != nil {
		panic("Not finished successfully.")
	}

	fmt.Println("Running PLINK 2 finished.")
	fmt.Println(time.Now().Format(time.UnixDate))

	labels := make(map[string]int)
	labels["EAS"] = 1
	labels["EUR"] = 2
	labels["AFR"] = 3
	labels["AMR"] = 4
	labels["SAS"] = 5

	labelsFileContent, err := ioutil.ReadFile(filepath.Join(outputDirectory, "Downloaded_files", "integrated_call_samples_v3.20130502.ALL.panel"))
	if err != nil {
		panic("Not finished successfully.")
	}
	labelsFileLines := strings.Split(string(labelsFileContent), "\n")[1:]

	labelNumbers := make(map[string]int)

	for _, labelsFileLine := range labelsFileLines {
		line := strings.TrimSpace(labelsFileLine)
		if len(line) == 0 {
			continue
		}

		labelNumbers[strings.Split(line, "\t")[0]] = labels[strings.Split(line, "\t")[2]] - 1
	}

	matrixIDsFileContent, err := ioutil.ReadFile(filepath.Join(outputDirectory, "Temporary_files", "matrix.rel.id"))
	if err != nil {
		panic("Not finished successfully.")
	}
	matrixIDsLines := strings.Split(string(matrixIDsFileContent), "\n")[1:]

	matrixIDs := make([]string, 0, len(matrixIDsLines))

	for _, matrixIDLine := range matrixIDsLines {
		matrixID := strings.TrimSpace(matrixIDLine)
		if len(matrixID) != 0 {
			matrixIDs = append(matrixIDs, matrixID)
		}
	}

	matrixFile := filepath.Join(outputDirectory, "Temporary_files", "matrix.rel")
	matrixFileContent, err := ioutil.ReadFile(matrixFile)
	if err != nil {
		panic("Not finished successfully.")
	}
	matrixLines := strings.Split(string(matrixFileContent), "\n")
	maximumFloat := -math.MaxFloat64
	floatNumbers := make([][]float64, 0, len(matrixLines))
	for i := 0; i < len(matrixLines); i++ {
		matrixLine := matrixLines[i]
		if len(strings.TrimSpace(matrixLine)) == 0 {
			continue
		}
		floatNumbers = append(floatNumbers, make([]float64, 0, len(matrixLines)))
		numbers := strings.Split(matrixLine, "\t")
		for _, number := range numbers {
			if len(strings.TrimSpace(number)) == 0 {
				continue
			}

			floatNumber, _ := strconv.ParseFloat(strings.TrimSpace(number), 64)
			maximumFloat = math.Max(maximumFloat, floatNumber)
			floatNumbers[i] = append(floatNumbers[i], floatNumber)
		}
	}

	distancesFile := filepath.Join(outputDirectory, "Final_files", "distances.csv")
	distancesCSV := strings.Builder{}

	for i := 0; i < len(floatNumbers); i++ {
		distancesCSV.WriteString(strconv.Itoa(labelNumbers[matrixIDs[i]]))
		for j := 0; j < len(floatNumbers); j++ {
			distancesCSV.WriteString(",")
			distancesCSV.WriteString(strconv.FormatFloat(maximumFloat-floatNumbers[i][j], 'f', 10, 64))
		}
		distancesCSV.WriteString("\r\n")
	}
	ioutil.WriteFile(distancesFile, []byte(distancesCSV.String()), 600)
}

func PrepareGenomeDistances1(outputDirectory string) {
	dataSetsPreparationInformation := new(helpers.DataSetPreparationInformation)

	dataSetsPreparationInformation.PrefixOfInputDownloadURLs = ""
	dataSetsPreparationInformation.InputDownloadURLs = []string{}

	dataSetsPreparationInformation.Parameters = nil
	dataSetsPreparationInformation.Preparation = GenomesDistancesPreparation1{}
	dataSetsPreparationInformation.OnlySpecificPreparation = true

	dataSetsPreparationInformation.Prepare(outputDirectory)
}
