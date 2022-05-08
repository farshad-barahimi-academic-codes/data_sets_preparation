/*
	Copyright (c) 2022 Farshad Barahimi. Licensed under the MIT license.

	This file (this code) is written by Farshad Barahimi.

	The purpose of writing this code is academic.
*/

package emails_features_1

import (
	"fmt"
	pq "github.com/emirpasic/gods/queues/priorityqueue"
	"github.com/emirpasic/gods/utils"
	"io/fs"
	"io/ioutil"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

import helpers "github.com/farshad-barahimi-academic-codes/data_sets_preparation/helpers"

type EmailFeaturesPreparation struct {
}

func (emailFeaturesPreparation EmailFeaturesPreparation) Prepare(dataSetPreparationInformation *helpers.DataSetPreparationInformation, outputDirectory string) {
	fmt.Println()
	fmt.Println("Note: all numbers reported may be subject to rounding or truncation rounding. Assumption of exact value should not be made without looking at the source code.")
	fmt.Println()

	selectAndCopyEmails(dataSetPreparationInformation, outputDirectory)

	emailsDirectory := filepath.Join(outputDirectory, "Temporary_files", "emails")

	numberOfEmails,
		initialParsedWords,
		emailsDirectoryNumbers,
		insideEmailWordFreqNormalized,
		numberOfEmailsContainingWord := parseFilteredEmailsDirectoryAndCalculateInitialWordStats(emailsDirectory)

	if numberOfEmails != len(emailsDirectoryNumbers) {
		panic("Not finished successfully.")
	}

	fmt.Println("Number of initial parsed words:", len(initialParsedWords))
	stopWordsFilteredWords := filterStopWords(initialParsedWords)
	fmt.Println("Number of stop words filtered words:", len(initialParsedWords))
	basicFilteredWords := filterWordsLexical(stopWordsFilteredWords)
	fmt.Println("Number of basic filtered words:", len(basicFilteredWords))

	perEmailSignificanceForBasicFilteredWords := computePerEmailSignificanceForBasicFilteredWords(numberOfEmails, basicFilteredWords, insideEmailWordFreqNormalized, numberOfEmailsContainingWord)
	perEmailSignificanceRanksForBasicFilteredWords := computePerEmailSignificanceRanksForBasicFilteredWords(numberOfEmails, perEmailSignificanceForBasicFilteredWords)
	fmt.Println("Significance and significance ranks for basic filtered words calculated.")

	firstFreqFilteredWords := filterWordsWithLowNumberOfEmailsContainingWord(basicFilteredWords, numberOfEmailsContainingWord)
	fmt.Println("Number of first frequency filtered words", len(firstFreqFilteredWords))

	secondFreqFilteredWords := filterWordsWithLowNumberOfOccurrencesInPerEmailHighRankingWords(numberOfEmails, basicFilteredWords, firstFreqFilteredWords, perEmailSignificanceRanksForBasicFilteredWords)
	sort.Strings(secondFreqFilteredWords)
	fmt.Println("Number of second frequency filtered words", len(secondFreqFilteredWords))

	perEmailSignificanceForSecondFreqFilteredWords,
		perEmailSignificanceRanksForSecondFreqFilteredWords := extractPerEmailSignificanceAndSignificanceRanksForSecondFreqFilteredWords(numberOfEmails, basicFilteredWords, secondFreqFilteredWords, perEmailSignificanceForBasicFilteredWords, perEmailSignificanceRanksForBasicFilteredWords)
	fmt.Println("Significance and significance ranks for second frequency filtered words extracted.")
	fmt.Println()

	perEmailCosineTailoredFeatures := computePerEmailCosineTailoredFeatures(numberOfEmails, secondFreqFilteredWords, perEmailSignificanceForSecondFreqFilteredWords, perEmailSignificanceRanksForSecondFreqFilteredWords, emailsDirectoryNumbers)
	perEmailCosineTailoredFeaturesAndDirectoryNumber := combineFeaturesWithEmailDirectoryNumber(perEmailCosineTailoredFeatures, emailsDirectoryNumbers)
	scrambleTheSortingOfEmailsAndWriteToFile(numberOfEmails, perEmailCosineTailoredFeaturesAndDirectoryNumber, filepath.Join(outputDirectory, "Temporary_files", "email_features.csv"))
	computeKnnClassificationAccuracy(perEmailCosineTailoredFeaturesAndDirectoryNumber, 10)
}

func selectAndCopyEmails(dataSetPreparationInformation *helpers.DataSetPreparationInformation, outputDirectory string) {
	labelNumbers := make(map[string]int)

	for _, parameter := range dataSetPreparationInformation.Parameters {
		labelNumbers[parameter] = len(labelNumbers) + 1
	}

	minimumEmailsCount := 300
	maximumEmailsCount := 300

	firstDirectory := make(map[string]string)
	directorySelectedEmailsCount := make(map[string]int)
	totalSelectedEmailsCount := 0

	filepath.Walk(filepath.Join(outputDirectory, "Uncompressed_downloaded_files"), func(path string, info fs.FileInfo, err error) error {
		if !info.IsDir() {
			directory := filepath.Base(filepath.Dir(path))
			labelNumber := labelNumbers[directory] - 1
			if labelNumber == -1 {
				return nil
			}

			if firstDirectory[directory] == "" {
				filePaths, _ := filepath.Glob(filepath.Dir(path) + "/*")
				emailsCount := 0
				for _, filePath := range filePaths {
					stat, _ := os.Stat(filePath)
					if !stat.IsDir() {
						emailsCount++
					}
				}

				if emailsCount < minimumEmailsCount {
					return nil
				}
				firstDirectory[directory] = filepath.Dir(path)
			} else if firstDirectory[directory] != filepath.Dir(path) {
				return nil
			}

			copyDirectory := filepath.Join(outputDirectory, "Temporary_files", "emails", strconv.Itoa(labelNumber))
			_, err = os.Stat(copyDirectory)
			if os.IsNotExist(err) {
				os.MkdirAll(copyDirectory, 600)
			}

			if directorySelectedEmailsCount[directory] >= maximumEmailsCount {
				return nil
			}

			copyPath := filepath.Join(copyDirectory, filepath.Base(path))

			bytes, err := ioutil.ReadFile(path)
			if err != nil {
				panic("Not finished successfully.")
			}

			err = ioutil.WriteFile(copyPath, bytes, 600)
			if err != nil {
				panic("Not finished successfully.")
			}

			directorySelectedEmailsCount[directory]++
			totalSelectedEmailsCount++

			if totalSelectedEmailsCount%100 == 0 {
				fmt.Println("Current number of selected emails:", totalSelectedEmailsCount)
			}
		}
		return nil
	})

	fmt.Println("Total number of emails selected:", totalSelectedEmailsCount)
	fmt.Println("Directories, directory numbers and number of emails selected:")
	for directory, directoryNumber := range labelNumbers {
		relativeDirectory := strings.TrimPrefix(firstDirectory[directory], filepath.Join(outputDirectory, "Uncompressed_downloaded_files"))
		relativeDirectory = strings.ReplaceAll(relativeDirectory, "\\", "/")
		relativeDirectory = strings.TrimPrefix(relativeDirectory, "/")
		fmt.Println("\t", directory, "(", relativeDirectory, "):", "(Number:", directoryNumber-1, ") , (Number of emails:", directorySelectedEmailsCount[directory], ")")
	}
}

func parseFilteredEmailsDirectoryAndCalculateInitialWordStats(emailsDirectory string) (
	numberOfEmails int,
	words []string,
	emailsDirectoryNumber []int,
	insideEmailWordFreqNormalized []map[string]float64,
	numberOfEmailContainingWord map[string]int) {

	emailsDirectoryNumber = make([]int, 0)
	insideEmailWordFreqNormalized = make([]map[string]float64, 0)
	numberOfEmailContainingWord = make(map[string]int)
	words = make([]string, 0)

	numberOfEmails = 0

	filepath.Walk(emailsDirectory, func(path string, info fs.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}

		directoryNumber, err := strconv.Atoi(filepath.Base(filepath.Dir(path)))
		if err != nil {
			panic("Not finished successfully.")
		}

		bytes, err := ioutil.ReadFile(path)
		if err != nil {
			panic("Not finished successfully.")
		}

		numberOfEmails++

		lines := strings.Split(strings.ToLower(string(bytes)), "\n")

		emailsDirectoryNumber = append(emailsDirectoryNumber, directoryNumber)
		emailWordFreq := make(map[string]int)
		emailWordFreqNormalized := make(map[string]float64)
		insideEmailWordFreqNormalized = append(insideEmailWordFreqNormalized, emailWordFreqNormalized)
		emailWordFreqSum := 0

		trimLineNumber := -1
		subjectLine := []string{""}
		for lineNumber, line := range lines {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "x-filename:") {
				trimLineNumber = lineNumber
			}

			if strings.HasPrefix(line, "subject:") && trimLineNumber == -1 {
				subjectLine[0] = line
			}
		}

		if trimLineNumber == -1 {
			panic("Not finished successfully.")
		}

		lines = lines[trimLineNumber+1:]
		lines = append(subjectLine, lines...)

		for _, line := range lines {
			if strings.HasPrefix(strings.TrimSpace(line), "---") {
				continue
			}
			initialDelimiters := " ,:!=;'>[]()"

			linePiecesInitial := strings.FieldsFunc(line, func(r rune) bool { return strings.ContainsRune(initialDelimiters, r) })
			linePiecesSecondary := make([]string, 0)
			for _, linePiece := range linePiecesInitial {
				if !strings.Contains(linePiece, "@") {
					linePiecesSecondary = append(linePiecesSecondary, linePiece)
				}
			}

			line = strings.Join(linePiecesSecondary, " ")

			linePiecesThird := strings.FieldsFunc(line, func(r rune) bool { return strings.ContainsRune(initialDelimiters+".", r) })
			linePieces := make([]string, 0)
			for _, linePiece := range linePiecesThird {
				if strings.TrimSpace(linePiece) != "" {
					linePieces = append(linePieces, strings.TrimSpace(linePiece))
				}
			}

			for _, linePiece := range linePieces {
				if emailWordFreq[linePiece] == 0 {
					if numberOfEmailContainingWord[linePiece] == 0 {
						words = append(words, linePiece)
					}

					numberOfEmailContainingWord[linePiece]++
				}
				emailWordFreq[linePiece]++
				emailWordFreqSum++
			}

			for _, linePiece := range linePieces {
				emailWordFreqNormalized[linePiece] = float64(emailWordFreq[linePiece]) / float64(emailWordFreqSum)
			}
		}

		return nil
	})

	sort.Strings(words)

	return numberOfEmails, words, emailsDirectoryNumber, insideEmailWordFreqNormalized, numberOfEmailContainingWord
}

