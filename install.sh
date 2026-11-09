#!/usr/bin/env bash
# Pasang / perbarui SCH (Sistem Informasi Sekolah) di server Ubuntu atau Debian.
#
#   Pasang baru (repo publik):
#     curl -fsSL https://raw.githubusercontent.com/reeltech-in/sch/main/install.sh | sudo bash
#
#   Dari folder hasil git clone (repo privat juga bisa):
#     sudo ./install.sh
#
#   Perbarui ke versi terbaru: jalankan perintah yang sama lagi. Database dibackup dulu,
#   dan bila versi baru gagal jalan, otomatis kembali ke versi sebelumnya.
#
#   Hapus layanan (data TIDAK dihapus):  sudo ./install.sh --uninstall
#
# Semua pengaturan lewat variabel lingkungan (opsional), contoh:
#   curl -fsSL .../install.sh | sudo SCH_PORT=8080 SCH_PROXY_IP=10.0.0.5 bash
#
#   SCH_DIR=/opt/sch          folder program & data
#   SCH_SERVICE=sch           nama layanan systemd
#   SCH_USER=sch              user sistem yang menjalankan program
#   SCH_PORT=3220             port HTTP
#   SCH_HTTPS_PORT=3221       port HTTPS lokal (kamera HP di jaringan sekolah); kosongkan = mati
#   SCH_PROXY=0               1 = online lewat Cloudflare Tunnel / reverse proxy
#   SCH_PROXY_IP=             IP proxy yang dipercaya (mis. IP mesin cloudflared); otomatis SCH_PROXY=1
#   SCH_DEMO=0                1 = isi data contoh (hanya bila database masih kosong)
#   SCH_UFW=0                 1 = buka port di firewall ufw
#   SCH_UNIT=keep             rewrite = tulis ulang file layanan systemd saat memperbarui
#   SCH_TEST=0                1 = jalankan tes otomatis sebelum memasang
#   SCH_REPO=https://github.com/reeltech-in/sch.git   SCH_BRANCH=main
#   SCH_SRC=                  pakai kode sumber dari folder ini (tanpa git clone)
#   GO_ROOT=/usr/local/go     lokasi Go untuk build
set -euo pipefail

SCH_DIR="${SCH_DIR:-/opt/sch}"
SCH_SERVICE="${SCH_SERVICE:-sch}"
SCH_USER="${SCH_USER:-sch}"
SCH_PORT="${SCH_PORT:-3220}"
SCH_HTTPS_PORT="${SCH_HTTPS_PORT-3221}"
SCH_PROXY_IP="${SCH_PROXY_IP:-}"
SCH_PROXY="${SCH_PROXY:-0}"
SCH_DEMO="${SCH_DEMO:-0}"
SCH_UFW="${SCH_UFW:-0}"
SCH_UNIT="${SCH_UNIT:-keep}"
SCH_TEST="${SCH_TEST:-0}"
SCH_REPO="${SCH_REPO:-https://github.com/reeltech-in/sch.git}"
SCH_BRANCH="${SCH_BRANCH:-main}"
SCH_SRC="${SCH_SRC:-}"
GO_ROOT="${GO_ROOT:-/usr/local/go}"
UNIT_FILE="/etc/systemd/system/${SCH_SERVICE}.service"

info() { printf '\033[1;34m== %s\033[0m\n' "$*"; }
ok()   { printf '\033[1;32m%s\033[0m\n' "$*"; }
die()  { printf '\033[1;31mGAGAL: %s\033[0m\n' "$*" >&2; exit 1; }

[ "$(id -u)" -eq 0 ] || die "jalankan sebagai root, contoh: sudo ./install.sh"
command -v apt-get >/dev/null || die "skrip ini untuk Ubuntu/Debian (butuh apt-get)"
command -v systemctl >/dev/null || die "butuh systemd"
case "$SCH_PORT" in ''|*[!0-9]*) die "SCH_PORT harus angka";; esac
[ -n "$SCH_PROXY_IP" ] && SCH_PROXY=1

# ---------------------------------------------------------------- uninstall
if [ "${1:-}" = "--uninstall" ]; then
  info "menghapus layanan ${SCH_SERVICE}"
  systemctl disable --now "$SCH_SERVICE" 2>/dev/null || true
  rm -f "$UNIT_FILE"; systemctl daemon-reload
  rm -rf "${SCH_DIR}/releases" "${SCH_DIR}/sch-server" "${SCH_DIR}/src"
  ok "Layanan dihapus. Data (database, upload, backup) masih ada di ${SCH_DIR}/data."
  echo "Untuk menghapus SEMUA data secara permanen: sudo rm -rf ${SCH_DIR} && sudo userdel ${SCH_USER}"
  exit 0
fi

# ---------------------------------------------------------------- 1. dependensi
info "memasang dependensi sistem (apt)"
export DEBIAN_FRONTEND=noninteractive
apt-get update -qq
apt-get install -y -qq ca-certificates curl git tar sqlite3 >/dev/null

