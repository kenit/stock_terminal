package main

import (
	"fmt"
	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"stock_terminal/source"
	"stock_terminal/source/binance"
	"strconv"
)

var (
	widgetArray []ui.Drawable
	currentSymbol string
)

func main() {

	c := source.GetConn()
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

	depthChan := make(chan *DepthData)
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
	debug.SetRect(200, 1, 100, 5)
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

	messageChan := c.PollMessage()
	for {
		select {
		case <-ch:
			return
		case symbol := <-symbolChan:
			currentSymbol = symbol
			symbolName.Text = currentSymbol
			c.SetFocus(symbol)
			break
		case message := <-messageChan:
			msg := message.(*binance.Message)
			switch msg.Data.Type {
			case "kline":
				open, _ := strconv.ParseFloat(msg.Data.Kline.Open, 64)
				close, _ := strconv.ParseFloat(msg.Data.Kline.Close, 64)
				high, _ := strconv.ParseFloat(msg.Data.Kline.High, 64)
				low, _ := strconv.ParseFloat(msg.Data.Kline.Low, 64)

				klineChan <- &KlineData{
					msg.Data.Symbol,
					msg.Data.Kline.StartTime,
					msg.Data.Kline.EndTime,
					open,
					close,
					high,
					low,
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
					debug.Text = fmt.Sprintf("%f, %f", lastPrice, openPrice)
					t2.Text = fmt.Sprintf("Low: %s   High: %s", msg.Data.Low, msg.Data.High)
				}
			case "depthUpdate":
				if msg.Data.Symbol == currentSymbol {
					depth := &DepthData{}
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
		ui.Render(widgetArray...)
	}
}
