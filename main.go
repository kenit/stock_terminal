package main

import (
	"fmt"
	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"stock_terminal/source"
	"stock_terminal/source/binance"
	"strconv"
	"sync"
	"time"
)

type Snapshot struct {
	UpdateId int
	Symbol   string
	Bids     [][2]float64
	Asks     [][2]float64
}

var (
	widgetArray     []ui.Drawable
	currentSymbol   string
	currentSnapshot *Snapshot
	lock            sync.Mutex
)

func main() {

	c := source.GetSource()
	if err := c.Conn(); err != nil {
		fmt.Printf("failed to connect to source, %s", err)
	}
	defer c.Close()

	if err := ui.Init(); err != nil {
		fmt.Printf("failed to initialize termui: %v", err)
	}
	defer ui.Close()

	klineChan := make(chan *KlineData)
	klineWidget := NewKlineWidget(klineChan)
	widgetArray = append(widgetArray, klineWidget)

	depthChan := make(chan *DepthData, 1000)
	askWidget, bidWidget := NewDepthWidget(depthChan)
	widgetArray = append(widgetArray, askWidget, bidWidget)

	t1 := widgets.NewParagraph()
	t2 := widgets.NewParagraph()
	t1.SetRect(10, 0, 35, 4)
	t2.SetRect(35, 0, 65, 4)
	t1.Border = false
	t2.Title = "24H"
	widgetArray = append(widgetArray, t1, t2)

	debug := widgets.NewParagraph()
	debug.SetRect(200, 0, 100, 4)
	//widgetArray = append(widgetArray, debug)

	selection := &Selection{}
	selection.StartYCoordinate = 20
	symbolChan, selectWidgets := selection.Init(c.GetSymbols())
	currentSymbol = selection.GetValue()

	for _, w := range selectWidgets {
		widgetArray = append(widgetArray, w)
	}

	symbolName := widgets.NewParagraph()
	symbolName.SetRect(0, 0, 10, 4)
	symbolName.Border = false
	symbolName.Text = currentSymbol
	widgetArray = append(widgetArray, symbolName)

	ch := make(chan int)
	go func() {
		for e := range ui.PollEvents() {
			if e.Type == ui.KeyboardEvent && (e.ID == "<C-c>" || e.ID == "q") {
				ch <- 1
				return
			}
		}
	}()

	go func(){
		for{
			ui.Render(widgetArray...)
			time.Sleep(100 * time.Millisecond)
		}
	}()

	messageChan := c.PollMessage()
	for {
		select {
		case <-ch:
			return
		case symbol := <-symbolChan:
			currentSymbol = symbol
			symbolName.Text = currentSymbol
			c.SetFocus(symbol)
			currentSnapshot = nil
		case message := <-messageChan:

			msg := message.(*binance.Message)
			switch msg.Data.Type {
			case "kline":
				openPrice, _ := strconv.ParseFloat(msg.Data.Kline.Open, 64)
				closePrice, _ := strconv.ParseFloat(msg.Data.Kline.Close, 64)
				highPrice, _ := strconv.ParseFloat(msg.Data.Kline.High, 64)
				lowPrice, _ := strconv.ParseFloat(msg.Data.Kline.Low, 64)

				klineChan <- &KlineData{
					msg.Data.Symbol,
					msg.Data.Kline.StartTime,
					msg.Data.Kline.EndTime,
					openPrice,
					closePrice,
					highPrice,
					lowPrice,
				}
			case "markPriceUpdate":
				price, _ := strconv.ParseFloat(msg.Data.Price, 64)
				selection.SetPrice(msg.Data.Symbol, price)
			case "24hrTicker":
				if msg.Data.Symbol == currentSymbol {
					lastPrice, _ := strconv.ParseFloat(msg.Data.LastPrice, 64)
					openPrice, _ := strconv.ParseFloat(msg.Data.Open, 64)

					if lastPrice > openPrice {
						t1.Text = fmt.Sprintf("%4.3f %s %s", lastPrice, msg.Data.Price, msg.Data.ChangePercent)
						t1.TextStyle.Fg = ui.ColorRed
					} else {
						t1.Text = fmt.Sprintf("%4.3f %s %s", lastPrice, msg.Data.Price, msg.Data.ChangePercent)
						t1.TextStyle.Fg = ui.ColorGreen
					}

					t2.Text = fmt.Sprintf("Low: %s   High: %s", msg.Data.Low, msg.Data.High)
				}
			case "depthUpdate":
				//fmt.Printf("%i, %i, %i\n", msg.Data.FirstUpdateId, msg.Data.FinalUpdateId, msg.Data.OldUpdateId)
				if msg.Data.Symbol == currentSymbol {
					depth := &DepthData{
						Symbol:        msg.Data.Symbol,
						FirstUpdateId: msg.Data.FirstUpdateId,
						FinalUpdateId: msg.Data.FinalUpdateId,
						OldUpdateId:   msg.Data.OldUpdateId,
					}
					for _, askData := range msg.Data.Ask {
						price, _ := strconv.ParseFloat(askData[0], 64)
						qty, _ := strconv.ParseFloat(askData[1], 64)
						depth.Ask = append(depth.Ask, [2]float64{price, qty})
					}

					for _, bidData := range msg.Data.Bid {
						price, _ := strconv.ParseFloat(bidData[0], 64)
						qty, _ := strconv.ParseFloat(bidData[1], 64)
						depth.Bid = append(depth.Bid, [2]float64{price, qty})
					}

					depthChan <- depth
				}
			}

		}

	}
}

func GetSnapshot() {
	c := source.GetSource()
	if s, err := c.GetSnapshot(); err == nil {
		lock.Lock()
		snapshot := s.(binance.Snapshot)

		currentSnapshot = &Snapshot{
			UpdateId: snapshot.LastUpdateId,
			Symbol:   snapshot.Symbol,
		}

		for _, ask := range snapshot.Asks {
			price, _ := strconv.ParseFloat(ask[0], 64)
			qty, _ := strconv.ParseFloat(ask[1], 64)
			currentSnapshot.Asks = append(currentSnapshot.Asks, [2]float64{price, qty})
		}

		for _, bid := range snapshot.Bids {
			price, _ := strconv.ParseFloat(bid[0], 64)
			qty, _ := strconv.ParseFloat(bid[1], 64)
			currentSnapshot.Bids = append(currentSnapshot.Bids, [2]float64{price, qty})
		}

		lock.Unlock()
	} else {
		fmt.Println(err)
	}
}
