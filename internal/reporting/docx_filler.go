package reporting

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"regexp"
	"strings"
)

// TableCellCoord cible une cellule précise d'un tableau Word (0-indexed).
type TableCellCoord struct {
	TableIndex int
	RowIndex   int
	ColIndex   int
	Value      string
}

// DOCXFillOptions définit les injections à appliquer sur un gabarit .docx officiel ANPP.
type DOCXFillOptions struct {
	// LabelValues associe un libellé de cellule (normalisé) à la valeur à écrire dans la cellule adjacente.
	LabelValues map[string]string
	// TextReplacements remplace des motifs exacts dans le texte des paragraphes/cellules (ex: "N°.…", "☐ Classe I").
	TextReplacements map[string]string
	// StrikeTerms barre (w:strike) les mentions exactes inutiles dans les paragraphes (ex: statuts non applicables).
	StrikeTerms []string
	// CellUpdates écrit des valeurs dans des coordonnées exactes (TableIndex, RowIndex, ColIndex).
	CellUpdates []TableCellCoord
}

var (
	reWT     = regexp.MustCompile(`(?s)<w:t(?:\s[^>]*)?>.*?</w:t>`)
	reWTag   = regexp.MustCompile(`<[^>]+>`)
	reWSpace = regexp.MustCompile(`\s+`)
)

func normalizeLabel(s string) string {
	s = strings.ReplaceAll(s, "\u00a0", " ")
	s = strings.ReplaceAll(s, "’", "'")
	s = reWSpace.ReplaceAllString(strings.TrimSpace(s), " ")
	return strings.ToLower(s)
}

func escapeXMLText(s string) string {
	var b bytes.Buffer
	_ = xml.EscapeText(&b, []byte(s))
	return b.String()
}

func extractXMLText(fragment string) string {
	matches := reWT.FindAllString(fragment, -1)
	if len(matches) == 0 {
		return ""
	}
	var sb strings.Builder
	for _, m := range matches {
		inner := reWTag.ReplaceAllString(m, "")
		sb.WriteString(unescapeBasicXML(inner))
	}
	return sb.String()
}

func unescapeBasicXML(s string) string {
	r := strings.NewReplacer(
		"&amp;", "&",
		"&lt;", "<",
		"&gt;", ">",
		"&quot;", "\"",
		"&apos;", "'",
	)
	return r.Replace(s)
}

// setCellText injecte ou remplace le texte dans une cellule <w:tc>...</w:tc> en préservant ses propriétés <w:tcPr> et <w:pPr>.
func setCellText(tcXML, value string) string {
	if value == "" {
		return tcXML
	}
	escaped := escapeXMLText(value)
	// Si la cellule possède déjà un <w:r> avec <w:t>, remplacer le premier <w:t> et vider les suivants si besoin,
	// sinon insérer un <w:r><w:t> dans le premier <w:p>.
	if reWT.MatchString(tcXML) {
		replacedFirst := false
		return reWT.ReplaceAllStringFunc(tcXML, func(match string) string {
			if !replacedFirst {
				replacedFirst = true
				return `<w:t xml:space="preserve">` + escaped + `</w:t>`
			}
			return `<w:t></w:t>`
		})
	}
	// Chercher </w:p> pour y insérer un run propre
	idx := strings.Index(tcXML, "</w:p>")
	if idx != -1 {
		run := `<w:r><w:t xml:space="preserve">` + escaped + `</w:t></w:r>`
		return tcXML[:idx] + run + tcXML[idx:]
	}
	return tcXML
}

// applyTextReplacementsToParagraph remplace des chaînes cibles à travers les <w:t> d'un paragraphe <w:p>.
func applyTextReplacementsToElement(elemXML string, replacements map[string]string) string {
	if len(replacements) == 0 {
		return elemXML
	}
	plain := extractXMLText(elemXML)
	if plain == "" {
		return elemXML
	}
	updated := plain
	changed := false
	for oldStr, newStr := range replacements {
		if strings.Contains(updated, oldStr) {
			updated = strings.ReplaceAll(updated, oldStr, newStr)
			changed = true
		}
	}
	if !changed {
		return elemXML
	}
	// Réinjecter le texte consolidé dans le premier <w:t> du bloc et vider les suivants du même bloc
	first := true
	return reWT.ReplaceAllStringFunc(elemXML, func(m string) string {
		if first {
			first = false
			return `<w:t xml:space="preserve">` + escapeXMLText(updated) + `</w:t>`
		}
		return `<w:t></w:t>`
	})
}

