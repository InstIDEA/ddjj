package extract

import (
	"bufio"
	"regexp"
	"strconv"
	"strings"
	"fmt"
)

type Extractor struct {
	Scanner *bufio.Scanner
	RawData string

	PrevToken string
	CurrToken string
	NextToken string

	CurrLine int
	SavedLine int
	
	Buffer []string
	BufferHookFunc func(tokens []string) []string

	Flags ExtractorFlag
}

type TokenType int

const (
	PrevToken = iota
	CurrToken
	NextToken
	MaxTokens
)

type ExtractorFlag int

const (
	// the tokens skip blank lines
	EXTRACTOR_FLAG_1 = 1<<(iota + 1)
	
	// remove begin and end spaces from tokens
	EXTRACTOR_FLAG_2

	// line tokenizer
	EXTRACTOR_FLAG_3
)

func NewExtractor(raw string) *Extractor {
	return &Extractor{
		RawData: raw,
		Scanner: bufio.NewScanner(strings.NewReader(raw)),
		BufferHookFunc: func(tokens []string) []string {
			return tokens
		},
	}
}

func (e *Extractor) Scan() bool {

	scan := func(s *bufio.Scanner) (string, bool) {
		text := ""

		if e.Flags & EXTRACTOR_FLAG_3 != 0 &&
		len(e.Buffer) > 1 {
			text = e.Buffer[1]
			e.Buffer = e.Buffer[1:]
			return text, true
		}

		for s.Scan() {
			text = s.Text()
			if e.Flags & EXTRACTOR_FLAG_1 != 0 {
				if text == "" {
					continue
				}
			}

			if e.Flags & EXTRACTOR_FLAG_2 != 0 {
				text = strings.TrimSpace(text)
			}

			if e.Flags & EXTRACTOR_FLAG_3 != 0 &&
			text != "" {
				e.Buffer = e.BufferHookFunc(tokenize(text, 3))
				text = e.Buffer[0]
			}

			return text, true
		}

		return "", false
	}

	e.PrevToken = e.CurrToken
	e.CurrToken = e.NextToken
	val, status := scan(e.Scanner)
	e.NextToken = val

	// EOF
	if !status && 
	e.CurrToken == "" && 
	e.Scanner.Err() == nil {
		return false
	}

	e.CurrLine++
	return true
}

func (e *Extractor) MoveUntilContains(t TokenType, s string) bool {
	tokens := [MaxTokens]*string { &e.PrevToken, &e.CurrToken, &e.NextToken }
	for e.Scan() {
		if strings.Contains(*tokens[t], s) {
			return true
		}
	}
	return false
}

func (e *Extractor) MoveUntilStartWith(t TokenType, s string) bool {
	tokens := [MaxTokens]*string { &e.PrevToken, &e.CurrToken, &e.NextToken }
	for e.Scan() {
		if isCurrLine(*tokens[t], s) {
			return true
		}
	}
	return false
}

func (e *Extractor) MoveUntilSavedLine() {
	e.Rewind()
	for e.Scan() {
		if e.CurrLine == e.SavedLine {
			break	
		}
	}
}

func (e *Extractor) Rewind() {
	e.Scanner = bufio.NewScanner(strings.NewReader(e.RawData))
	e.Buffer = []string{}
	e.CurrLine = 0
	e.PrevToken = ""
	e.CurrToken = ""
	e.NextToken = ""
}

func (e *Extractor) PrevLineNum() int {
	value := e.CurrLine - 1
	if value < 0 {
		return 0
	}
	return value
}

func (e *Extractor) CurrLineNum() int {
	return e.CurrLine
}

func (e *Extractor) NextLineNum() int {
	return e.CurrLine + 1
}

func (e *Extractor) SaveLine() {
	e.SavedLine = e.CurrLine
}

func (e *Extractor) BindFlag(flag ExtractorFlag) {
	e.Flags |= flag
}

func (e *Extractor) UnbindFlag(flag ExtractorFlag) {
	e.Flags &= flag
}

func (e *Extractor) UnbindAllFlags() {
	e.Flags = 0
}

func ContainsItem(s []string, e string) bool {
	for _, a := range s {
		if strings.Contains(e, a) {
			return true
		}
	}
	return false
}

func ContainsIntItem(s []int, e int) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func StringToInt64(line string) int64 {
	r := strings.NewReplacer(".", "", ",", "")
	i, _ := strconv.ParseInt(r.Replace(line), 10, 64)

	return i
}

func stringToInt(line string) int {
	r := strings.NewReplacer(".", "", ",", "")
	i, _ := strconv.Atoi(r.Replace(line))

	return i
}

func isCurrLine(line string, startwith string) bool {
	pattern := fmt.Sprintf(`^(%s).*$`, startwith)
	matched, _ := regexp.MatchString(pattern, line)
	return matched
}

func endsWith(line string, s string) bool {
	pattern := fmt.Sprintf(`.*(%s)$`, s)
	matched, _ := regexp.MatchString(pattern, line)
	return matched
}

func hasTrailingSpaces(line string, s string) bool {
	pattern := fmt.Sprintf(`(%s)\s\s*`, s)
	matched, _ := regexp.MatchString(pattern, line)
	return matched
} 

func hasLeadingSpaces(line string, s string) bool {
	pattern := fmt.Sprintf(`\s\s*(%s)`, s)
	matched, _ := regexp.MatchString(pattern, line)
	return matched
}

