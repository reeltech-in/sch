package main

import (
	"archive/zip"
	"bytes"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// Pembuat dokumen Word (.docx) tanpa pustaka luar: .docx adalah arsip zip berisi XML (Office Open XML).
// Masukan: teks Markdown sederhana yang sama dengan yang ditampilkan aplikasi (heading, paragraf, daftar,
// tabel, **tebal**, *miring*, `kode`, garis ---), ditambah kop sekolah dan kolom tanda tangan opsional.

type docxOpts struct {
	Judul      string
	Kop        bool
	NamaSek    string
	AlamatSek  string
	TeleponSek string
	LogoPath   string // file logo lokal (png/jpg) atau kosong
	TTD        bool
	Tempat     string
	Tanggal    string // sudah diformat, mis. "10 November 2026"
	Guru       string
	NIPGuru    string
	Kepsek     string
	NIPKepsek  string
}

func xmlEsc(s string) string {
	var b bytes.Buffer
	for _, r := range s {
		switch r {
		case '&':
			b.WriteString("&amp;")
		case '<':
			b.WriteString("&lt;")
		case '>':
			b.WriteString("&gt;")
		case '"':
			b.WriteString("&quot;")
		default:
			if r == 0x9 || r == 0xA || r == 0xD || (r >= 0x20 && r <= 0xD7FF) || (r >= 0xE000 && r <= 0xFFFD) || r >= 0x10000 {
				b.WriteRune(r)
			}
		}
	}
	return b.String()
}

var mdInlineRe = regexp.MustCompile("(\\*\\*[^*]+\\*\\*|`[^`]+`|\\*[^*\\n]+\\*)")

// runs: teks satu baris → <w:r> dengan tebal/miring/kode.
func docxRuns(s string, bold bool) string {
	var b strings.Builder
	run := func(t string, bo, it, code bool) {
		if t == "" {
			return
		}
		b.WriteString("<w:r><w:rPr>")
		if bo {
			b.WriteString("<w:b/>")
		}
		if it {
			b.WriteString("<w:i/>")
		}
		if code {
			b.WriteString(`<w:rFonts w:ascii="Consolas" w:hAnsi="Consolas"/>`)
		}
		b.WriteString(`</w:rPr><w:t xml:space="preserve">` + xmlEsc(t) + "</w:t></w:r>")
	}
	last := 0
	for _, m := range mdInlineRe.FindAllStringIndex(s, -1) {
		run(s[last:m[0]], bold, false, false)
		tok := s[m[0]:m[1]]
		switch {
		case strings.HasPrefix(tok, "**"):
			run(tok[2:len(tok)-2], true, false, false)
		case strings.HasPrefix(tok, "`"):
			run(tok[1:len(tok)-1], bold, false, true)
		default:
			run(tok[1:len(tok)-1], bold, true, false)
		}
		last = m[1]
	}
	run(s[last:], bold, false, false)
	return b.String()
}

func docxPara(style, runs, extraPPr string) string {
	p := "<w:p><w:pPr>"
	if style != "" {
		p += `<w:pStyle w:val="` + style + `"/>`
	}
	return p + extraPPr + "</w:pPr>" + runs + "</w:p>"
}

var (
	mdHeadRe  = regexp.MustCompile(`^(#{1,4})\s+(.*)$`)
	mdListRe  = regexp.MustCompile(`^(\s*)([-*•]|\d+[.)])\s+(.*)$`)
	mdTableRe = regexp.MustCompile(`^\s*\|.*\|\s*$`)
	mdSepRe   = regexp.MustCompile(`^\s*\|?\s*:?-{2,}`)
	mdHrRe    = regexp.MustCompile(`^\s*(-{3,}|_{3,}|\*{3,})\s*$`)
)

func docxTable(rows [][]string, header bool) string {
	cols := 0
	for _, r := range rows {
		if len(r) > cols {
			cols = len(r)
		}
	}
	border := `<w:top w:val="single" w:sz="4" w:color="808080"/><w:left w:val="single" w:sz="4" w:color="808080"/><w:bottom w:val="single" w:sz="4" w:color="808080"/><w:right w:val="single" w:sz="4" w:color="808080"/><w:insideH w:val="single" w:sz="4" w:color="808080"/><w:insideV w:val="single" w:sz="4" w:color="808080"/>`
	var b strings.Builder
	colW := docxLebar / max(1, cols)
	b.WriteString(fmt.Sprintf(`<w:tbl><w:tblPr><w:tblW w:w="%d" w:type="dxa"/><w:tblLayout w:type="fixed"/><w:tblBorders>`, docxLebar) + border + `</w:tblBorders><w:tblCellMar><w:left w:w="80" w:type="dxa"/><w:right w:w="80" w:type="dxa"/></w:tblCellMar></w:tblPr><w:tblGrid>`)
	for i := 0; i < cols; i++ {
		b.WriteString(fmt.Sprintf(`<w:gridCol w:w="%d"/>`, colW))
	}
	b.WriteString(`</w:tblGrid>`)
	for i, r := range rows {
		b.WriteString("<w:tr>")
		if i == 0 && header {
			b.WriteString(`<w:trPr><w:tblHeader/></w:trPr>`)
		}
		for c := 0; c < cols; c++ {
			cell := ""
			if c < len(r) {
				cell = r[c]
			}
			shade := ""
			if i == 0 && header {
				shade = `<w:shd w:val="clear" w:color="auto" w:fill="E7EBF0"/>`
			}
			b.WriteString(fmt.Sprintf(`<w:tc><w:tcPr><w:tcW w:w="%d" w:type="dxa"/>`, colW) + shade + `</w:tcPr>` + docxPara("TabelIsi", docxRuns(cell, i == 0 && header), "") + `</w:tc>`)
		}
		b.WriteString("</w:tr>")
	}
	b.WriteString(`</w:tbl>` + docxPara("", "", ""))
	return b.String()
}

// docxBody: Markdown → isi <w:body> (tanpa sectPr).
func docxBody(md string) string {
	lines := strings.Split(strings.ReplaceAll(md, "\r\n", "\n"), "\n")
	var b strings.Builder
	cell := func(r string) []string {
		r = strings.TrimSpace(r)
		r = strings.TrimPrefix(strings.TrimSuffix(r, "|"), "|")
		parts := strings.Split(r, "|")
		for i := range parts {
			parts[i] = strings.TrimSpace(parts[i])
		}
		return parts
	}
	for i := 0; i < len(lines); {
		l := lines[i]
		switch {
		case strings.HasPrefix(strings.TrimSpace(l), "```"):
			i++
			for i < len(lines) && !strings.HasPrefix(strings.TrimSpace(lines[i]), "```") {
				b.WriteString(docxPara("Kode", `<w:r><w:t xml:space="preserve">`+xmlEsc(lines[i])+`</w:t></w:r>`, ""))
				i++
			}
			i++
		case mdHeadRe.MatchString(l):
			m := mdHeadRe.FindStringSubmatch(l)
			b.WriteString(docxPara(fmt.Sprintf("Heading%d", min(3, len(m[1]))), docxRuns(m[2], false), ""))
			i++
		case mdTableRe.MatchString(l) && i+1 < len(lines) && mdSepRe.MatchString(lines[i+1]):
			rows := [][]string{cell(l)}
			i += 2
			for i < len(lines) && mdTableRe.MatchString(lines[i]) {
				rows = append(rows, cell(lines[i]))
				i++
			}
			b.WriteString(docxTable(rows, true))
		case mdListRe.MatchString(l):
			for i < len(lines) && mdListRe.MatchString(lines[i]) {
				m := mdListRe.FindStringSubmatch(lines[i])
				indent := 360 + 360*min(3, len(strings.ReplaceAll(m[1], "\t", "  "))/2)
				mark := "•"
				if m[2][0] >= '0' && m[2][0] <= '9' {
					mark = m[2]
				}
				ppr := fmt.Sprintf(`<w:ind w:left="%d" w:hanging="300"/>`, indent)
				b.WriteString(docxPara("Daftar", `<w:r><w:t xml:space="preserve">`+xmlEsc(mark)+"\t</w:t></w:r>"+docxRuns(m[3], false), ppr+`<w:tabs><w:tab w:val="left" w:pos="`+fmt.Sprint(indent)+`"/></w:tabs>`))
				i++
			}
		case mdHrRe.MatchString(l):
			b.WriteString(docxPara("", "", `<w:pBdr><w:bottom w:val="single" w:sz="6" w:color="999999"/></w:pBdr>`))
			i++
		case strings.TrimSpace(l) == "":
			i++
		default:
			var runs []string
			for i < len(lines) && strings.TrimSpace(lines[i]) != "" && !mdHeadRe.MatchString(lines[i]) && !mdListRe.MatchString(lines[i]) &&
				!mdTableRe.MatchString(lines[i]) && !strings.HasPrefix(strings.TrimSpace(lines[i]), "```") && !mdHrRe.MatchString(lines[i]) {
				runs = append(runs, docxRuns(strings.TrimSpace(lines[i]), false))
				i++
			}
			if len(runs) == 0 {
				runs = append(runs, docxRuns(strings.TrimSpace(l), false))
				i++
			}
			b.WriteString(docxPara("", strings.Join(runs, "<w:r><w:br/></w:r>"), ""))
		}
	}
	return b.String()
}

const emuPerPx = 9525

// lebar isi halaman A4 dalam twip (11906 − margin kiri 1418 − kanan 1134); lebar tabel absolut lebih andal daripada persen
const docxLebar = 9354

// docxKop: kop sekolah (logo kiri bila ada, nama & alamat), diakhiri garis tebal.
func docxKop(o docxOpts, logoRel string, logoW, logoH int) string {
	text := docxPara("KopNama", docxRuns(o.NamaSek, true), "")
	alamat := strings.TrimSpace(o.AlamatSek)
	if o.TeleponSek != "" {
		if alamat != "" {
			alamat += " · "
		}
		alamat += "Telp. " + o.TeleponSek
	}
	if alamat != "" {
		text += docxPara("KopAlamat", docxRuns(alamat, false), "")
	}
	line := `<w:tblBorders><w:bottom w:val="thick" w:sz="18" w:color="000000"/></w:tblBorders>`
	logoCell := ""
	if logoRel != "" {
		// logo dimuat di kotak 80×64 px (kolom kop 1300 twip) tanpa mengubah perbandingan sisi
		w, h := 80, 80*logoH/max(1, logoW)
		if h > 64 {
			w, h = 64*logoW/max(1, logoH), 64
		}
		cx, cy := w*emuPerPx, h*emuPerPx
		logoCell = `<w:tc><w:tcPr><w:tcW w:w="1300" w:type="dxa"/><w:vAlign w:val="center"/></w:tcPr><w:p><w:r><w:drawing><wp:inline distT="0" distB="0" distL="0" distR="0">` +
			fmt.Sprintf(`<wp:extent cx="%d" cy="%d"/><wp:docPr id="1" name="Logo"/>`, cx, cy) +
			`<a:graphic xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main"><a:graphicData uri="http://schemas.openxmlformats.org/drawingml/2006/picture"><pic:pic xmlns:pic="http://schemas.openxmlformats.org/drawingml/2006/picture">` +
			`<pic:nvPicPr><pic:cNvPr id="0" name="logo"/><pic:cNvPicPr/></pic:nvPicPr><pic:blipFill><a:blip r:embed="` + logoRel + `"/><a:stretch><a:fillRect/></a:stretch></pic:blipFill>` +
			fmt.Sprintf(`<pic:spPr><a:xfrm><a:off x="0" y="0"/><a:ext cx="%d" cy="%d"/></a:xfrm><a:prstGeom prst="rect"><a:avLst/></a:prstGeom></pic:spPr>`, cx, cy) +
			`</pic:pic></a:graphicData></a:graphic></wp:inline></w:drawing></w:r></w:p></w:tc>`
	}
	kolLogo := 0
	if logoCell != "" {
		kolLogo = 1300
	}
	grid := fmt.Sprintf(`<w:tblGrid><w:gridCol w:w="%d"/><w:gridCol w:w="%d"/></w:tblGrid>`, kolLogo, docxLebar-kolLogo)
	if logoCell == "" {
		grid = fmt.Sprintf(`<w:tblGrid><w:gridCol w:w="%d"/></w:tblGrid>`, docxLebar)
	}
	return fmt.Sprintf(`<w:tbl><w:tblPr><w:tblW w:w="%d" w:type="dxa"/><w:tblLayout w:type="fixed"/>`, docxLebar) + line + `</w:tblPr>` + grid + `<w:tr>` + logoCell +
		fmt.Sprintf(`<w:tc><w:tcPr><w:tcW w:w="%d" w:type="dxa"/><w:vAlign w:val="center"/></w:tcPr>`, docxLebar-kolLogo) + text + `</w:tc></w:tr></w:tbl>` + docxPara("", "", "")
}

// docxTTD: Mengetahui Kepala Sekolah (kiri) & Guru Mata Pelajaran (kanan), tanpa garis tabel.
func docxTTD(o docxOpts) string {
	or := func(s, def string) string {
		if strings.TrimSpace(s) == "" {
			return def
		}
		return s
	}
	nip := func(n string) string { return "NIP. " + or(n, "……………………") }
	col := func(atas1, atas2, nama, nipv string) string {
		p := func(t string, bold, underline bool) string {
			rpr := ""
			if bold {
				rpr += "<w:b/>"
			}
			if underline {
				rpr += `<w:u w:val="single"/>`
			}
			return docxPara("TTD", `<w:r><w:rPr>`+rpr+`</w:rPr><w:t xml:space="preserve">`+xmlEsc(t)+`</w:t></w:r>`, "")
		}
		return fmt.Sprintf(`<w:tc><w:tcPr><w:tcW w:w="%d" w:type="dxa"/></w:tcPr>`, docxLebar/2) + p(atas1, false, false) + p(atas2, false, false) +
			p("", false, false) + p("", false, false) + p("", false, false) + p(or(nama, "……………………………"), true, true) + p(nip(nipv), false, false) + `</w:tc>`
	}
	tgl := or(o.Tempat, "……………") + ", " + or(o.Tanggal, "……………")
	return docxPara("", "", "") + fmt.Sprintf(`<w:tbl><w:tblPr><w:tblW w:w="%d" w:type="dxa"/><w:tblLayout w:type="fixed"/><w:tblBorders>`, docxLebar) + `<w:top w:val="nil"/><w:left w:val="nil"/><w:bottom w:val="nil"/><w:right w:val="nil"/><w:insideH w:val="nil"/><w:insideV w:val="nil"/></w:tblBorders></w:tblPr>` +
		fmt.Sprintf(`<w:tblGrid><w:gridCol w:w="%d"/><w:gridCol w:w="%d"/></w:tblGrid>`, docxLebar/2, docxLebar-docxLebar/2) + `<w:tr><w:trPr><w:cantSplit/></w:trPr>` + col("Mengetahui,", "Kepala "+or(o.NamaSek, "Sekolah"), o.Kepsek, o.NIPKepsek) +
		col(tgl, "Guru Mata Pelajaran", o.Guru, o.NIPGuru) + `</w:tr></w:tbl>`
}

const docxStyles = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:styles xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
<w:docDefaults><w:rPrDefault><w:rPr><w:rFonts w:ascii="Calibri" w:hAnsi="Calibri" w:cs="Calibri"/><w:sz w:val="22"/><w:lang w:val="id-ID"/></w:rPr></w:rPrDefault>
<w:pPrDefault><w:pPr><w:spacing w:after="100" w:line="276" w:lineRule="auto"/></w:pPr></w:pPrDefault></w:docDefaults>
<w:style w:type="paragraph" w:default="1" w:styleId="Normal"><w:name w:val="Normal"/></w:style>
<w:style w:type="paragraph" w:styleId="Heading1"><w:name w:val="heading 1"/><w:basedOn w:val="Normal"/><w:next w:val="Normal"/><w:pPr><w:keepNext/><w:spacing w:before="240" w:after="120"/><w:outlineLvl w:val="0"/></w:pPr><w:rPr><w:b/><w:sz w:val="30"/></w:rPr></w:style>
<w:style w:type="paragraph" w:styleId="Heading2"><w:name w:val="heading 2"/><w:basedOn w:val="Normal"/><w:next w:val="Normal"/><w:pPr><w:keepNext/><w:spacing w:before="200" w:after="80"/><w:outlineLvl w:val="1"/></w:pPr><w:rPr><w:b/><w:sz w:val="26"/></w:rPr></w:style>
<w:style w:type="paragraph" w:styleId="Heading3"><w:name w:val="heading 3"/><w:basedOn w:val="Normal"/><w:next w:val="Normal"/><w:pPr><w:keepNext/><w:spacing w:before="160" w:after="60"/><w:outlineLvl w:val="2"/></w:pPr><w:rPr><w:b/><w:sz w:val="23"/></w:rPr></w:style>
<w:style w:type="paragraph" w:styleId="Daftar"><w:name w:val="Daftar"/><w:basedOn w:val="Normal"/><w:pPr><w:spacing w:after="40"/></w:pPr></w:style>
<w:style w:type="paragraph" w:styleId="TabelIsi"><w:name w:val="Tabel Isi"/><w:basedOn w:val="Normal"/><w:pPr><w:spacing w:before="20" w:after="20" w:line="240" w:lineRule="auto"/></w:pPr><w:rPr><w:sz w:val="20"/></w:rPr></w:style>
<w:style w:type="paragraph" w:styleId="Kode"><w:name w:val="Kode"/><w:basedOn w:val="Normal"/><w:pPr><w:spacing w:after="0"/></w:pPr><w:rPr><w:rFonts w:ascii="Consolas" w:hAnsi="Consolas"/><w:sz w:val="19"/></w:rPr></w:style>
<w:style w:type="paragraph" w:styleId="KopNama"><w:name w:val="Kop Nama"/><w:basedOn w:val="Normal"/><w:pPr><w:jc w:val="center"/><w:spacing w:after="20"/></w:pPr><w:rPr><w:b/><w:sz w:val="32"/></w:rPr></w:style>
<w:style w:type="paragraph" w:styleId="KopAlamat"><w:name w:val="Kop Alamat"/><w:basedOn w:val="Normal"/><w:pPr><w:jc w:val="center"/><w:spacing w:after="40"/></w:pPr><w:rPr><w:sz w:val="20"/></w:rPr></w:style>
<w:style w:type="paragraph" w:styleId="TTD"><w:name w:val="Tanda Tangan"/><w:basedOn w:val="Normal"/><w:pPr><w:jc w:val="center"/><w:spacing w:after="0"/></w:pPr></w:style>
</w:styles>`

// buildDocx menyusun file .docx lengkap.
func buildDocx(md string, o docxOpts) ([]byte, error) {
	var logo []byte
	var logoExt string
	var lw, lh int
	if o.Kop && o.LogoPath != "" {
		if b, err := os.ReadFile(o.LogoPath); err == nil {
			if cfg, format, err := image.DecodeConfig(bytes.NewReader(b)); err == nil && (format == "png" || format == "jpeg") {
				logo, lw, lh = b, cfg.Width, cfg.Height
				logoExt = map[string]string{"png": "png", "jpeg": "jpeg"}[format]
			}
		}
	}
	body := ""
	rel := ""
	if logo != nil {
		rel = "rIdLogo"
	}
	if o.Kop {
		body += docxKop(o, rel, lw, lh)
	}
	body += docxBody(md)
	if o.TTD {
		body += docxTTD(o)
	}
	doc := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships" xmlns:wp="http://schemas.openxmlformats.org/drawingml/2006/wordprocessingDrawing" xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main" xmlns:pic="http://schemas.openxmlformats.org/drawingml/2006/picture"><w:body>` +
		body + `<w:sectPr><w:pgSz w:w="11906" w:h="16838"/><w:pgMar w:top="1134" w:right="1134" w:bottom="1134" w:left="1418" w:header="567" w:footer="567" w:gutter="0"/></w:sectPr></w:body></w:document>`

	ct := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types"><Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/><Default Extension="xml" ContentType="application/xml"/><Default Extension="png" ContentType="image/png"/><Default Extension="jpeg" ContentType="image/jpeg"/>` +
		`<Override PartName="/word/document.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"/><Override PartName="/word/styles.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.styles+xml"/><Override PartName="/docProps/core.xml" ContentType="application/vnd.openxmlformats-package.core-properties+xml"/></Types>`
	rels := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"><Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="word/document.xml"/><Relationship Id="rId2" Type="http://schemas.openxmlformats.org/package/2006/relationships/metadata/core-properties" Target="docProps/core.xml"/></Relationships>`
	docRels := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"><Relationship Id="rIdStyles" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/styles" Target="styles.xml"/>`
	if logo != nil {
		docRels += `<Relationship Id="rIdLogo" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/image" Target="media/logo.` + logoExt + `"/>`
	}
	docRels += `</Relationships>`
	core := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<cp:coreProperties xmlns:cp="http://schemas.openxmlformats.org/package/2006/metadata/core-properties" xmlns:dc="http://purl.org/dc/elements/1.1/" xmlns:dcterms="http://purl.org/dc/terms/" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">` +
		`<dc:title>` + xmlEsc(o.Judul) + `</dc:title><dc:creator>` + xmlEsc(o.Guru) + `</dc:creator><dcterms:created xsi:type="dcterms:W3CDTF">` + time.Now().UTC().Format(time.RFC3339) + `</dcterms:created></cp:coreProperties>`

	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	files := []struct {
		name string
		data []byte
	}{{"[Content_Types].xml", []byte(ct)}, {"_rels/.rels", []byte(rels)}, {"word/document.xml", []byte(doc)},
		{"word/styles.xml", []byte(docxStyles)}, {"word/_rels/document.xml.rels", []byte(docRels)}, {"docProps/core.xml", []byte(core)}}
	if logo != nil {
		files = append(files, struct {
			name string
			data []byte
		}{"word/media/logo." + logoExt, logo})
	}
	for _, f := range files {
		w, err := zw.Create(f.name)
		if err != nil {
			return nil, err
		}
		if _, err := w.Write(f.data); err != nil {
			return nil, err
		}
	}
	if err := zw.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// localLogo: path file logo sekolah bila tersimpan di server (/files/<nama>.png|jpg).
