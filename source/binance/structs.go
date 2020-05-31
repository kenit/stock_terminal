package binance

type Request struct {
	Method string   `json:"method"`
	Params []string `json:"params"`
	Id     int      `json:"id"`
}

type Message struct {
	Stream string `json:"stream"`
	Data   Data   `json:"data"`
}

type Data struct {
	Type      string `json:"e"`
	EventTime int    `json:"E"`
	Symbol    string `json:"s"`
	Price     string `json:"p"`
	Kline     Kline  `json:"k"`
	Ticker
	Depth
}

type Kline struct {
	StartTime   int    `json:"t"`
	EndTime     int    `json:"T"`
	Open        string `json:"o"`
	Close       string `json:"c"`
	High        string `json:"h"`
	Low         string `json:"l"`
	IsClosed    bool   `json:"x"`
	LastTradeId int    `json:"L"`
}

type Ticker struct {
	ChangePercent string `json:"P"`
	Open          string `json:"o"`
	High          string `json:"h"`
	Low           string `json:"l"`
	LastPrice     string `json:"c"`
	OpenTime      int    `json:"O"`
	CloseTime     int    `json:"C"`
	LastTradeId   int    `json:"L"`
}

type Depth struct {
	Bid [][2]string `json:"b"`
	Ask [][2]string `json:"a"`
}
