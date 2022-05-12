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
	"path"
	"path/filepath"
	"time"
)

type GenomesPreparation1 struct {
}

func (genomesPreparation1 GenomesPreparation1) Prepare(dataSetPreparationInformation *helpers.DataSetPreparationInformation, outputDirectory string) {
	plinkPath := filepath.Join(outputDirectory, "Uncompressed_downloaded_files", "plink2_win64_20220503", "plink2.exe")
	bcfFilePath := filepath.Join(outputDirectory, "Downloaded_files", path.Base(dataSetPreparationInformation.InputDownloadURLs[0]))
	cmd := exec.Command(plinkPath,
		"--make-pgen",
		"--bcf",
		"../Downloaded_files/"+filepath.Base(bcfFilePath),
		"--out",
		"genomes",
		"--split-par",
		/* "b37" */ "2699520", "154931044")

	cmd.Dir = filepath.Join(outputDirectory, "Temporary_files")

	fmt.Println(time.Now().Format(time.UnixDate))
	fmt.Println("Running PLINK 2 ...")

	var debug bool = false

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

func PrepareGenomes1(outputDirectory string) {
	dataSetsPreparationInformation := new(helpers.DataSetPreparationInformation)

	dataSetsPreparationInformation.PrefixOfInputDownloadURLs = "https://ftp-trace.ncbi.nih.gov/1000genomes/ftp/release/20130502/supporting/bcf_files/"
	dataSetsPreparationInformation.InputDownloadURLs = []string{
		"ALL.wgs.phase3_shapeit2_mvncall_integrated_v5.20130502.genotypes.bcf",
		"ALL.wgs.phase3_shapeit2_mvncall_integrated_v5.20130502.genotypes.bcf.csi",
		"https://ftp-trace.ncbi.nih.gov/1000genomes/ftp/release/20130502/integrated_call_samples_v3.20130502.ALL.panel",
		"https://s3.amazonaws.com/plink2-assets/plink2_win64_20220503.zip",
		"https://www.cog-genomics.org/static/bin/plink2_src_220503.zip"}

	dataSetsPreparationInformation.Parameters = nil
	dataSetsPreparationInformation.Preparation = GenomesPreparation1{}
	//dataSetsPreparationInformation.OnlySpecificPreparation = true

	dataSetsPreparationInformation.Prepare(outputDirectory)
}
