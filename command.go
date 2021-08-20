package go_ora

import (
	"database/sql/driver"
	"encoding/binary"
	"errors"
	"fmt"
	"time"

	"github.com/sijms/go-ora/converters"
	"github.com/sijms/go-ora/network"
	"regexp"
	"strings"
)

type StmtType int

const (
	SELECT StmtType = 1
	DML    StmtType = 2
	PLSQL  StmtType = 3
	OTHERS StmtType = 4
)

type StmtInterface interface {
	hasMoreRows() bool
	noOfRowsToFetch() int
	fetch(dataSet *DataSet) error
	hasBLOB() bool
	hasLONG() bool
	//write() error
	//getExeOption() int
	read(dataSet *DataSet) error
	//Close() error
	//Exec(args []driver.Value) (driver.Result, error)
	//Query(args []driver.Value) (driver.Rows, error)
}
type defaultStmt struct {
	connection         *Connection
	text               string
	arrayBindingCount  int
	disableCompression bool
	_hasLONG           bool
	_hasBLOB           bool
	_hasMoreRows       bool
	_hasReturnClause   bool
	_noOfRowsToFetch   int
	stmtType           StmtType
	cursorID           int
	queryID            uint64
	Pars               []ParameterInfo
	columns            []ParameterInfo
	scnFromExe         []int
	arrayBindCount     int
}

func (stmt *defaultStmt) hasMoreRows() bool {
	return stmt._hasMoreRows
}
func (stmt *defaultStmt) noOfRowsToFetch() int {
	return stmt._noOfRowsToFetch
}
func (stmt *defaultStmt) hasLONG() bool {
	return stmt._hasLONG
}
func (stmt *defaultStmt) hasBLOB() bool {
	return stmt._hasBLOB
}

func (stmt *defaultStmt) basicWrite(exeOp int, parse, define bool) error {
	session := stmt.connection.session
	session.PutBytes(3, 0x5E, 0)
	session.PutUint(exeOp, 4, true, true)
	session.PutUint(stmt.cursorID, 2, true, true)
	if stmt.cursorID == 0 {
		session.PutBytes(1)

	} else {
		session.PutBytes(0)
	}
	if parse {
		session.PutUint(len(stmt.connection.strConv.Encode(stmt.text)), 4, true, true)
		session.PutBytes(1)
	} else {
		session.PutBytes(0, 1)
	}
	session.PutUint(13, 2, true, true)
	session.PutBytes(0, 0)
	if exeOp&0x40 == 0 && exeOp&0x20 != 0 && exeOp&0x1 != 0 && stmt.stmtType == SELECT {
		session.PutBytes(0)
		session.PutUint(stmt._noOfRowsToFetch, 4, true, true)
	} else {
		session.PutUint(0, 4, true, true)
		session.PutUint(0, 4, true, true)
	}
	// add fetch size = max(int32)
	session.PutUint(0x7FFFFFFF, 4, true, true)
	if len(stmt.Pars) > 0 {
		session.PutBytes(1)
		session.PutUint(len(stmt.Pars), 2, true, true)
	} else {
		session.PutBytes(0, 0)

	}
	session.PutBytes(0, 0, 0, 0, 0)
	if define {
		session.PutBytes(1)
		session.PutUint(len(stmt.columns), 2, true, true)
	} else {
		session.PutBytes(0, 0)
	}
	if session.TTCVersion >= 4 {
		session.PutBytes(0, 0, 1)
	}
	if session.TTCVersion >= 5 {
		session.PutBytes(0, 0, 0, 0, 0)
	}
	if parse {
		session.PutBytes(stmt.connection.strConv.Encode(stmt.text)...)
	}
	if define {
		session.PutBytes(0)
		for x := 0; x < len(stmt.columns); x++ {
			stmt.columns[x].Flag = 3
			stmt.columns[x].CharsetForm = 1
			//stmt.columns[x].MaxLen = 0x7fffffff
			err := stmt.columns[x].write(session)
			if err != nil {
				return err
			}
			session.PutBytes(0)
		}
	} else {
		al8i4 := make([]int, 13)
		if exeOp&1 <= 0 {
			al8i4[0] = 0
		} else {
			al8i4[0] = 1
		}
		switch stmt.stmtType {
		case DML:
			fallthrough
		case PLSQL:
			if stmt.arrayBindCount <= 1 {
				al8i4[1] = 1
			} else {
				al8i4[1] = stmt.arrayBindCount
			}
		case OTHERS:
			al8i4[1] = 1
		default:
			al8i4[1] = stmt._noOfRowsToFetch
		}
		if len(stmt.scnFromExe) == 2 {
			al8i4[5] = stmt.scnFromExe[0]
			al8i4[6] = stmt.scnFromExe[1]
		} else {
			al8i4[5] = 0
			al8i4[6] = 0
		}
		if stmt.stmtType == SELECT {
			al8i4[7] = 1
		} else {
			al8i4[7] = 0
		}
		for x := 0; x < len(al8i4); x++ {
			session.PutUint(al8i4[x], 4, true, true)
		}
	}
	for _, par := range stmt.Pars {
		_ = par.write(session)
	}
	return nil
}

