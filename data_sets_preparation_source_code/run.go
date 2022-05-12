/*
	Copyright (c) 2022 Farshad Barahimi. Licensed under the MIT license.

	This file (this code) is written by Farshad Barahimi.

	The purpose of writing this code is academic.
*/

package main

import (
	"fmt"
	"github.com/farshad-barahimi-academic-codes/data_sets_preparation/emails_features_1"
	"github.com/farshad-barahimi-academic-codes/data_sets_preparation/genomes_distances"
	"os"
	"strings"
)

func main() {
	args := os.Args

	fmt.Println("Some data set preparation code written by Farshad Barahimi.")
	fmt.Println("Copyright (c) 2022 Farshad Barahimi. Licensed under the MIT license.")
	fmt.Println("The purpose of writing this code is academic.")
	fmt.Println()

	if len(args) < 2 {
		fmt.Println("Not finished successfully. Incorrect number of arguments.")
		return
	}

	dataSetPreparationType := args[1]

	if dataSetPreparationType == "emails_features_1" {
		if len(args) != 6 {
			fmt.Println("Not finished successfully. Incorrect number of arguments.")
			return
		}
		outputDirectory := args[2]
		prefixOfInputDownloadURLs := args[3]
		inputDownloadURLs := args[4]
		parameters := strings.Split(args[5], ",")
		emails_features_1.Run(outputDirectory, prefixOfInputDownloadURLs, inputDownloadURLs, parameters)
	} else if dataSetPreparationType == "genomes_preparation_1" {
		if len(args) != 3 && len(args) != 4 {
			fmt.Println("Not finished successfully. Incorrect number of arguments.")
			return
		}
		outputDirectory := args[2]
		if len(args) == 3 {
			genomes_distances.PrepareGenomes1(outputDirectory, nil)
		} else {
			if args[3] == "" {
				panic("Not finished successfully.")
			}
			genomes_distances.PrepareGenomes1(outputDirectory, args[3])
		}

	} else if dataSetPreparationType == "genomes_distances_1" {
		if len(args) != 3 {
			fmt.Println("Not finished successfully. Incorrect number of arguments.")
			return
		}
		outputDirectory := args[2]

		genomes_distances.PrepareGenomeDistances1(outputDirectory)
	} else {
		fmt.Println("Not finished successfully.")
		return
	}
}
