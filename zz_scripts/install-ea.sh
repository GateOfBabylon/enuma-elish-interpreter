#!/bin/bash

BINARY_NAME="enumago"
INSTALL_PATH="/usr/local/bin/$BINARY_NAME"

if ! command -v go &> /dev/null; then
    echo "Go is not installed! Install it first: https://golang.org/dl/"
    exit 1
fi

echo "Building the Go EA Interpreter..."
go build -o $BINARY_NAME main.go

if [ $? -ne 0 ]; then
    echo "Build failed unexpectedly"
    exit 1
fi

echo "Installing $BINARY_NAME to $INSTALL_PATH..."
sudo mv $BINARY_NAME $INSTALL_PATH

sudo chmod +x $INSTALL_PATH

if command -v $BINARY_NAME &> /dev/null; then
    echo "Installation successful! You can now run:"
    echo "   $BINARY_NAME <your-file.ea>"
else
    echo "Installation failed. Please check permissions."
fi
