package main

import (
	"archive/zip"
	"bytes"
	"fmt"
	"math"
	"strconv"
	"strings"
)

// Pembuat lembar kerja Excel (.xlsx) tanpa pustaka luar: .xlsx adalah arsip zip berisi XML (SpreadsheetML).
// Cukup untuk ekspor tabel: baris pertama judul kolom (tebal, dibekukan), teks disimpan sebagai teks
// (NISN "0012…" tidak kehilangan nol di depan), angka sebagai angka.

type xlsxSheet struct {
	Nama  string
	Lebar []float64 // lebar kolom dalam karakter; kosong = bawaan Excel
	Baris [][]any   // baris pertama = judul kolom; nilai: string, int, int64, float64, nil (sel kosong)
}

// gaya sel (urutan cellXfs di styles.xml)
const (
	xsJudul = 1
	xsTeks  = 2
	xsAngka = 3
)

const xlsxStyles = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<styleSheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main">
<fonts count="2"><font><sz val="11"/><name val="Calibri"/><family val="2"/></font><font><b/><sz val="11"/><name val="Calibri"/><family val="2"/></font></fonts>
<fills count="3"><fill><patternFill patternType="none"/></fill><fill><patternFill patternType="gray125"/></fill><fill><patternFill patternType="solid"><fgColor rgb="FFE8EEF7"/><bgColor indexed="64"/></patternFill></fill></fills>
<borders count="2"><border><left/><right/><top/><bottom/><diagonal/></border><border><left style="thin"><color rgb="FFBFBFBF"/></left><right style="thin"><color rgb="FFBFBFBF"/></right><top style="thin"><color rgb="FFBFBFBF"/></top><bottom style="thin"><color rgb="FFBFBFBF"/></bottom><diagonal/></border></borders>
<cellStyleXfs count="1"><xf numFmtId="0" fontId="0" fillId="0" borderId="0"/></cellStyleXfs>
<cellXfs count="4">
<xf numFmtId="0" fontId="0" fillId="0" borderId="0" xfId="0"/>
<xf numFmtId="0" fontId="1" fillId="2" borderId="1" xfId="0" applyFont="1" applyFill="1" applyBorder="1" applyAlignment="1"><alignment vertical="center" wrapText="1"/></xf>
<xf numFmtId="49" fontId="0" fillId="0" borderId="1" xfId="0" applyNumberFormat="1" applyBorder="1" applyAlignment="1"><alignment vertical="top" wrapText="1"/></xf>
<xf numFmtId="0" fontId="0" fillId="0" borderId="1" xfId="0" applyBorder="1" applyAlignment="1"><alignment vertical="top"/></xf>
</cellXfs>
<cellStyles count="1"><cellStyle name="Normal" xfId="0" builtinId="0"/></cellStyles>
</styleSheet>`

// kolomXlsx: 0 → A, 25 → Z, 26 → AA
func kolomXlsx(i int) string {
	s := ""
	for i >= 0 {
		s = string(rune('A'+i%26)) + s
		i = i/26 - 1
	}
	return s
}

// namaSheetXlsx: nama lembar yang sah di Excel (tanpa []:*?/\, maks. 31 karakter, unik tanpa beda huruf besar/kecil).
func namaSheetXlsx(nama string, dipakai map[string]bool) string {
	nama = strings.Map(func(r rune) rune {
		if strings.ContainsRune(`[]:*?/\`, r) || r < 32 {
			return ' '
		}
		return r
	}, nama)
	nama = strings.Trim(strings.Join(strings.Fields(nama), " "), "'")
	if nama == "" {
		nama = "Lembar"
	}
	potong := func(s string, n int) string {
		if r := []rune(s); len(r) > n {
			return strings.TrimSpace(string(r[:n]))
		}
		return s
	}
	hasil := potong(nama, 31)
	for i := 2; dipakai[strings.ToLower(hasil)]; i++ {
		akhir := fmt.Sprintf(" (%d)", i)
		hasil = potong(nama, 31-len(akhir)) + akhir
	}
	dipakai[strings.ToLower(hasil)] = true
	return hasil
}

func angkaXlsx(f float64) string {
	if math.IsNaN(f) || math.IsInf(f, 0) {
		return ""
	}
	return strconv.FormatFloat(f, 'f', -1, 64)
}

func sheetXML(sh xlsxSheet) []byte {
	var b bytes.Buffer
	b.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>` + "\n")
	b.WriteString(`<worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships">`)
	if len(sh.Baris) > 1 { // judul kolom tetap terlihat saat digulir
		b.WriteString(`<sheetViews><sheetView workbookViewId="0"><pane ySplit="1" topLeftCell="A2" activePane="bottomLeft" state="frozen"/></sheetView></sheetViews>`)
	}
	if len(sh.Lebar) > 0 {
		b.WriteString(`<cols>`)
		for i, w := range sh.Lebar {
			if w > 0 {
				fmt.Fprintf(&b, `<col min="%d" max="%d" width="%s" customWidth="1"/>`, i+1, i+1, angkaXlsx(w))
			}
		}
		b.WriteString(`</cols>`)
	}
	b.WriteString(`<sheetData>`)
	for ri, row := range sh.Baris {
		fmt.Fprintf(&b, `<row r="%d">`, ri+1)
		for ci, v := range row {
			ref := kolomXlsx(ci) + strconv.Itoa(ri+1)
			gaya := xsTeks
			if ri == 0 {
				gaya = xsJudul
			}
			var angka string
			switch x := v.(type) {
			case nil:
				fmt.Fprintf(&b, `<c r="%s" s="%d"/>`, ref, gaya)
				continue
			case int:
				angka = strconv.Itoa(x)
			case int64:
				angka = strconv.FormatInt(x, 10)
			case float64:
				angka = angkaXlsx(x)
			case string:
				fmt.Fprintf(&b, `<c r="%s" t="inlineStr" s="%d"><is><t xml:space="preserve">%s</t></is></c>`, ref, gaya, xmlEsc(x))
				continue
			default:
				fmt.Fprintf(&b, `<c r="%s" t="inlineStr" s="%d"><is><t xml:space="preserve">%s</t></is></c>`, ref, gaya, xmlEsc(fmt.Sprint(x)))
				continue
			}
			if ri > 0 {
				gaya = xsAngka
			}
			if angka == "" {
				fmt.Fprintf(&b, `<c r="%s" s="%d"/>`, ref, gaya)
			} else {
				fmt.Fprintf(&b, `<c r="%s" s="%d"><v>%s</v></c>`, ref, gaya, angka)
			}
		}
		b.WriteString(`</row>`)
	}
	b.WriteString(`</sheetData></worksheet>`)
	return b.Bytes()
}

