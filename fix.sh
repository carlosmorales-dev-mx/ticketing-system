#!/usr/bin/env bash
# Instala el README profesional de SDV (con capturas) y un arreglo del plano en celular.
# Uso:  bash readme.sh [ruta/al/sdv-readme.zip]
# Variables:  SDV_DIR=/ruta/al/proyecto  (por defecto /mnt/Datos/proyectos/SDV)
set -euo pipefail

PROJ="${SDV_DIR:-/mnt/Datos/proyectos/SDV}"
ZIP="${1:-}"
fail() { echo "✗ $1" >&2; exit 1; }

# 1. Localizar el zip
if [ -z "$ZIP" ]; then
  ZIP="$(find "$HOME/Descargas" "$HOME/Downloads" "$PWD" -maxdepth 1 -name 'sdv-readme*.zip' \
         -printf '%T@ %p\n' 2>/dev/null | sort -nr | head -n 1 | cut -d' ' -f2- || true)"
fi
[ -n "$ZIP" ] && [ -f "$ZIP" ] || fail "No encontré sdv-readme.zip. Pásalo así: bash readme.sh /ruta/al/zip"

# 2. Revisar el entorno
[ -d "$PROJ/api" ] && [ -d "$PROJ/frontend" ] && [ -f "$PROJ/Makefile" ] \
  || fail "$PROJ no parece el proyecto SDV. Usa SDV_DIR=/ruta"
command -v python3 >/dev/null || fail "python3 no está en el PATH"

# 3. Extraer a una carpeta temporal y validar
TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT
python3 -I - "$ZIP" "$TMP" <<'PY' || fail "No pude leer $ZIP (¿está corrupto?)"
import sys, zipfile, os
z = zipfile.ZipFile(sys.argv[1]); dest = os.path.realpath(sys.argv[2])
for n in z.namelist():
    if not os.path.realpath(os.path.join(dest, n)).startswith(dest + os.sep):
        sys.exit("ruta sospechosa en el zip: " + n)
z.extractall(dest)
PY
for f in README.md README.en.md docs/img/mapa.png docs/img/boleto-multiple.png \
         docs/img/reiniciar-mapa.png docs/img/sin-conexion.png frontend/components/StageSVG.vue; do
  [ -f "$TMP/$f" ] || fail "Al zip le falta $f"
done

# 4. Respaldo (fuera del repo)
BACKUP="$HOME/sdv-readme-backup-$(date +%Y%m%d-%H%M%S)"
mkdir -p "$BACKUP"
for f in README.md README.en.md frontend/components/StageSVG.vue; do
  if [ -f "$PROJ/$f" ]; then mkdir -p "$BACKUP/$(dirname "$f")"; cp "$PROJ/$f" "$BACKUP/$f"; fi
done
echo "→ Respaldo en $BACKUP"

# 5. Copiar
mkdir -p "$PROJ/docs/img"
cp "$TMP"/docs/img/*.png "$PROJ/docs/img/"
cp "$TMP/README.md" "$TMP/README.en.md" "$PROJ/"
cp "$TMP/frontend/components/StageSVG.vue" "$PROJ/frontend/components/StageSVG.vue"
rm -rf "$PROJ/frontend/.nuxt" "$PROJ/frontend/node_modules/.vite" 2>/dev/null || true

echo "✓ Listo:"
echo "  - README.md y README.en.md reemplazados (versión pro, con 4 capturas en docs/img/)"
echo "  - StageSVG.vue: arregla el plano en pantallas de celular (si tu dev server corre, recarga solo)"
echo "  Revisa los cambios con:  cd $PROJ && git status"