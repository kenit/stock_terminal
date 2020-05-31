package source

type Api interface {
	Conn() error
	Close()
	PollMessage() <-chan interface{}
	GetSymbols() []string
	SetFocus(string)
}

var (
	apiConn Api
)

func Register(api Api){
	apiConn = api
}

func GetConn() Api{
	return apiConn
}