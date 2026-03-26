@echo off
REM Build a 64-bit Windows PE for this repo. Use this if omnigraph.exe triggers
REM "Unsupported 16-bit application" (wrong arch, bad file, or confused PATH).
setlocal
if "%GOARCH%"=="" set GOARCH=amd64
set GOOS=windows
go build -trimpath -o omnigraph.exe ./cmd/omnigraph
if errorlevel 1 exit /b 1
echo Built omnigraph.exe for windows/%GOARCH%
endlocal
