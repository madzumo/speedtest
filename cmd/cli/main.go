package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/charmbracelet/lipgloss"
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
	configFileName    = "settings.json"
	lipHeaderStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("127"))
	lipConfigStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("112"))
	lipOutputStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("231")).Background(lipgloss.Color("22"))
	lipErrorStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("231")).Background(lipgloss.Color("196")) //231 white
	lipSystemMsgStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("232")).Background(lipgloss.Color("170")) //232 black
	lipResetStyle     = lipgloss.NewStyle()                                                                     // No styling
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
	helpers.SetPEMfiles()
	config, _ := getConfig()
	for {
		headerX := showHeaderPlusConfig(config)
		if menuSelect := bubbles.ShowMenuList("MENU", false, menuTOP, "170", headerX); menuSelect != "" {
			menuSelection(menuSelect, config)
		} else {
			break
		}
	}
}

func menuSelection(menuSelect string, c *configSettings) {
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
				for {
					if complete, errorcode := runIperf(c.IperfS, false, c.IperfP, 0); !complete {
						if errorcode == 1 {
							time.Sleep(10 * time.Second)
						} else {
							break
						}
					} else {
						if complete, errorcode := runIperf(c.IperfS, true, c.IperfP, 0); !complete {
							if errorcode == 1 {
								time.Sleep(10 * time.Second)
							} else {
								break
							}
						} else {
							break
						}
					}
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

			headerX := showHeaderPlusConfig(c)
			if menuSelect := bubbles.ShowMenuList("CHANGE SETTINGS", true, menuSettings, "111", headerX); menuSelect != "" {
				switch menuSelect {
				case menuSettings[0]:
					c.IperfS = getUserInputString("Enter Iperf Server IP and hit 'enter'")
					helpers.ClearTerminalScreen()
				case menuSettings[1]:
					c.IperfP = getUserInputInt("Enter Iperf Port Number  and hit 'enter'")
					helpers.ClearTerminalScreen()
				case menuSettings[2]:
					c.Interval = getUserInputInt("Enter Repeat Test Interval in Minutes and hit 'enter'")
					helpers.ClearTerminalScreen()
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

	case menuTOP[4]:
		err := saveConfig(c)
		if err != nil {
			fmt.Println(lipErrorStyle.Render(fmt.Sprintf("Error Saving Config. %s", err)))
		} else {
			fmt.Println(lipSystemMsgStyle.Render("Config saved"))
		}
		helpers.PauseTerminalScreen()
		helpers.ClearTerminalScreen()
	}
}

func installPlaywright() (greatSuccess bool) {
	greatSuccess = true
	if err := playwright.Install(); err != nil {
		fmt.Println(lipErrorStyle.Render(fmt.Sprintf("could not install Playwright: %v\n", err)))
		helpers.PauseTerminalScreen()
		greatSuccess = false
	}
	helpers.ClearTerminalScreen()
	return greatSuccess
}

func showHeaderPlusConfig(config *configSettings) string {
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
	header := lipHeaderStyle.Render(helpers.MenuHeader) + "\n" +
		lipConfigStyle.Render(fmt.Sprintf("     Iperf:%s->%v  Tests:%s  Browser:%v  Repeat:%vmin\n\n",
			config.IperfS, config.IperfP, isps, config.ShowBrowser, config.Interval))
	return header
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
	fmt.Println(lipSystemMsgStyle.Render(msg))
	fmt.Scanln(&input)
	return input
}

func getUserInputInt(msg string) int {
	var input string
	fmt.Println(lipSystemMsgStyle.Render(msg))
	fmt.Scanln(&input)
	num, err := strconv.Atoi(input)
	if err != nil {
		fmt.Println(lipErrorStyle.Render("Entry must be a numeric number"))
		return 0
	}
	return num
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
