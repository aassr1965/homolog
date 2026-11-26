package reporting

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
)

// XLSXCellUpdate définit une valeur à écrire dans une cellule Excel (ex: "C4", "A11").
type XLSXCellUpdate struct {
	SheetPath string // Optionnel : défaut "xl/worksheets/sheet1.xml"
	CellRef   string
	Value     string
	IsNumber  bool
}

var reRowRef = regexp.MustCompile(`^([A-Z]+)([0-9]+)$`)

// FillXLSXTemplate remplit des cellules d'un gabarit .xlsx officiel ANPP en mémoire (sheet1.xml).
func FillXLSXTemplate(templateBytes []byte, updates []XLSXCellUpdate) ([]byte, error) {
	zr, err := zip.NewReader(bytes.NewReader(templateBytes), int64(len(templateBytes)))
	if err != nil {
		return nil, fmt.Errorf("lecture archive xlsx impossible: %w", err)
	}

	bySheet := make(map[string][]XLSXCellUpdate)
	for _, u := range updates {
		sp := u.SheetPath
		if sp == "" {
			sp = "xl/worksheets/sheet1.xml"
		}
		bySheet[sp] = append(bySheet[sp], u)
	}

	var outBuf bytes.Buffer
	zw := zip.NewWriter(&outBuf)

	for _, f := range zr.File {
		rc, err := f.Open()
		if err != nil {
			_ = zw.Close()
			return nil, fmt.Errorf("ouverture entrée xlsx %s: %w", f.Name, err)
		}
		content, err := io.ReadAll(rc)
		_ = rc.Close()
		if err != nil {
			_ = zw.Close()
			return nil, fmt.Errorf("lecture entrée xlsx %s: %w", f.Name, err)
		}

		if sheetUpdates := bySheet[f.Name]; len(sheetUpdates) > 0 {
			content = []byte(applyXLSXUpdates(string(content), sheetUpdates))
		}

		hdr := f.FileHeader
		hdr.UncompressedSize64 = uint64(len(content))
		hdr.Method = zip.Deflate
		w, err := zw.CreateHeader(&hdr)
		if err != nil {
			_ = zw.Close()
			return nil, fmt.Errorf("écriture entrée xlsx %s: %w", f.Name, err)
		}
		if _, err := w.Write(content); err != nil {
			_ = zw.Close()
			return nil, fmt.Errorf("écriture contenu xlsx %s: %w", f.Name, err)
		}
	}

	if err := zw.Close(); err != nil {
		return nil, fmt.Errorf("finalisation archive xlsx: %w", err)
	}
	return outBuf.Bytes(), nil
}

func buildXLSXCellXML(cellRef, styleAttr, value string, isNumber bool) string {
	stylePart := ""
	if styleAttr != "" {
		stylePart = ` s="` + styleAttr + `"`
	}
	if isNumber {
		return fmt.Sprintf(`<c r="%s"%s t="n"><v>%s</v></c>`, cellRef, stylePart, escapeXMLText(value))
	}
	return fmt.Sprintf(`<c r="%s"%s t="inlineStr"><is><t xml:space="preserve">%s</t></is></c>`, cellRef, stylePart, escapeXMLText(value))
}

func applyXLSXUpdates(sheetXML string, updates []XLSXCellUpdate) string {
	if len(updates) == 0 {
		return sheetXML
	}
	for _, u := range updates {
		if u.CellRef == "" || u.Value == "" {
			continue
		}
		// 1. Si la cellule <c r="REF" ...> existe déjà, la remplacer en conservant son style s="..."
		reCell := regexp.MustCompile(`(?s)<c\s+[^>]*r="` + regexp.QuoteMeta(u.CellRef) + `"[^>]*(?:/>|>.*?</c>)`)
		if loc := reCell.FindStringIndex(sheetXML); loc != nil {
			existing := sheetXML[loc[0]:loc[1]]
			styleAttr := ""
			if sm := regexp.MustCompile(`\bs="([^"]+)"`).FindStringSubmatch(existing); len(sm) == 2 {
				styleAttr = sm[1]
			}
			replacement := buildXLSXCellXML(u.CellRef, styleAttr, u.Value, u.IsNumber)
			sheetXML = sheetXML[:loc[0]] + replacement + sheetXML[loc[1]:]
			continue
		}

		// 2. Sinon, trouver ou créer la ligne <row r="ROWNUM" ...>
		m := reRowRef.FindStringSubmatch(u.CellRef)
		if len(m) != 3 {
			continue
		}
		rowNum, _ := strconv.Atoi(m[2])
		reRow := regexp.MustCompile(`(?s)<row\s+[^>]*r="` + strconv.Itoa(rowNum) + `"[^>]*>.*?</row>`)
		cellXML := buildXLSXCellXML(u.CellRef, "", u.Value, u.IsNumber)
		if rloc := reRow.FindStringIndex(sheetXML); rloc != nil {
			rowStr := sheetXML[rloc[0]:rloc[1]]
			endIdx := strings.LastIndex(rowStr, "</row>")
			if endIdx != -1 {
				updatedRow := rowStr[:endIdx] + cellXML + rowStr[endIdx:]
				sheetXML = sheetXML[:rloc[0]] + updatedRow + sheetXML[rloc[1]:]
			}
		} else if sdEnd := strings.LastIndex(sheetXML, "</sheetData>"); sdEnd != -1 {
			newRow := fmt.Sprintf(`<row r="%d">%s</row>`, rowNum, cellXML)
			sheetXML = sheetXML[:sdEnd] + newRow + sheetXML[sdEnd:]
		}
	}
	return sheetXML
}
