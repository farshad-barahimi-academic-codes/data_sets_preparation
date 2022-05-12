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

type GenomesDistancesPreparation1 struct {
}

func (genomesDistancesPreparation1 GenomesDistancesPreparation1) Prepare(dataSetPreparationInformation *helpers.DataSetPreparationInformation, outputDirectory string) {
	plinkPath := filepath.Join(outputDirectory, "Uncompressed_downloaded_files", "plink_win64_20210606", "plink.exe")
	bcfFilePath := filepath.Join(outputDirectory, "Downloaded_files", path.Base(dataSetPreparationInformation.InputDownloadURLs[0]))
	cmd := exec.Command(plinkPath,
		"--make-bed",
		"--bcf",
		"../Downloaded_files/"+filepath.Base(bcfFilePath),
		"--out",
		"genomes",
		"--split-x",
		"b37",
		"no-fail",
		"--keep-allele-order",
		"--allow-extra-chr",
		"0",
		"--const-fid",
		"--vcf-idspace-to",
		"_")

	cmd.Dir = filepath.Join(outputDirectory, "Temporary_files")

	fmt.Println(time.Now().Format(time.UnixDate))

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

	fmt.Println(time.Now().Format(time.UnixDate))

	cmd = exec.Command(plinkPath, "--bflile", "genomes", "--distance-matrix")

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

	err = cmd.Start()
	if err != nil {
		panic("Not finished successfully.")
	}

	err = cmd.Wait()

	if err != nil {
		panic("Not finished successfully.")
	}

	fmt.Println(time.Now().Format(time.UnixDate))
}

func PrepareGenomeDistances1(outputDirectory string) {
	dataSetsPreparationInformation := new(helpers.DataSetPreparationInformation)

	dataSetsPreparationInformation.PrefixOfInputDownloadURLs = "http://localhost/1000_genomes/"
	dataSetsPreparationInformation.InputDownloadURLs = []string{
		"ALL.wgs.phase3_shapeit2_mvncall_integrated_v5.20130502.genotypes.bcf",
		"ALL.wgs.phase3_shapeit2_mvncall_integrated_v5.20130502.genotypes.bcf.csi",
		"https://ftp-trace.ncbi.nih.gov/1000genomes/ftp/release/20130502/integrated_call_samples_v3.20130502.ALL.panel",
		"https://s3.amazonaws.com/plink1-assets/plink_win64_20210606.zip"}

	dataSetsPreparationInformation.Parameters = nil
	dataSetsPreparationInformation.Preparation = GenomesDistancesPreparation1{}

	dataSetsPreparationInformation.BasePreparation(outputDirectory)
}
