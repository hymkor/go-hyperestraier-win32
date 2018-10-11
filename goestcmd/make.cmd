@echo off
setlocal
set GOARCH=386
call :"%1"
endlocal
exit /b

:""
:"all"
:"build"
    go fmt
    go build
    exit /b

:"upgrade"
    for %%I in (%CD%) do set "EXE=%%~nI.exe"
    for /F %%I in ('where %EXE%') do copy /-Y /V "%EXE%" "%%I"
    exit /b
