package source

type Broadcast struct {
	CurrChan chan *Broadcast
	NextChan chan *Broadcast
	Message  interface{}
}

type Source interface {
	Conn() error
	Close()
	PollMessage() <-chan interface{}
	GetSymbols() []string
	SetFocus(string)
	GetSnapshot() (interface{},error)
}

var (
	sourceConn Source
)

func Register(api Source){
	sourceConn = api
}

func GetSource() Source{
	return sourceConn
}