# ---------------------------------------------------------------- 2. kode sumber
is_src() { [ -f "$1/go.mod" ] && head -1 "$1/go.mod" | grep -qx 'module sch'; }
if [ -z "$SCH_SRC" ]; then
  here="$(cd "$(dirname "${BASH_SOURCE[0]:-/nonexistent}")" 2>/dev/null && pwd || true)"
  if [ -n "$here" ] && is_src "$here"; then
    SCH_SRC="$here"
  else
    SCH_SRC="${SCH_DIR}/src"
    if [ -d "$SCH_SRC/.git" ]; then
      info "mengambil kode terbaru (${SCH_BRANCH})"
      git -C "$SCH_SRC" fetch -q origin "$SCH_BRANCH"
      git -C "$SCH_SRC" reset -q --hard "origin/${SCH_BRANCH}"
    else
      info "mengunduh kode dari ${SCH_REPO}"
      mkdir -p "$SCH_DIR"
      GIT_TERMINAL_PROMPT=0 git clone -q --depth 1 -b "$SCH_BRANCH" "$SCH_REPO" "$SCH_SRC" ||
        die "git clone gagal. Repo privat? Clone dulu secara manual lalu jalankan: sudo ./install.sh"
    fi
  fi
fi
is_src "$SCH_SRC" || die "folder ${SCH_SRC} bukan kode sumber SCH"
VER="$(tr -d '[:space:]' < "$SCH_SRC/VERSION")"
BUILD="$(date -u +%Y%m%d%H%M)"
info "kode sumber: ${SCH_SRC} (SCH v${VER})"

# ---------------------------------------------------------------- 3. Go
GO_WANT="$(awk '$1=="go"{print $2; exit}' "$SCH_SRC/go.mod")"
case "$(uname -m)" in
  x86_64|amd64) GOARCH_DL=amd64;;
  aarch64|arm64) GOARCH_DL=arm64;;
  armv7l|armv6l) GOARCH_DL=armv6l;;
  *) die "arsitektur $(uname -m) belum didukung";;
esac
ver_ge() { [ "$(printf '%s\n%s\n' "$2" "$1" | sort -V | head -1)" = "$2" ]; }
GO_HAVE="$("$GO_ROOT/bin/go" env GOVERSION 2>/dev/null | sed 's/^go//' || true)"
if [ -z "$GO_HAVE" ] || ! ver_ge "$GO_HAVE" "$GO_WANT"; then
  info "memasang Go ${GO_WANT} ke ${GO_ROOT}"
  tgz="go${GO_WANT}.linux-${GOARCH_DL}.tar.gz"
  tmp="$(mktemp -d)"; trap 'rm -rf "$tmp"' EXIT
  curl -fsSL "https://go.dev/dl/${tgz}" -o "$tmp/$tgz" || die "unduh Go gagal (cek internet)"
  sum="$(curl -fsSL 'https://go.dev/dl/?mode=json&include=all' | tr -d ' \n' |
         grep -o "\"filename\":\"${tgz}\"[^}]*" | grep -o '"sha256":"[0-9a-f]*"' | cut -d'"' -f4 | head -1 || true)"
  [ -n "$sum" ] || die "checksum Go tidak ditemukan"
  echo "${sum}  $tmp/$tgz" | sha256sum -c --quiet - || die "checksum Go tidak cocok"
  rm -rf "$GO_ROOT"; mkdir -p "$(dirname "$GO_ROOT")"
  tar -C "$tmp" -xzf "$tmp/$tgz"; mv "$tmp/go" "$GO_ROOT"
fi
export PATH="$GO_ROOT/bin:$PATH" GOTOOLCHAIN=local
export GOPATH="${GOPATH:-/root/go}" GOCACHE="${GOCACHE:-/root/.cache/go-build}"
info "Go: $(go version | cut -d' ' -f3)"

# ---------------------------------------------------------------- 4. build
REL="sch-server-v${VER}"
mkdir -p "${SCH_DIR}/releases" "${SCH_DIR}/data/backups"
if [ "$SCH_TEST" = 1 ]; then
  info "tes otomatis"
  (cd "$SCH_SRC" && go test -count=1 ./... >/dev/null) || die "tes gagal, pemasangan dibatalkan"
fi
info "build ${REL}"
(cd "$SCH_SRC" && CGO_ENABLED=0 go build -trimpath \
  -ldflags "-s -w -X main.version=${VER} -X main.buildTime=${BUILD}" -o "${SCH_DIR}/releases/${REL}.new" .)
chmod 755 "${SCH_DIR}/releases/${REL}.new"

# ---------------------------------------------------------------- 5. user & folder
if ! id "$SCH_USER" >/dev/null 2>&1; then
  info "membuat user sistem ${SCH_USER}"
  useradd --system --home-dir "$SCH_DIR" --no-create-home --shell /usr/sbin/nologin "$SCH_USER"
fi
chmod 700 "${SCH_DIR}/data"

# ---------------------------------------------------------------- 6. backup bila memperbarui
if [ -f "${SCH_DIR}/data/sch.db" ]; then
  info "backup database sebelum memperbarui"
  "${SCH_DIR}/releases/${REL}.new" -db "${SCH_DIR}/data/sch.db" \
    -backup "${SCH_DIR}/data/backups/sebelum-v${VER}-${BUILD}.db"