type Stmt struct {
	defaultStmt
	//reExec           bool
	reSendParDef bool
	parse        bool // means parse the command in the server this occur if the stmt is not cached
	execute      bool
	define       bool

	//noOfDefCols        int
}

type QueryResult struct {
	lastInsertedID int64
	rowsAffected   int64
}

func (rs *QueryResult) LastInsertId() (int64, error) {
	return rs.lastInsertedID, nil
}

func (rs *QueryResult) RowsAffected() (int64, error) {
	return rs.rowsAffected, nil
}

func NewStmt(text string, conn *Connection) *Stmt {
	ret := &Stmt{
		//connection:         conn,
		//text:               text,
		//reExec:             false,
		reSendParDef: false,
		parse:        true,
		execute:      true,
		define:       false,
		//hasLONG:            false,
		//hasBLOB:            false,
		//disableCompression: true,
		//arrayBindCount:     1,
		//parse:              true,
		//execute:            true,
		//define:             false,
		//scnFromExe:         make([]int, 2),
	}
	ret.connection = conn
	ret.text = text
	ret._hasBLOB = false
	ret._hasLONG = false
	ret.disableCompression = true
	ret.arrayBindCount = 1
	ret.scnFromExe = make([]int, 2)
	// get stmt type
	uCmdText := strings.TrimSpace(strings.ToUpper(text))
	if strings.HasPrefix(uCmdText, "SELECT") || strings.HasPrefix(uCmdText, "WITH") {
		ret.stmtType = SELECT
	} else if strings.HasPrefix(uCmdText, "UPDATE") ||
		strings.HasPrefix(uCmdText, "INSERT") ||
		strings.HasPrefix(uCmdText, "DELETE") {
		ret.stmtType = DML
	} else if strings.HasPrefix(uCmdText, "DECLARE") || strings.HasPrefix(uCmdText, "BEGIN") {
		ret.stmtType = PLSQL
	} else {
		ret.stmtType = OTHERS
	}

	// returning clause
	var err error
	ret._hasReturnClause, err = regexp.MatchString(`\bRETURNING\b`, uCmdText)
	if err != nil {
		ret._hasReturnClause = false
	}
	return ret
}

func (stmt *Stmt) write(session *network.Session) error {
	if !stmt.parse && stmt.stmtType == DML && !stmt.reSendParDef {
		exeOf := 0
		execFlag := 0
		count := stmt.arrayBindCount
		if stmt.stmtType == SELECT {
			session.PutBytes(3, 0x4E, 0)
			count = stmt._noOfRowsToFetch
			exeOf = 0x20
			if stmt._hasReturnClause || stmt.stmtType == PLSQL || stmt.disableCompression {
				exeOf |= 0x40000
			}

		} else {
			session.PutBytes(3, 4, 0)
		}
		if stmt.connection.autoCommit {
			execFlag = 1
		}
		session.PutUint(stmt.cursorID, 2, true, true)
		session.PutUint(count, 2, true, true)
		session.PutUint(exeOf, 2, true, true)
		session.PutUint(execFlag, 2, true, true)
	} else {
		//stmt.reExec = true
		err := stmt.basicWrite(stmt.getExeOption(), stmt.parse, stmt.define)
		if err != nil {
			return err
		}
		stmt.parse = false
		stmt.define = false
		stmt.reSendParDef = false
	}
	//if !stmt.reExec {
	//exeOp := stmt.getExeOption()
	//session.PutBytes(3, 0x5E, 0)
	//session.PutUint(exeOp, 4, true, true)
	//session.PutUint(stmt.cursorID, 2, true, true)
	//if stmt.cursorID == 0 {
	//	session.PutBytes(1)
	//	//session.PutUint(1, 1, false, false)
	//} else {
	//	session.PutBytes(0)
	//	//session.PutUint(0, 1, false, false)
	//}
	//session.PutUint(len(stmt.text), 4, true, true)
	//session.PutBytes(1)
	//session.PutUint(1, 1, false, false)
	//session.PutUint(13, 2, true, true)
	//session.PutBytes(0, 0)
	//if exeOp&0x40 == 0 && exeOp&0x20 != 0 && exeOp&0x1 != 0 && stmt.stmtType == SELECT {
	//	session.PutBytes(0)
	//	//session.PutUint(0, 1, false, false)
	//	session.PutUint(stmt.noOfRowsToFetch, 4, true, true)
	//} else {
	//	session.PutUint(0, 4, true, true)
	//	session.PutUint(0, 4, true, true)
	//}
	// longFetchSize == 0 marshal 1 else marshal longFetchSize
	//session.PutUint(1, 4, true, true)
	//if len(stmt.Pars) > 0 {
	//	session.PutBytes(1)
	//	//session.PutUint(1, 1, false, false)
	//	session.PutUint(len(stmt.Pars), 2, true, true)
	//} else {
	//	session.PutBytes(0, 0)
	//	//session.PutUint(0, 1, false, false)
	//	//session.PutUint(0, 1, false, false)
	//}
	//session.PutBytes(0, 0, 0, 0, 0)
	//if stmt.define {
	//	session.PutBytes(1)
	//	//session.PutUint(1, 1, false, false)
	//	session.PutUint(stmt.noOfDefCols, 2, true, true)
	//} else {
	//	session.PutBytes(0, 0)
	//	//session.PutUint(0, 1, false, false)
	//	//session.PutUint(0, 1, false, false)
	//}
	//if session.TTCVersion >= 4 {
	//	session.PutBytes(0, 0, 1)
	//	//session.PutUint(0, 1, false, false) // dbChangeRegisterationId
	//	//session.PutUint(0, 1, false, false)
	//	//session.PutUint(1, 1, false, false)
	//}
	//if session.TTCVersion >= 5 {
	//	session.PutBytes(0, 0, 0, 0, 0)
	//	//session.PutUint(0, 1, false, false)
	//	//session.PutUint(0, 1, false, false)
	//	//session.PutUint(0, 1, false, false)
	//	//session.PutUint(0, 1, false, false)
	//	//session.PutUint(0, 1, false, false)
	//}

	//session.PutBytes([]byte(stmt.text)...)
	//for x := 0; x < len(stmt.al8i4); x++ {
	//	session.PutUint(stmt.al8i4[x], 2, true, true)
	//}
	//for _, par := range stmt.Pars {
	//	_ = par.write(session)
	//}
	//stmt.reExec = true
	//stmt.parse = false
	//} else {
	//
	//}

	if len(stmt.Pars) > 0 {
		for x := 0; x < stmt.arrayBindCount; x++ {
			session.PutBytes(7)
			for _, par := range stmt.Pars {
				if par.DataType != RAW {
					if par.DataType == REFCURSOR {
						session.PutBytes(1, 0)
					} else {
						session.PutClr(par.BValue)
					}
				}
			}
			for _, par := range stmt.Pars {
				if par.DataType == RAW {
					session.PutClr(par.BValue)
				}
			}
		}
		//session.PutUint(7, 1, false, false)
		//for _, par := range stmt.Pars {
		//	session.PutClr(par.BValue)
		//}
	}
	return session.Write()
}

