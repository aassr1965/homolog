package reporting

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"path"
	"sort"
	"strconv"
	"strings"
)

// PreviewBlock représente un bloc structuré (paragraphe ou tableau) d'un document Word (.docx).
type PreviewBlock struct {
	Type string     `json:"type"` // "paragraph" | "table"
	Text string     `json:"text,omitempty"`
	Rows [][]string `json:"rows,omitempty"`
}

// PreviewSheet représente une feuille de calcul Excel (.xlsx).
type PreviewSheet struct {
	Name string     `json:"name"`
	Rows [][]string `json:"rows"`
}

// ANPPTemplatePreview contient l'aperçu structuré d'un gabarit officiel (.docx ou .xlsx).
type ANPPTemplatePreview struct {
	Name   string         `json:"name"`
	Kind   string         `json:"kind"` // "docx" | "xlsx"
	Blocks []PreviewBlock `json:"blocks,omitempty"`
	Sheets []PreviewSheet `json:"sheets,omitempty"`
}

// ExtractANPPTemplatePreview extrait le contenu lisible (paragraphes, tableaux ou feuilles) d'un gabarit .docx ou .xlsx.
func ExtractANPPTemplatePreview(filename string, raw []byte) (*ANPPTemplatePreview, error) {
	ext := strings.ToLower(path.Ext(filename))
	zr, err := zip.NewReader(bytes.NewReader(raw), int64(len(raw)))
	if err != nil {
		return nil, fmt.Errorf("archive Office invalide: %w", err)
	}

	files := make(map[string]*zip.File, len(zr.File))
	for _, f := range zr.File {
		files[f.Name] = f
	}

	switch ext {
	case ".docx":
		docFile, ok := files["word/document.xml"]
		if !ok {
			return nil, fmt.Errorf("word/document.xml introuvable")
		}
		blocks, err := parseDOCXBodyBlocks(docFile)
		if err != nil {
			return nil, err
		}
		return &ANPPTemplatePreview{
			Name:   path.Base(filename),
			Kind:   "docx",
			Blocks: blocks,
		}, nil

	case ".xlsx":
		sheets, err := parseXLSXSheets(files)
		if err != nil {
			return nil, err
		}
		return &ANPPTemplatePreview{
			Name:   path.Base(filename),
			Kind:   "xlsx",
			Sheets: sheets,
		}, nil

	default:
		return nil, fmt.Errorf("format non supporté pour l'aperçu: %s", ext)
	}
}

func readZipFileBytes(f *zip.File) ([]byte, error) {
	rc, err := f.Open()
	if err != nil {
		return nil, err
	}
	defer rc.Close()
	return io.ReadAll(io.LimitReader(rc, 16<<20))
}

func parseDOCXBodyBlocks(docFile *zip.File) ([]PreviewBlock, error) {
	data, err := readZipFileBytes(docFile)
	if err != nil {
		return nil, err
	}

	dec := xml.NewDecoder(bytes.NewReader(data))
	var blocks []PreviewBlock

	inBody := false
	inTopPara := false
	inTable := false
	inRow := false
	inCell := false
	inCellPara := false
	inText := false

	var paraBuf strings.Builder
	var cellBuf strings.Builder
	var curRow []string
	var curTable [][]string

	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		switch t := tok.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "body":
				inBody = true
			case "tbl":
				if inBody && !inTable {
					inTable = true
					curTable = nil
				}
			case "tr":
				if inTable {
					inRow = true
					curRow = nil
				}
			case "tc":
				if inRow {
					inCell = true
					cellBuf.Reset()
				}
			case "p":
				if inCell {
					inCellPara = true
					if cellBuf.Len() > 0 {
						cellBuf.WriteString("\n")
					}
				} else if inBody && !inTable {
					inTopPara = true
					paraBuf.Reset()
				}
			case "t":
				if inTopPara || inCellPara {
					inText = true
				}
			case "br":
				if inTopPara {
					paraBuf.WriteString("\n")
				} else if inCellPara {
					cellBuf.WriteString("\n")
				}
			case "tab":
				if inTopPara {
					paraBuf.WriteString("\t")
				} else if inCellPara {
					cellBuf.WriteString(" ")
				}
			}

		case xml.CharData:
			if inText {
				if inCellPara {
					cellBuf.Write(t)
				} else if inTopPara {
					paraBuf.Write(t)
				}
			}

		case xml.EndElement:
			switch t.Name.Local {
			case "body":
				inBody = false
			case "t":
				inText = false
			case "p":
				if inCellPara {
					inCellPara = false
				} else if inTopPara {
					inTopPara = false
					txt := strings.TrimSpace(paraBuf.String())
					if txt != "" {
						blocks = append(blocks, PreviewBlock{Type: "paragraph", Text: txt})
					}
				}
			case "tc":
				if inCell {
					inCell = false
					curRow = append(curRow, strings.TrimSpace(cellBuf.String()))
				}
			case "tr":
				if inRow {
					inRow = false
					if len(curRow) > 0 {
						curTable = append(curTable, curRow)
					}
				}
			case "tbl":
				if inTable {
					inTable = false
					if len(curTable) > 0 {
						blocks = append(blocks, PreviewBlock{Type: "table", Rows: curTable})
					}
				}
			}
		}
	}
	return blocks, nil
}

