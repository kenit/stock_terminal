package binance

import (
	//"fmt"
	"fmt"
	"github.com/gorilla/websocket"
	"stock_terminal/source"
	"strings"
)

const URL = "wss://fstream.binance.com/stream"

type broadcast struct {
	currChan chan *broadcast
	nextChan chan *broadcast
	message  *Message
}

type Binance struct {
	conn        *websocket.Conn
	close       chan interface{}
	messageChan chan *broadcast
	focus       string
}

var (
	requestId = 0
	symbols   = []string{
		"BTCUSDT",
		"ETHUSDT",
		"BNBUSDT",
		"ADAUSDT",
		"BCHUSDT",
		"LTCUSDT",
		"XRPUSDT",
	}
)

func (b *Binance) Conn() error {
	c, _, err := websocket.DefaultDialer.Dial(URL, nil)
	if err != nil {
		return err
	}
	b.conn = c

	var subParams []string
	for _, symbol := range symbols {
		symbol = strings.ToLower(symbol)
		subParams = append(subParams, []string{
			symbol + "@markPrice",
			symbol + "@kline_1m",
		}...)

		if symbol == strings.ToLower(b.focus) {
			subParams = append(subParams, []string{
				symbol + "@ticker",
				symbol + "@depth5",
			}...)
		}
	}

	request := &Request{
		Method: "SUBSCRIBE",
		Params: subParams,
		Id:     requestId,
	}
	c.WriteJSON(request)
	requestId++

	m := make(chan *broadcast, 1)
	b.messageChan = m
	go func() {
		var message Message
		currChan := m
		for {
			select {
			case <-b.close:
				c.Close()
				close(currChan)
				return
			default:
				if err := c.ReadJSON(&message); err == nil {
					nextChan := make(chan *broadcast, 1)
					currChan <- &broadcast{
						currChan: currChan,
						nextChan: nextChan,
						message:  &message,
					}
					currChan = nextChan
					b.messageChan = currChan
				} else {
					fmt.Println(err)
					close(currChan)
					return
				}
			}
		}
	}()

	return nil
}

func (b *Binance) PollMessage() <-chan interface{} {
	m := make(chan interface{}, 10)
	currChan := b.messageChan
	go func() {
		for {
			if b, ok := <-currChan; ok {
				m <- b.message
				currChan = b.nextChan
				b.currChan <- b
			} else {
				close(m)
				return
			}
		}
	}()
	return m
}

func (b *Binance) Close() {
	b.close <- 1
}

func (b *Binance) GetSymbols() []string {
	return symbols
}

func (b *Binance) SetFocus(symbol string) {

	request := &Request{
		Method: "UNSUBSCRIBE",
		Params: []string{
			 strings.ToLower(b.focus) + "@ticker",
			 strings.ToLower(b.focus) + "@depth5",
		},
		Id:     requestId,
	}
	b.conn.WriteJSON(request)
	requestId++

	request = &Request{
		Method: "SUBSCRIBE",
		Params: []string{
			strings.ToLower(symbol) + "@ticker",
			strings.ToLower(symbol) + "@depth5",
		},
		Id:     requestId,
	}
	b.conn.WriteJSON(request)
	requestId++

	b.focus = symbol
}

func init() {
	var api source.Api = &Binance{close: make(chan interface{}), focus: symbols[0]}
	source.Register(api)
}
