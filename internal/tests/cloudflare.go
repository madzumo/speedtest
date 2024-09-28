package tests

import (
	"fmt"
	"time"

	hp "github.com/madzumo/speedtest/internal/helpers"
	"github.com/playwright-community/playwright-go"
)

func CFTest(showBrowser bool) (resultOverview string) {
	// Start Playwright
	pw, err := playwright.Run()
	if err != nil {
		return fmt.Sprintf("could not start Playwright: %v", err)
	}
	defer func() {
		if err := pw.Stop(); err != nil {
		}
	}()

	// Launch browser. Set headless. non-headless mode = false
	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(!showBrowser),
	})
	if err != nil {
		return fmt.Sprintf("could not launch browser: %v", err)
	}

	defer func() {
		if err := browser.Close(); err != nil {
		}
	}()

	// Create a new browser context
	context, err := browser.NewContext()
	if err != nil {
		return fmt.Sprintf("could not create context: %v", err)
	}

	// Create a new page within the context
	page, err := context.NewPage()
	if err != nil {
		return fmt.Sprintf("could not create page: %v", err)
	}

	// Navigate
	_, err = page.Goto("https://speed.cloudflare.com")
	if err != nil {
		return fmt.Sprintf("could not navigate to speed.cloudflare.com: %v", err)
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
		return fmt.Sprintf("START button not found or not visible: %v", err)
	}

	// Click the "START" button
	if err := startButton.Click(); err != nil {
		return fmt.Sprintf("could not click START button: %v", err)
	}

	// Wait for 1 minute (60 seconds) to allow the test to complete
	// fmt.Println("Waiting for test to complete...")
	time.Sleep(35 * time.Second)

	// Find the first occurrence of "Mbps" and get the preceding div's text content
	locator := page.Locator(`text="Mbps"`).First().Locator("xpath=preceding-sibling::div[1]")
	textDown, err := locator.InnerText()
	if err != nil {
		return fmt.Sprintf("could not get inner text from locator: %v", err)
	}

	// Find the second occurrence of "Mbps" and get the preceding div's text content
	locator2 := page.Locator(`text="Mbps"`).Nth(1).Locator("xpath=preceding-sibling::div[1]")
	textUp, err := locator2.InnerText()
	if err != nil {
		return fmt.Sprintf("could not get inner text from locator: %v", err)
	}

	if err := browser.Close(); err != nil {
		return fmt.Sprintf("could not close browser: %v", err)
	}
	if err := pw.Stop(); err != nil {
		return fmt.Sprintf("could not stop Playwright: %v", err)
	}

	testResult := fmt.Sprintf("Cloudflare-> Down:%s  Up:%s", textDown, textUp)
	fmt.Println(hp.LipOutputStyle.Render(testResult))
	hp.WriteLogFile(fmt.Sprintf("🐇%s", testResult))

	return testResult
}
