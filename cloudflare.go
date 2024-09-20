package main

import (
	"fmt"
	"log"
	"time"

	"github.com/playwright-community/playwright-go"
)

func cfTest() (testResult string) {
	// Install Playwright if not already installed
	if err := playwright.Install(); err != nil {
		log.Fatalf("could not install Playwright: %v", err)
	}

	// Start Playwright
	pw, err := playwright.Run()
	if err != nil {
		log.Fatalf("could not start Playwright: %v", err)
	}
	defer func() {
		if err := pw.Stop(); err != nil {
			log.Fatalf("could not stop Playwright: %v", err)
		}
	}()

	// Launch the Chromium browser in non-headless mode
	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(false),
	})
	if err != nil {
		log.Fatalf("could not launch browser: %v", err)
	}
	defer func() {
		if err := browser.Close(); err != nil {
			log.Fatalf("could not close browser: %v", err)
		}
	}()

	// Create a new browser context
	context, err := browser.NewContext()
	if err != nil {
		log.Fatalf("could not create context: %v", err)
	}

	// Create a new page within the context
	page, err := context.NewPage()
	if err != nil {
		log.Fatalf("could not create page: %v", err)
	}

	// Navigate to speed.cloudflare.com
	_, err = page.Goto("https://speed.cloudflare.com")
	if err != nil {
		log.Fatalf("could not navigate to speed.cloudflare.com: %v", err)
	}

	// Locate the "START" button by its role and name
	startButton := page.GetByRole("button", playwright.PageGetByRoleOptions{
		Name: "START",
	})

	// Wait for the "START" button to be visible (up to 30 seconds)
	if err := startButton.WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(30000), // 30 seconds
	}); err != nil {
		log.Fatalf("START button not found or not visible: %v", err)
	}

	// Click the "START" button
	if err := startButton.Click(); err != nil {
		log.Fatalf("could not click START button: %v", err)
	}

	// Wait for 1 minute (60 seconds) to allow the test to complete
	fmt.Println("Waiting for test to complete...")
	time.Sleep(40 * time.Second)

	// Define the selector for the target <div> element
	selectorDownload := "div.gp.ev.gq.c.gr.gs"
	selectorUpload := "div.gt.ev.gq.c.gw.gs"

	//Download
	// Create a Locator for the target element
	targetElementDown := page.Locator(selectorDownload)

	// Wait for the element to be visible to ensure it's present in the DOM
	if err := targetElementDown.WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(30000), // 30 seconds
	}); err != nil {
		log.Fatalf("Target element '%s' not found or not visible: %v", selectorDownload, err)
	}

	// Retrieve the text content of the target element
	textDown, err := targetElementDown.TextContent()
	if err != nil {
		log.Fatalf("Could not retrieve text content from '%s': %v", selectorDownload, err)
	}

	//Upload
	// Create a Locator for the target element
	targetElementUpload := page.Locator(selectorUpload).First()

	// Wait for the element to be visible to ensure it's present in the DOM
	if err := targetElementUpload.WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(30000), // 30 seconds
	}); err != nil {
		log.Fatalf("Target element '%s' not found or not visible: %v", selectorUpload, err)
	}

	// Retrieve the text content of the target element
	textUp, err := targetElementUpload.TextContent()
	if err != nil {
		log.Fatalf("Could not retrieve text content from '%s': %v", selectorUpload, err)
	}

	testResult = fmt.Sprintf("🐇CloudFlare Test-> Down:%s, Up:%s", textDown, textUp)
	return testResult
	// fmt.Printf("The value is: %s\n", textDown)
	// fmt.Printf("The value is: %s\n", textUp)
}
