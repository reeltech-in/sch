package main

import (
	"database/sql"
	"fmt"
)

// Indeks untuk query yang tumbuh bersama data (bertahun-tahun absensi, nilai, log). Dibuat lewat moduleMigrations
// agar berjalan setelah semua tabel modul dibuat; CREATE INDEX IF NOT EXISTS aman dijalankan setiap start.
// Catatan: filter bulanan memakai rentang (tanggal >= ? AND tanggal <= ?), bukan LIKE '2026-11%', karena
// SQLite hanya memakai indeks untuk LIKE bila kolomnya ber-COLLATE NOCASE (dokumentasi "The LIKE Optimization").

func init() {
	moduleMigrations = append(moduleMigrations, func(db *sql.DB) error {
		for _, q := range []string{
			`CREATE INDEX IF NOT EXISTS idx_attend_course_tgl ON attendance(course_id, tanggal, status)`,
			`CREATE INDEX IF NOT EXISTS idx_attend_siswa_tgl ON attendance(siswa_id, tanggal)`,
			`DROP INDEX IF EXISTS idx_attend_siswa`, // digantikan idx_attend_siswa_tgl
			`CREATE INDEX IF NOT EXISTS idx_progress_siswa ON material_progress(siswa_id)`,
			`CREATE INDEX IF NOT EXISTS idx_journal_tgl ON teaching_journal(tanggal)`,
			`CREATE INDEX IF NOT EXISTS idx_notes_tgl ON student_notes(tanggal)`,
			`CREATE INDEX IF NOT EXISTS idx_guestbook_created ON guestbook(created_at)`,
			`CREATE INDEX IF NOT EXISTS idx_sessions_user ON sessions(user_id)`,
			`CREATE INDEX IF NOT EXISTS idx_wa_created ON wa_outbox(created_at)`,
			`CREATE INDEX IF NOT EXISTS idx_daily_foto ON daily_attendance(tanggal) WHERE foto IS NOT NULL`,
			// kunci tamu (foreign key) yang dipindai SQLite saat pengguna dihapus (ON DELETE SET NULL/CASCADE);
			// tanpa indeks, menghapus satu guru memindai ratusan ribu baris absensi
			`CREATE INDEX IF NOT EXISTS idx_daily_petugas ON daily_attendance(petugas_id) WHERE petugas_id IS NOT NULL`,
			`CREATE INDEX IF NOT EXISTS idx_loans_petugas ON loans(petugas_id) WHERE petugas_id IS NOT NULL`,
			`CREATE INDEX IF NOT EXISTS idx_notes_created_by ON student_notes(created_by)`,
			`CREATE INDEX IF NOT EXISTS idx_exams_created_by ON exams(created_by)`,
			// cari buku per ISBN (pinjam dengan scan ISBN, cegah duplikat impor e-book)
			`CREATE INDEX IF NOT EXISTS idx_books_isbn ON books(isbn) WHERE isbn IS NOT NULL`,
			// jumlah peserta selesai per ujian (daftar CBT guru/admin) tanpa membaca baris jawaban
			`CREATE INDEX IF NOT EXISTS idx_exam_att_done ON exam_attempts(exam_id) WHERE selesai=1`,
		} {
			if _, err := db.Exec(q); err != nil {
				return fmt.Errorf("%s: %w", q, err)
			}
		}
		return nil
	})
}

// memberCount: jumlah siswa sebuah kelas (tahun berjalan + arsip), setara memberOf tetapi bisa memakai indeks
// (users.kelas_id & class_members.kelas_id) alih-alih memindai seluruh tabel users per baris.
func memberCount(kelasExpr string) string {
	return fmt.Sprintf(`((SELECT COUNT(*) FROM users mu WHERE mu.kelas_id=%[1]s AND mu.role='siswa' AND mu.status='aktif')
		+ (SELECT COUNT(*) FROM class_members mm JOIN users mu ON mu.id=mm.siswa_id AND mu.role='siswa'
		   WHERE mm.kelas_id=%[1]s AND NOT (mu.kelas_id IS %[1]s AND mu.status='aktif')))`, kelasExpr)
}