func (stmt *Stmt) getExeOption() int {
	op := 0
	if stmt.stmtType == PLSQL || stmt._hasReturnClause {
		op |= 0x40000
	}
	if stmt.arrayBindCount > 1 {
		op |= 0x80000
	}
	if stmt.connection.autoCommit && stmt.stmtType == DML {
		op |= 0x100
	}
	if stmt.parse {
		op |= 1
	}
	if stmt.execute {
		op |= 0x20
	}
	if !stmt.parse && !stmt.execute {
		op |= 0x40
	}
	if len(stmt.Pars) > 0 {
		op |= 0x8
		if stmt.stmtType == PLSQL || stmt._hasReturnClause {
			op |= 0x400
		}
	}
	if stmt.stmtType != PLSQL && !stmt._hasReturnClause {
		op |= 0x8000
	}
	if stmt.define {
		op |= 0x10
	}
	return op

	/* HasReturnClause
	if  stmt.PLSQL or cmdText == "" return false
	Regex.IsMatch(cmdText, "\\bRETURNING\\b"
	*/
}

func (stmt *defaultStmt) fetch(dataSet *DataSet) error {
	stmt.connection.session.ResetBuffer()
	stmt.connection.session.PutBytes(3, 5, 0)
	stmt.connection.session.PutInt(stmt.cursorID, 2, true, true)
	stmt.connection.session.PutInt(stmt._noOfRowsToFetch, 2, true, true)
	err := stmt.connection.session.Write()
	if err != nil {
		return err
	}
	return stmt.read(dataSet)
}
func (stmt *defaultStmt) read(dataSet *DataSet) error {
	loop := true
	after7 := false
	containOutputPars := false
	dataSet.parent = stmt
	session := stmt.connection.session
	for loop {
		msg, err := session.GetByte()
		if err != nil {
			return err
		}
		switch msg {
		case 4:
			stmt.connection.session.Summary, err = network.NewSummary(session)
			if err != nil {
				return err
			}
			stmt.connection.connOption.Tracer.Printf("Summary: RetCode:%d, Error Message:%q", stmt.connection.session.Summary.RetCode, string(stmt.connection.session.Summary.ErrorMessage))

			stmt.cursorID = stmt.connection.session.Summary.CursorID
			stmt.disableCompression = stmt.connection.session.Summary.Flags&0x20 != 0
			if stmt.connection.session.HasError() {
				if stmt.connection.session.Summary.RetCode == 1403 {
					stmt._hasMoreRows = false
					stmt.connection.session.Summary = nil
				} else {
					return errors.New(stmt.connection.session.GetError())
				}

			}
			loop = false
		case 6:
			//_, err = session.GetByte()
			err = dataSet.load(session)
			if err != nil {
				return err
			}
			if !after7 {
				if stmt.stmtType == SELECT {

				}
			}
		case 7:
			after7 = true
			if stmt._hasReturnClause && containOutputPars {
				for x := 0; x < len(stmt.Pars); x++ {
					if stmt.Pars[x].Direction == Output {
						num, err := session.GetInt(4, true, true)
						if err != nil {
							return err
						}
						if num > 1 {
							return errors.New("more than one row affected with return clause")
						}
						if num == 0 {
							stmt.Pars[x].BValue = nil
							stmt.Pars[x].Value = nil
						} else {
							stmt.Pars[x].BValue, err = session.GetClr()
							if err != nil {
								return err
							}
							err = stmt.calculateParameterValue(&stmt.Pars[x])
							if err != nil {
								return err
							}
							_, err = session.GetInt(2, true, true)
							if err != nil {
								return err
							}
						}
					}
				}
			} else {
				if containOutputPars {
					for x := 0; x < len(stmt.Pars); x++ {
						if stmt.Pars[x].DataType == REFCURSOR {
							cursor := RefCursor{}
							cursor.connection = stmt.connection
							cursor.parent = stmt
							err = cursor.load(session)
							if err != nil {
								return err
							}
							stmt.Pars[x].Value = cursor

						} else {
							if stmt.Pars[x].Direction != Input {
								stmt.Pars[x].BValue, err = session.GetClr()
								if err != nil {
									return err
								}
								err = stmt.calculateParameterValue(&stmt.Pars[x])
								if err != nil {
									return err
								}
								_, err = session.GetInt(2, true, true)
								if err != nil {
									return err
								}
							} else {
								//_, err = session.GetClr()
							}

						}
					}
				} else {
					// see if it is re-execute
					//fmt.Println(dataSet.Cols)
					if len(dataSet.Cols) == 0 && len(stmt.columns) > 0 {
						dataSet.Cols = make([]ParameterInfo, len(stmt.columns))
						copy(dataSet.Cols, stmt.columns)
					}
					for x := 0; x < len(dataSet.Cols); x++ {
						if dataSet.Cols[x].getDataFromServer {
							if dataSet.Cols[x].DataType == ROWID {
								rowid, err := newRowID(session)
								if err != nil {
									return err
								}
								if rowid == nil {
									dataSet.Cols[x].Value = nil
								} else {
									dataSet.Cols[x].Value = string(rowid.getBytes())
								}
								continue
							}
							dataSet.Cols[x].BValue, err = session.GetClr()
							//fmt.Println("buffer: ", temp)
							if err != nil {
								return err
							}
							err = stmt.calculateParameterValue(&dataSet.Cols[x])
							if err != nil {
								return err
							}
							if dataSet.Cols[x].DataType == LONG || dataSet.Cols[x].DataType == LongRaw {
								_, err = session.GetInt(4, true, true)
								if err != nil {
									return err
								}
								_, err = session.GetInt(4, true, true)
								if err != nil {
									return err
								}
							}
							//if temp == nil {
							//	dataSet.currentRow[x] = nil
							//	if dataSet.Cols[x].DataType == LONG || dataSet.Cols[x].DataType == LongRaw {
							//		_, err = session.GetBytes(2)
							//		if err != nil {
							//			return err
							//		}
							//		_, err = session.GetInt(4, true, true)
							//		if err != nil {
							//			return err
							//		}
							//	}
							//} else {
							//	//switch (this.m_definedColumnType)
							//	//{
							//	//case OraType.ORA_TIMESTAMP_DTY:
							//	//case OraType.ORA_TIMESTAMP:
							//	//case OraType.ORA_TIMESTAMP_LTZ_DTY:
							//	//case OraType.ORA_TIMESTAMP_LTZ:
							//	//	this.m_marshallingEngine.UnmarshalCLR_ColData(11);
							//	//	break;
							//	//case OraType.ORA_TIMESTAMP_TZ_DTY:
							//	//case OraType.ORA_TIMESTAMP_TZ:
							//	//	this.m_marshallingEngine.UnmarshalCLR_ColData(13);
							//	//	break;
							//	//case OraType.ORA_INTERVAL_YM_DTY:
							//	//case OraType.ORA_INTERVAL_DS_DTY:
							//	//case OraType.ORA_INTERVAL_YM:
							//	//case OraType.ORA_INTERVAL_DS:
							//	//case OraType.ORA_IBFLOAT:
							//	//case OraType.ORA_IBDOUBLE:
							//	//case OraType.ORA_RAW:
							//	//case OraType.ORA_CHAR:
							//	//case OraType.ORA_CHARN:
							//	//case OraType.ORA_VARCHAR:
							//	//	this.m_marshallingEngine.UnmarshalCLR_ColData(this.m_colMetaData.m_maxLength);
							//	//	break;
							//	//case OraType.ORA_RESULTSET:
							//	//	throw new InvalidOperationException();
							//	//case OraType.ORA_NUMBER:
							//	//case OraType.ORA_FLOAT:
							//	//case OraType.ORA_VARNUM:
							//	//	this.m_marshallingEngine.UnmarshalCLR_ColData(21);
							//	//	break;
							//	//case OraType.ORA_DATE:
							//	//	this.m_marshallingEngine.UnmarshalCLR_ColData(7);
							//	//	break;
							//	//default:
							//	//	throw new Exception("UnmarshalColumnData: Unimplemented type");
							//	//}
							//	//fmt.Println("type: ", dataSet.Cols[x].DataType)
							//	switch dataSet.Cols[x].DataType {
							//	case NCHAR, CHAR, LONG:
							//		if stmt.connection.strConv.GetLangID() != dataSet.Cols[x].CharsetID {
							//			tempCharset := stmt.connection.strConv.GetLangID()
							//			stmt.connection.strConv.SetLangID(dataSet.Cols[x].CharsetID)
							//			dataSet.currentRow[x] = stmt.connection.strConv.Decode(temp)
							//			stmt.connection.strConv.SetLangID(tempCharset)
							//		} else {
							//			dataSet.currentRow[x] = stmt.connection.strConv.Decode(temp)
							//		}
							//
							//	case NUMBER:
							//		dataSet.currentRow[x] = converters.DecodeNumber(temp)
							//		// if dataSet.Cols[x].Scale == 0 {
							//		// 	dataSet.currentRow[x] = int64(converters.DecodeInt(temp))
							//		// } else {
							//		// 	dataSet.currentRow[x] = converters.DecodeDouble(temp)
							//		// 	//base := math.Pow10(int(dataSet.Cols[x].Scale))
							//		// 	//if dataSet.Cols[x].Scale < 0x80 {
							//		// 	//	dataSet.currentRow[x] = math.Round(converters.DecodeDouble(temp)*base) / base
							//		// 	//} else {
							//		// 	//	dataSet.currentRow[x] = converters.DecodeDouble(temp)
							//		// 	//}
							//		// }
							//	case TimeStamp:
							//		fallthrough
							//	case TimeStampDTY:
							//		fallthrough
							//	case TimeStampeLTZ:
							//		fallthrough
							//	case TimeStampLTZ_DTY:
							//		fallthrough
							//	case TimeStampTZ:
							//		fallthrough
							//	case TimeStampTZ_DTY:
							//		fallthrough
							//	case DATE:
							//		dateVal, err := converters.DecodeDate(temp)
							//		if err != nil {
							//			return err
							//		}
							//		dataSet.currentRow[x] = dateVal
							//	//case :
							//	//	data, err := session.GetClr()
							//	//	if err != nil {
							//	//		return err
							//	//	}
							//	//	lob := &Lob{
							//	//		sourceLocator: data,
							//	//	}
							//	//	session.SaveState()
							//	//	dataSize, err := lob.getSize(session)
							//	//	if err != nil {
							//	//		return err
							//	//	}
							//	//	lobData, err := lob.getData(session)
							//	//	if err != nil {
							//	//		return err
							//	//	}
							//	//	if dataSize != int64(len(lobData)) {
							//	//		return errors.New("error reading lob data")
							//	//	}
							//	//	session.LoadState()
							//	//
							//	case OCIBlobLocator, OCIClobLocator:
							//		data, err := session.GetClr()
							//		if err != nil {
							//			return err
							//		}
							//		lob := &Lob{
							//			sourceLocator: data,
							//		}
							//		session.SaveState()
							//		dataSize, err := lob.getSize(stmt.connection)
							//		if err != nil {
							//			return err
							//		}
							//		lobData, err := lob.getData(stmt.connection)
							//		if err != nil {
							//			return err
							//		}
							//		session.LoadState()
							//		if dataSet.Cols[x].DataType == OCIBlobLocator {
							//			if dataSize != int64(len(lobData)) {
							//				return errors.New("error reading lob data")
							//			}
							//			dataSet.currentRow[x] = lobData
							//		} else {
							//			tempCharset := stmt.connection.strConv.GetLangID()
							//			if lob.variableWidthChar() {
							//				if stmt.connection.dBVersion.Number < 10200 && lob.littleEndianClob() {
							//					stmt.connection.strConv.SetLangID(2002)
							//				} else {
							//					stmt.connection.strConv.SetLangID(2000)
							//				}
							//			} else {
							//				stmt.connection.strConv.SetLangID(dataSet.Cols[x].CharsetID)
							//			}
							//			resultClobString := stmt.connection.strConv.Decode(lobData)
							//			stmt.connection.strConv.SetLangID(tempCharset)
							//			if dataSize != int64(len([]rune(resultClobString))) {
							//				return errors.New("error reading clob data")
							//			}
							//			dataSet.currentRow[x] = resultClobString
							//		}
							//	default:
							//		dataSet.currentRow[x] = temp
							//	}
							//	if dataSet.Cols[x].DataType == LONG || dataSet.Cols[x].DataType == LongRaw {
							//		_, err = session.GetInt(4, true, true)
							//		if err != nil {
							//			return err
							//		}
							//		_, err = session.GetInt(4, true, true)
							//		if err != nil {
							//			return err
							//		}
							//	}
							//}
						}
					}
					newRow := make(Row, dataSet.ColumnCount)
					for x := 0; x < len(dataSet.Cols); x++ {
						newRow[x] = dataSet.Cols[x].Value
					}
					//copy(newRow, dataSet.currentRow)
					dataSet.Rows = append(dataSet.Rows, newRow)
				}
			}
		case 8:
			size, err := session.GetInt(2, true, true)
			if err != nil {
				return err
			}
			for x := 0; x < 2; x++ {
				stmt.scnFromExe[x], err = session.GetInt(4, true, true)
				if err != nil {
					return err
				}
			}
			for x := 2; x < size; x++ {
				_, err = session.GetInt(4, true, true)
				if err != nil {
					return err
				}
			}
			_, err = session.GetInt(2, true, true)
			//fmt.Println(num)
			//if (num > 0)
			//	this.m_marshallingEngine.UnmarshalNBytes_ScanOnly(num);
			// get session timezone
			size, err = session.GetInt(2, true, true)
			for x := 0; x < size; x++ {
				_, val, num, err := session.GetKeyVal()
				if err != nil {
					return err
				}
				//fmt.Println(key, val, num)
				if num == 163 {
					session.TimeZone = val
					//fmt.Println("session time zone", session.TimeZone)
				}
			}
			if session.TTCVersion >= 4 {
				// get queryID
				size, err = session.GetInt(4, true, true)
				if err != nil {
					return err
				}
				if size > 0 {
					bty, err := session.GetBytes(size)
					if err != nil {
						return err
					}
					if len(bty) >= 8 {
						stmt.queryID = binary.LittleEndian.Uint64(bty[size-8:])
						fmt.Println("query ID: ", stmt.queryID)
					}
				}
			}

		case 11:
			err = dataSet.load(session)
			if err != nil {
				return err
			}
			//dataSet.BindDirections = make([]byte, dataSet.ColumnCount)
			for x := 0; x < dataSet.ColumnCount; x++ {
				direction, err := session.GetByte()
				switch direction {
				case 32:
					stmt.Pars[x].Direction = Input
				case 16:
					stmt.Pars[x].Direction = Output
					containOutputPars = true
				case 48:
					stmt.Pars[x].Direction = InOut
					containOutputPars = true
				}
				if err != nil {
					return err
				}
			}
		case 16:
			size, err := session.GetByte()
			if err != nil {
				return err
			}
			_, err = session.GetBytes(int(size))
			if err != nil {
				return err
			}
			dataSet.MaxRowSize, err = session.GetInt(4, true, true)
			if err != nil {
				return err
			}
			dataSet.ColumnCount, err = session.GetInt(4, true, true)
			if err != nil {
				return err
			}
			if dataSet.ColumnCount > 0 {
				_, err = session.GetByte() // session.GetInt(1, false, false)
			}
			dataSet.Cols = make([]ParameterInfo, dataSet.ColumnCount)
			for x := 0; x < dataSet.ColumnCount; x++ {
				err = dataSet.Cols[x].load(session)
				if err != nil {
					return err
				}
				if dataSet.Cols[x].DataType == LONG || dataSet.Cols[x].DataType == LongRaw {
					stmt._hasLONG = true
				}
				if dataSet.Cols[x].DataType == OCIClobLocator || dataSet.Cols[x].DataType == OCIBlobLocator {
					stmt._hasBLOB = true
				}
			}
			stmt.columns = make([]ParameterInfo, dataSet.ColumnCount)
			copy(stmt.columns, dataSet.Cols)
			_, err = session.GetDlc()
			if session.TTCVersion >= 3 {
				_, err = session.GetInt(4, true, true)
				_, err = session.GetInt(4, true, true)
			}
			if session.TTCVersion >= 4 {
				_, err = session.GetInt(4, true, true)
				_, err = session.GetInt(4, true, true)
			}
			if session.TTCVersion >= 5 {
				_, err = session.GetDlc()
			}
		case 21:
			_, err := session.GetInt(2, true, true) // noOfColumnSent
			if err != nil {
				return err
			}
			bitVectorLen := dataSet.ColumnCount / 8
			if dataSet.ColumnCount%8 > 0 {
				bitVectorLen++
			}
			bitVector := make([]byte, bitVectorLen)
			for x := 0; x < bitVectorLen; x++ {
				bitVector[x], err = session.GetByte()
				if err != nil {
					return err
				}
			}
			dataSet.setBitVector(bitVector)

		default:
			loop = false
		}
	}
	if stmt.connection.connOption.Tracer.IsOn() {
		dataSet.Trace(stmt.connection.connOption.Tracer)
	}
	return nil
}
func (stmt *defaultStmt) calculateParameterValue(param *ParameterInfo) error {
	//var ret driver.Value
	session := stmt.connection.session
	if param.BValue == nil {
		param.Value = nil
		return nil
	}
	switch param.DataType {
	case NCHAR, CHAR, LONG:
		if stmt.connection.strConv.GetLangID() != param.CharsetID {
			tempCharset := stmt.connection.strConv.GetLangID()
			stmt.connection.strConv.SetLangID(param.CharsetID)
			param.Value = stmt.connection.strConv.Decode(param.BValue)
			stmt.connection.strConv.SetLangID(tempCharset)
		} else {
			param.Value = stmt.connection.strConv.Decode(param.BValue)
		}
	case NUMBER:
		param.Value = converters.DecodeNumber(param.BValue)
	case TimeStamp:
		fallthrough
	case TimeStampDTY:
		fallthrough
	case TimeStampeLTZ:
		fallthrough
	case TimeStampLTZ_DTY:
		fallthrough
	case TimeStampTZ:
		fallthrough
	case TimeStampTZ_DTY:
		fallthrough
	case DATE:
		dateVal, err := converters.DecodeDate(param.BValue)
		if err != nil {
			return err
		}
		param.Value = dateVal
	case OCIBlobLocator, OCIClobLocator:
		data, err := session.GetClr()
		if err != nil {
			return err
		}
		lob := &Lob{
			sourceLocator: data,
		}
		session.SaveState()
		dataSize, err := lob.getSize(stmt.connection)
		if err != nil {
			return err
		}
		lobData, err := lob.getData(stmt.connection)
		if err != nil {
			return err
		}
		session.LoadState()
		if param.DataType == OCIBlobLocator {
			if dataSize != int64(len(lobData)) {
				return errors.New("error reading lob data")
			}
			param.Value = lobData
		} else {
			tempCharset := stmt.connection.strConv.GetLangID()
			if lob.variableWidthChar() {
				if stmt.connection.dBVersion.Number < 10200 && lob.littleEndianClob() {
					stmt.connection.strConv.SetLangID(2002)
				} else {
					stmt.connection.strConv.SetLangID(2000)
				}
			} else {
				stmt.connection.strConv.SetLangID(param.CharsetID)
			}
			resultClobString := stmt.connection.strConv.Decode(lobData)
			stmt.connection.strConv.SetLangID(tempCharset)
			if dataSize != int64(len([]rune(resultClobString))) {
				return errors.New("error reading clob data")
			}
			param.Value = resultClobString
		}
	default:
		param.Value = param.BValue
	}
	return nil
}