// splitTopLevelTags découpe toutes les occurrences non imbriquées d'un tag donné (ex: <w:tbl> ou <w:tr> ou <w:tc>).
func findTagRanges(xmlStr, tagName string) [][2]int {
	openTok := "<" + tagName
	closeTok := "</" + tagName + ">"
	var ranges [][2]int
	pos := 0
	for pos < len(xmlStr) {
		idx := strings.Index(xmlStr[pos:], openTok)
		if idx == -1 {
			break
		}
		start := pos + idx
		// Vérifier que le caractère après openTok est ' ' ou '>'
		afterOpen := start + len(openTok)
		if afterOpen >= len(xmlStr) || (xmlStr[afterOpen] != ' ' && xmlStr[afterOpen] != '>') {
			pos = afterOpen
			continue
		}
		// Suivre la profondeur d'imbrication
		depth := 1
		scan := afterOpen
		for scan < len(xmlStr) && depth > 0 {
			nextOpen := strings.Index(xmlStr[scan:], openTok)
			nextClose := strings.Index(xmlStr[scan:], closeTok)
			if nextClose == -1 {
				break
			}
			if nextOpen != -1 && nextOpen < nextClose {
				cAfter := scan + nextOpen + len(openTok)
				if cAfter < len(xmlStr) && (xmlStr[cAfter] == ' ' || xmlStr[cAfter] == '>') {
					depth++
				}
				scan = cAfter
			} else {
				depth--
				scan = scan + nextClose + len(closeTok)
			}
		}
		if depth == 0 {
			ranges = append(ranges, [2]int{start, scan})
			pos = scan
		} else {
			pos = afterOpen
		}
	}
	return ranges
}

// FillDOCXTemplate remplit un gabarit .docx ANPP en mémoire et renvoie les octets du .docx généré.
func FillDOCXTemplate(templateBytes []byte, opts DOCXFillOptions) ([]byte, error) {
	zr, err := zip.NewReader(bytes.NewReader(templateBytes), int64(len(templateBytes)))
	if err != nil {
		return nil, fmt.Errorf("lecture archive docx impossible: %w", err)
	}

	normLabels := make(map[string]string, len(opts.LabelValues))
	for k, v := range opts.LabelValues {
		normLabels[normalizeLabel(k)] = v
	}

	coordMap := make(map[[3]int]string, len(opts.CellUpdates))
	for _, cu := range opts.CellUpdates {
		coordMap[[3]int{cu.TableIndex, cu.RowIndex, cu.ColIndex}] = cu.Value
	}

	var outBuf bytes.Buffer
	zw := zip.NewWriter(&outBuf)

	for _, f := range zr.File {
		rc, err := f.Open()
		if err != nil {
			_ = zw.Close()
			return nil, fmt.Errorf("ouverture entrée zip %s: %w", f.Name, err)
		}
		content, err := io.ReadAll(rc)
		_ = rc.Close()
		if err != nil {
			_ = zw.Close()
			return nil, fmt.Errorf("lecture entrée zip %s: %w", f.Name, err)
		}

		if f.Name == "word/document.xml" {
			content = []byte(transformDocumentXML(string(content), normLabels, opts.TextReplacements, opts.StrikeTerms, coordMap))
		}

		hdr := f.FileHeader
		hdr.UncompressedSize64 = uint64(len(content))
		hdr.Method = zip.Deflate
		w, err := zw.CreateHeader(&hdr)
		if err != nil {
			_ = zw.Close()
			return nil, fmt.Errorf("écriture entrée zip %s: %w", f.Name, err)
		}
		if _, err := w.Write(content); err != nil {
			_ = zw.Close()
			return nil, fmt.Errorf("écriture contenu zip %s: %w", f.Name, err)
		}
	}

	if err := zw.Close(); err != nil {
		return nil, fmt.Errorf("finalisation archive docx: %w", err)
	}
	return outBuf.Bytes(), nil
}

