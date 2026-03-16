// Package utils provides utility functions for QR code generation.
package utils

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image/png"

	"github.com/skip2/go-qrcode"
)

// GenerateQRBase64 generates a QR code and returns it as a base64 encoded string.
func GenerateQRBase64(content string) (string, error) {
	qr, err := qrcode.New(content, qrcode.Medium)
	if err != nil {
		return "", err
	}

	img := qr.Image(256)
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}

// PrintQRToTerminal prints a QR code to the terminal.
func PrintQRToTerminal(content string) {
	PrintQRToTerminalWithName(content, "")
}

// PrintQRToTerminalWithName prints a QR code to the terminal with a session name.
func PrintQRToTerminalWithName(content, sessionName string) {
	qr, err := qrcode.New(content, qrcode.Low)
	if err != nil {
		fmt.Printf("❌ Error generando QR: %v\n", err)
		return
	}

	fmt.Println()
	fmt.Println("╔══════════════════════════════════════════════════════╗")
	fmt.Println("║              📱 ESCANEA CON TELEGRAM                 ║")
	fmt.Println("╠══════════════════════════════════════════════════════╣")
	if sessionName != "" {
		fmt.Printf("║  Sesión: %-43s ║\n", sessionName)
		fmt.Println("╠══════════════════════════════════════════════════════╣")
	}
	fmt.Println("╚══════════════════════════════════════════════════════╝")
	fmt.Println(qr.ToSmallString(false))
	fmt.Println("╔══════════════════════════════════════════════════════╗")
	fmt.Printf("║ URL: %-48s ║\n", truncate(content, 48))
	fmt.Println("╚══════════════════════════════════════════════════════╝")
	fmt.Println()
}

func truncate(s string, maxLength int) string {
	if len(s) <= maxLength {
		return s
	}
	return s[:maxLength-3] + "..."
}
