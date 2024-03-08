@echo off

echo Building and signing application for Windows

set /p PASSWORD="Insert the certificate password and press Enter:"

set CSC_LINK=as-win.p12
set CSC_KEY_PASSWORD=%PASSWORD%
npm run release