func transformDocumentXML(docXML string, normLabels map[string]string, textRepl map[string]string, strikeTerms []string, coordMap map[[3]int]string) string {
	tblRanges := findTagRanges(docXML, "w:tbl")
	var out strings.Builder
	last := 0

	for ti, tr := range tblRanges {
		// Paragraphes hors tableaux avant ce tableau (anti-titre orphelin activé)
		segment := docXML[last:tr[0]]
		segment = transformParagraphsInSegment(segment, textRepl, strikeTerms, true)
		out.WriteString(segment)

		tblXML := transformTableXML(docXML[tr[0]:tr[1]], ti, normLabels, textRepl, strikeTerms, coordMap)
		out.WriteString(tblXML)
		last = tr[1]
	}

	tail := docXML[last:]
	tail = transformParagraphsInSegment(tail, textRepl, strikeTerms, false)
	out.WriteString(tail)
	return out.String()
}

func transformParagraphsInSegment(seg string, textRepl map[string]string, strikeTerms []string, beforeTable bool) string {
	pRanges := findTagRanges(seg, "w:p")
	if len(pRanges) == 0 {
		return seg
	}
	// Repérer le dernier paragraphe non vide avant un tableau pour éviter un titre orphelin
	lastNonEmptyIdx := -1
	if beforeTable {
		for i := len(pRanges) - 1; i >= 0; i-- {
			if strings.TrimSpace(extractXMLText(seg[pRanges[i][0]:pRanges[i][1]])) != "" {
				lastNonEmptyIdx = i
				break
			}
		}
	}
	var out strings.Builder
	last := 0
	for i, pr := range pRanges {
		out.WriteString(seg[last:pr[0]])
		pXML := seg[pr[0]:pr[1]]
		if len(textRepl) > 0 {
			pXML = applyTextReplacementsToElement(pXML, textRepl)
		}
		if len(strikeTerms) > 0 {
			pXML = applyStrikeTermsToParagraph(pXML, strikeTerms)
		}
		if isHeadingParagraph(pXML) || i == lastNonEmptyIdx {
			pXML = ensureKeepNext(pXML)
		}
		out.WriteString(pXML)
		last = pr[1]
	}
	out.WriteString(seg[last:])
	return out.String()
}

var reFirstRPr = regexp.MustCompile(`(?s)<w:rPr>.*?</w:rPr>`)

// applyStrikeTermsToParagraph barre (w:strike) uniquement les termes listés dans strikeTerms tout en conservant le reste du paragraphe intact.
func applyStrikeTermsToParagraph(pXML string, strikeTerms []string) string {
	plain := extractXMLText(pXML)
	if plain == "" {
		return pXML
	}
	var active []string
	for _, t := range strikeTerms {
		if t != "" && strings.Contains(plain, t) {
			active = append(active, t)
		}
	}
	if len(active) == 0 {
		return pXML
	}

	// Récupérer le <w:rPr> du premier run pour préserver la police/taille/couleur
	baseRPr := ""
	if m := reFirstRPr.FindString(pXML); m != "" {
		baseRPr = m
	}
	strikeRPr := `<w:rPr><w:strike/></w:rPr>`
	if baseRPr != "" {
		if strings.Contains(baseRPr, "<w:strike") {
			strikeRPr = baseRPr
		} else {
			strikeRPr = strings.Replace(baseRPr, "</w:rPr>", "<w:strike/></w:rPr>", 1)
		}
	}

	// Découper plain en segments normaux / barrés
	type seg struct {
		text   string
		struck bool
	}
	segs := []seg{{text: plain, struck: false}}
	for _, term := range active {
		var next []seg
		for _, s := range segs {
			if s.struck || !strings.Contains(s.text, term) {
				next = append(next, s)
				continue
			}
			parts := strings.Split(s.text, term)
			for idx, part := range parts {
				if part != "" {
					next = append(next, seg{text: part, struck: false})
				}
				if idx < len(parts)-1 {
					next = append(next, seg{text: term, struck: true})
				}
			}
		}
		segs = next
	}

	var runsBuilder strings.Builder
	for _, s := range segs {
		rPr := baseRPr
		if s.struck {
			rPr = strikeRPr
		}
		runsBuilder.WriteString(`<w:r>`)
		runsBuilder.WriteString(rPr)
		runsBuilder.WriteString(`<w:t xml:space="preserve">`)
		runsBuilder.WriteString(escapeXMLText(s.text))
		runsBuilder.WriteString(`</w:t></w:r>`)
	}

	// Remplacer tous les <w:r>...</w:r> du paragraphe tout en conservant <w:p ...><w:pPr>...</w:pPr> et </w:p>
	rRanges := findTagRanges(pXML, "w:r")
	if len(rRanges) == 0 {
		return pXML
	}
	var out strings.Builder
	out.WriteString(pXML[:rRanges[0][0]])
	out.WriteString(runsBuilder.String())
	out.WriteString(pXML[rRanges[len(rRanges)-1][1]:])
	return out.String()
}

