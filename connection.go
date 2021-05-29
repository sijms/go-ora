package go_ora

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"os"
	"os/user"
	"strconv"
	"strings"

	"github.com/sijms/go-ora/converters"
	"github.com/sijms/go-ora/network"
	"github.com/sijms/go-ora/trace"
)

type ConnectionState int

const (
	Closed ConnectionState = 0
	Opened ConnectionState = 1
)

type LogonMode int

const (
	NoNewPass LogonMode = 0x1
	//WithNewPass LogonMode = 0x2
	SysDba      LogonMode = 0x20 // no verify response from server
	SysOper     LogonMode = 0x40 // no verify response from server
	UserAndPass LogonMode = 0x100
	//PROXY       LogonMode = 0x400
)

type NLSData struct {
	Calender        string
	Comp            string
	LengthSemantics string
	NCharConvExcep  string
	DateLang        string
	Sort            string
	Currency        string
	DateFormat      string
	IsoCurrency     string
	NumericChars    string
	DualCurrency    string
	Timestamp       string
	TimestampTZ     string
}
type Connection struct {
	State             ConnectionState
	LogonMode         LogonMode
	autoCommit        bool
	conStr            *ConnectionString
	connOption        *network.ConnectionOption
	session           *network.Session
	tcpNego           *TCPNego
	dataNego          *DataTypeNego
	authObject        *AuthObject
	SessionProperties map[string]string
	dBVersion         *DBVersion
	sessionID         int
	serialID          int
	strConv           converters.IStringConverter
	NLSData           NLSData
}

type oracleDriver struct {
}

func init() {
	sql.Register("oracle", &oracleDriver{})
}
func (drv *oracleDriver) Open(name string) (driver.Conn, error) {

	conn, err := NewConnection(name)
	if err != nil {
		return nil, err
	}

	return conn, conn.Open()
}

func (conn *Connection) SetStringConveter(conveter converters.IStringConverter) {
	conn.strConv = conveter
	conn.session.StrConv = conveter
}

