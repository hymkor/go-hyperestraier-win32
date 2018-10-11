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
    for /F %%I in ('where goestcmd') do copy /-Y /V goestcmd.exe "%%I"
    exit /b
