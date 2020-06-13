# Stock Terminal
## Introduction
Use unix terminal to monitor cryptocurrency price in Binance.


## Install 
1. Make you have the latest go and dep installed on your PC.
2. `go get -u github.com/kenit/stock_terminal`
3. `cd $GOPATH/src/github.com/kenit/stock_terminal`
4. `dep ensure`
3. `go install`

## Execute
`$GOPATH/bin/stock_terminal`


## Todo
-[ ] Allow to add currency
-[X] Change width of Kline chart when size of terminal is changed

All ideas are welcome, please provide it in issue.

## Known issue
1. User will encounter error if switch bewteeen currency to fastly.