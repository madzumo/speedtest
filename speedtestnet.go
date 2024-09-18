package main

import (
	"fmt"

	"github.com/showwin/speedtest-go/speedtest"
)

func getServerLists() {
	var speedtestClient = speedtest.New()

	user, _ := speedtestClient.FetchUserInfo()

	serverList, _ := speedtestClient.FetchServers()

	fmt.Println(user)
	fmt.Println("*********************************")
	fmt.Println(serverList)
	fmt.Println("*********************************")
}

func runSpeedTestNet() (testresult string) {
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
		testresult = fmt.Sprintf("🌎Internet Test-> Down:%s, Up:%s, Latency:%s", s.DLSpeed, s.ULSpeed, s.Latency)
		// fmt.Println(result)
		s.Context.Reset() // reset counter
	}
	return testresult
}
