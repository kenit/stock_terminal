package main

import (
	"fmt"
	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
)

type DepthData struct{
	Bid [][2]float64
	Ask [][2]float64
}

func NewDepthWidget(depthChan chan *DepthData) (*widgets.List, *widgets.List){
	askWidget := widgets.NewList()
	askWidget.SetRect(0, 4, 30, 12)
	askWidget.TextStyle = ui.NewStyle(ui.ColorRed)
	bidWidget := widgets.NewList()
	bidWidget.SetRect(0, 12, 30, 20)
	bidWidget.TextStyle = ui.NewStyle(ui.ColorGreen)

	go func(){
		for depth := range depthChan{
			askRow := make([]string, 6)
			bidRow := make([]string, 6)
			askRow[0] = "Price         Amount"
			bidRow[0] = "Price         Amount"
			for i, price := range depth.Ask{
				askRow[5 - i] = fmt.Sprintf("%4.3f      %.3f", price[0], price[1])
			}

			for i, price := range depth.Bid{
				bidRow[i+1] = fmt.Sprintf("%4.3f      %.3f", price[0], price[1])
			}

			askWidget.Rows = askRow
			bidWidget.Rows = bidRow
		}
	}()
	return askWidget, bidWidget
}
