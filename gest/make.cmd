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
