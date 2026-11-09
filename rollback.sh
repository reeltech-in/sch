#!/usr/bin/env bash
# Kembali ke rilis SCH sebelumnya di server.
#   ./rollback.sh            # tampilkan daftar rilis
#   ./rollback.sh 1.3.0      # aktifkan sch-server-v1.3.0
# Catatan: database TIDAK dikembalikan otomatis. Backup sebelum tiap rilis ada di /opt/sch/data/backups.
set -euo pipefail
HOST="${SCH_HOST:-root@10.32.10.112}"
KEY="${SCH_KEY:-$HOME/.ssh/sch_vm_ed25519}"
SSH=(ssh -i "$KEY" -o IdentitiesOnly=yes -o BatchMode=yes "$HOST")

if [ $# -eq 0 ]; then
  "${SSH[@]}" 'echo "Aktif: $(readlink /opt/sch/sch-server)"; echo "Rilis tersedia:"; ls -1t /opt/sch/releases'
  exit 0
fi
VER="$1"
"${SSH[@]}" "set -e; cd /opt/sch; test -f releases/sch-server-v${VER} || { echo 'Rilis v${VER} tidak ada' >&2; exit 1; }
  ln -sfn releases/sch-server-v${VER} sch-server; systemctl restart sch; sleep 2
  curl -fs http://127.0.0.1:\${SCH_PORT:-3220}/api/version; echo"
