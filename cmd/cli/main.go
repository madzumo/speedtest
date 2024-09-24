package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/madzumo/speedtest/internal/bubbles"
	hp "github.com/madzumo/speedtest/internal/helpers"
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
		"Set MSS Size",
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
	hp.SetPEMfiles()
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
			if !hp.InstallPlaywright() {
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
				loopcount := 0
				for {
					if complete, errorcode := runIperf(c.IperfS, false, c.IperfP, c.MSS); !complete {
						if errorcode == 1 {
							time.Sleep(10 * time.Second)
						} else {
							break
						}
					} else {
						if complete, errorcode := runIperf(c.IperfS, true, c.IperfP, c.MSS); !complete {
							if errorcode == 1 {
								time.Sleep(10 * time.Second)
							} else {
								break
							}
						} else {
							break
						}
					}
					loopcount += 1
					if loopcount >= 4 {
						fmt.Println(hp.LipErrorStyle.Render("To many retries. Iperf test will exit."))
						break
					}
				}
			}

			if c.Interval > 0 {
				time.Sleep(time.Duration(c.Interval) * time.Minute)
			} else {
				break
			}
		}
		hp.PauseTerminalScreen()
		hp.ClearTerminalScreen()
	case menuTOP[3]:
		for {

			headerX := showHeaderPlusConfig(c)
			if menuSelect := bubbles.ShowMenuList("CHANGE SETTINGS", true, menuSettings, "111", headerX); menuSelect != "" {
				switch menuSelect {
				case menuSettings[0]:
					c.IperfS = getUserInputString("Enter Iperf Server IP and hit 'enter'")
					hp.ClearTerminalScreen()
				case menuSettings[1]:
					c.IperfP = getUserInputInt("Enter Iperf Port Number  and hit 'enter'")
					hp.ClearTerminalScreen()
				case menuSettings[2]:
					c.Interval = getUserInputInt("Enter Repeat Test Interval in Minutes and hit 'enter'")
					hp.ClearTerminalScreen()
				case menuSettings[3]:
					c.MSS = getUserInputInt("Enter desired MSS size and hit 'enter'")
					hp.ClearTerminalScreen()
				case menuSettings[4]:
					if c.CloudFrontTest {
						c.CloudFrontTest = false
					} else {
						c.CloudFrontTest = true
					}
				case menuSettings[5]:
					if c.MLabTest {
						c.MLabTest = false
					} else {
						c.MLabTest = true
					}
				case menuSettings[6]:
					if c.NetTest {
						c.NetTest = false
					} else {
						c.NetTest = true
					}
				case menuSettings[7]:
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
			fmt.Println(hp.LipErrorStyle.Render(fmt.Sprintf("Error Saving Config. %s", err)))
		} else {
			fmt.Println(hp.LipSystemMsgStyle.Render("Config saved"))
		}
		hp.PauseTerminalScreen()
		hp.ClearTerminalScreen()
	}
}

func showHeaderPlusConfig(config *configSettings) string {
	var isps, mssCustom string
	if config.CloudFrontTest {
		isps += "CF,"
	}
	if config.MLabTest {
		isps += "ML,"
	}
	if config.NetTest {
		isps += "NET"
	}

	if config.MSS == 0 {
		mssCustom = "Auto"
	} else {
		mssCustom = strconv.Itoa(config.MSS)
	}

	header := hp.LipHeaderStyle.Render(hp.MenuHeader) + "\n" +
		hp.LipConfigStyle.Render(fmt.Sprintf("Iperf:%s->%v  MSS:%s  Tests:%s  Browser:%v  Repeat:%vmin\n\n",
			config.IperfS, config.IperfP, mssCustom, isps, config.ShowBrowser, config.Interval))
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
	fmt.Println(hp.LipSystemMsgStyle.Render(msg))
	fmt.Scanln(&input)
	return input
}

func getUserInputInt(msg string) int {
	var input string
	fmt.Println(hp.LipSystemMsgStyle.Render(msg))
	fmt.Scanln(&input)
	num, err := strconv.Atoi(input)
	if err != nil {
		fmt.Println(hp.LipErrorStyle.Render("Entry must be a numeric number"))
		return 0
	}
	return num
}