func isHeadingParagraph(pXML string) bool {
	txt := strings.TrimSpace(extractXMLText(pXML))
	if txt == "" || len([]rune(txt)) > 180 {
		return false
	}
	if strings.Contains(pXML, `w:val="Heading`) || strings.Contains(pXML, `w:val="Titre`) || strings.Contains(pXML, `w:val="Title`) {
		return true
	}
	// Titre court en gras (ex: "Informations générales...", "1. Identification...") ne se terminant pas par ":" ou "."
	if strings.Contains(pXML, "<w:b") && !strings.HasSuffix(txt, ":") && !strings.HasSuffix(txt, ".") {
		return true
	}
	return false
}

func ensureKeepNext(pXML string) string {
	if strings.Contains(pXML, "<w:keepNext") {
		return pXML
	}
	if strings.Contains(pXML, "<w:pPr>") {
		return strings.Replace(pXML, "<w:pPr>", "<w:pPr><w:keepNext/><w:keepLines/>", 1)
	}
	idx := strings.Index(pXML, ">")
	if idx == -1 {
		return pXML
	}
	return pXML[:idx+1] + "<w:pPr><w:keepNext/><w:keepLines/></w:pPr>" + pXML[idx+1:]
}

func transformTableXML(tblXML string, tblIdx int, normLabels map[string]string, textRepl map[string]string, strikeTerms []string, coordMap map[[3]int]string) string {
	rowRanges := findTagRanges(tblXML, "w:tr")
	if len(rowRanges) == 0 {
		return tblXML
	}
	var out strings.Builder
	last := 0
	for ri, rr := range rowRanges {
		out.WriteString(tblXML[last:rr[0]])
		out.WriteString(transformRowXML(tblXML[rr[0]:rr[1]], tblIdx, ri, normLabels, textRepl, strikeTerms, coordMap))
		last = rr[1]
	}
	out.WriteString(tblXML[last:])
	return out.String()
}

func transformRowXML(rowXML string, tblIdx, rowIdx int, normLabels map[string]string, textRepl map[string]string, strikeTerms []string, coordMap map[[3]int]string) string {
	cellRanges := findTagRanges(rowXML, "w:tc")
	if len(cellRanges) == 0 {
		return rowXML
	}

	cells := make([]string, len(cellRanges))
	for ci, cr := range cellRanges {
		cells[ci] = rowXML[cr[0]:cr[1]]
	}

	// 1. Remplacements textuels et mentions barrées dans chaque cellule
	if len(textRepl) > 0 || len(strikeTerms) > 0 {
		for ci := range cells {
			cells[ci] = transformParagraphsInSegment(cells[ci], textRepl, strikeTerms, false)
		}
	}

	// 2. Correspondance par libellé -> écriture dans la cellule suivante si elle est vide
	if len(normLabels) > 0 && len(cells) >= 2 {
		for ci := 0; ci < len(cells)-1; ci++ {
			lbl := normalizeLabel(extractXMLText(cells[ci]))
			if lbl == "" {
				continue
			}
			for key, val := range normLabels {
				if val == "" {
					continue
				}
				if lbl == key || strings.HasPrefix(lbl, key) {
					nextTxt := strings.TrimSpace(extractXMLText(cells[ci+1]))
					if nextTxt == "" {
						cells[ci+1] = setCellText(cells[ci+1], val)
					}
					break
				}
			}
		}
	}

	// 3. Coordonnées explicites (prioritaires)
	for ci := range cells {
		if val, ok := coordMap[[3]int{tblIdx, rowIdx, ci}]; ok && val != "" {
			cells[ci] = setCellText(cells[ci], val)
		}
	}

	var out strings.Builder
	last := 0
	for ci, cr := range cellRanges {
		out.WriteString(rowXML[last:cr[0]])
		out.WriteString(cells[ci])
		last = cr[1]
	}
	out.WriteString(rowXML[last:])
	return out.String()
}
