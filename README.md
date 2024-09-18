# Speed Test with Iperf3

This is a self running app that runs iperf3 speed test to a remote server host. It runs every 10 minutes and saves the result to a log file in the same location as the executable. You can run this utility on multiple clients with just one iperf3 server. Options you can change: port number, frequency time interval, and MSS segment size for various testing.

## How to Run
- Setup iperf on a remote server 
```
sudo apt install iperf3
iperf3 -s
```

- Run latest release
- Option 1: Enter the severs Server IP address
- Option 5: Run the test with default settings

![Menu](readme.png)