func filterStopWords(initialParsedWords []string) []string {
	stopWordList := []string{}
	stopWordList = append(stopWordList, []string{"i", "we", "you", "they", "she", "he", "it", "this", "that", "these", "those"}...)
	stopWordList = append(stopWordList, []string{"my", "our", "your", "their", "her", "his", "its"}...)
	stopWordList = append(stopWordList, []string{"me", "us", "you", "them", "her", "him"}...)
	stopWordList = append(stopWordList, []string{"was", "were", "am", "are", "is", "be", "been", "being", "will", "would", "could", "can", "had", "has", "have", "may", "might", "should"}...)
	stopWordList = append(stopWordList, []string{"and", "or", "but", "also", "however", "so", "because", "not", "if", "then"}...)
	stopWordList = append(stopWordList, []string{"a", "an", "the"}...)
	stopWordList = append(stopWordList, []string{"what", "who", "which", "where", "when", "how", "did", "do", "does", "why"}...)
	stopWordList = append(stopWordList, []string{"only", "all", "just", "any", "few", "some", "other", "much", "very"}...)
	stopWordList = append(stopWordList, []string{"one", "two", "three", "four", "five", "six", "seven", "eight", "nine", "ten"}...)
	stopWordList = append(stopWordList, []string{"get", "let", "want", "like"}...)
	stopWordList = append(stopWordList, []string{"here", "there", "in", "of", "for", "at", "as", "with", "by", "on"}...)
	stopWordList = append(stopWordList, []string{"http", "https", "www", "com"}...)
	stopWordList = append(stopWordList, []string{"to", "from", "subject", "cc", "bcc", "re", "fw", "forwarded", "sent", "am", "pm", "attached", "regards", "best", "find", "email", "following", "thanks", "thank", "dear", "hi", "hello", "fax", "phone", "address", "e-mail", "below", "fyi"}...)
	stopWordList = append(stopWordList, []string{"eol"}...)
	stopWordList = append(stopWordList, []string{"monday", "tuesday", "wednesday", "thursday", "friday", "saturday", "sunday"}...)
	stopWordList = append(stopWordList, []string{"january", "february", "march", "april", "may", "june", "july", "august", "september", "october", "november", "december"}...)

	stopWords := make(map[string]bool)
	for _, stopWord := range stopWordList {
		stopWords[stopWord] = true
	}

	filteredWords := make(map[string]bool)

	for _, word := range initialParsedWords {
		if !stopWords[word] {
			filteredWords[word] = true
		}
	}

	filteredWordsList := make([]string, 0, len(filteredWords))
	for word := range filteredWords {
		filteredWordsList = append(filteredWordsList, word)
	}

	return filteredWordsList
}

