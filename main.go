package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
)

const API_URL = "https://languagetool.org/api/v2/check"

type RepValue struct {
	Value string `json:"value"`
}

type Match struct {
	ShortMessage string     `json:"shortMessage"`
	Message      string     `json:"message"`
	Replacements []RepValue `json:"replacements"`
	Offset       int        `json:"offset"`
	Length       int        `json:"length"`
}
type Matches []Match

type CheckResult struct {
	Matches Matches `json:"matches"`
}

func NewCheckResult() *CheckResult {
	return &CheckResult{
		Matches: Matches{},
	}
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Please pass argument spellcheck words that you want.")
		os.Exit(1)
	}

	text := strings.Join(os.Args[1:], " ")
	checked, err := doSpellCheck(text)
	if err != nil {
		fmt.Println("[Error]", err)
	} else {
		fmt.Println(format([]rune(text), checked))
	}
}

func format(text []rune, c *CheckResult) string {
	separator := strings.Repeat("=", 40)
	var line []string
	for i, m := range c.Matches {
		line = append(line, "")
		if i > 0 {
			line = append(line, separator)
			line = append(line, "")
		}

		if m.ShortMessage != "" {
			line = append(line, fmt.Sprintf("\033[93m!! %s !!\033[0m", m.ShortMessage))
		} else {
			line = append(line, fmt.Sprintf("\033[93m!! %s !!\033[0m", m.Message))
		}
		line = append(line, fmt.Sprintf("Word: \033[1m\"%s\"\033[0m, at offset: %d", string(text[m.Offset:(m.Offset+m.Length)]), m.Offset))
		reps := []string{}
		for _, r := range m.Replacements {
			reps = append(reps, r.Value)
		}
		line = append(line, fmt.Sprintf("Suggested: \033[92m%s\033[0m", strings.Join(reps, "\033[0m, \033[92m")))
	}
	line = append(line, "")

	return strings.Join(line, "\n")
}

func doSpellCheck(text string) (*CheckResult, error) {
	post := url.Values{}
	post.Add("text", text)
	post.Add("language", "en-US")
	post.Add("enableOnly", "false")

	req, _ := http.NewRequest("POST", API_URL, strings.NewReader(post.Encode()))
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if ret, err := ioutil.ReadAll(resp.Body); err != nil {
		return nil, err
	} else {
		r := NewCheckResult()

		if err := json.Unmarshal(ret, r); err != nil {
			return nil, err
		}

		return r, nil
	}
}
