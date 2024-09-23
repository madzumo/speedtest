package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/madzumo/speedtest/internal/bubbles"
	"github.com/madzumo/speedtest/internal/helpers"
	"github.com/playwright-community/playwright-go"
)

var (
	menuTOP = []string{
		"Run ALL Tests",
		"Run Internet Speed Tests Only",
		"Run Iperf Test Only",
		"Change Settings",
		"Save Settings",
	}
	menuSettings = []string{
		"Set Iperf Server IP",
		"Set Iperf Port number",
		"Set Repeat Test Interval in Minutes",
		"Toggle: Use CloudFront",
		"Toggle: Use M-Labs",
		"Toggle: Use Speedtest.net",
		"Toggle: Show Browser on Speed Tests",
	}
	configFileName = "settings.json"
)

type configSettings struct {
	IperfS         string `json:"iperfServer"`
	IperfP         int    `json:"iperfPort"`
	Interval       int    `json:"repeatInterval"`
	MSS            int    `json:"MSS"`
	CloudFrontTest bool   `json:"CloudFrontTest"`
	MLabTest       bool   `json:"MLabTest"`
	NetTest        bool   `json:"SpeedNetTest"`
	ShowBrowser    bool   `json:"showBrowser"`
}

func main() {
	setPEMfiles()
	config, _ := getConfig()
	for {
		showHeaderPlusConfig(config)
		if menuSelect := bubbles.ShowMenuList("MENU", false, menuTOP, "170"); menuSelect != "" {
			menuSelection(menuSelect, config)
		} else {
			break
		}
	}
}

func menuSelection(menuSelect string, c *configSettings) {
	helpers.ClearTerminalScreen()
	switch menuSelect {
	case menuTOP[0], menuTOP[1], menuTOP[2]:
		if c.CloudFrontTest || c.MLabTest {
			if !installPlaywright() {
				return
			}
		}
		for {
			if menuSelect == menuTOP[0] || menuSelect == menuTOP[1] {
				if c.CloudFrontTest {
					cfTest(c.ShowBrowser)
				}
				if c.MLabTest {
					mlTest(c.ShowBrowser)
				}
				if c.NetTest {
					netTest()
				}
			}
			if menuSelect == menuTOP[0] || menuSelect == menuTOP[2] {
				if runIperf(c.IperfS, false, c.IperfP, 0) {
					runIperf(c.IperfS, true, c.IperfP, 0)
				}
			}

			if c.Interval > 0 {
				time.Sleep(time.Duration(c.Interval) * time.Minute)
			} else {
				break
			}
		}
		helpers.PauseTerminalScreen()
		helpers.ClearTerminalScreen()
	case menuTOP[3]:
		for {
			helpers.ClearTerminalScreen()
			showHeaderPlusConfig(c)
			if menuSelect := bubbles.ShowMenuList("CHANGE SETTINGS", true, menuSettings, "111"); menuSelect != "" {
				switch menuSelect {
				case menuSettings[0]:
					c.IperfS = getUserInputString("Enter Iperf Server IP and hit 'enter'")
				case menuSettings[1]:
					c.IperfP = getUserInputInt("Enter Iperf Port Number  and hit 'enter'")
				case menuSettings[2]:
					c.Interval = getUserInputInt("Enter Repeat Test Interval in Minutes and hit 'enter'")
				case menuSettings[3]:
					if c.CloudFrontTest {
						c.CloudFrontTest = false
					} else {
						c.CloudFrontTest = true
					}
				case menuSettings[4]:
					if c.MLabTest {
						c.MLabTest = false
					} else {
						c.MLabTest = true
					}
				case menuSettings[5]:
					if c.NetTest {
						c.NetTest = false
					} else {
						c.NetTest = true
					}
				case menuSettings[6]:
					if c.ShowBrowser {
						c.ShowBrowser = false
					} else {
						c.ShowBrowser = true
					}
				}
			} else {
				break
			}
		}
		helpers.ClearTerminalScreen()
	case menuTOP[4]:
		// err := saveConfig(c)
		// cp := helpers.NewPromptColor()
		// if err != nil {
		// 	cp.Error.Printf("Error Saving Config. %s\n", err)
		// } else {
		// 	cp.Notify2.Printf("Config saved\n")
		// }
		// helpers.PauseTerminalScreen()
		helpers.ClearTerminalScreen()
	}
}