func filterWordsLexical(stopWordsFilteredWords []string) []string {

	filteredWords := make(map[string]bool)

	for _, word := range stopWordsFilteredWords {
		if strings.ContainsAny(word, "/@&-0123456789") {
			continue
		}

		if len(word) == 1 {
			continue
		}

		filteredWords[word] = true
	}

	filteredWordsList := make([]string, 0, len(filteredWords))
	for word := range filteredWords {
		filteredWordsList = append(filteredWordsList, word)
	}

	return filteredWordsList
}

func filterWordsWithLowNumberOfEmailsContainingWord(basicFilteredWords []string, numberOfEmailsContainingWord map[string]int) map[string]bool {
	firstFreqFilteredWords := make(map[string]bool)

	for _, word := range basicFilteredWords {
		if numberOfEmailsContainingWord[word] >= 10 {
			firstFreqFilteredWords[word] = true
		}
	}

	return firstFreqFilteredWords
}

func computePerEmailSignificanceForBasicFilteredWords(numberOfEmails int, basicFilteredWords []string, insideEmailWordFreqNormalized []map[string]float64, numberOfEmailsContainingWord map[string]int) [][]float64 {
	perEmailSignificanceForBasicFilteredWords := make([][]float64, 0)

	for emailNumber := 0; emailNumber < numberOfEmails; emailNumber++ {
		emailWordFreqNormalized := insideEmailWordFreqNormalized[emailNumber]
		thisEmailSignificanceForBasicFilteredWords := make([]float64, len(basicFilteredWords))
		perEmailSignificanceForBasicFilteredWords = append(perEmailSignificanceForBasicFilteredWords, thisEmailSignificanceForBasicFilteredWords)

		emailWords := make(map[string]bool)
		for word := range emailWordFreqNormalized {
			emailWords[word] = true
		}

		for wordNumber, word := range basicFilteredWords {
			if emailWords[word] {
				thisEmailSignificanceForBasicFilteredWords[wordNumber] = emailWordFreqNormalized[word] * math.Log(float64(numberOfEmails)/float64(numberOfEmailsContainingWord[word]))
			} else {
				thisEmailSignificanceForBasicFilteredWords[wordNumber] = -1
			}
		}
	}

	return perEmailSignificanceForBasicFilteredWords
}

