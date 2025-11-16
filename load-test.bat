@echo off
chcp 65001 > nul
setlocal enabledelayedexpansion

set URL=http://localhost:8080
set REQUESTS=50

echo Load testing %URL%
echo Total requests: %REQUESTS%
echo.

set SUCCESS=0
set ERROR=0

for /l %%i in (1,1,%REQUESTS%) do (
    curl -s -f "%URL%/health" > nul 2> nul
    
    if !errorlevel! equ 0 (
        set /a SUCCESS+=1
        echo [%%i] OK
    ) else (
        set /a ERROR+=1  
        echo [%%i] FAIL
    )
    
    timeout /t 1 /nobreak > nul
)

echo.
echo === RESULTS ===
echo Successful: !SUCCESS!
echo Errors: !ERROR!
set /a RATE=!SUCCESS!*100/!REQUESTS!
echo Success rate: !RATE!%%
echo.
pause