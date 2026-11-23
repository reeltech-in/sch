#!/usr/bin/env bash
# Kembali ke rilis SCH sebelumnya di server.
#   ./rollback.sh            # tampilkan daftar rilis
#   ./rollback.sh 1.3.0      # aktifkan sch-server-v1.3.0
# Catatan: database TIDAK dikembalikan otomatis. Backup sebelum tiap rilis ada di /opt/sch/data/backups.
# Alamat server dibaca dari deploy.env, sama seperti deploy.sh.
set -euo pipefail
cd "$(dirname "$0")"
[ -f deploy.env ] && . ./deploy.env
HOST="${SCH_HOST:?isi SCH_HOST di deploy.env (salin dari deploy.env.example)}"
KEY="${SCH_KEY:-$HOME/.ssh/id_ed25519}"
SSH=(ssh -i "$KEY" -o IdentitiesOnly=yes -o BatchMode=yes "$HOST")

if [ $# -eq 0 ]; then
  "${SSH[@]}" 'echo "Aktif: $(readlink /opt/sch/sch-server)"; echo "Rilis tersedia:"; ls -1t /opt/sch/releases'
  exit 0
fi
VER="$1"
"${SSH[@]}" "set -e; cd /opt/sch; test -f releases/sch-server-v${VER} || { echo 'Rilis v${VER} tidak ada' >&2; exit 1; }
  ln -sfn releases/sch-server-v${VER} sch-server; systemctl restart sch; sleep 2
  curl -fs http://127.0.0.1:\${SCH_PORT:-3220}/api/version; echo"