func computePerEmailSignificanceRanksForBasicFilteredWords(numberOfEmails int, perEmailSignificanceForBasicFilteredWords [][]float64) [][]int {
	insideEmailsWordRanks := make([][]int, 0)

	numberOfBasicFilteredWords := len(perEmailSignificanceForBasicFilteredWords[0])

	for emailNumber := 0; emailNumber < numberOfEmails; emailNumber++ {
		thisEmailSignificanceForBasicFilteredWords := perEmailSignificanceForBasicFilteredWords[emailNumber]
		insideEmailWordRanks := make([]int, numberOfBasicFilteredWords)
		insideEmailsWordRanks = append(insideEmailsWordRanks, insideEmailWordRanks)

		wordIndicesToBeSorted := make([]int, numberOfBasicFilteredWords)
		for i := 0; i < numberOfBasicFilteredWords; i++ {
			wordIndicesToBeSorted[i] = i
		}

		sort.Slice(wordIndicesToBeSorted, func(i int, j int) bool {
			return thisEmailSignificanceForBasicFilteredWords[wordIndicesToBeSorted[i]] > thisEmailSignificanceForBasicFilteredWords[wordIndicesToBeSorted[j]]
		})

		for rank, wordIndex := range wordIndicesToBeSorted {
			if thisEmailSignificanceForBasicFilteredWords[wordIndex] < -0.5 {
				insideEmailWordRanks[wordIndex] = -1
			} else {
				insideEmailWordRanks[wordIndex] = rank + 1
			}
		}
	}

	return insideEmailsWordRanks
}

func filterWordsWithLowNumberOfOccurrencesInPerEmailHighRankingWords(numberOfEmails int, basicFilteredWords []string, firstFreqFilteredWords map[string]bool, perEmailSignificanceRanksForBasicFilteredWords [][]int) []string {
	wordHighRankOccurrences := make(map[string]int)

	numberOfBasicFilteredWords := len(basicFilteredWords)

	for emailNumber := 0; emailNumber < numberOfEmails; emailNumber++ {
		thisEmailSignificanceRanksForBasicFilteredWords := perEmailSignificanceRanksForBasicFilteredWords[emailNumber]

		for i := 0; i < numberOfBasicFilteredWords; i++ {
			word := basicFilteredWords[i]

			if thisEmailSignificanceRanksForBasicFilteredWords[i] <= 100 && thisEmailSignificanceRanksForBasicFilteredWords[i] != -1 {
				wordHighRankOccurrences[word]++
			}
		}
	}

	secondFreqFilteredWords := make(map[string]bool)
	for word := range firstFreqFilteredWords {
		if wordHighRankOccurrences[word] > 50 {
			secondFreqFilteredWords[word] = true
		}
	}

	secondFreqFilteredList := make([]string, 0, len(secondFreqFilteredWords))

	for word := range secondFreqFilteredWords {
		secondFreqFilteredList = append(secondFreqFilteredList, word)
	}

	return secondFreqFilteredList
}

