# stock_terminal
## Introduction
Use unix terminal to monitor cryptocurrency price in Binance

## Install 
1. Make you have the latest go and dep installed on your PC.
2. `go get -u github.com/kenit/stock_terminal`
3. `cd $GOPATH/src/github.com/kenit/stock_terminal`
4. `dep ensure`
3. `go install`

## Execute
`$GOPATH/bin/stock_terminal`


## Todo
1. Allow to add currency
2. Change width of Kline chart when size of terminal is changed

## Known issue
1. User will encounter error if switch bewteeen currency to fastly.
