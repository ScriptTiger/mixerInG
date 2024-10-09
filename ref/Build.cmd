@echo off

set APP=mixerInG
if not exist "Release" md "Release"

choice /m "Dev build?"
if %errorlevel% == 1 (set dev=1) else set dev=0

set GOARCH=amd64
call :Build_OS

if %dev% == 1 goto Exit

set GOARCH=386
call :Build_OS

:Exit
pause
exit /b

:Build_OS

set GOOS=windows
set EXT=.exe
call :Build

if %dev% == 1 exit /b

set GOOS=linux
set EXT=
call :Build

if %GOARCH% == 386 exit /b

set GOOS=darwin
set EXT=.app
call :Build

exit /b

:Build
echo Building %APP%_%GOOS%_%GOARCH%%EXT%...
go build -ldflags="-s -w" -o "Release/%APP%_%GOOS%_%GOARCH%%EXT%" ref.go

exit /b