func parseXLSXSheets(files map[string]*zip.File) ([]PreviewSheet, error) {
	var sharedStrings []string
	if ssFile, ok := files["xl/sharedStrings.xml"]; ok {
		sharedStrings = parseXLSXSharedStrings(ssFile)
	}

	sheetNames := parseXLSXWorkbookNames(files["xl/workbook.xml"])

	var sheetPaths []string
	for name := range files {
		if strings.HasPrefix(name, "xl/worksheets/sheet") && strings.HasSuffix(name, ".xml") {
			sheetPaths = append(sheetPaths, name)
		}
	}
	sort.Strings(sheetPaths)

	sheets := make([]PreviewSheet, 0, len(sheetPaths))
	for i, sp := range sheetPaths {
		rows := parseXLSXSheetRows(files[sp], sharedStrings)
		title := fmt.Sprintf("Feuille %d", i+1)
		if i < len(sheetNames) && sheetNames[i] != "" {
			title = sheetNames[i]
		}
		sheets = append(sheets, PreviewSheet{Name: title, Rows: rows})
	}
	return sheets, nil
}

func parseXLSXSharedStrings(f *zip.File) []string {
	data, err := readZipFileBytes(f)
	if err != nil {
		return nil
	}
	dec := xml.NewDecoder(bytes.NewReader(data))
	var list []string
	inSI := false
	inT := false
	var buf strings.Builder

	for {
		tok, err := dec.Token()
		if err != nil {
			break
		}
		switch t := tok.(type) {
		case xml.StartElement:
			if t.Name.Local == "si" {
				inSI = true
				buf.Reset()
			} else if t.Name.Local == "t" && inSI {
				inT = true
			}
		case xml.CharData:
			if inT {
				buf.Write(t)
			}
		case xml.EndElement:
			if t.Name.Local == "t" {
				inT = false
			} else if t.Name.Local == "si" {
				inSI = false
				list = append(list, buf.String())
			}
		}
	}
	return list
}

func parseXLSXWorkbookNames(f *zip.File) []string {
	if f == nil {
		return nil
	}
	data, err := readZipFileBytes(f)
	if err != nil {
		return nil
	}
	dec := xml.NewDecoder(bytes.NewReader(data))
	var names []string
	for {
		tok, err := dec.Token()
		if err != nil {
			break
		}
		if se, ok := tok.(xml.StartElement); ok && se.Name.Local == "sheet" {
			for _, attr := range se.Attr {
				if attr.Name.Local == "name" {
					names = append(names, attr.Value)
					break
				}
			}
		}
	}
	return names
}

func xlsxColIndex(cellRef string) int {
	col := 0
	for i := 0; i < len(cellRef); i++ {
		ch := cellRef[i]
		if ch >= 'A' && ch <= 'Z' {
			col = col*26 + int(ch-'A'+1)
		} else if ch >= 'a' && ch <= 'z' {
			col = col*26 + int(ch-'a'+1)
		} else {
			break
		}
	}
	if col > 0 {
		return col - 1
	}
	return 0
}

func parseXLSXSheetRows(f *zip.File, shared []string) [][]string {
	data, err := readZipFileBytes(f)
	if err != nil {
		return nil
	}
	dec := xml.NewDecoder(bytes.NewReader(data))
	var rows [][]string
	const maxRows = 200
	const maxCols = 30

	inRow := false
	inCell := false
	inVal := false
	inInlineT := false
	cellType := ""
	cellCol := 0
	var valBuf strings.Builder
	var curRow []string

	for {
		tok, err := dec.Token()
		if err != nil {
			break
		}
		switch t := tok.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "row":
				if len(rows) >= maxRows {
					return rows
				}
				inRow = true
				curRow = nil
			case "c":
				if inRow {
					inCell = true
					cellType = ""
					cellCol = len(curRow)
					valBuf.Reset()
					for _, a := range t.Attr {
						if a.Name.Local == "t" {
							cellType = a.Value
						} else if a.Name.Local == "r" {
							cellCol = xlsxColIndex(a.Value)
						}
					}
				}
			case "v":
				if inCell {
					inVal = true
					valBuf.Reset()
				}
			case "t":
				if inCell && cellType == "inlineStr" {
					inInlineT = true
				}
			}
		case xml.CharData:
			if inVal || inInlineT {
				valBuf.Write(t)
			}
		case xml.EndElement:
			switch t.Name.Local {
			case "v":
				inVal = false
			case "t":
				inInlineT = false
			case "c":
				if inCell {
					inCell = false
					raw := strings.TrimSpace(valBuf.String())
					val := raw
					if cellType == "s" {
						if idx, convErr := strconv.Atoi(raw); convErr == nil && idx >= 0 && idx < len(shared) {
							val = shared[idx]
						}
					}
					if cellCol >= 0 && cellCol < maxCols {
						for len(curRow) <= cellCol {
							curRow = append(curRow, "")
						}
						curRow[cellCol] = val
					}
				}
			case "row":
				if inRow {
					inRow = false
					hasContent := false
					for _, c := range curRow {
						if strings.TrimSpace(c) != "" {
							hasContent = true
							break
						}
					}
					if hasContent {
						rows = append(rows, curRow)
					}
				}
			}
		}
	}
	return rows
}