func (stmt *defaultStmt) Close() error {
	if stmt.cursorID != 0 {
		session := stmt.connection.session
		session.ResetBuffer()
		session.PutBytes(17, 105, 0, 1, 1, 1)
		session.PutInt(stmt.cursorID, 4, true, true)
		return (&simpleObject{
			session:     session,
			operationID: 0x93,
			data:        nil,
			err:         nil,
		}).write().read()
	}
	return nil
}

func (stmt *Stmt) Exec(args []driver.Value) (driver.Result, error) {
	stmt.connection.connOption.Tracer.Printf("Exec:\n%s", stmt.text)
	for x := 0; x < len(args); x++ {
		par := *stmt.NewParam("", args[x], 0, Input)
		if x < len(stmt.Pars) {
			if par.MaxLen > stmt.Pars[x].MaxLen {
				stmt.reSendParDef = true
			}
			stmt.Pars[x] = par
		} else {
			stmt.Pars = append(stmt.Pars, par)
		}
		stmt.connection.connOption.Tracer.Printf("    %d:\n%v", x, args[x])
	}
	session := stmt.connection.session
	//if len(args) > 0 {
	//	stmt.Pars = nil
	//}
	//for x := 0; x < len(args); x++ {
	//	stmt.AddParam("", args[x], 0, Input)
	//}
	session.ResetBuffer()
	err := stmt.write(session)
	if err != nil {
		return nil, err
	}
	dataSet := new(DataSet)
	err = stmt.read(dataSet)
	if err != nil {
		return nil, err
	}
	result := new(QueryResult)
	if session.Summary != nil {
		result.rowsAffected = int64(session.Summary.CurRowNumber)
	}
	return result, nil
}
func (stmt *Stmt) NewParam(name string, val driver.Value, size int, direction ParameterDirection) *ParameterInfo {
	param := &ParameterInfo{
		Name:        name,
		Direction:   direction,
		Flag:        3,
		CharsetID:   871,
		CharsetForm: 1,
	}
	if val == nil {
		param.DataType = NCHAR
		param.BValue = nil
		param.ContFlag = 16
		param.MaxCharLen = 0
		param.MaxLen = 1
		param.CharsetForm = 1
	} else {
		switch val := val.(type) {
		case int64:
			param.BValue = converters.EncodeInt64(val)
			param.DataType = NUMBER
		case int32:
			param.BValue = converters.EncodeInt(int(val))
			param.DataType = NUMBER
		case int16:
			param.BValue = converters.EncodeInt(int(val))
			param.DataType = NUMBER
		case int8:
			param.BValue = converters.EncodeInt(int(val))
			param.DataType = NUMBER
		case int:
			param.BValue = converters.EncodeInt(val)
			param.DataType = NUMBER
		case float32:
			param.BValue, _ = converters.EncodeDouble(float64(val))
			param.DataType = NUMBER
		case float64:
			param.BValue, _ = converters.EncodeDouble(val)
			param.DataType = NUMBER
		case time.Time:
			param.BValue = converters.EncodeDate(val)
			param.DataType = DATE
			param.ContFlag = 0
			param.MaxLen = 11
			param.MaxCharLen = 11
		case string:
			param.DataType = NCHAR
			param.ContFlag = 16
			param.MaxCharLen = len(val)
			param.CharsetForm = 1
			if val == "" && direction == Input {
				param.BValue = nil
				param.MaxLen = 1
			} else {
				param.BValue = stmt.connection.strConv.Encode(val)
				if size > len(val) {
					param.MaxCharLen = size
				}
				param.MaxLen = param.MaxCharLen * converters.MaxBytePerChar(param.CharsetID)
			}
		case []byte:
			param.BValue = val
			param.DataType = RAW
			param.MaxLen = len(val)
			param.ContFlag = 0
			param.MaxCharLen = 0
			param.CharsetForm = 0
		}
		if param.DataType == NUMBER {
			param.ContFlag = 0
			param.MaxCharLen = 0
			param.MaxLen = 22
			param.CharsetForm = 0
		}
		if direction == Output {
			param.BValue = nil
		}
	}
	return param
}
func (stmt *Stmt) AddParam(name string, val driver.Value, size int, direction ParameterDirection) {
	stmt.Pars = append(stmt.Pars, *stmt.NewParam(name, val, size, direction))

}
func (stmt *Stmt) AddRefCursorParam(_ string) {
	par := stmt.NewParam("1", nil, 0, Output)
	par.DataType = REFCURSOR
	par.ContFlag = 0
	par.CharsetForm = 0
	stmt.Pars = append(stmt.Pars, *par)
}

