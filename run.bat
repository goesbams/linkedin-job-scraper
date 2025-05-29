@echo off
if "%~3"=="" (
    echo Usage: %0 ^<country^> ^<job_title^> [limit]
    echo Examples:
    echo   %0 "Germany" "golang developer" 50
    echo   %0 "Austria" "software engineer" 25
    exit /b 1
)

set COUNTRY=%~1
set JOB_TITLE=%~2
set LIMIT=%~3
if "%LIMIT%"=="" set LIMIT=25

echo 🔍 Searching for '%JOB_TITLE%' jobs in %COUNTRY% (limit: %LIMIT%)
echo 🇮🇩 Checking for Indonesian employees...

go run main.go "%COUNTRY%" "%JOB_TITLE%" "%LIMIT%"

if exist linkedin_jobs_*.json (
    echo 📁 Moving results to results directory...
    if not exist results mkdir results
    move linkedin_jobs_*.json results\
    echo ✅ Results saved in results directory
) else (
    echo ⚠️  No results file found
)