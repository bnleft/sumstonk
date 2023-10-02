# sumstonk

### Goal
Give people a general feel of a stock's performance

### Why
- Make a faster decision to invest
- No need to read a bunch of articles
- No need to sift through social media

### Technologies
- Go
- Metaphor API
- SMMRY API
- Google Cloud Natural Language API

### Current Features
- Get today's sentiment for any stonk

### Setup
- Install [Go](https://go.dev/doc/install)
- Get [Metaphor](https://dashboard.metaphor.systems/) API key
- Get [Smmry](https://smmry.com/api) API key
- Setup [GCP](https://cloud.google.com/docs/authentication/provide-credentials-adc#local-dev)
- Populate keys in `.env`
- Run `go run main.go`

### Demo
![image](https://github.com/bnleft/sumstonk/assets/80173797/c6fdc598-9bbf-48b3-8260-b18b01dc252f)
![image](https://github.com/bnleft/sumstonk/assets/80173797/2fd63ab6-5106-4aa0-aa85-f5787aca0c73)

### Future Features
- Configure number of sources
- Sophisticated summary API
- Support for relevant sites (like Reddit, CNBC, Bloomberg)
- Site cleaning (we don't want "Top 10s" articles)
- Bypass paywall
- Summary for any stonk
- Input by company name
- Website App (preferred, but would have exceeded API usage)
- Filtered by date
    - Today
    - Last Week
    - Last Month
    - Last 3 Months
    - Last Year
- Filtered by industry
- Colored by price fluctuations
- Add cryptocurrency
