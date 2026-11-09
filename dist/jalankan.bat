@echo off
title SCH LMS Sekolah
cd /d "%~dp0"
set EXE=sch-server-windows-amd64.exe
if not exist "%EXE%" set EXE=sch-server-windows-386.exe
"%EXE%" -addr :3220 -https :3221 -db data\sch.db
echo.
echo Server berhenti. Tekan tombol apa saja untuk menutup.
pause >nul
