package controllers

import (
	"fmt"
	"github.com/striker2000/petrovich"
	"github.com/unidoc/unioffice/color"
	"github.com/unidoc/unioffice/common/license"
	"github.com/unidoc/unioffice/document"
	"github.com/unidoc/unioffice/measurement"
	"github.com/unidoc/unioffice/schema/soo/wml"
	"github.com/unidoc/unioffice/spreadsheet"
	"os/exec"
	"path"
	"unicode/utf8"

	"os"
	"scinceHub/app/models"
	"strings"
)

const (
	italic  = "\x1b[3m"
	regular = "\x1b[0m"
)

func GetFileWithPublicationList(userID uint64, filters models.PublicationDownloadFiltres) (string, error) {
	publications, err := models.GetPublicationListByFilters(userID, filters)
	if err != nil {
		return "", err
	}
	var filename string
	switch filters.Type {
	case models.Word:
		filename, err = createWordDocument(userID, publications)
	case models.Exel:
		filename, err = createExcelDocument(userID, publications)
	case models.LibraWord:
		filename, err = createLibraWordDocument(userID, publications)
	case models.LibraExcel:
		filename, err = createLibraExcelDocument(userID, publications)
	}
	return filename, err
}

func createWordDocument(userID uint64, publications []models.Publications) (string, error) {
	doc := document.New()
	defer doc.Close()

	addStrokeCenter(doc, 16, "СПИСОК")
	addStrokeCenter(doc, 12, "учебно-методических и научных работ")
	addStrokeCenter(doc, 12, getStrokeYearsInterval(publications))
	addStrokeCenter(doc, 12, getFormattedNameByID(userID))

	doc.AddParagraph()
	table := doc.AddTable()
	table.Properties().SetWidthPercent(100)

	borders := table.Properties().Borders()
	borders.SetAll(wml.ST_BorderSingle, color.Auto, 1*measurement.Point)

	row := table.AddRow()
	AddRow(&row, "№", true)
	AddRow(&row, "Наименование работы", true)
	AddRow(&row, "Форма работы", true)
	AddRow(&row, "Дата публикации", true)
	AddRow(&row, "Соавторы", true)
	for index, publication := range publications {
		row = table.AddRow()
		AddRow(&row, fmt.Sprint(index+1), false)
		AddRow(&row, publication.Title, false)
		AddRow(&row, "Печатные", false)
		AddRow(&row, fmt.Sprint(publication.CreatedAt.Format("02.01.2006")), false)
		AddRow(&row, getAuthorsFromPublication(userID, publication), false)
	}
	randomNum, _ := GenerateRandomNumber()
	filename := fmt.Sprintf("public/uploads/%d_%d_list.docx", userID, randomNum)
	err := doc.SaveToFile(filename)
	if err != nil {
		return "", err
	}
	return filename, nil
}

func getStrokeYearsInterval(publications []models.Publications) string {
	stroke := ""
	if len(publications) == 0 {
		stroke = fmt.Sprintf("за - г.г.")
	} else {
		startYear := publications[len(publications)-1].CreatedAt.Format("2006")
		endYear := publications[0].CreatedAt.Format("2006")
		stroke = fmt.Sprintf("за %s%s-%s%s г.г.", italic, startYear, endYear, regular)
	}
	return stroke
}

func getFormattedNameByID(ID uint64) string {
	profile, _ := models.GetProfileNameByID(ID)
	var gender petrovich.Gender
	switch profile.Gender {
	case 1:
		gender = petrovich.Female
	case 2:
		gender = petrovich.Male
	default:
		gender = petrovich.Androgynous
	}
	fname := petrovich.FirstName(profile.FirstName, gender, petrovich.Genitive)
	lname := petrovich.LastName(profile.LastName, gender, petrovich.Genitive)
	mname := " " + petrovich.MiddleName(profile.MiddleName, gender, petrovich.Genitive)
	if mname == " " {
		mname = ""
	}

	return fmt.Sprintf("%s %s%s", lname, fname, mname)
}

func addStrokeCenter(doc *document.Document, fontSize float64, stroke string) {
	p := doc.AddParagraph()
	run := p.AddRun()
	run.Properties().SetSize(measurement.Distance(fontSize))
	run.Properties().SetFontFamily("Times New Roman")
	run.AddText(stroke)
	p.Properties().SetAlignment(wml.ST_JcCenter)
}

