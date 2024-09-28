package tests

import (
	"fmt"

	hp "github.com/madzumo/speedtest/internal/helpers"
	"github.com/showwin/speedtest-go/speedtest"
)

// func getServerLists() {
// 	var speedtestClient = speedtest.New()

// 	user, _ := speedtestClient.FetchUserInfo()

// 	serverList, _ := speedtestClient.FetchServers()

// 	fmt.Println(user)
// 	fmt.Println("*********************************")
// 	fmt.Println(serverList)
// 	fmt.Println("*********************************")
// }

func NETTest() {
	var testResult string
	// quit := make(chan struct{})
	// go bubbles.ShowSpinner(quit, "Speedtest.NET Test....", "196") // Run spinner in a goroutine

	var speedtestClient = speedtest.New()
	// Get user's network information
	// user, _ := speedtestClient.FetchUserInfo()

	// Get a list of servers near a specified location
	// user.SetLocationByCity("Tokyo")
	// user.SetLocation("Osaka", 34.6952, 135.5006)

	serverList, _ := speedtestClient.FetchServers()
	targets, _ := serverList.FindServer([]int{})

	for _, s := range targets {
		s.PingTest(nil)
		s.DownloadTest()
		s.UploadTest()
		// Note: The unit of s.DLSpeed, s.ULSpeed is bytes per second, this is a float64.
		// testResult = fmt.Sprintf("🌎SpeedTest.Net-> Down:%s, Up:%s, Latency:%s", s.DLSpeed, s.ULSpeed, s.Latency)
		testResult = fmt.Sprintf("SpeedTest.Net-> Down:%s, Up:%s", s.DLSpeed, s.ULSpeed)

		s.Context.Reset() // reset counter
	}
	// close(quit)
	// time.Sleep(1 * time.Second)
	fmt.Println(hp.LipOutputStyle.Render(testResult))
	hp.WriteLogFile(fmt.Sprintf("⚡%s", testResult))
}
