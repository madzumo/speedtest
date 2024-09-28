package tests

import (
	"fmt"
	"strings"
	"time"

	hp "github.com/madzumo/speedtest/internal/helpers"
	"github.com/playwright-community/playwright-go"
)

func MLTest(showBrowser bool) string {
	// Start Playwright
	pw, err := playwright.Run()
	if err != nil {
		return fmt.Sprintf("could not start Playwright: %v", err)
	}
	defer func() {
		if err := pw.Stop(); err != nil {
		}
	}()

	// Launch the Chromium browser in non-headless mode
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

	// Navigate to speed.cloudflare.com
	_, err = page.Goto("https://speed.measurementlab.net/#/")
	if err != nil {
		return fmt.Sprintf("could not navigate to ml/google speedtest: %v", err)
	}

	//	INPUT CODE HERE
	if err := page.Locator("text=I agree to the data policy").Click(); err != nil {
		return fmt.Sprintf("could not click on agree text: %v", err)
	}

	if err := page.Locator("text=Begin Again Testing").Click(); err != nil {
		return fmt.Sprintf("could not click on text: %v", err)
	}

	// Wait to allow the test to complete
	time.Sleep(35 * time.Second)

	// Locating the first instance of an element with class "ng-binding" containing text "Mb/s"
	speedLocator := page.Locator(".ng-binding:text-matches(\"[0-9]+\\.?[0-9]* Mb/s\", \"i\")").First()
	if err := speedLocator.WaitFor(); err != nil {
		return fmt.Sprintf("could not wait for speed result: %v", err)
	}
	textDown, err := speedLocator.TextContent()
	if err != nil {
		return fmt.Sprintf("could not get text content: %v", err)
	}
	// log.Println("Download speed:", textDown)

	// Locating the first instance of an element with class "ng-binding" containing text "Mb/s"
	speedLocator2 := page.Locator(".ng-binding:text-matches(\"[0-9]+\\.?[0-9]* Mb/s\", \"i\")").Nth(2)
	if err := speedLocator2.WaitFor(); err != nil {
		return fmt.Sprintf("could not wait for speed result: %v", err)
	}
	textUp, err := speedLocator2.TextContent()
	if err != nil {
		return fmt.Sprintf("could not get text content: %v", err)
	}
	// log.Println("Upload speed:", textUp)

	if err := browser.Close(); err != nil {
		return fmt.Sprintf("could not close browser: %v", err)
	}
	if err := pw.Stop(); err != nil {
		return fmt.Sprintf("could not stop Playwright: %v", err)
	}

	testResult := fmt.Sprintf("M-Labs Test-> Down:%s Up:%s", strings.Replace(textDown, "Mb/s", "", -1), strings.Replace(textUp, "Mb/s", "", -1))
	hp.WriteLogFile(fmt.Sprintf("🧪%s", testResult))
	return testResult
}