func extractPerEmailSignificanceAndSignificanceRanksForSecondFreqFilteredWords(numberOfEmails int, basicFilteredWords []string, secondFreqFilteredWords []string, perEmailSignificanceForBasicFilteredWords [][]float64, perEmailSignificanceRanksForBasicFilteredWords [][]int) ([][]float64, [][]int) {
	wordBasicFilteredWordsIndices := make(map[string]int)
	for wordNumber, word := range basicFilteredWords {
		wordBasicFilteredWordsIndices[word] = wordNumber
	}

	perEmailSignificanceForSecondFreqFilteredWords := make([][]float64, numberOfEmails)

	for emailNumber := 0; emailNumber < numberOfEmails; emailNumber++ {
		perEmailSignificanceForSecondFreqFilteredWords[emailNumber] = make([]float64, len(secondFreqFilteredWords))
		for wordNumber, word := range secondFreqFilteredWords {
			perEmailSignificanceForSecondFreqFilteredWords[emailNumber][wordNumber] = perEmailSignificanceForBasicFilteredWords[emailNumber][wordBasicFilteredWordsIndices[word]]
		}
	}

	perEmailSignificanceRanksForSecondFreqFilteredWords := make([][]int, numberOfEmails)
	for emailNumber := 0; emailNumber < numberOfEmails; emailNumber++ {
		perEmailSignificanceRanksForSecondFreqFilteredWords[emailNumber] = make([]int, len(secondFreqFilteredWords))
		for wordNumber, word := range secondFreqFilteredWords {
			thisEmailWordSignificanceRankForSecondFreqFilteredWords := 1
			rankBefore := perEmailSignificanceRanksForBasicFilteredWords[emailNumber][wordBasicFilteredWordsIndices[word]]

			if rankBefore != -1 && rankBefore <= 100 {
				for _, wordB := range secondFreqFilteredWords {
					rankBeforeB := perEmailSignificanceRanksForBasicFilteredWords[emailNumber][wordBasicFilteredWordsIndices[wordB]]
					if rankBefore > rankBeforeB && rankBeforeB != -1 {
						thisEmailWordSignificanceRankForSecondFreqFilteredWords++
					}
				}
			} else {
				thisEmailWordSignificanceRankForSecondFreqFilteredWords = -1
			}

			perEmailSignificanceRanksForSecondFreqFilteredWords[emailNumber][wordNumber] = thisEmailWordSignificanceRankForSecondFreqFilteredWords
		}
	}

	return perEmailSignificanceForSecondFreqFilteredWords, perEmailSignificanceRanksForSecondFreqFilteredWords

}