func (conn *Connection) GetNLS() (*NLSData, error) {

	// we read from nls_session_parameters ONCE
	cmdText := `
DECLARE
	err_code VARCHAR2(2000);
	err_msg  VARCHAR2(2000);
	BEGIN
		SELECT 
			MAX(CASE WHEN PARAMETER='NLS_CALENDAR' THEN VALUE END) AS NLS_CALENDAR,
			MAX(CASE WHEN PARAMETER='NLS_COMP' THEN VALUE END) AS NLS_COMP,
			MAX(CASE WHEN PARAMETER='NLS_LENGTH_SEMANTICS' THEN VALUE END) AS NLS_LENGTH_SEMANTICS,
			MAX(CASE WHEN PARAMETER='NLS_NCHAR_CONV_EXCP' THEN VALUE END) AS NLS_NCHAR_CONV_EXCP,
			MAX(CASE WHEN PARAMETER='NLS_DATE_LANGUAGE' THEN VALUE END) AS NLS_DATE_LANGUAGE,
			MAX(CASE WHEN PARAMETER='NLS_SORT' THEN VALUE END) AS NLS_SORT,
			MAX(CASE WHEN PARAMETER='NLS_CURRENCY' THEN VALUE END) AS NLS_CURRENCY,
			MAX(CASE WHEN PARAMETER='NLS_DATE_FORMAT' THEN VALUE END) AS NLS_DATE_FORMAT,
			MAX(CASE WHEN PARAMETER='NLS_ISO_CURRENCY' THEN VALUE END) AS NLS_ISO_CURRENCY,
			MAX(CASE WHEN PARAMETER='NLS_NUMERIC_CHARACTERS' THEN VALUE END) AS NLS_NUMERIC_CHARACTERS,
			MAX(CASE WHEN PARAMETER='NLS_DUAL_CURRENCY' THEN VALUE END) AS NLS_DUAL_CURRENCY,
			MAX(CASE WHEN PARAMETER='NLS_TIMESTAMP_FORMAT' THEN VALUE END) AS NLS_TIMESTAMP_FORMAT,
			MAX(CASE WHEN PARAMETER='NLS_TIMESTAMP_TZ_FORMAT' THEN VALUE END) AS NLS_TIMESTAMP_TZ_FORMAT,
			'0' AS p_err_code,
			'0' AS p_err_msg
			into :p_nls_calendar, :p_nls_comp, :p_nls_length_semantics, :p_nls_nchar_conv_excep, 
				:p_nls_date_lang, :p_nls_sort, :p_nls_currency, :p_nls_date_format, :p_nls_iso_currency,
				:p_nls_numeric_chars, :p_nls_dual_currency, :p_nls_timestamp, :p_nls_timestamp_tz,
				:p_err_code, :p_err_msg
		FROM
			nls_session_parameters
		;
	END;`
	stmt := NewStmt(cmdText, conn)
	stmt.AddParam("p_nls_calendar", "", 40, Output)
	stmt.AddParam("p_nls_comp", "", 40, Output)
	stmt.AddParam("p_nls_length_semantics", "", 40, Output)
	stmt.AddParam("p_nls_nchar_conv_excep", "", 40, Output)
	stmt.AddParam("p_nls_date_lang", "", 40, Output)
	stmt.AddParam("p_nls_sort", "", 40, Output)
	stmt.AddParam("p_nls_currency", "", 40, Output)
	stmt.AddParam("p_nls_date_format", "", 40, Output)
	stmt.AddParam("p_nls_iso_currency", "", 40, Output)
	stmt.AddParam("p_nls_numeric_chars", "", 40, Output)
	stmt.AddParam("p_nls_dual_currency", "", 40, Output)
	stmt.AddParam("p_nls_timestamp", "", 48, Output)
	stmt.AddParam("p_nls_timestamp_tz", "", 56, Output)
	stmt.AddParam("p_err_code", "", 2000, Output)
	stmt.AddParam("p_err_msg", "", 2000, Output)
	defer stmt.Close()
	//fmt.Println(stmt.Pars)
	_, err := stmt.Exec(nil)
	if err != nil {
		return nil, err
	}

	if len(stmt.Pars) >= 10 {
		conn.NLSData.Calender = conn.strConv.Decode(stmt.Pars[0].BValue)
		conn.NLSData.Comp = conn.strConv.Decode(stmt.Pars[1].BValue)
		conn.NLSData.LengthSemantics = conn.strConv.Decode(stmt.Pars[2].BValue)
		conn.NLSData.NCharConvExcep = conn.strConv.Decode(stmt.Pars[3].BValue)
		conn.NLSData.DateLang = conn.strConv.Decode(stmt.Pars[4].BValue)
		conn.NLSData.Sort = conn.strConv.Decode(stmt.Pars[5].BValue)
		conn.NLSData.Currency = conn.strConv.Decode(stmt.Pars[6].BValue)
		conn.NLSData.DateFormat = conn.strConv.Decode(stmt.Pars[7].BValue)
		conn.NLSData.IsoCurrency = conn.strConv.Decode(stmt.Pars[8].BValue)
		conn.NLSData.NumericChars = conn.strConv.Decode(stmt.Pars[9].BValue)
		conn.NLSData.DualCurrency = conn.strConv.Decode(stmt.Pars[10].BValue)
		conn.NLSData.Timestamp = conn.strConv.Decode(stmt.Pars[11].BValue)
		conn.NLSData.TimestampTZ = conn.strConv.Decode(stmt.Pars[12].BValue)
	}

	/*
		for _, par := range stmt.Pars {
			if par.Name == "p_nls_calendar" {
				conn.NLSData.Calender = conn.strConv.Decode(par.BValue)
			} else if par.Name == "p_nls_comp" {
				conn.NLSData.Comp = conn.strConv.Decode(par.BValue)
			} else if par.Name == "p_nls_length_semantics" {
				conn.NLSData.LengthSemantics = conn.strConv.Decode(par.BValue)
			} else if par.Name == "p_nls_nchar_conv_excep" {
				conn.NLSData.NCharConvExcep = conn.strConv.Decode(par.BValue)
			} else if par.Name == "p_nls_date_lang" {
				conn.NLSData.DateLang = conn.strConv.Decode(par.BValue)
			} else if par.Name == "p_nls_sort" {
				conn.NLSData.Sort = conn.strConv.Decode(par.BValue)
			} else if par.Name == "p_nls_currency" {
				conn.NLSData.Currency = conn.strConv.Decode(par.BValue)
			} else if par.Name == "p_nls_date_format" {
				conn.NLSData.DateFormat = conn.strConv.Decode(par.BValue)
			} else if par.Name == "p_nls_iso_currency" {
				conn.NLSData.IsoCurrency = conn.strConv.Decode(par.BValue)
			} else if par.Name == "p_nls_numeric_chars" {
				conn.NLSData.NumericChars = conn.strConv.Decode(par.BValue)
			} else if par.Name == "p_nls_dual_currency" {
				conn.NLSData.DualCurrency = conn.strConv.Decode(par.BValue)
			} else if par.Name == "p_nls_timestamp" {
				conn.NLSData.Timestamp = conn.strConv.Decode(par.BValue)
			} else if par.Name == "p_nls_timestamp_tz" {
				conn.NLSData.TimestampTZ = conn.strConv.Decode(par.BValue)
			}
		}
	*/

	return &conn.NLSData, nil
}