fi

# ---------------------------------------------------------------- 7. layanan systemd
if [ ! -f "$UNIT_FILE" ] || [ "$SCH_UNIT" = rewrite ]; then
  info "menulis layanan ${UNIT_FILE}"
  args="-addr :${SCH_PORT} -https '${SCH_HTTPS_PORT:+:${SCH_HTTPS_PORT}}' -db ${SCH_DIR}/data/sch.db"
  [ "$SCH_PROXY" = 1 ] && args="$args -proxy"
  [ -n "$SCH_PROXY_IP" ] && args="$args -proxy-ip ${SCH_PROXY_IP}"
  [ "$SCH_DEMO" = 1 ] && args="$args -demo"
  cat > "$UNIT_FILE" <<EOF
[Unit]
Description=SCH Sistem Informasi Sekolah (${SCH_SERVICE})
After=network-online.target
Wants=network-online.target

[Service]
User=${SCH_USER}
Group=${SCH_USER}
WorkingDirectory=${SCH_DIR}
ExecStart=${SCH_DIR}/sch-server ${args}
Restart=always
RestartSec=3
LimitNOFILE=65536
NoNewPrivileges=true
ProtectSystem=strict
ProtectHome=true
PrivateTmp=true
ReadWritePaths=${SCH_DIR}/data

[Install]
WantedBy=multi-user.target
EOF
  systemctl daemon-reload
fi
grep -q -- "-addr :${SCH_PORT}\b" "$UNIT_FILE" || {
  p="$(grep -o -- '-addr :[0-9]*' "$UNIT_FILE" | cut -d: -f2)"; [ -n "$p" ] && SCH_PORT="$p"; }

# ---------------------------------------------------------------- 8. aktifkan rilis
info "mengaktifkan v${VER}"
cd "$SCH_DIR"
PREV="$(readlink -f sch-server 2>/dev/null || true)"
mv -f "releases/${REL}.new" "releases/${REL}"
ln -sfn "releases/${REL}" sch-server
cp -f "$SCH_SRC/README.md" PANDUAN.md 2>/dev/null || true
cp -f "$SCH_SRC/CHANGELOG.md" CHANGELOG.md 2>/dev/null || true
chown -R "${SCH_USER}:${SCH_USER}" "$SCH_DIR"
systemctl enable -q "$SCH_SERVICE"
systemctl restart "$SCH_SERVICE"
healthy=0
for _ in $(seq 1 60); do
  sleep 1
  resp="$(curl -fs "http://127.0.0.1:${SCH_PORT}/api/version" 2>/dev/null || true)"
  case "$resp" in *"\"version\":\"${VER}\""*) healthy=1; break;; esac
done
if [ "$healthy" != 1 ]; then
  journalctl -u "$SCH_SERVICE" -n 20 --no-pager -o cat | grep -iv password >&2 || true
  if [ -n "$PREV" ] && [ "$PREV" != "$(readlink -f sch-server)" ]; then
    ln -sfn "$PREV" sch-server; systemctl restart "$SCH_SERVICE"
    die "v${VER} tidak merespons; dikembalikan ke $(basename "$PREV")"
  fi
  die "v${VER} tidak merespons di port ${SCH_PORT}. Cek: journalctl -u ${SCH_SERVICE} -n 50"
fi
(cd releases && ls -1t sch-server-v* 2>/dev/null | tail -n +6 | xargs -r rm -f --)

# ---------------------------------------------------------------- 9. firewall (opsional)
if [ "$SCH_UFW" = 1 ] && command -v ufw >/dev/null; then
  ufw allow "${SCH_PORT}/tcp" >/dev/null
  [ -n "$SCH_HTTPS_PORT" ] && ufw allow "${SCH_HTTPS_PORT}/tcp" >/dev/null
  info "ufw: port ${SCH_PORT} ${SCH_HTTPS_PORT} dibuka"
fi

# ---------------------------------------------------------------- selesai
IP="$(hostname -I 2>/dev/null | awk '{print $1}')"
echo
ok "SCH v${VER} berjalan."
echo "  Buka di browser : http://${IP:-IP-SERVER}:${SCH_PORT}"
[ -n "$SCH_HTTPS_PORT" ] && echo "  HTTPS lokal     : https://${IP:-IP-SERVER}:${SCH_HTTPS_PORT}  (untuk kamera HP)"
echo "  Status layanan  : systemctl status ${SCH_SERVICE}"
echo "  Log             : journalctl -u ${SCH_SERVICE} -f"
echo "  Data & backup   : ${SCH_DIR}/data"
if [ -f "${SCH_DIR}/data/ADMIN-PASSWORD.txt" ]; then
  echo
  ok "Login pertama (segera ganti password, lalu hapus file ini):"
  sed 's/^/  /' "${SCH_DIR}/data/ADMIN-PASSWORD.txt"
  echo "  File: ${SCH_DIR}/data/ADMIN-PASSWORD.txt"
fi
