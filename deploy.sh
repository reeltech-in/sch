#!/usr/bin/env bash
# Rilis SCH ke server Linux (systemd) dengan nomor versi dari file VERSION.
#
#   ./deploy.sh                 # rilis versi di file VERSION
#
# Langkah:
#   1. build linux/amd64 dengan versi tertanam
#   2. backup database di server (konsisten, lewat sch-server -backup)
#   3. unggah ke /opt/sch/releases/sch-server-v<versi>, arahkan /opt/sch/sch-server ke rilis itu
#   4. restart layanan dan cek /api/version; bila gagal, otomatis kembali ke rilis sebelumnya
#
# Alamat server dibaca dari file deploy.env (tidak ikut di-commit; contoh: deploy.env.example)
# atau variabel lingkungan: SCH_HOST=root@IP-SERVER  SCH_KEY=~/.ssh/kunci_ssh  SCH_PORT=3220
set -euo pipefail
cd "$(dirname "$0")"
[ -f deploy.env ] && . ./deploy.env

HOST="${SCH_HOST:?isi SCH_HOST di deploy.env (salin dari deploy.env.example)}"
KEY="${SCH_KEY:-$HOME/.ssh/id_ed25519}"
PORT="${SCH_PORT:-3220}"
VER="$(tr -d '[:space:]' < VERSION)"
BUILD="$(date -u +%Y%m%d%H%M)"
REL="sch-server-v${VER}"
SSH=(ssh -i "$KEY" -o IdentitiesOnly=yes -o BatchMode=yes "$HOST")

if ! grep -q "^## v${VER} " CHANGELOG.md; then
  echo "Tambahkan catatan '## v${VER} — tanggal' di CHANGELOG.md terlebih dahulu." >&2
  exit 1
fi

echo "== build SCH v${VER} (build ${BUILD})"
go vet ./...
echo "== tes otomatis"
go test -count=1 ./... >/dev/null || { echo "TES GAGAL: rilis dibatalkan. Jalankan: go test ./..."; exit 1; }
mkdir -p dist
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath \
  -ldflags "-s -w -X main.version=${VER} -X main.buildTime=${BUILD}" -o "dist/${REL}" .

echo "== unggah & backup database"
"${SSH[@]}" "mkdir -p /opt/sch/releases /opt/sch/data/backups"
scp -i "$KEY" -o IdentitiesOnly=yes -q "dist/${REL}" "${HOST}:/opt/sch/releases/${REL}.new"
scp -i "$KEY" -o IdentitiesOnly=yes -q README.md "${HOST}:/opt/sch/PANDUAN.md"
scp -i "$KEY" -o IdentitiesOnly=yes -q CHANGELOG.md "${HOST}:/opt/sch/CHANGELOG.md"
# backup dibuat dengan biner BARU (punya pilihan -backup), sebelum layanan diganti
"${SSH[@]}" "set -e; chmod 755 /opt/sch/releases/${REL}.new
  if [ -f /opt/sch/data/sch.db ]; then
    /opt/sch/releases/${REL}.new -db /opt/sch/data/sch.db -backup /opt/sch/data/backups/sebelum-v${VER}-${BUILD}.db
  fi"

echo "== aktifkan rilis"
"${SSH[@]}" "set -e; cd /opt/sch
  PREV=\$(readlink -f sch-server 2>/dev/null || true)
  # migrasi satu kali: biner lama (bukan symlink) disimpan sebagai rilis
  if [ -f sch-server ] && [ ! -L sch-server ]; then mv sch-server releases/sch-server-sebelumnya; PREV=/opt/sch/releases/sch-server-sebelumnya; fi
  mv -f releases/${REL}.new releases/${REL}; chmod 755 releases/${REL}
  ln -sfn releases/${REL} sch-server
  chown -R sch:sch /opt/sch
  systemctl restart sch
  for i in \$(seq 1 60); do
    sleep 1
    if curl -fs http://127.0.0.1:${PORT}/api/version | grep -q '\"version\":\"${VER}\"'; then echo 'SEHAT: v${VER} berjalan'; exit 0; fi
  done
  echo 'GAGAL: v${VER} tidak merespons, kembali ke rilis sebelumnya' >&2
  if [ -n \"\$PREV\" ]; then ln -sfn \"\$PREV\" sch-server; systemctl restart sch; fi
  journalctl -u sch -n 20 --no-pager -o cat | grep -iv password >&2
  exit 1"

# simpan 5 rilis terakhir saja
"${SSH[@]}" "cd /opt/sch/releases && ls -1t sch-server-v* 2>/dev/null | tail -n +6 | xargs -r rm -f --"
echo "Selesai: SCH v${VER} aktif di ${HOST}"
