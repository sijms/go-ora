package go_ora

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"go-ora/converters"
	"go-ora/network"
	"os"
	"os/user"
	"strconv"
	"strings"
)

type ConnectionState int

const (
	Closed ConnectionState = 0
	Opened ConnectionState = 1
)

type LogonMode int

const (
	NoNewPass   LogonMode = 0x1
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
	strConv           *converters.StringConverter
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

func (conn *Connection) GetNLS() (*NLSData, error) {
	cmdText := `DECLARE
	err_code VARCHAR2(2000);
	err_msg  VARCHAR2(2000);
	BEGIN
		SELECT VALUE into :p_nls_calendar from nls_session_parameters where PARAMETER='NLS_CALENDAR';
		SELECT VALUE into :p_nls_comp from nls_session_parameters where PARAMETER='NLS_COMP';
		SELECT VALUE into :p_nls_length_semantics from nls_session_parameters where PARAMETER='NLS_LENGTH_SEMANTICS';
		SELECT VALUE into :p_nls_nchar_conv_excep from nls_session_parameters where PARAMETER='NLS_NCHAR_CONV_EXCP';
		SELECT VALUE into :p_nls_date_lang from nls_session_parameters where PARAMETER='NLS_DATE_LANGUAGE';
		SELECT VALUE into :p_nls_sort from nls_session_parameters where PARAMETER='NLS_SORT';
		SELECT VALUE into :p_nls_currency from nls_session_parameters where PARAMETER='NLS_CURRENCY';
		SELECT VALUE into :p_nls_date_format from nls_session_parameters where PARAMETER='NLS_DATE_FORMAT';
		SELECT VALUE into :p_nls_iso_currency from nls_session_parameters where PARAMETER='NLS_ISO_CURRENCY';
		SELECT VALUE into :p_nls_numeric_chars from nls_session_parameters where PARAMETER='NLS_NUMERIC_CHARACTERS';
		SELECT VALUE into :p_nls_dual_currency from nls_session_parameters where PARAMETER='NLS_DUAL_CURRENCY';
		SELECT VALUE into :p_nls_timestamp from nls_session_parameters where PARAMETER='NLS_TIMESTAMP_FORMAT';
		SELECT VALUE into :p_nls_timestamp_tz from nls_session_parameters where PARAMETER='NLS_TIMESTAMP_TZ_FORMAT';
		SELECT '0' into :p_err_code from dual;
		SELECT '0' into :p_err_msg from dual;
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
	stmt.AddParam("p_nls_timestamp", "", 40, Output)
	stmt.AddParam("p_nls_timestamp_tz", "", 40, Output)
	stmt.AddParam("p_err_code", "", 2000, Output)
	stmt.AddParam("p_err_msg", "", 2000, Output)
	//fmt.Println(stmt.Pars)
	_, err := stmt.Exec(nil)
	if err != nil {
		return nil, err
	}
	for _, par := range stmt.Pars {
		if par.Name == "p_nls_calendar" {
			conn.NLSData.Calender = conn.strConv.Decode(par.Value)
		} else if par.Name == "p_nls_comp" {
			conn.NLSData.Comp = conn.strConv.Decode(par.Value)
		} else if par.Name == "p_nls_length_semantics" {
			conn.NLSData.LengthSemantics = conn.strConv.Decode(par.Value)
		} else if par.Name == "p_nls_nchar_conv_excep" {
			conn.NLSData.NCharConvExcep = conn.strConv.Decode(par.Value)
		} else if par.Name == "p_nls_date_lang" {
			conn.NLSData.DateLang = conn.strConv.Decode(par.Value)
		} else if par.Name == "p_nls_sort" {
			conn.NLSData.Sort = conn.strConv.Decode(par.Value)
		} else if par.Name == "p_nls_currency" {
			conn.NLSData.Currency = conn.strConv.Decode(par.Value)
		} else if par.Name == "p_nls_date_format" {
			conn.NLSData.DateFormat = conn.strConv.Decode(par.Value)
		} else if par.Name == "p_nls_iso_currency" {
			conn.NLSData.IsoCurrency = conn.strConv.Decode(par.Value)
		} else if par.Name == "p_nls_numeric_chars" {
			conn.NLSData.NumericChars = conn.strConv.Decode(par.Value)
		} else if par.Name == "p_nls_dual_currency" {
			conn.NLSData.DualCurrency = conn.strConv.Decode(par.Value)
		} else if par.Name == "p_nls_timestamp" {
			conn.NLSData.Timestamp = conn.strConv.Decode(par.Value)
		} else if par.Name == "p_nls_timestamp_tz" {
			conn.NLSData.TimestampTZ = conn.strConv.Decode(par.Value)
		}
	}
	return &conn.NLSData, nil
}

func (conn *Connection) Prepare(query string) (driver.Stmt, error) {
	return NewStmt(query, conn), nil
}

func (conn *Connection) Ping(ctx context.Context) error {
	conn.session.ResetBuffer()
	return (&simpleObject{
		session:     conn.session,
		operationID: 0x93,
		data:        nil,
	}).write().read()
}

func (conn *Connection) Logoff() error {
	session := conn.session
	session.ResetBuffer()
	session.PutBytes([]byte{0x11, 0x87, 0, 0, 0, 0x2, 0x1, 0x11,
		0x1, 0, 0, 0, 0x1, 0, 0, 0,
		0, 0, 0x1, 0, 0, 0, 0, 0,
		3, 9, 0})
	err := session.Write()
	if err != nil {
		return err
	}
	loop := true
	for loop {
		msg, err := session.GetByte()
		if err != nil {
			return err
		}
		switch msg {
		case 4:
			session.Summary, err = network.NewSummary(session)
			if err != nil {
				return err
			}
			loop = false
		case 9:
			if session.HasEOSCapability {
				if session.Summary == nil {
					session.Summary = new(network.SummaryObject)
				}
				session.Summary.EndOfCallStatus, err = session.GetInt(4, true, true)
				if err != nil {
					return err
				}
			}
			if session.HasFSAPCapability {
				if session.Summary == nil {
					session.Summary = new(network.SummaryObject)
				}
				session.Summary.EndToEndECIDSequence, err = session.GetInt(2, true, true)
				if err != nil {
					return err
				}
			}
			loop = false
		default:
			return errors.New(fmt.Sprintf("message code error: received code %d and expected code is 4, 9", msg))
		}
	}
	if session.HasError() {
		return errors.New(session.GetError())
	}
	return nil
}

func (conn *Connection) Open() error {
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

	conn.tcpNego, err = NewTCPNego(conn.session)
	if err != nil {
		return err
	}
	// create string converter object
	conn.strConv = converters.NewStringConverter(conn.tcpNego.ServerCharset)
	conn.session.StrConv = conn.strConv

	conn.dataNego, err = buildTypeNego(conn.tcpNego, conn.session)
	if err != nil {
		return err
	}

	conn.session.TTCVersion = conn.dataNego.CompileTimeCaps[7]

	if conn.tcpNego.ServerCompileTimeCaps[7] < conn.session.TTCVersion {
		conn.session.TTCVersion = conn.tcpNego.ServerCompileTimeCaps[7]
	}

	//if (((int) this.m_serverCompiletimeCapabilities[15] & 1) != 0)
	//	this.m_marshallingEngine.HasEOCSCapability = true;
	//if (((int) this.m_serverCompiletimeCapabilities[16] & 16) != 0)
	//	this.m_marshallingEngine.HasFSAPCapability = true;

	err = conn.doAuth()
	if err != nil {
		return err
	}
	conn.State = Opened
	conn.dBVersion, err = GetDBVersion(conn.session)
	if err != nil {
		return err
	}

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
	_, err = conn.GetNLS()
	if err != nil {
		return err
	}
	return nil
}

func (conn *Connection) Begin() (driver.Tx, error) {
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
		userName = User.Name
	}
	hostName, _ := os.Hostname()
	indexOfSlash := strings.LastIndex(os.Args[0], "/")
	indexOfSlash += 1
	if indexOfSlash < 0 {
		indexOfSlash = 0
	}
	//userName = "samy"
	//hostName = "LABMANAGER"
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
	return &Connection{
		State:      Closed,
		conStr:     conStr,
		connOption: connOption,
	}, nil
}

func (conn *Connection) Close() (err error) {
	//var err error = nil
	if conn.session != nil {
		//err = conn.Logoff()
		conn.session.Disconnect()
		conn.session = nil
	}
	return
}

func (conn *Connection) doAuth() error {
	conn.session.ResetBuffer()
	conn.session.PutBytes([]byte{3, 118, 0, 1})
	conn.session.PutUint(len(conn.conStr.UserID), 4, true, true)
	conn.LogonMode = conn.LogonMode | NoNewPass
	conn.session.PutUint(int(conn.LogonMode), 4, true, true)
	conn.session.PutBytes([]byte{1, 1, 5, 1, 1})
	conn.session.PutBytes([]byte(conn.conStr.UserID))
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