func computePerEmailCosineTailoredFeatures(numberOfEmails int, secondFilteredFreqWords []string, featuresSecondRound [][]float64, perEmailSignificanceRanksForSecondFreqFilteredWords [][]int, emailsDirectoryNumber []int) [][]uint8 {
	averageSecondFreqFilteredWordsOccurrenceInPerEmailTopRankingsForBasicFilteredWords := 0

	for emailNumber := 0; emailNumber < numberOfEmails; emailNumber++ {
		for wordNumber, _ := range secondFilteredFreqWords {
			insideEmailRank := perEmailSignificanceRanksForSecondFreqFilteredWords[emailNumber][wordNumber]
			if insideEmailRank != -1 {
				averageSecondFreqFilteredWordsOccurrenceInPerEmailTopRankingsForBasicFilteredWords++
			}
		}
	}

	averageSecondFreqFilteredWordsOccurrenceInPerEmailTopRankingsForBasicFilteredWords = averageSecondFreqFilteredWordsOccurrenceInPerEmailTopRankingsForBasicFilteredWords / numberOfEmails

	fmt.Println("Average second frequency filtered words occurrence in per email top rankings for basic filtered words:", averageSecondFreqFilteredWordsOccurrenceInPerEmailTopRankingsForBasicFilteredWords)

	numberOfNonZeroPrimaryFeaturesPerEmailUpperBound := averageSecondFreqFilteredWordsOccurrenceInPerEmailTopRankingsForBasicFilteredWords * 2

	fmt.Println("Number of non zero primary features per email upper bound:", numberOfNonZeroPrimaryFeaturesPerEmailUpperBound)

	var averageNumberOfNonZeroPrimaryFeaturesPerEmail float64 = 0
	numberOfNonZeroPrimaryFeaturesPerEmail := make([]float64, numberOfEmails)
	averageNumberOfNonZeroPrimaryFeaturesPerEmailPerDirectoryNumber := make(map[int]float64)
	numberOfSelectedEmailPerDirectoryNumber := make(map[int]int)

	var standardDeviationOfNumberOfNonZeroPrimaryFeaturesPerEmail float64 = 0
	standardDeviationOfNumberOfNonZeroPrimaryFeaturesPerEmailPerDirectoryNumber := make(map[int]float64)

	selectedWordsPerEmail := make([]map[string]bool, 0)
	for emailNumber := 0; emailNumber < numberOfEmails; emailNumber++ {
		emailSelectedWords := make(map[string]bool)
		for i := 0; i < len(secondFilteredFreqWords); i++ {
			emailWordRank := perEmailSignificanceRanksForSecondFreqFilteredWords[emailNumber][i]

			if emailWordRank <= numberOfNonZeroPrimaryFeaturesPerEmailUpperBound && emailWordRank != -1 {
				emailSelectedWords[secondFilteredFreqWords[i]] = true
			}
		}
		selectedWordsPerEmail = append(selectedWordsPerEmail, emailSelectedWords)
		numberOfNonZeroPrimaryFeaturesPerEmail[emailNumber] = float64(len(emailSelectedWords))
		averageNumberOfNonZeroPrimaryFeaturesPerEmail += float64(len(emailSelectedWords))
		averageNumberOfNonZeroPrimaryFeaturesPerEmailPerDirectoryNumber[emailsDirectoryNumber[emailNumber]] += float64(len(emailSelectedWords))
		numberOfSelectedEmailPerDirectoryNumber[emailsDirectoryNumber[emailNumber]]++
	}

	averageNumberOfNonZeroPrimaryFeaturesPerEmail /= float64(numberOfEmails)
	maximumDirectoryNumber := -1

	for emailNumber := 0; emailNumber < numberOfEmails; emailNumber++ {
		standardDeviationOfNumberOfNonZeroPrimaryFeaturesPerEmail += math.Pow(numberOfNonZeroPrimaryFeaturesPerEmail[emailNumber]-averageNumberOfNonZeroPrimaryFeaturesPerEmail, 2)
		standardDeviationOfNumberOfNonZeroPrimaryFeaturesPerEmailPerDirectoryNumber[emailsDirectoryNumber[emailNumber]] += math.Pow(numberOfNonZeroPrimaryFeaturesPerEmail[emailNumber]-averageNumberOfNonZeroPrimaryFeaturesPerEmailPerDirectoryNumber[emailsDirectoryNumber[emailNumber]]/float64(numberOfSelectedEmailPerDirectoryNumber[emailsDirectoryNumber[emailNumber]]), 2)
		if emailsDirectoryNumber[emailNumber] > maximumDirectoryNumber {
			maximumDirectoryNumber = emailsDirectoryNumber[emailNumber]
		}
	}

	standardDeviationOfNumberOfNonZeroPrimaryFeaturesPerEmail /= float64(numberOfEmails)
	standardDeviationOfNumberOfNonZeroPrimaryFeaturesPerEmail = math.Sqrt(standardDeviationOfNumberOfNonZeroPrimaryFeaturesPerEmail)
	fmt.Println()
	fmt.Println("Average number of non zero primary features per email:", averageNumberOfNonZeroPrimaryFeaturesPerEmail)
	fmt.Println("Standard deviation of number of non zero primary features per email:", standardDeviationOfNumberOfNonZeroPrimaryFeaturesPerEmail)
	fmt.Println("Per directory number:")
	for i := 0; i <= maximumDirectoryNumber; i++ {
		averageNumberOfNonZeroPrimaryFeaturesPerEmailPerDirectoryNumber[i] /= float64(numberOfSelectedEmailPerDirectoryNumber[i])
		standardDeviationOfNumberOfNonZeroPrimaryFeaturesPerEmailPerDirectoryNumber[i] = math.Sqrt(standardDeviationOfNumberOfNonZeroPrimaryFeaturesPerEmailPerDirectoryNumber[i] / float64(numberOfSelectedEmailPerDirectoryNumber[i]))
		fmt.Println("\t", i, ":", "(average:", averageNumberOfNonZeroPrimaryFeaturesPerEmailPerDirectoryNumber[i], ") , (standard deviation:", standardDeviationOfNumberOfNonZeroPrimaryFeaturesPerEmailPerDirectoryNumber[i], ")")
	}
	fmt.Println()
	/////////////////////////

	toBeShuffled := make([][]uint8, numberOfEmails)
	numberOfPrimaryFeatures := len(secondFilteredFreqWords)
	currentNumberOfSecondaryFeatures := 0
	numberOfSecondaryFeaturesUpperBound := 50000

	for emailNumber := 0; emailNumber < numberOfEmails; emailNumber++ {
		toBeShuffled[emailNumber] = make([]uint8, numberOfPrimaryFeatures+numberOfSecondaryFeaturesUpperBound)
		featureNumber := -1
		numberOfTopRankingSelectedWords := 0
		emailSelectedWords := selectedWordsPerEmail[emailNumber]

		for _, word := range secondFilteredFreqWords {
			featureNumber++
			var emailFeature uint8 = 0
			if emailSelectedWords[word] {
				emailFeature = 1
				numberOfTopRankingSelectedWords++
			} else {
				emailFeature = 0
			}

			toBeShuffled[emailNumber][featureNumber] = emailFeature
		}

		for i := numberOfTopRankingSelectedWords; i < numberOfNonZeroPrimaryFeaturesPerEmailUpperBound; i++ {
			currentNumberOfSecondaryFeatures++
			toBeShuffled[emailNumber][numberOfPrimaryFeatures-1+currentNumberOfSecondaryFeatures] = 1
		}
	}

	if currentNumberOfSecondaryFeatures > numberOfSecondaryFeaturesUpperBound {
		panic("Not finished successfully")
	} else {
		finalNumberOfFeatures := numberOfPrimaryFeatures + currentNumberOfSecondaryFeatures
		fmt.Println("Number of features per email:", finalNumberOfFeatures)
		fmt.Println("Number of primary features per email:", numberOfPrimaryFeatures)
		fmt.Println("Number of secondary features per email:", currentNumberOfSecondaryFeatures)
		fmt.Println("Average number of secondary non zero features per email:", float64(currentNumberOfSecondaryFeatures)/float64(numberOfEmails))

		for emailNumber := 0; emailNumber < numberOfEmails; emailNumber++ {
			toBeShuffled[emailNumber] = toBeShuffled[emailNumber][:finalNumberOfFeatures]
		}
	}

	return toBeShuffled
}

