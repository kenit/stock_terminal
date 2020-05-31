package main

import (
	"fmt"
	"github.com/gizak/termui/v3/widgets"
)

func NewPriceWidget(p chan float64) *widgets.Paragraph{
	widget := widgets.NewParagraph()
	widget.SetRect(0, 1, 20, 5)

	go func() {
		for price := range p {
			widget.Text = fmt.Sprintf("%4.2f", price)
		}
	}()
	return widget
}