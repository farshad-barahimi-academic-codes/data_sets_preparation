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
	"os/exec"
	"path/filepath"
	"time"
)

type GenomesDistancesPreparation1 struct {
}

func (genomesDistancesPreparation1 GenomesDistancesPreparation1) Prepare(dataSetPreparationInformation *helpers.DataSetPreparationInformation, outputDirectory string) {
	plinkPath := filepath.Join(outputDirectory, "Uncompressed_downloaded_files", "plink2_win64_20220503", "plink2.exe")

	fmt.Println(time.Now().Format(time.UnixDate))
	fmt.Println("Running PLINK 2 ...")

	var debug bool = false

	cmd := exec.Command(plinkPath, "--pfile", "genomes", "--make-rel", "--memory", "30000")

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