//func (stmt *Stmt) AddParam(name string, val driver.BValue, size int, direction ParameterDirection) {
//	param := ParameterInfo{
//		Name:        name,
//		Direction:   direction,
//		Flag:        3,
//		CharsetID:   871,
//		CharsetForm: 1,
//	}
//	//if param.Direction == Output {
//	//	if _, ok := val.(string); ok {
//	//		param.MaxCharLen = size
//	//		param.MaxLen = size * converters.MaxBytePerChar(stmt.connection.strConv.LangID)
//	//	}
//	//	stmt.Pars = append(stmt.Pars, param)
//	//	return
//	//}
//	if val == nil {
//		param.DataType = NCHAR
//		param.BValue = nil
//		param.ContFlag = 16
//		param.MaxCharLen = 0
//		param.MaxLen = 1
//		param.CharsetForm = 1
//	} else {
//		switch val := val.(type) {
//		case int64:
//			param.BValue = converters.EncodeInt64(val)
//			param.DataType = NUMBER
//		case int32:
//			param.BValue = converters.EncodeInt(int(val))
//			param.DataType = NUMBER
//		case int16:
//			param.BValue = converters.EncodeInt(int(val))
//			param.DataType = NUMBER
//		case int8:
//			param.BValue = converters.EncodeInt(int(val))
//			param.DataType = NUMBER
//		case int:
//			param.BValue = converters.EncodeInt(val)
//			param.DataType = NUMBER
//		case float32:
//			param.BValue, _ = converters.EncodeDouble(float64(val))
//			param.DataType = NUMBER
//		case float64:
//			param.BValue, _ = converters.EncodeDouble(val)
//			param.DataType = NUMBER
//		case time.Time:
//			param.BValue = converters.EncodeDate(val)
//			param.DataType = DATE
//			param.ContFlag = 0
//			param.MaxLen = 11
//			param.MaxCharLen = 11
//		case string:
//			param.BValue = stmt.connection.strConv.Encode(val)
//			param.DataType = NCHAR
//			param.ContFlag = 16
//			param.MaxCharLen = len(val)
//			if size > len(val) {
//				param.MaxCharLen = size
//			}
//			param.MaxLen = param.MaxCharLen * converters.MaxBytePerChar(stmt.connection.strConv.LangID)
//			param.CharsetForm = 1
//		}
//		if param.DataType == NUMBER {
//			param.ContFlag = 0
//			param.MaxCharLen = 22
//			param.MaxLen = 22
//			param.CharsetForm = 1
//		}
//		if direction == Output {
//			param.BValue = nil
//		}
//	}
//	stmt.Pars = append(stmt.Pars, param)
//}