func combineFeaturesWithEmailDirectoryNumber(features [][]uint8, emailDirectoryNumbers []int) [][]uint8 {
	numberOfEmails := len(emailDirectoryNumbers)
	numberOfFeaturesPerEmail := len(features[0])

	combined := make([][]uint8, numberOfEmails)

	for emailNumber := 0; emailNumber < numberOfEmails; emailNumber++ {
		combined[emailNumber] = make([]uint8, 0, numberOfFeaturesPerEmail+1)
		combined[emailNumber] = append(combined[emailNumber], uint8(emailDirectoryNumbers[emailNumber]))

		for i := 0; i < numberOfFeaturesPerEmail; i++ {
			combined[emailNumber] = append(combined[emailNumber], features[emailNumber][i])
		}
	}

	return combined
}

func scrambleTheSortingOfEmailsAndWriteToFile(numberOfEmails int, toBeShuffled [][]uint8, outputFilePath string) {
	randomGenerator := rand.New(rand.NewSource(5665343934110297328))
	randomGenerator.Shuffle(len(toBeShuffled), func(i, j int) {
		toBeShuffled[i], toBeShuffled[j] = toBeShuffled[j], toBeShuffled[i]
	})

	shuffled := toBeShuffled

	csv := strings.Builder{}

	for i := 0; i < numberOfEmails; i++ {
		csv.WriteString(strconv.Itoa(int(shuffled[i][0])))
		for j := 1; j < len(shuffled[i]); j++ {
			csv.WriteString(",")
			csv.WriteString(strconv.Itoa(int(shuffled[i][j])))
		}
		csv.WriteString("\r\n")
	}

	bytesCsv, err := ioutil.ReadAll(strings.NewReader(csv.String()))
	if err != nil {
		panic("Not finished successfully.")
	}
	ioutil.WriteFile(outputFilePath, bytesCsv, 600)
}