func (a *App) localLogo(logoURL string) string {
	name := strings.TrimPrefix(logoURL, "/files/")
	if !fileNameRe.MatchString(name) || strings.HasSuffix(name, ".pdf") {
		return ""
	}
	return filepath.Join(a.uploadDir(), name)
}

var bulanIndo = [...]string{"", "Januari", "Februari", "Maret", "April", "Mei", "Juni", "Juli", "Agustus", "September", "Oktober", "November", "Desember"}

// tanggalPanjang: "2026-11-10" → "10 November 2026"
func tanggalPanjang(d string) string {
	t, err := time.Parse("2006-01-02", d)
	if err != nil {
		return ""
	}
	return fmt.Sprintf("%d %s %d", t.Day(), bulanIndo[t.Month()], t.Year())
}

// apiAIDocDocx: unduh dokumen AI sebagai Word (.docx), dengan kop & tanda tangan sesuai pilihan di layar cetak.
func apiAIDocDocx(c *Ctx) error {
	var judul, output string
	if err := c.A.reader().QueryRow(`SELECT judul, output FROM ai_documents WHERE id=? AND user_id=?`, c.ID("id"), c.U.ID).Scan(&judul, &output); err != nil {
		return notFound()
	}
	q := c.R.URL.Query()
	s, err := c.A.settingsData()
	if err != nil {
		return err
	}
	str := func(k string) string { v, _ := s[k].(string); return v }
	cut := func(v string, n int) string {
		if r := []rune(strings.TrimSpace(v)); len(r) > n {
			return string(r[:n])
		}
		return strings.TrimSpace(v)
	}
	o := docxOpts{Judul: judul, Kop: q.Get("kop") != "0", NamaSek: str("nama_sekolah"), AlamatSek: str("alamat"), TeleponSek: str("telepon"),
		LogoPath: c.A.localLogo(str("logo_url")), TTD: q.Get("ttd") == "1", Tempat: cut(q.Get("tempat"), 60), Tanggal: tanggalPanjang(q.Get("tanggal")),
		Guru: cut(q.Get("guru"), 100), NIPGuru: cut(q.Get("nip_guru"), 30), Kepsek: cut(q.Get("kepsek"), 100), NIPKepsek: cut(q.Get("nip_kepsek"), 30)}
	b, err := buildDocx(output, o)
	if err != nil {
		return err
	}
	name := strings.Map(func(r rune) rune {
		if strings.ContainsRune(`/\:*?"<>|`, r) || r < 32 {
			return '-'
		}
		return r
	}, judul)
	if len([]rune(name)) > 100 {
		name = string([]rune(name)[:100])
	}
	h := c.W.Header()
	h.Set("Content-Type", "application/vnd.openxmlformats-officedocument.wordprocessingml.document")
	h.Set("Content-Disposition", `attachment; filename="dokumen.docx"; filename*=UTF-8''`+url.PathEscape(name+".docx"))
	h.Set("Cache-Control", "no-store")
	_, err = c.W.Write(b)
	return err
}
