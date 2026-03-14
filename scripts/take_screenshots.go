package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"time"

	"github.com/chromedp/chromedp"
)

func main() {
	// Create a context
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	// Run tasks
	// 1. Capture Desktop Screenshot
	// 2. Capture Mobile Screenshot
	// Note: We assume the app is running on localhost:3281
	url := "http://localhost:3281/projects/overdrive/jobs"

	fmt.Printf("Navigating to %s...\n", url)

	var desktopBuf []byte
	var mobileBuf []byte

	// 1. Desktop Task
	if err := chromedp.Run(ctx,
		// Navigate
		chromedp.Navigate(url),
		// Set Viewport to standard Desktop
		chromedp.EmulateViewport(1920, 1080),
		// Wait for body to be visible
		chromedp.WaitVisible(`body`, chromedp.ByQuery),
		// Sleep a bit to ensure rendering
		chromedp.Sleep(2*time.Second),
		// Capture screenshot
		chromedp.CaptureScreenshot(&desktopBuf),
	); err != nil {
		log.Fatal(err)
	}

	// 2. Mobile Task
	if err := chromedp.Run(ctx,
		// Set Viewport to iPhone X
		chromedp.EmulateViewport(375, 812, chromedp.EmulateMobile),
		// Wait
		chromedp.Sleep(1*time.Second),
		// Capture screenshot
		chromedp.CaptureScreenshot(&mobileBuf),
	); err != nil {
		log.Fatal(err)
	}

	// Save the files
	if err := ioutil.WriteFile("static/img/app-desktop.png", desktopBuf, 0644); err != nil {
		log.Fatal(err)
	}
	if err := ioutil.WriteFile("static/img/app-mobile.png", mobileBuf, 0644); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Screenshots saved to static/img/app-desktop.png and static/img/app-mobile.png")
}
