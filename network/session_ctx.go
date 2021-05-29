package network

//internal static readonly int NSVSNDHS = 311;
//internal static readonly int NSVSNRDS = 312;
//internal static readonly int NSVSNRDR = 312;
//internal static readonly int NSVSNDHO = 312;
//internal static readonly int NSVSNDHE = 314;
//internal static readonly int NSVSNIP6 = 314;
//internal static readonly int NSVSNSRN = 313;
//internal static readonly int NSVSNPPP = 313;

type SessionContext struct {
	//conn net.Conn
	connOption ConnectionOption
	//PortNo int
	//InstanceName string
	//HostName string
	//IPAddress string
	//Protocol string
	//ServiceName string
	SID []byte
	//internal Stream m_socketStream;
	//internal Socket m_socket;
	//internal ReaderStream m_readerStream;
	//internal WriterStream m_writerStream;
	//internal ITransportAdapter m_transportAdapter;
	//ConnectData string
	Version           uint16
	LoVersion         uint16
	Options           uint16
	NegotiatedOptions uint16
	OurOne            uint16
	Histone           uint16
	ReconAddr         string
	//internal Ano m_ano;
	//internal bool m_bAnoEnabled;
	ACFL0               uint8
	ACFL1               uint8
	SessionDataUnit     uint16
	TransportDataUnit   uint16
	UsingAsyncReceivers bool
	IsNTConnected       bool
	OnBreakReset        bool
	GotReset            bool
}

func NewSessionContext(connOption ConnectionOption) *SessionContext {
	return &SessionContext{
		SessionDataUnit:   connOption.SessionDataUnitSize,
		TransportDataUnit: connOption.TransportDataUnitSize,
		Version:           312,
		LoVersion:         300,
		Options:           1 | 1024 | 2048,
		OurOne:            1,
		connOption:        connOption,
	}
}
