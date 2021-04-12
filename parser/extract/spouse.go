package extract

import (
	"github.com/pkg/errors"
	"strings"
)

// TODO
// check 'ESTADO CIVIL'

func Spouse(e *Extractor) (string, error) {
	var spouse string
	var firstNameCase int
	var lastNameCase int

	e.BindFlag(EXTRACTOR_FLAG_1)

	if e.MoveUntilContains(CurrToken, "NYUGE") {
		e.SaveLine()
		for e.Scan() {
			if spouse == "" {
				fname, err := getSpouseFirstName(e, firstNameCase)

				if err != nil {
					return "", err
				}

				if fname != "" &&
				len(fname) <= 32 {
					spouse += fname
					continue
				}

				firstNameCase++
				e.MoveUntilSavedLine()
				continue
			}

			lname, err := getSpouseLastName(e, lastNameCase)

			if err != nil {
				// no more cases for last name
				// so return the first name

				// maybe concat the last name 
				// of who declared the ddjj
				return spouse, err
			}

			if lname != "" &&
			len(lname) <= 32 {
				spouse = spouse + " " + lname
				break
			}

			lastNameCase++
			e.MoveUntilSavedLine()
		}
	}

	return spouse, nil
}

func getSpouseFirstName(e *Extractor, index int) (string, error) {
	switch index {
	case 0:
		e.MoveUntilContains(NextToken, "APELLIDOS")
		if isCurrLine(e.NextToken, "APELLIDOS") &&
		isCurrLine(e.PrevToken, "NOMBRE") &&
		!isNumber(e.CurrToken) {
			return getSpouseData(e.CurrToken), nil
		}
		return "", nil
	}
	return "", errors.New("failed when extracting spouse first name or doesn't have")
}

func getSpouseLastName(e *Extractor, index int) (string, error) {
	exclude := []string { "OBS", "ACTIVIDAD", "PUBLICACI", "FECHA" }
	switch index {
	case 0:
		e.MoveUntilContains(CurrToken, "ACTIVIDAD")
		if !ContainsItem(exclude, e.NextToken) &&
		!isNumber(e.NextToken) {
			return getSpouseData(e.NextToken), nil
		}
		return "", nil
	case 1:
		e.MoveUntilContains(NextToken, "DATOS LABORALES")
		if isCurrLine(e.NextToken, "DATOS LABORALES") &&
		!ContainsItem(exclude, e.PrevToken) &&
		!isNumber(e.PrevToken) {
			return getSpouseData(e.PrevToken), nil
		}
		return "", nil
	case 2: // last case, low occurrence
		if e.MoveUntilContains(CurrToken, "APELLIDOS") &&
		!strings.Contains(e.NextToken, "CÉDULA") {
			return getSpouseData(e.NextToken), nil
		}
		return "", nil
	}

	return "", errors.New("failed when extracting spouse last name or doesn't have")
}

func getSpouseData(s string) string {
	if s == "" {
		return "NO TIENE"
	}
	return s
}

