package main

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
)

var verbose bool
var program *Program
var tokenDelimiters, punctuation, comma, colon, questionMark string = " \n\t", ".;!?", ",", ":", "?"
var act, scene, and string = "Act", "Scene", "and"
var enter, exit, exeunt, closebrace string = "[Enter", "[Exit", "[Exeunt", "]"
var outputs, inputs, parameters map[string]int
var gotos, conditions, pushes, pops map[string]int
var NOUNS, ADJECTIVES, PERSON_NOUNS, PERSON_ADJECTIVES map[string]int
var operations map[string](func([]int) int)

func LoadKeywords() {
	outputs = LoadMapFromFile("keywords/outputs.kws")
	inputs = LoadMapFromFile("keywords/inputs.kws")
	parameters = LoadMapFromFile("keywords/operations.kws")
	operations = map[string](func([]int) int){
		"sum":        Sum,
		"difference": Difference,
		"product":    Product,
		"division":   Quotient,
		"quotient":   Quotient,
		"square":     Square,
		"cube":       Cube,
		"root":       SquareRoot,
		"twice":      Double,
	}
	gotos = map[string]int{
		"Let": 1,
	}
	conditions = map[string]int{
		"If": 1,
	}
	pushes = map[string]int{
		"Remember": 1,
	}
	pops = map[string]int{
		"Recall": 1,
	}
	NOUNS = LoadMapFromFile("keywords/nouns.kws")
	ADJECTIVES = LoadMapFromFile("keywords/adjectives.kws")
	PERSON_NOUNS = LoadMapFromFile("keywords/personalnouns.kws")
	PERSON_ADJECTIVES = LoadMapFromFile("keywords/personaladjectives.kws")
	//fmt.Println(outputs)
}

func LoadMapFromFile(filename string) (mapping map[string]int) {
	mapping = map[string]int{}
	file, _ := os.Open(filename)
	reader := bufio.NewScanner(file)
	reader.Split(bufio.ScanWords)
	for reader.Scan() {
		key := reader.Text()
		if reader.Scan() {
			value, _ := strconv.ParseInt(reader.Text(), 10, 64)
			mapping[key] = int(value)
		}
	}
	return
}

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
	Description      string
	Output           string
	DramatisPersonae map[string]*Character
	Acts             []*Act
}

func (p *Program) Parse(words []string) (msg string, newWords []string) {
	p.Description, words = ParseTo(words, punctuation)
	p.DramatisPersonae = map[string]*Character{}
	if verbose {
		fmt.Println(p.Description)
	}
	for words[0] != "Act" {
		if verbose {
			fmt.Println("Character:", words[0])
		}
		c := new(Character)
		_, words = c.Parse(words)
		p.DramatisPersonae[c.Name] = c
	}
	for words[0] == act {
		p.Acts = append(p.Acts, new(Act))
		_, words = p.Acts[len(p.Acts)-1].Parse(words)
	}
	msg, newWords = "", words
	return
}
func (p *Program) Eval(environ map[string]int) map[string]int {
	if verbose {
		fmt.Println("Evaluating", p.Description)
	}
	environ["Act"] = 1
	counter := 1
	for _, c := range p.DramatisPersonae {
		c.Stack = append(c.Stack, 0)
		environ[c.Name] = counter
		counter++
	}
	for environ["Act"]-1 < len(p.Acts) {
		environ = p.Acts[environ["Act"]-1].Eval(environ)
		environ["Act"]++
		if environ["Goto"] != 0 {
			delete(environ, "Goto")
		}
	}
	return environ
}

type Character struct {
	Name        string
	Description string
	Stack       []int
	OnStage     bool
}

func (c *Character) Parse(words []string) (msg string, newWords []string) {
	c.Name, words = ParseTo(words, comma)
	c.Name = c.Name[:len(c.Name)-1]
	c.Description, words = ParseTo(words, punctuation)
	msg = c.Name + " " + c.Description
	newWords = words
	return
}
func (c *Character) Eval(environ map[string]int) map[string]int { return environ }
func (c *Character) Value() int {
	return c.Stack[len(c.Stack)-1]
}

type Act struct {
	Number      *Roman
	Description string
	Scenes      []*Scene
}

func (a *Act) Parse(words []string) (msg string, newWords []string) {
	a.Number = new(Roman)
	_, words = a.Number.Parse(words)
	if verbose {
		fmt.Println("Parsing Act", a.Number.Content)
	}
	a.Description, words = ParseTo(words[1:], punctuation)
	for words[0] == scene {
		a.Scenes = append(a.Scenes, new(Scene))
		_, words = a.Scenes[len(a.Scenes)-1].Parse(words)
	}
	msg = ""
	newWords = words
	return
}
func (a *Act) Eval(environ map[string]int) map[string]int {
	if verbose {
		fmt.Println("Evaluating Act", a.Number.Content+":", a.Description)
	}
	environ["Scene"] = 1
	act := environ["Act"]
	for environ["Act"] == act && environ["Scene"]-1 < len(a.Scenes) && environ["Goto"] < 1 {
		environ = a.Scenes[environ["Scene"]-1].Eval(environ)
		environ["Scene"]++
	}
	return environ
}

