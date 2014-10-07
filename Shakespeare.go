package main

import (
	"fmt"
	"bufio"
	"os"
	"strings"
	"strconv"
)

var tokenDelimiters, punctuation, comma, colon string = " \n\t", ".;!?", ",", ":"
var act, scene, and string = "Act", "Scene", "and"
var  enter, exit, exeunt, closebrace string = "[Enter", "[Exit", "[Exeunt", "]"

// Parsing interface
type Interpreter interface {
	Parse(words []string) (msg string, newWords []string)
	Eval(environ map[string]int) map[string]int
}

func ParseTo(words []string, delim string) (toPeriod string, newWords []string) {
	// Strip out the section of words that we've used.
	start := 0
	for !strings.Contains(delim, words[start]) {
		start++
	}
	toPeriod = strings.Join(words[:start], " ") + words[start]
	newWords = words[start+1:]
	return
}

// Program is composed of Title, DramatisPersonae, and 1 or more Acts
type Program struct {
	Description string
	DramatisPersonae []*Character
	Acts []*Act
}

func (p *Program) Parse(words []string) (msg string, newWords []string) {
	p.Description, words = ParseTo(words, punctuation)
	fmt.Println(p.Description)
	for words[0] != "Act" {
		fmt.Println(words[0])
		p.DramatisPersonae = append(p.DramatisPersonae, new(Character))
		_, words = p.DramatisPersonae[len(p.DramatisPersonae)-1].Parse(words)
	}
	for words[0] == act {
		p.Acts = append(p.Acts, new(Act))
		_, words = p.Acts[len(p.Acts)-1].Parse(words)
	}
	msg, newWords = "", words
	return
}

type Character struct {
	Name string
	Description string
	Stack []int
	OnStage bool
}
func (c *Character) Parse(words []string) (msg string, newWords []string) {
	c.Name, words = ParseTo(words, comma)
	c.Name = c.Name[:len(c.Name)-1]
	c.Description, words = ParseTo(words, punctuation)
	msg = c.Name + " " + c.Description
	newWords = words
	return
}

type Act struct {
	Number *Roman
	Description string
	Scenes []*Scene
}
func (a *Act) Parse(words []string) (msg string, newWords []string) {
	a.Number = new(Roman)
	_, words = a.Number.Parse(words)
	fmt.Println("Parsing Act", a.Number.Content)
	a.Description, words = ParseTo(words[1:], punctuation)
	for words[0] == scene {
		a.Scenes = append(a.Scenes, new(Scene))
		_, words = a.Scenes[len(a.Scenes)-1].Parse(words)
	}
	msg = ""
	newWords = words
	return
}

type Scene struct {
	Number *Roman
	Description string
	Lines []*Line
}
func (s *Scene) Parse(words []string) (msg string, newWords []string) {
	s.Number = new(Roman)
	_, words = s.Number.Parse(words)
	fmt.Println("Parsing Scene", s.Number.Content)
	s.Description, words = ParseTo(words, punctuation)
	for words[0] != act && words[0] != scene && len(words) > 1 {
		s.Lines = append(s.Lines, new(Line))
		_, words = s.Lines[len(s.Lines)-1].Parse(words)
	}
	msg, newWords = "", words
	return 
}

type Line struct {
	Name string
	Sentences []*Sentence
}
func (l *Line) Parse(words []string) (msg string, newWords []string) {
	var end int
	l.Name = strings.Trim(words[0],colon)
	switch words[0] {
	case enter:
		fallthrough
	case exit:
		fallthrough
	case exeunt:
		s := new(Sentence)
		for end=0; words[end] != closebrace; end++ {
			if words[end] != and {
				s.Words = append(s.Words, words[end])
			}
		}
		end++
		l.Sentences = append(l.Sentences, s)
		words = words[end:]
	default:
		fmt.Println("Parsing line: "+words[0])
		words = words[2:]
		for end=0; (words[end+1] != colon && words[end][0] != '[') && end+1 < len(words); {
			s := new(Sentence)
			var w string
			w, words = s.Parse(words)
			l.Sentences = append(l.Sentences, s)
			end = 0
			fmt.Println("Sentence: "+w)
		}
		if end+1 >= len(words) {
			// This is an error.  All scenes should end in a line stating "[Exeunt]"
		}
	}
	msg, newWords = "", words
	return
}

type Sentence struct {
	Words []string
}
func (s *Sentence) Parse(words []string) (msg string, newWords []string) {
	var end int
	for end=0; !strings.Contains(punctuation, words[end]); end++ {}
	s.Words = words[:end]
	newWords = words[end+1:]
	msg = strings.Join(s.Words, " ")
	return
}

type Roman struct {
	Content string
	Value int
}
func (r *Roman) Parse(words []string) (msg string, newWords []string) {
	r.Content = words[1]
	switch r.Content {
	case "I":
		r.Value = 1
	case "II":
		r.Value = 2
	case "III":
		r.Value = 3
	case "IV":
		r.Value = 4
	case "V":
		r.Value = 5		
	case "VI":
		r.Value = 6
	case "VII":
		r.Value = 7
	case "VIII":
		r.Value = 8
	case "IX":
		r.Value = 9
	case "X":
		r.Value = 10
	}
	newWords = words[2:]
	msg = r.Content + ": " + strconv.Itoa(r.Value)
	//fmt.Println(msg)
	return
}

func main() {
	// Setup
	var tokens bufio.Scanner
	tokens = *bufio.NewScanner(os.Stdin)
	tokens.Split(bufio.ScanWords)
	var words []string
	for tokens.Scan() {
		word := tokens.Text()
		if strings.Contains(punctuation+comma+colon+closebrace, word[len(word)-1:]) {
			words = append(words, word[:len(word)-1])
			words = append(words, word[len(word)-1:])
		} else {
			words = append(words, word)
		}
	}
	// Add an extra token to the end.
	words = append(words, string(byte(0)))
	fmt.Printf("%q\n", words)
	p := new(Program)
	p.Parse(words)
	return
}