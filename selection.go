package main

import (
	"fmt"
	ui "github.com/gizak/termui/v3"
	uiWidgets"github.com/gizak/termui/v3/widgets"
)

type Selection struct {
	SelectedIndex    int
	options          []string
	widgets          []*uiWidgets.Paragraph
	StartYCoordinate int
}

func (self *Selection) Init(options []string) (<-chan string, []*uiWidgets.Paragraph) {
	self.options = options
	for i, option := range options {
		widget := uiWidgets.NewParagraph()
		startYCoordinate := self.StartYCoordinate + 2 * i
		widget.SetRect(0, startYCoordinate, 30, startYCoordinate + 3)
		widget.Text = fmt.Sprintf("%s         ", option)
		widget.Border = false
		self.widgets = append(self.widgets, widget)
	}

	self.widgets[0].TextStyle = ui.NewStyle(ui.ColorBlack, ui.ColorWhite)
	c := make(chan string, 50)

	go func() {
		for e := range ui.PollEvents() {
			if e.Type == ui.KeyboardEvent {
				self.widgets[self.SelectedIndex].TextStyle = ui.NewStyle(ui.ColorWhite, ui.ColorClear)
				switch e.ID{
				case "<Up>":
					if self.SelectedIndex > 0 {
						self.SelectedIndex--
						c <- self.options[self.SelectedIndex]
					}
				case "<Down>":
					if self.SelectedIndex < len(self.options) - 1 {
						self.SelectedIndex++
						c <- self.options[self.SelectedIndex]

					}
				}
				self.widgets[self.SelectedIndex].TextStyle = ui.NewStyle(ui.ColorBlack, ui.ColorWhite)
			}

		}
	}()

	return c, self.widgets
}

func (self *Selection) GetValue() string{
	return self.options[self.SelectedIndex]
}

func (self *Selection) SetPrice(symbol string, price float64){
	for i, option := range self.options{
		if option == symbol{
			self.widgets[i].Text = fmt.Sprintf("%s           %4.2f", option, price)
			return
		}
	}
}