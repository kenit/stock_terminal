package main

import (
	"fmt"
	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"github.com/kenit/stock_terminal/source"
	"github.com/kenit/stock_terminal/source/binance"
	"sort"
	"strconv"
)

type DepthData struct {
	Symbol        string
	FirstUpdateId int
	FinalUpdateId int
	OldUpdateId   int
	Bid           [][2]float64
	Ask           [][2]float64
}

func NewDepthWidget(depthChan chan *DepthData) (*widgets.List, *widgets.List) {
	askWidget := widgets.NewList()
	askWidget.SetRect(0, 4, 30, 12)
	askWidget.TextStyle = ui.NewStyle(ui.ColorRed)
	bidWidget := widgets.NewList()
	bidWidget.SetRect(0, 12, 30, 20)
	bidWidget.TextStyle = ui.NewStyle(ui.ColorGreen)

	go func() {

		started := false
		lastFinalUpdateId := 0
		qq := 0

		for depth := range depthChan {

			if depth.Symbol != currentSymbol {
				continue
			}

			if currentSnapshot == nil || currentSnapshot.Symbol != currentSymbol {
				GetSnapshot()
				started = false
				lastFinalUpdateId = 0
			}

			if !started && (depth.FinalUpdateId < currentSnapshot.UpdateId || depth.FirstUpdateId > currentSnapshot.UpdateId) {
				qq++
				if qq > 30 {
					GetSnapshot()
					qq = 0
				}
				continue
			} else {
				qq = 0
				started = true
			}

			if lastFinalUpdateId != 0 && lastFinalUpdateId != depth.OldUpdateId {
				GetSnapshot()
				started = false
				lastFinalUpdateId = 0
				continue
			}

			lastFinalUpdateId = depth.FinalUpdateId

			lock.Lock()
			for _, ask := range depth.Ask {
				index := sort.Search(len(currentSnapshot.Asks), func(i int) bool { return currentSnapshot.Asks[i][0] >= ask[0] })
				switch {
				case index < len(currentSnapshot.Asks) && ask[0] == currentSnapshot.Asks[index][0]:
					currentSnapshot.Asks[index][0] = ask[0]
				default:
					if index >= len(currentSnapshot.Asks) {

						currentSnapshot.Asks = append(currentSnapshot.Asks, ask)
					} else {
						currentSnapshot.Asks = append(currentSnapshot.Asks, [2]float64{})
						copy(currentSnapshot.Asks[index+1:], currentSnapshot.Asks[index:])
						currentSnapshot.Asks[index] = ask
					}
				}
			}

			for _, bid := range depth.Bid {
				index := sort.Search(len(currentSnapshot.Bids), func(i int) bool { return currentSnapshot.Bids[i][0] <= bid[0] })
				switch {
				case index < len(currentSnapshot.Bids) && bid[0] == currentSnapshot.Bids[index][0]:
					currentSnapshot.Bids[index][0] = bid[0]
				default:
					if index >= len(currentSnapshot.Bids) {
						currentSnapshot.Bids = append(currentSnapshot.Bids, bid)
					} else {
						currentSnapshot.Bids = append(currentSnapshot.Bids, [2]float64{})
						copy(currentSnapshot.Bids[index+1:], currentSnapshot.Bids[index:])
						currentSnapshot.Bids[index] = bid
					}
				}
			}
			lock.Unlock()
		}
	}()

	go func() {
		for message := range source.GetSource().PollMessage() {

			if currentSnapshot == nil {
				continue
			}

			var askRow, bidRow []string
			askRow = append(askRow, "Price         Amount")
			bidRow = append(bidRow, "Price         Amount")
			msg := message.(*binance.Message)
			if msg.Data.Type == "kline" && msg.Data.Symbol == currentSymbol {

				closePrice, _ := strconv.ParseFloat(msg.Data.Kline.Close, 64)
				lock.Lock()
				index := sort.Search(len(currentSnapshot.Asks), func(i int) bool { return currentSnapshot.Asks[i][0] >= closePrice })
				for i := index; i < len(currentSnapshot.Asks) && len(askRow) < 6; i++ {
					price := currentSnapshot.Asks[i]
					if price[1] == 0 {
						continue
					}
					askRow = append(askRow, "")
					copy(askRow[2:], askRow[1:])
					askRow[1] = fmt.Sprintf("%4.3f      %.3f", price[0], price[1])
				}

				index = sort.Search(len(currentSnapshot.Bids), func(i int) bool { return currentSnapshot.Bids[i][0] <= closePrice })

				for i := index; i < len(currentSnapshot.Bids) && len(bidRow) < 6; i++ {
					price := currentSnapshot.Bids[i]
					if price[1] == 0 {
						continue
					}
					bidRow = append(bidRow, fmt.Sprintf("%4.3f      %.3f", price[0], price[1]))
				}
				lock.Unlock()

				askWidget.Rows = askRow
				bidWidget.Rows = bidRow
			}
		}

	}()
	return askWidget, bidWidget
}
