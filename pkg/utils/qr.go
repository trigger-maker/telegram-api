package utils

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image/png"

	"github.com/skip2/go-qrcode"
)

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

func PrintQRToTerminal(content string) {
	PrintQRToTerminalWithName(content, "")
}

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

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}