func getAuthorsFromPublication(userID uint64, publication models.Publications) string {
	authors := make([]string, 0)
	for _, profile := range publication.Profiles {
		if profile.ID == userID {
			continue
		}
		if profile.MiddleName == "" {
			authors = append(authors, fmt.Sprintf("%s %s", profile.LastName, profile.FirstName))
		} else {
			authors = append(authors, fmt.Sprintf("%s %s %s", profile.LastName, profile.FirstName, profile.MiddleName))
		}
	}
	return strings.Join(authors, ", ")
}

func createExcelDocument(userID uint64, publications []models.Publications) (string, error) {
	exel := spreadsheet.New()
	defer exel.Close()
	sheet := exel.AddSheet()

	boldStyle := exel.StyleSheet.AddCellStyle()
	boldFont := exel.StyleSheet.AddFont()
	boldFont.SetName("Times New Roman")
	boldFont.SetBold(true)
	boldFont.SetSize(12)
	boldStyle.SetFont(boldFont)

	style := exel.StyleSheet.AddCellStyle()
	font := exel.StyleSheet.AddFont()
	font.SetName("Times New Roman")
	font.SetSize(12)
	style.SetFont(font)

	headers := []string{
		"№",
		"Наименование работы",
		"Форма работы",
		"Дата публикации",
		"Соавторы",
	}
	row := sheet.AddRow()

	for i, header := range headers {
		SetCellParams(row.AddCell(), boldStyle, header)
		width := measurement.Distance(utf8.RuneCountInString(header))
		if i == 0 {
			sheet.Column(uint32(i) + 1).SetWidth(width * 40)
		} else {
			sheet.Column(uint32(i) + 1).SetWidth(width * 12)
		}
		row.SetHeightAuto()
	}
	for i, publication := range publications {
		row = sheet.AddRow()
		SetCellParams(row.AddCell(), style, fmt.Sprint(i+1))
		SetCellParams(row.AddCell(), style, publication.Title)
		SetCellParams(row.AddCell(), style, "Печатные")
		SetCellParams(row.AddCell(), style, fmt.Sprint(publication.CreatedAt.Format("02.01.2006")))
		SetCellParams(row.AddCell(), style, getAuthorsFromPublication(userID, publication))
		row.SetHeightAuto()
	}
	err := exel.Validate()
	if err != nil {
		return "", err
	}
	randomNum, _ := GenerateRandomNumber()
	filename := fmt.Sprintf("public/uploads/%d_%d_list.xlsx", userID, randomNum)
	err = exel.SaveToFile(filename)
	if err != nil {
		return "", err
	}

	return filename, err
}

func createLibraWordDocument(userID uint64, publications []models.Publications) (string, error) {
	filename, err := createWordDocument(userID, publications)
	if err != nil {
		return "", err
	}
	dir := path.Dir(filename)
	err = exec.Command("libreoffice", "--headless", "--convert-to", "odt", "--outdir", dir, filename).Run()

	if err != nil {
		return "", err
	}
	os.Remove(filename)
	filename = strings.Replace(filename, "docx", "odt", 1)
	return filename, nil
}

func createLibraExcelDocument(userID uint64, publications []models.Publications) (string, error) {
	filename, err := createExcelDocument(userID, publications)
	if err != nil {
		return "", err
	}
	dir := path.Dir(filename)
	err = exec.Command("libreoffice", "--headless", "--convert-to", "ods", "--outdir", dir, filename).Run()

	if err != nil {
		return "", err
	}
	os.Remove(filename)
	filename = strings.Replace(filename, "xlsx", "ods", 1)
	return filename, nil
}

func SetCellParams(cell spreadsheet.Cell, style spreadsheet.CellStyle, text string) {
	cell.SetString(text)
	cell.SetStyle(style)
}

func AddRow(row *document.Row, text string, bold bool) {
	run := row.AddCell().AddParagraph().AddRun()
	run.Properties().SetFontFamily("Times New Roman")
	run.Properties().SetSize(14)
	if bold {
		run.Properties().SetBold(true)
	}
	run.AddText(text)
}

func InitLicense() {
	err := license.SetMeteredKey(os.Getenv("UNIDOC_LICENSE_API_KEY"))
	if err != nil {
		panic(err)
	}
}
