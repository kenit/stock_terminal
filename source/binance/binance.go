package binance

import (
	"encoding/json"
	//"fmt"
	"fmt"
	"github.com/gorilla/websocket"
	"net/http"
	"github.com/kenit/stock_terminal/source"
	"strings"
)

const URL = "wss://fstream.binance.com/stream"


type Binance struct {
	conn        *websocket.Conn
	close       chan interface{}
	messageChan chan *source.Broadcast
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
				symbol + "@depth@0ms",
			}...)
		}
	}

	request := &Request{
		Method: "SUBSCRIBE",
		Params: subParams,
		Id:     requestId,
	}
	if err := c.WriteJSON(request); err != nil{
		return err
	}
	requestId++

	m := make(chan *source.Broadcast, 1)
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
					nextChan := make(chan *source.Broadcast, 1)
					currChan <- &source.Broadcast{
						CurrChan: currChan,
						NextChan: nextChan,
						Message:  &message,
					}
					currChan = nextChan
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
				m <- b.Message
				currChan = b.NextChan
				b.CurrChan <- b
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

func (b *Binance) GetSnapshot() (interface{},error){
	var snapshot Snapshot
	symbol := b.focus
	if response, err := http.Get("https://fapi.binance.com/fapi/v1/depth?limit=1000&symbol=" + symbol); err == nil{

		if err = json.NewDecoder(response.Body).Decode(&snapshot); err == nil{
			snapshot.Symbol = symbol
			return snapshot, nil
		}else{
			return nil, err
		}
	}else{
		return nil, err
	}
}

func (b *Binance) SetFocus(symbol string) {

	request := &Request{
		Method: "UNSUBSCRIBE",
		Params: []string{
			 strings.ToLower(b.focus) + "@ticker",
			 strings.ToLower(b.focus) + "@depth@0ms",
		},
		Id:     requestId,
	}
	b.conn.WriteJSON(request)
	requestId++



	request = &Request{
		Method: "SUBSCRIBE",
		Params: []string{
			strings.ToLower(symbol) + "@ticker",
			strings.ToLower(symbol) + "@depth@0ms",
		},
		Id:     requestId,
	}
	b.conn.WriteJSON(request)
	requestId++

	b.focus = symbol
}

func init() {
	var api source.Source = &Binance{close: make(chan interface{}), focus: symbols[0]}
	source.Register(api)
}
