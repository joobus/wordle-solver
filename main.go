package main

import (
	_ "embed"
	"fmt"
	"log"
	"os/exec"
	"regexp"
)

//go:embed wordlist.txt
var wordlist []byte

// Matcher is the state object which holds information on character matches
type Matcher struct {
	Guesses      []string
	Found        []rune
	Excluded     string
	IndexMatches [5]*IndexMatch
	RegexString  string
}

// IndexMatch is the state object for the letter at a specific index in the word
type IndexMatch struct {
	Char     rune
	Excluded []rune
}

// buildMatcher creates the regex string and list of matched characters used to
// execute against the wordlist
func buildMatcher(gs []string) (*Matcher, error) {
	m := Matcher{
		Guesses: gs,
	}

	// Instantiate pointers so we can skip nil checks in this func
	for i := range m.IndexMatches {
		m.IndexMatches[i] = &IndexMatch{}
	}

	// For each guess
	for _, g := range gs {
		rs := []rune(g) // slice of runes in word/input
		ci := 0         // character index in actual word
		// For each character in guess
		for i := 0; i < len(rs); i++ {
			im := m.IndexMatches
			if rs[i] == '-' { // Right letter, wrong spot
				i++
				im[ci].Excluded = append(im[ci].Excluded, rs[i])
				m.Found = append(m.Found, rs[i])
			} else if rs[i] == '+' { // Right letter, right spot
				i++
				im[ci].Char = rs[i]
			} else {
				// Character is not in word
				m.Excluded += string(rs[i])
			}
			ci++
		}
	}

	rx := "^" // regex string
	// For each character position, generate a regex character set
	for _, v := range m.IndexMatches {
		if v.Char != 0 { // Character already known for this position
			rx += string(v.Char)
		} else {
			// Generate negative regex of characters excluded globally, and at
			// current position
			exs := m.Excluded
			// Iterate through local excludes, characters which are know to be
			// in string, but not at this position
			for _, c := range v.Excluded {
				exs += string(c)
			}
			rx += fmt.Sprintf("[^%s]", exs)
		}
	}
	rx += "$"
	// Make sure generated regex is valid
	_, err := regexp.Compile(rx)
	if err != nil {
		return nil, err
	}

	m.RegexString = rx

	return &m, nil
}

// RunRgRegex calls the ripgrep command with given arguments
func RunRgRegex(rgx string, cmdInput *[]byte) (*[]byte, error) {
	cmd := exec.Command("rg", rgx)

	// cmdInput is what should be sent to stdin
	if cmdInput != nil {
		stdin, err := cmd.StdinPipe()
		if err != nil {
			return nil, err
		}

		_, err = stdin.Write(*cmdInput)
		if err != nil {
			return nil, err
		}

		err = stdin.Close()
		if err != nil {
			return nil, err
		}
	}

	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}

	return &out, nil
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	guesses := []string{}

	count := 6
	for i := 0; i < count; i++ {
		var guess string
		log.Println("Enter guess result:")
		fmt.Scanln(&guess)
		log.Printf("You guessed '%s'", guess)

		guesses = append(guesses, guess)

		matcher, err := buildMatcher(guesses)
		if err != nil {
			log.Fatal(err)
		}

		log.Printf("regex: %s", matcher.RegexString)
		log.Println("Possible matches:")

		// Execute main regex which excludes letters at each position
		out, err := RunRgRegex(matcher.RegexString, &wordlist)
		if err != nil {
			log.Fatal(err)
		}

		// Since there is not a way to do an "AND" regex, after running main
		// regex, run rg again for each character known to be in word but with
		// position unknown
		for _, c := range matcher.Found {
			// For each letter, pipe in wordlist from last rg command to
			// progressively filter word list
			out, err = RunRgRegex(string(c), out)
			if err != nil {
				log.Fatal(err)
			}
		}

		if out != nil {
			fmt.Printf("\n%s\n", *out)
		} else {
			log.Fatal("No command output!")
		}
	}
}
