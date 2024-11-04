package driver

import (
	"encoding/json"
	"log"
	"os"
	"regexp"
)

type Step struct {
	Input           string `json:"input"`
	ExpectedContent string `json:"expectedContent"`
}

func (s *Step) MatchesExpectedContent(content []byte) (bool, error) {
	pattern := regexp.QuoteMeta(s.ExpectedContent)
	re, err := regexp.Compile(pattern)
	if err != nil {
		return false, err
	}
	if re.Match([]byte(content)) {
		return true, nil
	}
	return false, nil
}

// Group represents a group of steps
type Group struct {
	Name  string `json:"name"`
	Steps []Step `json:"steps"`
}

type TestCase struct {
	Name            string
	Input           string
	ExpectedContent string
}

func (s *TestCase) MatchesExpectedContent(content []byte) (bool, error) {
	pattern := regexp.QuoteMeta(s.ExpectedContent)
	re, err := regexp.Compile(pattern)
	if err != nil {
		return false, err
	}
	// Check if the content matches the regex pattern
	if re.Match(content) {
		return true, nil
	}
	return false, nil
}

// DataGroup represents the overall structure of the JSON.
type DataGroup struct {
	Groups []Group `json:"groups"`
}

type Session struct {
	Name   string  `json:"name"`
	Groups []Group `json:"groups"`
}

func ReadData() []Session {
	data, err := os.ReadFile("test_setup.json")
	if err != nil {
		log.Fatalf("Failed to read file: %v", err)
	}
	// Unmarshal JSON data
	var sessions []Session
	err = json.Unmarshal(data, &sessions)
	if err != nil {
		log.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	return sessions
}

func FilterGroupsByName(groups []Group, name string) []Group {
	var filteredGroups []Group
	for _, group := range groups {
		if group.Name == name {
			filteredGroups = append(filteredGroups, group)
		}
	}
	return filteredGroups
}

func LoadTestGroups(filePath string) (DataGroup, error) {
	var sessionsData DataGroup
	data, err := os.ReadFile(filePath)
	if err != nil {
		return sessionsData, err
	}
	err = json.Unmarshal(data, &sessionsData)
	return sessionsData, err
}

func CreateTestCases(group DataGroup) []TestCase {
	var tests []TestCase
	for _, group := range group.Groups {
		for _, step := range group.Steps {
			// Create a test case for each group
			tests = append(tests, TestCase{
				Name:            group.Name,
				Input:           step.Input,
				ExpectedContent: step.ExpectedContent,
			})
		}
	}

	return tests
}