type Scene struct {
	Number      *Roman
	Description string
	Lines       []*Line
}

func (s *Scene) Parse(words []string) (msg string, newWords []string) {
	s.Number = new(Roman)
	_, words = s.Number.Parse(words)
	if verbose {
		fmt.Println("Parsing Scene", s.Number.Content)
	}
	s.Description, words = ParseTo(words, punctuation)
	for words[0] != act && words[0] != scene && len(words) > 1 {
		s.Lines = append(s.Lines, new(Line))
		_, words = s.Lines[len(s.Lines)-1].Parse(words)
	}
	msg, newWords = "", words
	return
}
func (s *Scene) Eval(environ map[string]int) map[string]int {
	if verbose {
		fmt.Println("Evaluating Scene", s.Number.Content+":", s.Description)
	}
	act, scene := environ["Act"], environ["Scene"]
	line := 0
	for environ["Act"] == act && environ["Scene"] == scene && environ["Goto"] == 0 && line < len(s.Lines) {
		environ = s.Lines[line].Eval(environ)
		line++
	}
	return environ
}

type Line struct {
	Name      string
	Sentences []Interpreter
}

func (l *Line) Parse(words []string) (msg string, newWords []string) {
	var end int
	msg = ""
	l.Name = strings.Trim(words[0], colon)
	switch words[0] {
	case enter:
		fallthrough
	case exit:
		fallthrough
	case exeunt:
		s := new(Sentence)
		for end = 0; words[end] != closebrace; end++ {
			if words[end] != and {
				s.Words = append(s.Words, words[end])
			}
		}
		end++
		l.Sentences = append(l.Sentences, s)
		if verbose {
			fmt.Printf("%s\n", strings.Join(words[:end-1], " ")+words[end-1])
		}
		words = words[end:]
	default:
		if verbose {
			fmt.Println("Parsing line: " + words[0])
		}
		words, end = words[2:], 0
		for (words[1] != colon && words[0][0] != '[') && 1 < len(words) {
			var s Interpreter
			var w string
			w0, w2 := strings.ToLower(words[0]), strings.ToLower(words[2])
			if PERSON_NOUNS[w0] != 0 {
				temp := new(Assignment)
				temp.Name = l.Name
				s = temp
			} else if inputs[w0] == inputs[w2] && inputs[w0] != 0 {
				temp := new(Input)
				temp.Name = l.Name
				s = temp
			} else if outputs[w0] == outputs[w2] && outputs[w0] != 0 {
				temp := new(Output)
				temp.Name = l.Name
				s = temp
			} else if gotos[w0] != 0 {
				temp := new(Goto)
				temp.Name = l.Name
				s = temp
			} else if conditions[w0] != 0 {
				temp := new(Conditional)
				temp.Name = l.Name
				s = temp
			} else if pushes[w0] != 0 {
				temp := new(Push)
				temp.Name = l.Name
				s = temp
			} else if pops[w0] != 0 {
				temp := new(Pop)
				temp.Name = l.Name
				s = temp
			} else {
				temp := new(Sentence)
				temp.Name = l.Name
				s = temp
			}
			w, words = s.Parse(words)
			msg += w + "\n"
			if se, ok := s.(*Sentence); ok {
				if se.Terminator == questionMark {
					q := new(Query)
					q.Name = l.Name
					q.Words, q.Terminator = se.Words, se.Terminator
					s = q
				}
			}

			l.Sentences = append(l.Sentences, s)
			end = 0
			if verbose {
				switch t := s.(type) {
				default:
					fmt.Printf("%T: %s\n", t, w)
				}
			}
		}
		if end+1 >= len(words) {
			// This is an error.  All scenes should end in a line stating "[Exeunt]"
		}
	}
	newWords = words
	return
}
func (l *Line) Eval(environ map[string]int) map[string]int {
	dp := program.DramatisPersonae
	switch l.Name {
	case enter:
		fallthrough
	case exit:
		fallthrough
	case exeunt:
		s, count := l.Sentences[0].(*Sentence), 0
		for _, word := range s.Words {
			if character, ok := dp[word]; ok {
				count++
				if l.Name == enter {
					character.OnStage = true
					if verbose {
						fmt.Println(word, "enters the stage")
					}
				} else {
					character.OnStage = false
					if verbose {
						fmt.Println(word, "exits the stage")
					}
				}
			}
		}
	default:
		if _, ok := environ["Target"]; ok {
			delete(environ, "Target")
		}
		for name, c := range program.DramatisPersonae {
			if name != l.Name && c.OnStage {
				environ["Target"] = environ[name]
			}
		}
		for _, s := range l.Sentences {
			environ = s.Eval(environ)
			if environ["Goto"] != 0 {
				return environ
			}
		}
	}
	return environ
}

