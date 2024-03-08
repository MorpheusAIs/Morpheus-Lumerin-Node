#!/bin/bash

echo "Building and signing application for macOS"

echo "Insert the certificate password and press Enter:"
read -rs PASSWORD

CSC_LINK="file://as-mac.p12" CSC_KEY_PASSWORD=$PASSWORD npm run release
