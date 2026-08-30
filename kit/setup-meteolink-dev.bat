@echo off
:: meteolink.dev → local + VM — exécuter en tant qu'Administrateur (clic droit → Exécuter en tant qu'admin)
powershell -ExecutionPolicy Bypass -File "%~dp0setup-meteolink-dev.ps1"
pause
