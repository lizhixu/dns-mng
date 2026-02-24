@echo off
REM DNS Manager - Build and Push to Docker Hub
REM Usage: build-and-push.bat [version]

setlocal

set VERSION=%1
if "%VERSION%"=="" set VERSION=latest

set DOCKER_USERNAME=jacyli
set IMAGE_NAME=dns-mng

echo.
echo 🔨 Building DNS Manager Docker images...
echo 📦 Version: %VERSION%
echo.

REM Build backend image
echo 🔨 Building backend image...
docker build -t %DOCKER_USERNAME%/%IMAGE_NAME%:backend-%VERSION% ./backend
if errorlevel 1 goto :error
docker tag %DOCKER_USERNAME%/%IMAGE_NAME%:backend-%VERSION% %DOCKER_USERNAME%/%IMAGE_NAME%:backend
if errorlevel 1 goto :error

REM Build frontend image
echo 🔨 Building frontend image...
docker build -t %DOCKER_USERNAME%/%IMAGE_NAME%:frontend-%VERSION% ./frontend
if errorlevel 1 goto :error
docker tag %DOCKER_USERNAME%/%IMAGE_NAME%:frontend-%VERSION% %DOCKER_USERNAME%/%IMAGE_NAME%:frontend
if errorlevel 1 goto :error

echo.
echo ✅ Build completed!
echo.
echo 📤 Pushing images to Docker Hub...
echo.

REM Login to Docker Hub
echo 🔐 Please login to Docker Hub if prompted...
docker login
if errorlevel 1 goto :error

REM Push backend images
echo 📤 Pushing backend image...
docker push %DOCKER_USERNAME%/%IMAGE_NAME%:backend-%VERSION%
if errorlevel 1 goto :error
docker push %DOCKER_USERNAME%/%IMAGE_NAME%:backend
if errorlevel 1 goto :error

REM Push frontend images
echo 📤 Pushing frontend image...
docker push %DOCKER_USERNAME%/%IMAGE_NAME%:frontend-%VERSION%
if errorlevel 1 goto :error
docker push %DOCKER_USERNAME%/%IMAGE_NAME%:frontend
if errorlevel 1 goto :error

echo.
echo ✅ All images pushed successfully!
echo.
echo 📋 Images:
echo    Backend:  %DOCKER_USERNAME%/%IMAGE_NAME%:backend-%VERSION%
echo    Backend:  %DOCKER_USERNAME%/%IMAGE_NAME%:backend (latest)
echo    Frontend: %DOCKER_USERNAME%/%IMAGE_NAME%:frontend-%VERSION%
echo    Frontend: %DOCKER_USERNAME%/%IMAGE_NAME%:frontend (latest)
echo.
echo 🚀 To use these images:
echo    docker-compose pull
echo    docker-compose up -d
echo.
goto :end

:error
echo.
echo ❌ Error occurred during build or push!
echo.
exit /b 1

:end
endlocal