func (conn *Connection) Prepare(query string) (driver.Stmt, error) {
	conn.connOption.Tracer.Print("Prepare\n", query)
	return NewStmt(query, conn), nil
}

func (conn *Connection) Ping(_ context.Context) error {
	conn.connOption.Tracer.Print("Ping")
	conn.session.ResetBuffer()
	return (&simpleObject{
		session:     conn.session,
		operationID: 0x93,
		data:        nil,
	}).write().read()
}

//func (conn *Connection) Logoff() error {
//	conn.connOption.Tracer.Print("Logoff")
//	session := conn.session
//	session.ResetBuffer()
//	session.PutBytes(0x11, 0x87, 0, 0, 0, 0x2, 0x1, 0x11, 0x1, 0, 0, 0, 0x1, 0, 0, 0, 0, 0, 0x1, 0, 0, 0, 0, 0,
//		3, 9, 0)
//	err := session.Write()
//	if err != nil {
//		return err
//	}
//	loop := true
//	for loop {
//		msg, err := session.GetByte()
//		if err != nil {
//			return err
//		}
//		switch msg {
//		case 4:
//			session.Summary, err = network.NewSummary(session)
//			if err != nil {
//				return err
//			}
//			loop = false
//		case 9:
//			if session.HasEOSCapability {
//				if session.Summary == nil {
//					session.Summary = new(network.SummaryObject)
//				}
//				session.Summary.EndOfCallStatus, err = session.GetInt(4, true, true)
//				if err != nil {
//					return err
//				}
//			}
//			if session.HasFSAPCapability {
//				if session.Summary == nil {
//					session.Summary = new(network.SummaryObject)
//				}
//				session.Summary.EndToEndECIDSequence, err = session.GetInt(2, true, true)
//				if err != nil {
//					return err
//				}
//			}
//			loop = false
//		default:
//			return errors.New(fmt.Sprintf("message code error: received code %d and expected code is 4, 9", msg))
//		}
//	}
//	if session.HasError() {
//		return errors.New(session.GetError())
//	}
//	return nil
//}

