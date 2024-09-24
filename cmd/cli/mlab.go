package main

import (
	"fmt"
	"log"
	"time"

	"github.com/madzumo/speedtest/internal/bubbles"
	"github.com/playwright-community/playwright-go"
)

func mlTest(showBrowser bool) (testResult string) {
	quit := make(chan struct{})
	go bubbles.ShowSpinner(quit, "M-Lab Speed Test....", "57") // Run spinner in a goroutine

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
	_, err = page.Goto("https://speed.measurementlab.net/#/")
	if err != nil {
		log.Fatalf("could not navigate to ml/google speedtest: %v", err)
	}

	//	INPUT CODE HERE
	if err := page.Locator("text=I agree to the data policy").Click(); err != nil {
		log.Fatalf("could not click on agree text: %v", err)
	}

	if err := page.Locator("text=Begin Again Testing").Click(); err != nil {
		log.Fatalf("could not click on text: %v", err)
	}

	// Wait to allow the test to complete
	time.Sleep(35 * time.Second)

	// Locating the first instance of an element with class "ng-binding" containing text "Mb/s"
	speedLocator := page.Locator(".ng-binding:text-matches(\"[0-9]+\\.?[0-9]* Mb/s\", \"i\")").First()
	if err := speedLocator.WaitFor(); err != nil {
		log.Fatalf("could not wait for speed result: %v", err)
	}
	textDown, err := speedLocator.TextContent()
	if err != nil {
		log.Fatalf("could not get text content: %v", err)
	}
	// log.Println("Download speed:", textDown)

	// Locating the first instance of an element with class "ng-binding" containing text "Mb/s"
	speedLocator2 := page.Locator(".ng-binding:text-matches(\"[0-9]+\\.?[0-9]* Mb/s\", \"i\")").Nth(2)
	if err := speedLocator2.WaitFor(); err != nil {
		log.Fatalf("could not wait for speed result: %v", err)
	}
	textUp, err := speedLocator2.TextContent()
	if err != nil {
		log.Fatalf("could not get text content: %v", err)
	}
	// log.Println("Upload speed:", textUp)

	if err := browser.Close(); err != nil {
		log.Fatalf("could not close browser: %v", err)
	}
	if err := pw.Stop(); err != nil {
		log.Fatalf("could not stop Playwright: %v", err)
	}
	close(quit)
	time.Sleep(1 * time.Second)
	testResult = fmt.Sprintf("M-Labs Test-> Down:%s, Up:%s", textDown, textUp)
	fmt.Println(lipOutputStyle.Render(testResult))
	return testResult
}
