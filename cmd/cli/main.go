package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	hp "github.com/madzumo/speedtest/internal/helpers"
)

var (
	configFileName = "settings.json"
)

type configSettings struct {
	IperfS         string      `json:"iperfServer"`
	IperfP         int         `json:"iperfPort"`
	Interval       int         `json:"repeatInterval"`
	MSS            int         `json:"MSS"`
	CloudFrontTest bool        `json:"CloudFrontTest"`
	MLabTest       bool        `json:"MLabTest"`
	NetTest        bool        `json:"SpeedNetTest"`
	ShowBrowser    bool        `json:"showBrowser"`
	EmailSettings  hp.EmailJob `json:"emailSettings"`
	IperfTimeout   int         `json:"iperfTimeOut"`
}

func main() {
	// fig := figure.NewFigure("Welcome", "big", false)
	// fig.Print()
	// return
	hp.SetPEMfiles()
	config, _ := getConfigSettings()
	headerX, headerIP := showHeaderPlusConfigPlusIP(config, false, false)
	ShowMenuList(headerX, headerIP, config)
	// fmt.Println(filepath.Abs(hp.GetLogFileName()))
}

func showHeaderPlusConfigPlusIP(config *configSettings, settingsMenu bool, emailMenu bool) (string, string) {
	var header string
	myIP := hp.GetLocalIP()

	if emailMenu {
		var emailMethod string
		var fromMethod string
		var hostMethod string
		var portMethod string
		if config.EmailSettings.UseOutlook {
			emailMethod = "Outlook"
			fromMethod = "Outlook"
			hostMethod = "EXCH"
			portMethod = ""
		} else if config.EmailSettings.UseSMTP {
			emailMethod = "SMTP"
			fromMethod = config.EmailSettings.From
			hostMethod = config.EmailSettings.SMTPHost
			portMethod = config.EmailSettings.SMTPPort
		} else {
			emailMethod = "OFF"
			fromMethod = "-"
			hostMethod = "-"
			portMethod = ""
		}
		header = hp.LipHeaderStyle.Render(hp.MenuHeader) + "\n" +
			hp.LipConfigSMTPStyle.Render(fmt.Sprintf("Method:%s  Host:%s:%v  From:%s To:%s",
				emailMethod, hostMethod, portMethod,
				fromMethod, config.EmailSettings.To)) + "\n" +
			hp.LipFooterStyle.Render(fmt.Sprintf("Your IP:%s\n\n", myIP))
	} else {
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
		if settingsMenu {
			header = hp.LipHeaderStyle.Render(hp.MenuHeader) + "\n" +
				hp.LipConfigSettingsStyle.Render(fmt.Sprintf("Iperf:%s:%v  MSS:%s  Tests:%s  Browser:%v  Repeat:%vmin",
					config.IperfS, config.IperfP, mssCustom, isps, config.ShowBrowser, config.Interval)) + "\n" +
				hp.LipFooterStyle.Render(fmt.Sprintf("Your IP:%s\n\n", myIP))
		} else {
			header = hp.LipHeaderStyle.Render(hp.MenuHeader) + "\n" +
				hp.LipConfigStyle.Render(fmt.Sprintf("Iperf:%s:%v  MSS:%s  Tests:%s  Browser:%v  Repeat:%vmin",
					config.IperfS, config.IperfP, mssCustom, isps, config.ShowBrowser, config.Interval)) + "\n" +
				hp.LipFooterStyle.Render(fmt.Sprintf("Your IP:%s\n\n", myIP))
		}
	}
	return header, myIP
}

func getConfigSettings() (*configSettings, error) {

	configTemp := configSettings{
		IperfS:         "0.0.0.0",
		IperfP:         5201,
		Interval:       0,
		MSS:            0,
		CloudFrontTest: true,
		MLabTest:       true,
		NetTest:        true,
		ShowBrowser:    false,
		EmailSettings: hp.EmailJob{
			From:     "sender@domain.com",
			To:       "recipient@domain.com",
			Subject:  "Speed Test Report",
			Body:     "Speed Test Report Incoming!",
			SMTPHost: "smtp.domain.com",
			SMTPPort: "587",
			UserName: "user",
			PassWord: "password",
		},
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
