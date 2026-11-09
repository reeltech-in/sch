package main

import "testing"

// alamat e-book Kemendikbud: hanya https ke cermin resmi; host lama dialihkan; selain itu ditolak (cegah SSRF)
func TestSibiURL(t *testing.T) {
	cases := map[string]string{
		"https://static-sc.cloudapp.web.id/content/pdf/a.pdf":        "https://static-sc.cloudapp.web.id/content/pdf/a.pdf",
		"https://static.buku.kemdikbud.go.id/content/pdf/a.pdf":      "https://static-sc.cloudapp.web.id/content/pdf/a.pdf",
		"https://static.buku.kemendikdasmen.go.id/content/pdf/a.pdf": "https://static-sc.cloudapp.web.id/content/pdf/a.pdf",
		"http://static-sc.cloudapp.web.id/content/pdf/a.pdf":         "",
		"https://127.0.0.1/a.pdf":                                    "",
		"https://evil.example/static-sc.cloudapp.web.id/a.pdf":       "",
		"https://static-sc.cloudapp.web.id.evil.example/a.pdf":       "",
		"": "",
	}
	for in, want := range cases {
		if got := sibiURL(in); got != want {
			t.Errorf("sibiURL(%q) = %q, mau %q", in, got, want)
		}
	}
	if cacheName("x", ".pdf") == cacheName("y", ".pdf") || len(cacheName("x", ".pdf")) != 36 {
		t.Error("nama cache harus unik per alamat")
	}
}