type Sentence struct {
	Name       string
	Words      []string
	Terminator string
}

func (s *Sentence) Parse(words []string) (msg string, newWords []string) {
	var end int
	for end = 0; !strings.Contains(punctuation, words[end]); end++ {
	}
	s.Words = words[:end]
	s.Terminator = words[end]
	newWords = words[end+1:]
	msg = strings.Join(s.Words, " ") + s.Terminator
	return
}
func (s *Sentence) Eval(environ map[string]int) map[string]int { return environ }

type Assignment struct {
	Sentence
}

func (a *Assignment) Eval(environ map[string]int) map[string]int {
	var target *Character
	for name, c := range program.DramatisPersonae {
		if environ[name] == environ["Target"] {
			target = c
		}
	}
	start := 0
	if verbose {
		fmt.Print("Evaluating ", strings.Join(a.Words, " "), "...")
	}
	value := ParseMath(a.Words, &start, target.Stack[len(target.Stack)-1])
	if verbose {
		fmt.Println(value)
	}

	target.Stack[len(target.Stack)-1] = value
	return environ
}

type Output struct {
	Sentence
}

func (o *Output) Eval(environ map[string]int) map[string]int {
	var target *Character
	for name, c := range program.DramatisPersonae {
		if environ[name] == environ["Target"] {
			target = c
		}
	}
	value := target.Stack[len(target.Stack)-1]
	if outputs[strings.ToLower(o.Words[0])] == 1 {
		if verbose {
			fmt.Println(target.Name+"(asString):", string(byte(value)))
		}
		program.Output += string(byte(value))
	} else {
		if verbose {
			fmt.Println(target.Name+":", value)
		}
		program.Output += string(value)
	}
	return environ
}

type Input struct {
	Sentence
}

func (i *Input) Eval(environ map[string]int) map[string]int { return environ }

type Goto struct {
	Sentence
	ToAct  bool
	Number Roman
}

func (g *Goto) Eval(environ map[string]int) map[string]int { return environ }

type Query struct {
	Sentence
}

func (q *Query) Eval(environ map[string]int) map[string]int { return environ }

type Conditional struct {
	Sentence
	Question Query
}

func (c *Conditional) Eval(environ map[string]int) map[string]int { return environ }

type Push struct {
	Sentence
}

func (p *Push) Eval(environ map[string]int) map[string]int { return environ }

type Pop struct {
	Sentence
}

func (p *Pop) Eval(environ map[string]int) map[string]int { return environ }

type Roman struct {
	Content string
	Value   int
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
	return
}
func (r *Roman) Eval(environ map[string]int) map[string]int { return environ }

func ParseMath(words []string, start *int, startVal int) int {
	// Parse the array for values and math.
	value := 1
	for ; *start < len(words); *start++ {
		i := *start
		word := words[i]
		w := strings.ToLower(word)
		if _, ok := ADJECTIVES[w]; ok {
			value *= 2
		}
		if val, ok := NOUNS[w]; ok {
			value *= val
			*start++
			return value
		}
		if val, ok := program.DramatisPersonae[word]; ok {
			value *= val.Stack[len(val.Stack)-1]
			*start++
			return value
		}
		// find math
		w2 := w
		if i < len(words)-1 {
			w2 = strings.ToLower(words[i+1])
		}
		if val, ok := parameters[w]; ok {
			if _, ok2 := parameters[w2]; ok2 {
			} else {
				*start++
				var params []int
				for i := val; i > 0; i-- {
					params = append(params, ParseMath(words, start, startVal))
				}
				return operations[w](params)
			}
		}
	}
	return startVal
}

func Sum(x []int) int {
	return x[0] + x[1]
}

func Difference(x []int) int {
	return x[0] - x[1]
}

func Product(x []int) int {
	return x[0] * x[1]
}

func Quotient(x []int) int {
	return x[0] / x[1]
}

func Square(x []int) int {
	return x[0] * x[0]
}

func Cube(x []int) int {
	return x[0] * x[0] * x[0]
}

func SquareRoot(x []int) int {
	return int(math.Sqrt(float64(x[0])))
}

func Double(x []int) int {
	return x[0] * 2
}

func main() {
	// Setup
	var filename string
	for _, a := range os.Args[1:] {
		if a == "-v" || a == "-verbose" {
			verbose = true
		} else {
			filename = a
		}
	}
	if len(filename) == 0 {
		fmt.Println("Must provide a file!")
		return
	}

	LoadKeywords()
	var tokens bufio.Scanner
	file, err := os.Open(filename)
	if err != nil {
		fmt.Println(err)
		return
	}
	tokens = *bufio.NewScanner(file)
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
	if verbose {
		fmt.Printf("Tokens:\n%q\n", words)
	}
	program = new(Program)
	program.Parse(words)
	environ := map[string]int{}
	program.Eval(environ)
	if verbose {
		fmt.Println("Program output:")
	}
	fmt.Print(program.Output)
	return
}