func (conn *Connection) Open() error {
	tracer := conn.connOption.Tracer
	tracer.Print("Open :", conn.connOption.ConnectionData())

	switch conn.conStr.DBAPrivilege {
	case SYSDBA:
		conn.LogonMode |= SysDba
	case SYSOPER:
		conn.LogonMode |= SysOper
	default:
		conn.LogonMode = 0
	}
	conn.session = network.NewSession(*conn.connOption)
	err := conn.session.Connect()
	if err != nil {
		return err
	}

	tracer.Print("TCP Negotiation")
	conn.tcpNego, err = NewTCPNego(conn.session)
	if err != nil {
		return err
	}
	tracer.Print("Server Charset: ", conn.tcpNego.ServerCharset)
	tracer.Print("Server National Charset: ", conn.tcpNego.ServernCharset)
	// create string converter object
	conn.strConv = converters.NewStringConverter(conn.tcpNego.ServerCharset)
	conn.session.StrConv = conn.strConv
	tracer.Print("Data Type Negotiation")
	conn.dataNego, err = buildTypeNego(conn.tcpNego, conn.session)
	if err != nil {
		return err
	}

	conn.session.TTCVersion = conn.dataNego.CompileTimeCaps[7]

	if conn.tcpNego.ServerCompileTimeCaps[7] < conn.session.TTCVersion {
		conn.session.TTCVersion = conn.tcpNego.ServerCompileTimeCaps[7]
	}
	tracer.Print("TTC Version: ", conn.session.TTCVersion)
	err = conn.doAuth()
	if err != nil {
		return err
	}
	conn.State = Opened
	conn.dBVersion, err = GetDBVersion(conn.session)
	if err != nil {
		return err
	}
	tracer.Print("Connected")
	tracer.Print("Database Version: ", conn.dBVersion.Text)
	if len(conn.SessionProperties) == 0 {
		//return errors.New(fmt.Sprint("Session properties is null"))
	} else {
		sessionID, err := strconv.ParseUint(conn.SessionProperties["AUTH_SESSION_ID"], 10, 32)
		if err != nil {
			return err
		}
		conn.sessionID = int(sessionID)
		serialNum, err := strconv.ParseUint(conn.SessionProperties["AUTH_SERIAL_NUM"], 10, 32)
		if err != nil {
			return err
		}
		conn.serialID = int(serialNum)
		conn.connOption.InstanceName = conn.SessionProperties["AUTH_SC_INSTANCE_NAME"]
		conn.connOption.Host = conn.SessionProperties["AUTH_SC_SERVER_HOST"]
		conn.connOption.ServiceName = conn.SessionProperties["AUTH_SC_SERVICE_NAME"]
		conn.connOption.DomainName = conn.SessionProperties["AUTH_SC_DB_DOMAIN"]
		conn.connOption.DBName = conn.SessionProperties["AUTH_SC_DBUNIQUE_NAME"]
	}

	_, err = conn.GetNLS()
	if err != nil {
		return err
	}

	return nil
}

func (conn *Connection) Begin() (driver.Tx, error) {
	conn.connOption.Tracer.Print("Begin transaction")
	conn.autoCommit = false
	return &Transaction{conn: conn}, nil
}

func NewConnection(databaseUrl string) (*Connection, error) {
	//this.m_id = this.GetHashCode().ToString();
	conStr, err := newConnectionStringFromUrl(databaseUrl)
	if err != nil {
		return nil, err
	}
	userName := ""
	User, err := user.Current()
	if err == nil {
		userName = User.Username
	}
	hostName, _ := os.Hostname()
	indexOfSlash := strings.LastIndex(os.Args[0], "/")
	indexOfSlash += 1
	if indexOfSlash < 0 {
		indexOfSlash = 0
	}

	connOption := &network.ConnectionOption{
		Port:                  conStr.Port,
		TransportConnectTo:    0xFFFF,
		SSLVersion:            "",
		WalletDict:            "",
		TransportDataUnitSize: 0xFFFF,
		SessionDataUnitSize:   0xFFFF,
		Protocol:              "tcp",
		Host:                  conStr.Host,
		UserID:                conStr.UserID,
		//IP:                    "",
		SID: conStr.SID,
		//Addr:                  "",
		//Server:                conn.conStr.Host,
		ServiceName:  conStr.ServiceName,
		InstanceName: conStr.InstanceName,
		PrefetchRows: conStr.PrefetchRows,
		ClientData: network.ClientData{
			ProgramPath: os.Args[0],
			ProgramName: os.Args[0][indexOfSlash:],
			UserName:    userName,
			HostName:    hostName,
			DriverName:  "OracleClientGo",
			PID:         os.Getpid(),
		},
		//InAddrAny:             false,
	}

	if len(conStr.Trace) > 0 {
		tf, err := os.Create(conStr.Trace)
		if err != nil {
			//noinspection GoErrorStringFormat
			return nil, fmt.Errorf("Can't open trace file: %w", err)
		}
		connOption.Tracer = trace.NewTraceWriter(tf)
	} else {
		connOption.Tracer = trace.NilTracer()
	}
	return &Connection{
		State:      Closed,
		conStr:     conStr,
		connOption: connOption,
		autoCommit: true,
	}, nil
}

