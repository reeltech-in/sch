@echo off
title SCH (online lewat Cloudflare Tunnel)
cd /d "%~dp0"
set EXE=sch-server-windows-amd64.exe
if not exist "%EXE%" set EXE=sch-server-windows-386.exe
rem Hanya bisa dibuka dari komputer ini; cloudflared meneruskan pengunjung dari internet.
"%EXE%" -addr 127.0.0.1:3220 -https "" -proxy -db data\sch.db
echo.
echo Server berhenti. Tekan tombol apa saja untuk menutup.
pause >nul