// buildXlsx: satu berkas Excel dengan satu atau lebih lembar.
func buildXlsx(sheets []xlsxSheet) ([]byte, error) {
	if len(sheets) == 0 {
		sheets = []xlsxSheet{{Nama: "Lembar1"}}
	}
	var buf bytes.Buffer
	z := zip.NewWriter(&buf)
	put := func(name string, data []byte) error {
		w, err := z.Create(name)
		if err != nil {
			return err
		}
		_, err = w.Write(data)
		return err
	}
	var ct, wb, rel bytes.Buffer
	ct.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types"><Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/><Default Extension="xml" ContentType="application/xml"/><Override PartName="/xl/workbook.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.sheet.main+xml"/><Override PartName="/xl/styles.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.styles+xml"/>`)
	wb.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<workbook xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships"><sheets>`)
	rel.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">`)
	dipakai := map[string]bool{}
	for i, sh := range sheets {
		n := i + 1
		fmt.Fprintf(&ct, `<Override PartName="/xl/worksheets/sheet%d.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.worksheet+xml"/>`, n)
		fmt.Fprintf(&wb, `<sheet name="%s" sheetId="%d" r:id="rId%d"/>`, xmlEsc(namaSheetXlsx(sh.Nama, dipakai)), n, n)
		fmt.Fprintf(&rel, `<Relationship Id="rId%d" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/worksheet" Target="worksheets/sheet%d.xml"/>`, n, n)
	}
	ct.WriteString(`</Types>`)
	wb.WriteString(`</sheets></workbook>`)
	fmt.Fprintf(&rel, `<Relationship Id="rId%d" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/styles" Target="styles.xml"/></Relationships>`, len(sheets)+1)
	for _, f := range []struct {
		name string
		data []byte
	}{
		{"[Content_Types].xml", ct.Bytes()},
		{"_rels/.rels", []byte(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"><Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="xl/workbook.xml"/></Relationships>`)},
		{"xl/workbook.xml", wb.Bytes()},
		{"xl/_rels/workbook.xml.rels", rel.Bytes()},
		{"xl/styles.xml", []byte(xlsxStyles)},
	} {
		if err := put(f.name, f.data); err != nil {
			return nil, err
		}
	}
	for i, sh := range sheets { // [Content_Types].xml di urutan pertama arsip, lembar sesudahnya
		if err := put(fmt.Sprintf("xl/worksheets/sheet%d.xml", i+1), sheetXML(sh)); err != nil {
			return nil, err
		}
	}
	if err := z.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// namaBerkas: nama file unduhan tanpa karakter yang dilarang Windows/macOS.
func namaBerkas(s string) string {
	return strings.Map(func(x rune) rune {
		if strings.ContainsRune(`/\:*?"<>|`, x) || x < 32 {
			return '-'
		}
		return x
	}, s)
}
