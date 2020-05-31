package main

import (
	"fmt"
	"image"
	"math"
	rw "github.com/mattn/go-runewidth"
	. "github.com/gizak/termui/v3"
)

type KlineChart struct {
	Block
	BarColors    []Color
	LabelStyles  []Style
	NumStyles    []Style // only Fg and Modifier are used
	NumFormatter func(float64) string
	Data         []*KlineData
	Labels       []string
	BarWidth     int
	BarGap       int
}

type KlineData struct {
	Symbol    string
	StartTime int
	EndTime   int
	Open      float64
	Close     float64
	High      float64
	Low       float64
}

func NewKlineChart() *KlineChart {
	return &KlineChart{
		Block:        *NewBlock(),
		BarColors:    Theme.StackedBarChart.Bars,
		LabelStyles:  Theme.StackedBarChart.Labels,
		NumStyles:    Theme.StackedBarChart.Nums,
		NumFormatter: func(n float64) string { return fmt.Sprint(n) },
		BarGap:       1,
		BarWidth:     3,
	}
}

func (d *KlineData) Max() float64 {
	return math.Max(d.Open, math.Max(d.Close, d.High))
}

func (d *KlineData) Min() float64 {
	return math.Min(d.Open, math.Min(d.Close, d.Low))
}

func (k *KlineChart) Draw(buf *Buffer) {
	k.Block.Draw(buf)

	maxVal := 0.0
	minVal := 10000.0

	for _, data := range k.Data {
		maxVal = math.Max(maxVal, data.Max())
		minVal = math.Min(minVal, data.Min())
	}

	maxVal = maxVal * 1.0000005
	minVal = minVal * 0.9999995

	barXCoordinate := k.Inner.Min.X

	for i, bar := range k.Data {

		y0 := int(((maxVal - bar.High) / (maxVal - minVal)) * float64(k.Inner.Dy()-1))
		y1 := int(((maxVal - bar.Low) / (maxVal - minVal)) * float64(k.Inner.Dy()-1))

		for y := (k.Inner.Min.Y + 2) + y0; y < (k.Inner.Min.Y+2)+y1; y++ {
			c := NewCell(' ', NewStyle(ColorClear, ColorWhite))
			buf.SetCell(c, image.Pt(int(barXCoordinate+k.BarWidth/2), y))
		}

		y0 = int(((maxVal - math.Max(bar.Open, bar.Close)) / (maxVal - minVal)) * float64(k.Inner.Dy()-1))
		y1 = int(((maxVal - math.Min(bar.Open, bar.Close)) / (maxVal - minVal)) * float64(k.Inner.Dy()-1))

		for x := barXCoordinate; x < MinInt(barXCoordinate+k.BarWidth, k.Inner.Max.X); x++ {
			for y := (k.Inner.Min.Y + 2) + y0; y < (k.Inner.Min.Y+2)+y1; y++ {

				var color Color
				if bar.Close > bar.Open {
					color = ColorRed
				} else {
					color = ColorGreen
				}

				c := NewCell(' ', NewStyle(ColorClear, color))

				buf.SetCell(c, image.Pt(x, y))
			}
		}

		if i < len(k.Labels) {
			labelXCoordinate := barXCoordinate + MaxInt(
				int(float64(k.BarWidth)/2)-int(float64(rw.StringWidth(k.Labels[i]))/2),
				0,
			)
			buf.SetString(
				TrimString(k.Labels[i], k.BarWidth),
				SelectStyle(k.LabelStyles, i),
				image.Pt(labelXCoordinate, k.Inner.Max.Y-1),
			)
		}

		barXCoordinate += k.BarWidth + k.BarGap
	}
}

func NewKlineWidget(dataChan chan *KlineData) *KlineChart {
	allData := make(map[string][]*KlineData)
	widget := NewKlineChart()
	widget.SetRect(31, 4, 200, 40)
	widget.BarWidth = 5

	go func() {
		for data := range dataChan {

			if count := len(allData[data.Symbol]); count == 0 {
				allData[data.Symbol] = append(allData[data.Symbol], data)
			} else {
				latest := allData[data.Symbol][count-1]
				if latest.EndTime == data.EndTime {
					latest.High = data.High
					latest.Low = data.Low
					latest.Open = data.Open
					latest.Close = data.Close
				} else {
					if count > 30 {
						allData[data.Symbol] = append(allData[data.Symbol][1:], data)
					} else {
						allData[data.Symbol] = append(allData[data.Symbol], data)
					}
				}
			}

			widget.Data = allData[currentSymbol]

		}
	}()

	return widget
}