func (conn *Connection) Close() (err error) {
	conn.connOption.Tracer.Print("Close")
	//var err error = nil
	if conn.session != nil {
		//err = conn.Logoff()
		conn.session.Disconnect()
		conn.session = nil
	}
	conn.connOption.Tracer.Print("Connection Closed")
	conn.connOption.Tracer.Close()
	return
}

func (conn *Connection) doAuth() error {
	conn.connOption.Tracer.Print("doAuth")
	conn.session.ResetBuffer()
	conn.session.PutBytes(3, 118, 0, 1)
	conn.session.PutUint(len(conn.conStr.UserID), 4, true, true)
	conn.LogonMode = conn.LogonMode | NoNewPass
	conn.session.PutUint(int(conn.LogonMode), 4, true, true)
	conn.session.PutBytes(1, 1, 5, 1, 1)
	conn.session.PutBytes([]byte(conn.conStr.UserID)...)
	conn.session.PutKeyValString("AUTH_TERMINAL", conn.connOption.ClientData.HostName, 0)
	conn.session.PutKeyValString("AUTH_PROGRAM_NM", conn.connOption.ClientData.ProgramName, 0)
	conn.session.PutKeyValString("AUTH_MACHINE", conn.connOption.ClientData.HostName, 0)
	conn.session.PutKeyValString("AUTH_PID", fmt.Sprintf("%d", conn.connOption.ClientData.PID), 0)
	conn.session.PutKeyValString("AUTH_SID", conn.connOption.ClientData.UserName, 0)
	err := conn.session.Write()
	if err != nil {
		return err
	}

	conn.authObject, err = NewAuthObject(conn.conStr.UserID, conn.conStr.Password, conn.tcpNego, conn.session)
	if err != nil {
		return err
	}
	// if proxyAuth ==> mode |= PROXY
	err = conn.authObject.Write(conn.connOption, conn.LogonMode, conn.session)
	if err != nil {
		return err
	}
	stop := false
	for !stop {
		msg, err := conn.session.GetInt(1, false, false)
		if err != nil {
			return err
		}
		switch msg {
		case 4:
			conn.session.Summary, err = network.NewSummary(conn.session)
			if err != nil {
				return err
			}
			if conn.session.HasError() {
				return errors.New(conn.session.GetError())
			}
			stop = true
		case 8:
			dictLen, err := conn.session.GetInt(4, true, true)
			if err != nil {
				return err
			}
			conn.SessionProperties = make(map[string]string, dictLen)
			for x := 0; x < dictLen; x++ {
				key, val, _, err := conn.session.GetKeyVal()
				if err != nil {
					return err
				}
				conn.SessionProperties[string(key)] = string(val)
			}
		case 15:
			warning, err := network.NewWarningObject(conn.session)
			if err != nil {
				return err
			}
			if warning != nil {
				fmt.Println(warning)
			}
			stop = true
		default:
			return errors.New(fmt.Sprintf("message code error: received code %d and expected code is 8", msg))
		}
	}

	// if verifyResponse == true
	// conn.authObject.VerifyResponse(conn.SessionProperties["AUTH_SVR_RESPONSE"])
	return nil
}
