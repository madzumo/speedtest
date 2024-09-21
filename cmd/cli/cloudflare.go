package main

import (
	"fmt"
	"log"
	"time"

	"github.com/madzumo/speedtest/internal/bubbles"
	"github.com/madzumo/speedtest/internal/helpers"
	"github.com/playwright-community/playwright-go"
)

func cfTest(showBrowser bool) (testResult string) {
	quit := make(chan struct{})
	go bubbles.ShowSpinner(quit, "Cloudflare Speed Test....", "202") // Run spinner in a goroutine

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
		Headless: playwright.Bool(!showBrowser),
	})
	// browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
	// 	Headless:       playwright.Bool(!showBrowser),
	// 	ExecutablePath: playwright.String("./chrome-win/chrome.exe"),
	// })
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
	// fmt.Println("Waiting for test to complete...")
	time.Sleep(38 * time.Second)

	// Find the first occurrence of "Mbps" and get the preceding div's text content
	locator := page.Locator(`text="Mbps"`).First().Locator("xpath=preceding-sibling::div[1]")
	textDown, err := locator.InnerText()
	if err != nil {
		log.Fatalf("could not get inner text from locator: %v", err)
	}

	// Find the second occurrence of "Mbps" and get the preceding div's text content
	locator2 := page.Locator(`text="Mbps"`).Nth(1).Locator("xpath=preceding-sibling::div[1]")
	textUp, err := locator2.InnerText()
	if err != nil {
		log.Fatalf("could not get inner text from locator: %v", err)
	}

	// fmt.Println("Extracted text:", textDown)

	if err := browser.Close(); err != nil {
		log.Fatalf("could not close browser: %v", err)
	}
	if err := pw.Stop(); err != nil {
		log.Fatalf("could not stop Playwright: %v", err)
	}

	close(quit) // This will send a quitMsg to your spinner model
	testResult = fmt.Sprintf("Cloudflare Test -> Down:%s, Up:%s", textDown, textUp)
	cp := helpers.NewPromptColor()
	cp.Normal.Println(testResult)
	return testResult
}
