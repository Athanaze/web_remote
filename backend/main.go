package main

import (
	"fmt"
	"image/png"
	"log"
	"net"
	"net/http"
	"os"

	"github.com/fyne-io/fyne"
	"github.com/fyne-io/fyne/app"
	"github.com/fyne-io/fyne/widget"
	"github.com/gorilla/websocket"
	"github.com/tarm/serial"
	"github.com/skip2/go-qrcode"
)

var upgrader = websocket.Upgrader{}

func main() {
	// Get the IP address of the computer
	ipAddress, err := getLocalIP()
	if err != nil {
		log.Fatalf("Failed to get local IP address: %s", err)
	}

	// Generate the QR code
	qrCode, err := qrcode.New(fmt.Sprintf("http://%s:3000", ipAddress), qrcode.Medium)
	if err != nil {
		log.Fatalf("Failed to generate QR code: %s", err)
	}

	// Create a new Gorilla WebSocket server
	port := "/dev/ttyUSB0"
	baudRate := 9600
	serialConfig := serial.Config{Name: port, Baud: baudRate}
	serialPort, err := serial.OpenPort(&serialConfig)
	if err != nil {
		log.Fatalf("Failed to open serial port: %s", err)
	}

	http.HandleFunc("/data", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("Failed to upgrade WebSocket connection: %s", err)
			return
		}

		for {
			data := make([]byte, 128)
			n, err := serialPort.Read(data)
			if err != nil {
				log.Printf("Failed to read from serial port: %s", err)
				return
			}
			if n > 0 {
				if err := conn.WriteMessage(websocket.TextMessage, data[:n]); err != nil {
					log.Printf("Failed to send data to WebSocket client: %s", err)
					return
				}
			}
		}
	})

	go func() {
		log.Fatal(http.ListenAndServe(":3000", nil))
	}()

	// Create the Fyne application and window
	a := app.New()
	w := a.NewWindow("MISSION CONTROL RUNNING")

	// Create the QR code image
	qrImage, err := qrCode.PNG(256)
	if err != nil {
		log.Fatalf("Failed to generate QR code image: %s", err)
	}
	qrResource := fyne.NewStaticResource("qr.png", qrImage)

	// Create the widget for the QR code image and link
	qrWidget := widget.NewIcon(qrResource)
	linkWidget := widget.NewHyperlink("OPEN MISSION CONTROL", fmt.Sprintf("http://%s:3000", ipAddress))

	// Create the container for the widgets
	container := fyne.NewContainerWithLayout(
		fyne.NewGridLayoutWithColumns(1),
		qrWidget,
		linkWidget,
	)

	// Set the window content to the container
	w.SetContent(container)

	// Show the window
	w.ShowAndRun()
}

// getLocalIP returns the IP address of the computer.
func getLocalIP() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}
	for _, addr := range addrs {
	
		ipNet, ok := addr.(*net.IPNet)
		if ok && !ipNet.IP.IsLoopback() && ipNet.IP.To4() != nil {
			return ipNet.IP.String(), nil
		}
	}
	return "", fmt.Errorf("Failed to find local IP address")
}	