func installPlaywright() (greatSuccess bool) {
	greatSuccess = true
	if err := playwright.Install(); err != nil {
		// log.Fatalf("could not install Playwright: %v", err)
		cp := helpers.NewPromptColor()
		cp.Error.Printf("could not install Playwright: %v\n", err)
		helpers.PauseTerminalScreen()
		greatSuccess = false
	}
	helpers.ClearTerminalScreen()
	return greatSuccess
}

func showHeaderPlusConfig(config *configSettings) {
	helpers.ClearTerminalScreen()
	var isps string
	if config.CloudFrontTest {
		isps += "CF,"
	}
	if config.MLabTest {
		isps += "ML,"
	}
	if config.NetTest {
		isps += "NET"
	}
	cp := helpers.NewPromptColor()
	cp.Notify1.Println(helpers.MenuHeader)
	cp.Notify4.Printf("     Iperf:%s->%v  Tests:%s  Browser:%v  Repeat:%vmin\n", config.IperfS, config.IperfP, isps, config.ShowBrowser, config.Interval)
}

func getConfig() (*configSettings, error) {
	configTemp := configSettings{
		IperfS:         "0.0.0.0",
		IperfP:         5201,
		Interval:       0,
		MSS:            0,
		CloudFrontTest: true,
		MLabTest:       true,
		NetTest:        true,
		ShowBrowser:    false,
	}

	data, err := os.ReadFile(configFileName)
	if err != nil {
		return &configTemp, err
	}

	err = json.Unmarshal(data, &configTemp)
	return &configTemp, err
}

func saveConfig(config *configSettings) error {
	//convert to struct -> JSON
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(configFileName, data, 0644)
}

func getUserInputString(msg string) string {
	var input string
	cp := helpers.NewPromptColor()
	cp.Normal.Printf("%s\n", msg)
	fmt.Scanln(&input)
	return input
}

func getUserInputInt(msg string) int {
	var input string
	cp := helpers.NewPromptColor()
	cp.Normal.Printf("%s\n", msg)
	fmt.Scanln(&input)
	num, err := strconv.Atoi(input)
	if err != nil {
		cp.Error.Println("Entry must be a numeric number")
		return 0
	}
	return num
}

func setPEMfiles() {
	// Get the current working directory
	dir, err := os.Getwd()
	if err != nil {
		fmt.Println("Error getting current directory:", err)
		return
	}

	// Find .pem files in the directory
	matches, err := filepath.Glob(filepath.Join(dir, "*.pem"))
	if err != nil {
		// fmt.Println("Error searching for .pem files:", err)
		return
	}

	// If a .pem file was found, set the environment variable
	if len(matches) > 0 {
		err = os.Setenv("NODE_EXTRA_CA_CERTS", matches[0]) // Use the first .pem file found
		if err != nil {
			// fmt.Println("Error setting environment variable:", err)
		} else {
			// fmt.Println("Environment variable set:", os.Getenv("NODE_EXTRA_CA_CERTS"))
		}
	} else {
		// fmt.Println("No .pem files found.")
	}
}

func setLogFileName() string {
	hostname, err := os.Hostname()
	if err != nil {
		fmt.Println("Error getting hostname of client:", err)
	} else {
		return fmt.Sprintf("iperf3_%s.txt", hostname)
	}
	return ""
}

func writeLogFile(logData string) {
	logFileName := setLogFileName()
	fileWriter, err := os.OpenFile(logFileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("failed to create/open Log file: %v\n", err)
	}
	defer fileWriter.Close()

	currentTime := time.Now().Format("2006-01-02 15:04:05")
	if _, err := fmt.Fprintf(fileWriter, "[%s]%s\n", currentTime, logData); err != nil {
		fmt.Printf("failed to write to Log file: %v\n", err)
	}

	// fmt.Printf("[%s]%s\n", currentTime, logData)
}