func isDate(line string) bool {
	matched, _ := regexp.MatchString(`[0-9]{2}/[0-9]{2}/[0-9]{4}`, line)
	return matched
}

func isTimestamp(line string) bool {
	matched, _ := regexp.MatchString(`[0-9]{2}/[0-9]{2}/[0-9]{4}\s*[0-9]{2}:[0-9]{2}:[0-9]{2}`, line)
	return matched
}

func isAlpha(line string) bool {
	matched, _ := regexp.MatchString(`[aA-zZ].*$`, line)
	return matched
}

func isAlphaNum(line string) bool {
	matched, _ := regexp.MatchString(`[aA-zZ0-9].*$`, line)
	return matched
}

func isNumber(line string) bool {
	matched, _ := regexp.MatchString(`^[\+\-]*[0-9.,]*[0-9]$`, line)
	return matched
}

func isBarCode(line string) bool {
	matched, _ := regexp.MatchString(`[0-9]{5,6}-[0-9]{5,7}-[0-9]{1,3}`, line)
	return matched
}

func isKeyValuePair(key string, precedence string) (string, bool) {
	r := strings.NewReplacer(":", "")
	inline := strings.Split(r.Replace(key), precedence)

	if len(inline) > 1 {
		value := strings.TrimSpace(inline[len(inline) -1])
		if value != "" {
			return value, true
		}
	}
	return key, false
}

// call after isNumber
func isAddressStreet(s string) bool {
	contains := []string { "N°", "CASI", "E/", "CALLE", "C/", "AVDA.", 
	"AV.", "RUTA", "KM", "ENTRE", "ESQ.", "PISO", "BLOQUE", "PLANTA" }

	for _, value := range contains {
		if strings.Contains(s, value) {
			return true
		}
	}

	matched, _ := regexp.MatchString(`[0-9]{3,4}`, s)
	if matched {
		return true
	}

	return false
}

// call after isNumber
func isPhoneNumber(s string) bool {
	matched, _ := regexp.MatchString(`(\()[0-9].*(\))|[0-9\s]*[0-9\-\/\.]`, s)

	if matched {
		return true
	}
	return false
}

func removeAccents(s string) string {
	r := strings.NewReplacer("Á", "A", "É", "E", "Í", "I", "Ó", "O", "Ú", "U" )
	return r.Replace(s)
}

func removeSpaces(s string) string {
	return strings.ReplaceAll(s, " ", "")
}

// split a line into words that not exceed the max continuous spaces
func tokenize(line string, max int) []string {
	var tokens []string
	var buffer strings.Builder
	var spaces int

	line = strings.TrimSpace(line)
	for _, letter := range line {
		if letter == ' ' {
			spaces++
			buffer.WriteRune(letter)
			continue
		}

		if spaces >= max {
			token := strings.TrimSpace(buffer.String())
			if token != "" {	
				tokens = append(tokens, token)
			}
			buffer.Reset()
		}
		spaces = 0
		buffer.WriteRune(letter)
	}
	tokens = append(tokens, strings.TrimSpace(buffer.String()))
	return tokens
}

// join two or more array string if one of them contains the sep string
func combine(tokens []string, sep string) []string {
	var result []string
	length := len(tokens)

	for i := 0; i < length; i++ {
		if i + 1 < length &&
		strings.Contains(tokens[i], sep) &&
		!strings.Contains(tokens[i+1], sep) {
			result = append(result, tokens[i] + tokens[i+1])
			i += 1
			continue
		}
		result = append(result, tokens[i])
	}
	return result
}

func getTime(data string) string {
	re := regexp.MustCompile(`[0-9]{2}`)
	values := re.FindAllString(data, -1)
	length := len(values)

	if length < 3 {
		return ""
	}

	return fmt.Sprintf("%s:%s:%s", values[length -3], values[length -2], values[length -1])
}

func getDate(data string) string {
	re := regexp.MustCompile(`[0-9]{2}.[0-9]{2}.[0-9]{4}`)
	result := re.FindString(data)

	if result == "" {
		return ""
	}

	re = regexp.MustCompile(`[0-9]{2,4}`)
	values := re.FindAllString(result, -1)

	if len(values) < 3 {
		return ""
	}

	return fmt.Sprintf("%s/%s/%s", values[0], values[1], values[2])
}

/*
legacy code support
don't use these functions
use extractor struct and methods instead

the extractions that using these functions will be reviewed
*/

// MoveUntil finds a word and stops the scan there.
func MoveUntil(scanner *bufio.Scanner, search string, exact bool) *bufio.Scanner {
	for scanner.Scan() {
		line := scanner.Text()

		if exact {
			if line == search {
				break
			}
		} else {
			if strings.Contains(line, search) {
				break
			}
		}

	}

	return scanner
}

func getTotalInCategory(scanner *bufio.Scanner) int64 {
	scanner.Scan()
	scanner.Scan()
	r := strings.NewReplacer(".", "", ",", "")
	i, _ := strconv.ParseInt(r.Replace(scanner.Text()), 10, 64)

	return i
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func stringToInt64(line string) int64 {
	return StringToInt64(line)
}

func stringToYear(line string) int {
	year := stringToInt(line)

	if year == 0 {
		return 0
	}

	if year < 100 {
		return 2000 + year
	}

	return year
}
