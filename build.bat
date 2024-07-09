@echo off
setlocal enabledelayedexpansion

REM Check if .env file exists
if not exist ".env" (
    echo .env file not found!
    exit /b 1
)

REM Read the .env file and set variables
for /f "delims=" %%i in (.env) do (
    set "line=%%i"
    for /f "tokens=1,2 delims==" %%a in ("!line!") do (
        set %%a=%%b
    )
)

REM Check if the VERSION variable was set successfully
IF "%VERSION%"=="" (
    echo VERSION variable is not set in the .env file!
    exit /b 1
)

REM Define the image name
SET IMAGE_NAME=user-service

REM Build the Docker image
echo Building Docker image with version %VERSION%...
docker build -t dev-1:32000/%IMAGE_NAME%:dev-latest -t dev-1:32000/%IMAGE_NAME%:dev -t dev-1:32000/%IMAGE_NAME%:dev-%VERSION% .

REM Check if the build was successful
IF %ERRORLEVEL% NEQ 0 (
    echo Docker build failed!
    exit /b %ERRORLEVEL%
) ELSE (
    echo Docker build succeeded!
)

REM Push the Docker image to the registry
echo Pushing Docker image to the registry...
docker push dev-1:32000/%IMAGE_NAME%:dev-latest
docker push dev-1:32000/%IMAGE_NAME%:dev
docker push dev-1:32000/%IMAGE_NAME%:dev-%VERSION%

REM Check if the push was successful
IF %ERRORLEVEL% NEQ 0 (
    echo Docker push failed!
    exit /b %ERRORLEVEL%
) ELSE (
    echo Docker push succeeded!
)

REM Update Kubernetes deployment
echo Updating Kubernetes deployment...
kubectl set image deployment/%IMAGE_NAME% %IMAGE_NAME%=localhost:32000/%IMAGE_NAME%:dev-latest --namespace=default

REM Check if the update was successful
IF %ERRORLEVEL% NEQ 0 (
    echo Kubernetes deployment update failed!
    exit /b %ERRORLEVEL%
) ELSE (
    echo Kubernetes deployment updated successfully!
)

echo Done.
endlocal