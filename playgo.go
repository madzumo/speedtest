package main

import (
	"fmt"
	"time"

	"github.com/playwright-community/playwright-go"
)

func doPlay(urlTest string) {
	// Install browsers if not already installed
	if err := playwright.Install(); err != nil {
		fmt.Printf("Could not install playwright: %v\n", err)
		return
	}

	// Start Playwright
	pw, err := playwright.Run()
	if err != nil {
		fmt.Printf("Could not start Playwright: %v\n", err)
		return
	}
	defer pw.Stop()

	// Launch a new browser instance
	browser, err := pw.Chromium.Launch()
	if err != nil {
		fmt.Printf("Could not launch browser: %v\n", err)
		return
	}
	defer browser.Close()

	// Create a new browser context and page
	context, err := browser.NewContext()
	if err != nil {
		fmt.Printf("Could not create context: %v\n", err)
		return
	}
	//clear cookies & cache
	err = context.ClearCookies()
	if err != nil {
		fmt.Printf("Could not clear cookies: %v\n", err)
		return
	}

	page, err := context.NewPage()
	if err != nil {
		fmt.Printf("Could not create page: %v\n", err)
		return
	}

	// Record the start time
	startTime := time.Now()

	// Navigate to the URL and wait until the page is fully loaded
	_, err = page.Goto(urlTest, playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateLoad,
	})
	if err != nil {
		fmt.Printf("Could not navigate to the page: %v\n", err)
		return
	}

	// Calculate the duration
	duration := time.Since(startTime)

	// Print the time taken to load the page
	fmt.Printf("Page loaded in %v\n", duration)
}
