package coverage

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
)

type reportLine struct {
	file          string
	lineFrom      int
	lineTo        int
	numStatements int
	cntStatements int
}

type fileCoverage struct {
	linesCoverage map[string]reportLine
	lines         map[int]int
}

type codacyCoverageJSON struct {
	Total       int                      `json:"total"`
	FileReports []codacyFileCoverageJSON `json:"fileReports"`
}

type codacyFileCoverageJSON struct {
	Filename string      `json:"filename"`
	Total    int         `json:"total"`
	Coverage map[int]int `json:"coverage"`
}

const (
	ModeSet = "set"
)

var regex *regexp.Regexp
var regexpStringFilename = `([a-zA-Z\/\._\-\d]*)`
var regexpStringStat = `(\d+)`
var regexpStringMode = `mode: ([set|count|atomic]*)`
var regexpString = fmt.Sprintf(`%s:%s.*?,%s.* %s %s`, regexpStringFilename, regexpStringStat, regexpStringStat, regexpStringStat, regexpStringStat)

// GenerateCoverageJSON generates a json string containing
// coverage information in codacy's format
func GenerateCoverageJSON(coverageFile string) ([]byte, error) {
	regex, _ = regexp.Compile(regexpString)

	fileReader, err := os.Open(coverageFile)

	//defer func() { _ = fileReader.Close() }()
	if err != nil {
		return nil, err
	}

	reader := bufio.NewReader(fileReader)

	files := make(map[string]*fileCoverage)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}

		if isSkippableLine(line) {
			continue
		}

		parsed, err := parseLine(line)
		if err != nil {
			return nil, err
		}

		file := files[parsed.file]
		if file == nil {
			file = new(fileCoverage)
			files[parsed.file] = file
			file.linesCoverage = make(map[string]reportLine)
			file.lines = make(map[int]int)
		}

		if line, ok := file.linesCoverage[fmt.Sprintf("%d-%d", parsed.lineFrom, parsed.lineTo)]; ok {
			if line.cntStatements == 0 {
				file.linesCoverage[fmt.Sprintf("%d-%d", parsed.lineFrom, parsed.lineTo)] = parsed
			}
		} else {
			file.linesCoverage[fmt.Sprintf("%d-%d", parsed.lineFrom, parsed.lineTo)] = parsed
		}

		for i := parsed.lineFrom; i <= parsed.lineTo; i++ {
			file.lines[i] += parsed.cntStatements
		}
	}

	total, perFile := calculatePercentages(files)

	covJSON := codacyCoverageJSON{}
	covJSON.Total = total
	covJSON.FileReports = make([]codacyFileCoverageJSON, 0)

	for filename, fileCoverage := range perFile {
		fileCov := codacyFileCoverageJSON{}
		fileCov.Filename = filename
		fileCov.Total = fileCoverage
		fileCov.Coverage = files[filename].lines
		covJSON.FileReports = append(covJSON.FileReports, fileCov)
	}

	json, err := json.Marshal(covJSON)
	if err != nil {
		return nil, err
	}

	return json, nil
}

func parseLine(line string) (reportLine, error) {
	result := regex.FindStringSubmatch(line)
	if len(result) >= 6 {
		lineFrom, err := strconv.Atoi(result[2])
		if err != nil {
			return reportLine{}, err
		}

		lineTo, err := strconv.Atoi(result[3])
		if err != nil {
			return reportLine{}, err
		}

		numStatements, err := strconv.Atoi(result[4])
		if err != nil {
			return reportLine{}, err
		}
		cntStatements, err := strconv.Atoi(result[5])
		if err != nil {
			return reportLine{}, err
		}

		return reportLine{result[1], lineFrom, lineTo, numStatements, cntStatements}, nil
	}

	return reportLine{}, errors.New("invalid line format")
}

func isSkippableLine(line string) bool {
	return (len(strings.TrimSpace(line)) == 0) || strings.HasPrefix(line, "mode")
}

func calculatePercentages(files map[string]*fileCoverage) (int, map[string]int) {
	totalNumStatements := 0
	totalCoveredStatements := 0
	percentages := make(map[string]int)

	for file, coverage := range files {
		fileNumStatements := 0
		fileCoveredStatements := 0
		for _, line := range coverage.linesCoverage {
			fileNumStatements += line.numStatements
			if line.cntStatements > 0 {
				fileCoveredStatements += line.numStatements
			}
		}
		percentages[file] = calculatePercentage(fileNumStatements, fileCoveredStatements)
		totalNumStatements += fileNumStatements
		totalCoveredStatements += fileCoveredStatements
	}

	return calculatePercentage(totalNumStatements, totalCoveredStatements), percentages
}

func calculatePercentage(num int, cvd int) int {
	if num == 0 {
		return 0
	}

	return cvd * 100 / num
}

func compileRegexp() (*regexp.Regexp, error) {
	return regexp.Compile(regexpString)
}
