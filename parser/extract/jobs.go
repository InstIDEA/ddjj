package extract

import (
	"fmt"
	"strings"
	"github.com/InstIDEA/ddjj/parser/declaration"	
)

func Jobs(e *Extractor, parser *ParserData) []*declaration.Job {

	e.BindFlag(EXTRACTOR_FLAG_1)
	e.BindFlag(EXTRACTOR_FLAG_2)

	var instituciones []*declaration.Job
	var counter = countJobs(e)
	var successful int

	e.Rewind()
	e.BindFlag(EXTRACTOR_FLAG_3)

	job := &declaration.Job{ }

	/* tokenize the read text line and combine
	the tokens that are separed by the ':' symbol */
	e.BufferHookFunc = func(tokens []string) []string {
		return combine(tokens, ":")
	}

	if counter > 0 &&
	e.MoveUntilStartWith(CurrToken, "DATOS LABORALES") {

		for e.Scan() {
			if counter == successful {
				break
			}

			if job.Institucion == "" {
				value := getJobInst(e)

				if !isJobFormField(value) {
					job.Institucion = value
				}
			}

			if job.Cargo == "" &&
			job.Institucion != "" {
				value := getJobTitle(e)

				if !isJobFormField(value) {
					job.Cargo = value
				}
			}

			if job.Cargo != "" && job.Institucion != "" {
				successful++
				instituciones = append(instituciones, job)
				job = &declaration.Job{ }
			}
		}
	}

	if successful != counter {
		parser.addMessage(fmt.Sprintf("ignored jobs: %d/%d", counter - successful, counter))
	}

	if instituciones == nil {
		parser.addError(fmt.Errorf("failed when extracting jobs"))
		return nil
	}

	return instituciones
}

func getJobTitle(e *Extractor) string {

	if strings.Contains(e.CurrToken, "CARGO") {
		val, check := isKeyValuePair(e.CurrToken, "CARGO")
		if check {
			return val
		}
	}

	if strings.Contains(e.PrevToken, "CARGO") && 
	strings.Contains(e.CurrToken, "FECHA EGRESO") {
		return e.NextToken

	}

	return ""
}

func getJobInst(e *Extractor) string {

	if strings.Contains(e.CurrToken, "INSTITUCIÓN") &&
	strings.Contains(e.NextToken, "ACTO ADM. COM") {
		val, check := isKeyValuePair(e.CurrToken, "INSTITUCIÓN")
		if check {
			return val
		}
	}

	if strings.Contains(e.PrevToken, "DIRECCIÓN") &&
	isNumber(e.CurrToken) {
		return e.NextToken
	}

	return ""
}

func countJobs(e *Extractor) int {
	var counter int

	for e.Scan() {
		if strings.Contains(e.CurrToken, "CARGO:") {
			counter++
		}
	}
	return counter
}

func isJobFormField(s string) bool {
	formField := []string {
		"TIPO",
		"INSTITUCION:",
		"DIRECCION:",
		"DEPENDENCIA",
		"CATEGORIA",
		"NOMBRADO/CONTRATADO",
		"CARGO",
		"FECHA ASUNC./CESE/OTROS",
		"ACTO ADMINIST",
		"FECHA ACT. ADM",
		"TELEFONO",
		"COMISIONADO:",
		"FECHA INGRESO",
		"FECHA EGRESO",
		"ACTO ADM. COM",
	}

	s = removeAccents(s)
	for _, value := range formField {
		if isCurrLine(s, value) {
			return true
		}
	}

	return false
}
