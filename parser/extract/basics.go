package extract

import (
	"github.com/pkg/errors"
	"time"
)

// Date returns the date for the declaration.
func Date(e *Extractor) (time.Time, error) {
	var date string

	e.BindFlag(EXTRACTOR_FLAG_1)
	if e.MoveUntilContains(CurrToken, "DECLARACIÓN") {
		for e.Scan() {
			if isDate(e.CurrToken) &&
			len(e.CurrToken) <= 10 { // not substring (%dd/%mm/%yyyy)
				date = e.CurrToken
				break
			}
		}
	}

	if date == "" {
		return time.Time{}, errors.New("failed when extracting date")
	}

	t, err := time.Parse("02/01/2006", date)
	if err != nil {
		return time.Time{}, errors.New("Error parsing " + date + err.Error())
	}
	return t, nil
}

// Cedula returns the ID card number.
func Cedula(e *Extractor) (int, error) {
	var value int

	e.BindFlag(EXTRACTOR_FLAG_1)
	if e.MoveUntilStartWith(CurrToken, "CÉDULA") {
		if isNumber(e.NextToken) {
			value = stringToInt(e.NextToken)
		}
	}

	if value == 0 {
		return 0, errors.New("failed when extracting cedula")
	}
	return value, nil
}

// Name returns the official's name.
func Name(e *Extractor) (string, error) {
	var value string

	e.BindFlag(EXTRACTOR_FLAG_1)
	if e.MoveUntilStartWith(CurrToken, "NOMBRE") {
		if isAlpha(e.NextToken) {
			value = e.NextToken
		}
	}

	if value == "" {
		return "", errors.New("failed when extracting name")
	}
	return value, nil
}

// Lastname returns the official's lastname.
func Lastname(e *Extractor) (string, error) {
	var value string

	e.BindFlag(EXTRACTOR_FLAG_1)
	if e.MoveUntilStartWith(CurrToken, "APELLIDOS") {
		if isAlpha(e.NextToken) {
			value = e.NextToken
		}
	}

	if value == "" {
		return "", errors.New("failed when extracting lastname")
	}
	return value, nil
}

func ReceptionDate(e *Extractor) (time.Time, error) {
	var date string

	e.BindFlag(EXTRACTOR_FLAG_1)
	if e.MoveUntilStartWith(PrevToken, "RECEPCIONADO") {
		val, check := isKeyValuePair(e.PrevToken, "RECEPCIONADO")
		if check &&
		isDate(val) {
			date = getDate(val)
		}

		if date == "" &&
		isDate(e.NextToken) &&
		isBarCode(removeSpaces(e.CurrToken)) {
			date = getDate(e.NextToken)
		}
	}

	if date == "" {
		return time.Time{}, errors.New("failed when extracting reception date")
	}

	t, err := time.Parse("02/01/2006", date)
	if err != nil {
		return time.Time{}, errors.New("Error parsing " + date + err.Error())
	}
	return t, nil
}

func DownloadDate(e *Extractor) (time.Time, error) {
	var date string

	e.BindFlag(EXTRACTOR_FLAG_1)
	if e.MoveUntilStartWith(NextToken, "página") &&
	isCurrLine(e.CurrToken, "versión") {
		val := getDate(e.PrevToken)
		val += " "
		val += getTime(e.PrevToken)

		if isTimestamp(val) {
			date = val
		}
	}

	if date == "" {
		return time.Time{}, errors.New("failed when extracting download date")
	}

	// RFC3339 layout
	t, err := time.Parse("02/01/2006 15:04:05", date)
	if err != nil {
		return time.Time{}, errors.New("Error parsing " + date + err.Error())
	}
	return t, nil
}

func Version(e *Extractor) (string, error) {
	var version string

	e.BindFlag(EXTRACTOR_FLAG_1)
	if e.MoveUntilStartWith(CurrToken, "versión") {
		val, check := isKeyValuePair(e.CurrToken, "versión")
		if check {
			version = val
		}
	}

	if version == "" {
		return "", errors.New("failed when extracting version")
	}

	return version, nil
}