func computeKnnClassificationAccuracy(shuffled [][]uint8, k int) {
	fmt.Println("Computing KNN classification accuracy")

	numberOfEmails := len(shuffled)
	numberOfFeatures := len(shuffled[0]) - 1
	numberOfCorrects := 0
	numberOfCorrectsPerDirectoryNumber := make(map[uint8]int)
	confusionMatrix := make(map[uint8]map[uint8]int)

	var numberOfDirectories uint8 = 0
	for emailNumber := 0; emailNumber < numberOfEmails; emailNumber++ {
		if shuffled[emailNumber][0]+1 > numberOfDirectories {
			numberOfDirectories = shuffled[emailNumber][0] + 1
		}
	}

	var directoryNumber1, directoryNumber2 uint8
	for directoryNumber1 = 0; directoryNumber1 < numberOfDirectories; directoryNumber1++ {
		confusionMatrix[directoryNumber1] = make(map[uint8]int)
	}

	for emailNumber := 0; emailNumber < numberOfEmails; emailNumber++ {
		if emailNumber%100 == 0 {
			fmt.Println("Please wait...", emailNumber, "/", numberOfEmails)
		}
		emailFeatures := shuffled[emailNumber][1:]
		if shuffled[emailNumber][0]+1 > numberOfDirectories {
			numberOfDirectories = shuffled[emailNumber][0] + 1
		}

		priorityQueue := pq.NewWith(func(a interface{}, b interface{}) int {
			FeaturesA := a.([]uint8)[1:]
			FeaturesB := b.([]uint8)[1:]

			var cosineDistanceA float64 = 0
			var cosineDistanceB float64 = 0

			var divideBy float64 = 0
			var divideByA float64 = 0
			var divideByB float64 = 0

			for i := 0; i < numberOfFeatures; i++ {
				cosineDistanceA += float64(emailFeatures[i] * FeaturesA[i])
				cosineDistanceB += float64(emailFeatures[i] * FeaturesB[i])

				divideBy += float64(emailFeatures[i])
				divideByA += float64(FeaturesA[i])
				divideByB += float64(FeaturesB[i])
			}

			cosineDistanceA = 1.0 - cosineDistanceA/(math.Sqrt(divideBy)*math.Sqrt(divideByA))
			cosineDistanceB = 1.0 - cosineDistanceB/(math.Sqrt(divideBy)*math.Sqrt(divideByB))

			return utils.Float64Comparator(cosineDistanceA, cosineDistanceB)
		})

		for i := 0; i < numberOfEmails; i++ {
			if i == emailNumber {
				continue
			}
			priorityQueue.Enqueue(shuffled[i])
		}

		neighboursDirectoryNumberOccurrences := make(map[uint8]int)

		for i := 0; i < k; i++ {
			directoryNumberAndFeatures, _ := priorityQueue.Dequeue()
			directoryNumber := directoryNumberAndFeatures.([]uint8)[0]
			neighboursDirectoryNumberOccurrences[directoryNumber]++
		}

		var maxOccurrenceDirectoryNumber uint8 = 0
		maxOccurrences := -1
		for directoryNumber, occurrences := range neighboursDirectoryNumberOccurrences {
			if occurrences > maxOccurrences {
				maxOccurrenceDirectoryNumber = directoryNumber
				maxOccurrences = occurrences
			}
		}

		if maxOccurrences == -1 {
			panic("Not finished successfully.")
		}

		confusionMatrix[shuffled[emailNumber][0]][maxOccurrenceDirectoryNumber]++

		if maxOccurrenceDirectoryNumber == shuffled[emailNumber][0] {
			numberOfCorrects++
			numberOfCorrectsPerDirectoryNumber[shuffled[emailNumber][0]]++
		}
	}

	var accuracy float64 = 100.0 * float64(numberOfCorrects) / float64(numberOfEmails)
	fmt.Println("Accuracy:", accuracy, "%")
	fmt.Println("Confusion matrix:")
	for directoryNumber1 = 0; directoryNumber1 < numberOfDirectories; directoryNumber1++ {
		for directoryNumber2 = 0; directoryNumber2 < numberOfDirectories; directoryNumber2++ {
			var confusion float64 = float64(confusionMatrix[directoryNumber1][directoryNumber2]) / float64(numberOfEmails)
			fmt.Print(strconv.FormatFloat(confusion, 'f', 4, 64), " , ")
		}
		fmt.Println()
	}
}

func Run(outputDirectory string, prefixOfInputDownloadURL string, inputDownloadUrls string, parameters []string) {
	dataSetsPreparationInformation := new(helpers.DataSetPreparationInformation)

	dataSetsPreparationInformation.PrefixOfInputDownloadURLs = prefixOfInputDownloadURL
	dataSetsPreparationInformation.InputDownloadURLs = []string{inputDownloadUrls}
	dataSetsPreparationInformation.Parameters = parameters
	dataSetsPreparationInformation.Preparation = EmailFeaturesPreparation{}

	dataSetsPreparationInformation.BasePreparation(outputDirectory)
}