//func (stmt *Stmt) reExec() (driver.Rows, error) {
//
//}
func (stmt *Stmt) Query(args []driver.Value) (driver.Rows, error) {
	stmt.connection.connOption.Tracer.Printf("Query:\n%s", stmt.text)
	stmt._noOfRowsToFetch = stmt.connection.connOption.PrefetchRows
	stmt._hasMoreRows = true
	for x := 0; x < len(args); x++ {
		par := *stmt.NewParam("", args[x], 0, Input)
		if x < len(stmt.Pars) {
			if par.MaxLen > stmt.Pars[x].MaxLen {
				stmt.reSendParDef = true
			}
			stmt.Pars[x] = par
		} else {
			stmt.Pars = append(stmt.Pars, par)
		}
	}
	//stmt.Pars = nil
	//for x := 0; x < len(args); x++ {
	//	stmt.AddParam()
	//}
	stmt.connection.session.ResetBuffer()
	// if re-execute
	err := stmt.write(stmt.connection.session)
	if err != nil {
		return nil, err
	}
	//err = stmt.connection.session.Write()
	//if err != nil {
	//	return nil, err
	//}
	dataSet := new(DataSet)
	err = stmt.read(dataSet)
	if err != nil {
		return nil, err
	}
	return dataSet, nil
}

func (stmt *Stmt) NumInput() int {
	return -1
}

/*
parse = true
execute = true
fetch = true if hasReturn or PLSQL
define = false
*/
