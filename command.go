package go_ora

import (
	"database/sql/driver"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/sijms/go-ora/converters"
	"github.com/sijms/go-ora/network"
	"math"
	"time"

	//charmap "golang.org/x/text/encoding/charmap"
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

type Stmt struct {
	connection      *Connection
	Text            string
	stmtType        StmtType
	Pars            []ParameterInfo
	hasReturnClause bool
	parse           bool // means parse the command in the server this occur if the stmt is not cached
	execute         bool
	define          bool
	exeOption       int
	cursorID        int
	noOfRowsToFetch int
	noOfDefCols     int
	al8i4           []byte
	arrayBindCount  int
	queryID         uint64
	scnFromExe      []int
	hasMoreRows     bool
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
		connection: conn,
		Text:       text,
		parse:      true,
		execute:    true,
		define:     false,
		al8i4:      make([]byte, 13),
		scnFromExe: make([]int, 2),
	}

	// get stmt type
	uCmdText := strings.Trim(strings.ToUpper(text), " ")
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

	// returning cluase
	var err error
	ret.hasReturnClause, err = regexp.MatchString(`\bRETURNING\b`, uCmdText)
	if err != nil {
		ret.hasReturnClause = false
	}
	ret.al8i4[0] = 1
	switch ret.stmtType {
	case DML:
		fallthrough
	case PLSQL:
		if ret.arrayBindCount <= 1 {
			ret.al8i4[1] = 1
		} else {
			ret.al8i4[1] = uint8(ret.arrayBindCount)
		}
	case OTHERS:
		ret.al8i4[1] = 1
	default:
		ret.al8i4[1] = 0
	}
	if ret.stmtType == SELECT {
		ret.al8i4[7] = 1
	} else {
		ret.al8i4[7] = 0
	}
	return ret
}

func (stmt *Stmt) write(session *network.Session) error {
	exeOp := stmt.getExeOption()
	session.PutBytes([]byte{3, 0x5E, 0})
	session.PutUint(exeOp, 4, true, true)
	session.PutUint(stmt.cursorID, 2, true, true)
	if stmt.cursorID == 0 {
		session.PutUint(1, 1, false, false)
	} else {
		session.PutUint(0, 1, false, false)
	}
	session.PutUint(len(stmt.Text), 4, true, true)
	session.PutUint(1, 1, false, false)
	session.PutUint(13, 2, true, true)
	session.PutBytes([]byte{0, 0})
	if exeOp&0x40 == 0 && exeOp&0x20 != 0 && exeOp&0x1 != 0 && stmt.stmtType == SELECT {
		session.PutUint(0, 1, false, false)
		session.PutUint(stmt.noOfRowsToFetch, 4, true, true)
	} else {
		session.PutUint(0, 4, true, true)
		session.PutUint(0, 4, true, true)
	}
	// longFetchSize == 0 marshal 1 else marshal longFetchSize
	session.PutUint(1, 4, true, true)
	if len(stmt.Pars) > 0 {
		session.PutUint(1, 1, false, false)
		session.PutUint(len(stmt.Pars), 2, true, true)
	} else {
		session.PutUint(0, 1, false, false)
		session.PutUint(0, 1, false, false)
	}
	session.PutBytes([]byte{0, 0, 0, 0, 0})
	if stmt.define {
		session.PutUint(1, 1, false, false)
		session.PutUint(stmt.noOfDefCols, 2, true, true)
	} else {
		session.PutUint(0, 1, false, false)
		session.PutUint(0, 1, false, false)
	}
	if session.TTCVersion >= 4 {
		session.PutUint(0, 1, false, false) // dbChangeRegisterationId
		session.PutUint(0, 1, false, false)
		session.PutUint(1, 1, false, false)
	}
	if session.TTCVersion >= 5 {
		session.PutUint(0, 1, false, false)
		session.PutUint(0, 1, false, false)
		session.PutUint(0, 1, false, false)
		session.PutUint(0, 1, false, false)
		session.PutUint(0, 1, false, false)
	}

	session.PutBytes([]byte(stmt.Text))
	for x := 0; x < len(stmt.al8i4); x++ {
		session.PutUint(stmt.al8i4[x], 2, true, true)
	}
	for _, par := range stmt.Pars {
		_ = par.write(session)
	}
	if len(stmt.Pars) > 0 {
		session.PutUint(7, 1, false, false)
		for _, par := range stmt.Pars {
			session.PutClr(par.Value)
		}
	}
	return session.Write()
}

//func (stmt *Stmt) NoQuery() error {
//	stmt.autoCommit = true
//	stmt.connection.session.ResetBuffer()
//	err := stmt.write(stmt.connection.session)
//	if err != nil {
//		return err
//	}
//	err = stmt.connection.session.write()
//	if err != nil {
//		return err
//	}
//	dataSet := new(DataSet)
//	return stmt.read(dataSet)
//}

//func (stmt *Stmt) QueryN() (*DataSet, error) {
//	stmt.autoCommit = false
//	stmt.noOfRowsToFetch = 25
//	stmt.connection.session.ResetBuffer()
//	err := stmt.write(stmt.connection.session)
//	if err != nil {
//		return nil, err
//	}
//	err = stmt.connection.session.write()
//	if err != nil {
//		return nil, err
//	}
//	dataSet := new(DataSet)
//	err = stmt.read(dataSet)
//	if err != nil {
//		return nil, err
//	}
//	return dataSet, nil
//
//}

func (stmt *Stmt) getExeOption() int {
	op := 0
	if stmt.stmtType == PLSQL || stmt.hasReturnClause {
		op |= 0x40000
	}

	if stmt.connection.autoCommit {
		op |= 0x100
	}
	if stmt.parse {
		op |= 1
	}
	if stmt.execute {
		op |= 0x20
	}
	if len(stmt.Pars) > 0 {
		op |= 0x8
		if stmt.stmtType == PLSQL || stmt.hasReturnClause {
			op |= 0x400
		}
	}
	if stmt.stmtType != PLSQL && !stmt.hasReturnClause {
		op |= 0x8000
	}
	return op

	/* HasReturnClause
	if  stmt.PLSQL or cmdText == "" return false
	Regex.IsMatch(cmdText, "\\bRETURNING\\b"
	*/
}
func (stmt *Stmt) fetch(dataSet *DataSet) error {
	stmt.connection.session.ResetBuffer()
	stmt.connection.session.PutBytes([]byte{3, 5, 0})
	stmt.connection.session.PutInt(stmt.cursorID, 2, true, true)
	stmt.connection.session.PutInt(stmt.noOfRowsToFetch, 2, true, true)
	err := stmt.connection.session.Write()
	if err != nil {
		return err
	}
	return stmt.read(dataSet)
}
func (stmt *Stmt) read(dataSet *DataSet) error {
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
			//fmt.Println(stmt.connection.session.Summary)
			//fmt.Println(stmt.connection.session.Summary)
			stmt.cursorID = stmt.connection.session.Summary.CursorID
			if stmt.connection.session.HasError() {
				if stmt.connection.session.Summary.RetCode == 1403 {
					stmt.hasMoreRows = false
					stmt.connection.session.Summary = nil
				} else {
					return errors.New(stmt.connection.session.GetError())
				}

			}
			loop = false
		case 6:
			//_, err = session.GetByte()
			err = dataSet.read(session)
			if err != nil {
				return err
			}
			if !after7 {
				if stmt.stmtType == SELECT {

				}
			}
		case 7:
			after7 = true
			if stmt.hasReturnClause {
				//if (bHasReturningParams && bindAccessors != null)
				//{
				//	int paramLen = bindAccessors.Length;
				//	this.m_marshallingEngine.m_oraBufRdr.m_bHavingParameterData = true;
				//	for (int index1 = 0; index1 < paramLen; ++index1)
				//	{
				//		if (bindAccessors[index1] != null)
				//		{
				//			int num = (int) this.m_marshallingEngine.UnmarshalUB4(false);
				//			if (num > 1)
				//				bMoreThanOneRowAffectedByDmlWithRetClause = true;
				//			if (num == 0)
				//			{
				//				bindAccessors[index1].AddNullForData();
				//			}
				//			else
				//			{
				//				for (int index2 = 0; index2 < num; ++index2)
				//				{
				//					bindAccessors[index1].m_bReceivedOutValueFromServer = true;
				//					bindAccessors[index1].UnmarshalOneRow();
				//				}
				//			}
				//		}
				//	}
				//	this.m_marshallingEngine.m_oraBufRdr.m_currentOB = (OraBuf) null;
				//	this.m_marshallingEngine.m_oraBufRdr.m_bHavingParameterData = false;
				//	++noOfRowsFetched;
				//	continue;
				//}
			} else {
				if containOutputPars {
					for x := 0; x < dataSet.ColumnCount; x++ {
						if stmt.Pars[x].Direction != Input {
							stmt.Pars[x].Value, err = session.GetClr()
						} else {
							_, err = session.GetClr()
						}
						if err != nil {
							return err
						}
						_, err = session.GetInt(2, true, true)
					}
				} else {
					newRow := make(Row, dataSet.ColumnCount)
					for x := 0; x < dataSet.ColumnCount; x++ {
						if dataSet.Cols[x].getDataFromServer {

							temp, err := session.GetClr()
							if err != nil {
								return err
							}
							if temp == nil {
								newRow[x] = nil
							} else {
								//switch (this.m_definedColumnType)
								//{
								//case OraType.ORA_TIMESTAMP_DTY:
								//case OraType.ORA_TIMESTAMP:
								//case OraType.ORA_TIMESTAMP_LTZ_DTY:
								//case OraType.ORA_TIMESTAMP_LTZ:
								//	this.m_marshallingEngine.UnmarshalCLR_ColData(11);
								//	break;
								//case OraType.ORA_TIMESTAMP_TZ_DTY:
								//case OraType.ORA_TIMESTAMP_TZ:
								//	this.m_marshallingEngine.UnmarshalCLR_ColData(13);
								//	break;
								//case OraType.ORA_INTERVAL_YM_DTY:
								//case OraType.ORA_INTERVAL_DS_DTY:
								//case OraType.ORA_INTERVAL_YM:
								//case OraType.ORA_INTERVAL_DS:
								//case OraType.ORA_IBFLOAT:
								//case OraType.ORA_IBDOUBLE:
								//case OraType.ORA_RAW:
								//case OraType.ORA_CHAR:
								//case OraType.ORA_CHARN:
								//case OraType.ORA_VARCHAR:
								//	this.m_marshallingEngine.UnmarshalCLR_ColData(this.m_colMetaData.m_maxLength);
								//	break;
								//case OraType.ORA_RESULTSET:
								//	throw new InvalidOperationException();
								//case OraType.ORA_NUMBER:
								//case OraType.ORA_FLOAT:
								//case OraType.ORA_VARNUM:
								//	this.m_marshallingEngine.UnmarshalCLR_ColData(21);
								//	break;
								//case OraType.ORA_DATE:
								//	this.m_marshallingEngine.UnmarshalCLR_ColData(7);
								//	break;
								//default:
								//	throw new Exception("UnmarshalColumnData: Unimplemented type");
								//}
								switch dataSet.Cols[x].DataType {
								case NCHAR:
									//fmt.Println("string value:", stmt.connection.strConv.Decode(temp))
									newRow[x] = stmt.connection.strConv.Decode(temp)
								case NUMBER:
									if dataSet.Cols[x].Scale == 0 {
										newRow[x] = int64(converters.DecodeInt(temp))
									} else {
										base := math.Pow10(int(dataSet.Cols[x].Scale))
										newRow[x] = math.Round(converters.DecodeDouble(temp)*base) / base
									}
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
									dateVal, err := converters.DecodeDate(temp)
									if err != nil {
										return err
									}
									newRow[x] = dateVal
								default:
									newRow[x] = temp
								}
							}
						} else {
							// copy from last row
							if len(dataSet.Rows) > 0 {
								lastRow := dataSet.Rows[len(dataSet.Rows)-1]
								newRow[x] = lastRow[x]
							} else {
								newRow[x] = nil
							}
						}

					}
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
			err = dataSet.read(session)
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
			size, err := session.GetInt(1, false, false)
			if err != nil {
				return err
			}
			_, err = session.GetBytes(size)
			if err != nil {
				return err
			}
			dataSet.MaxRowSize, err = session.GetInt(4, true, true)
			if err != nil {
				return err
			}
			noOfColumns, err := session.GetInt(4, true, true)
			if err != nil {
				return err
			}
			if noOfColumns > 0 {
				_, err = session.GetInt(1, false, false)
			}
			dataSet.Cols = make([]ParameterInfo, noOfColumns)
			for x := 0; x < noOfColumns; x++ {
				err = dataSet.Cols[x].read(session)
				if err != nil {
					return err
				}
			}
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
			noOfColumnSent, err := session.GetInt(2, true, true)
			if err != nil {
				return err
			}
			bitVectorLen := noOfColumnSent / 8
			if noOfColumnSent%8 > 0 {
				bitVectorLen++
			}
			bitVector := make([]byte, bitVectorLen)
			for x := 0; x < bitVectorLen; x++ {
				bitVector[x], err = session.GetByte()
			}
			dataSet.setBitVector(bitVector)

		default:
			loop = false
		}
	}
	return nil
}

func (stmt *Stmt) Close() error {
	session := stmt.connection.session
	session.ResetBuffer()
	session.PutBytes([]byte{17, 105, 0, 1, 1, 1})
	session.PutInt(stmt.cursorID, 4, true, true)
	return (&simpleObject{
		session:     session,
		operationID: 0x93,
		data:        nil,
		err:         nil,
	}).write().read()
}

func (stmt *Stmt) Exec(args []driver.Value) (driver.Result, error) {
	session := stmt.connection.session
	for x := 0; x < len(args); x++ {
		stmt.AddParam("", args[x], 0, Input)
	}
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
func (stmt *Stmt) AddParam(name string, val driver.Value, size int, direction ParameterDirection) {
	param := ParameterInfo{
		Name:        name,
		Direction:   direction,
		Flag:        3,
		CharsetID:   871,
		CharsetForm: 1,
	}
	//if param.Direction == Output {
	//	if _, ok := val.(string); ok {
	//		param.MaxCharLen = size
	//		param.MaxLen = size * converters.MaxBytePerChar(stmt.connection.strConv.LangID)
	//	}
	//	stmt.Pars = append(stmt.Pars, param)
	//	return
	//}
	switch val := val.(type) {
	case int64:
		param.Value = converters.EncodeInt64(val)
		param.DataType = NUMBER
	case int32:
		param.Value = converters.EncodeInt(int(val))
		param.DataType = NUMBER
	case int16:
		param.Value = converters.EncodeInt(int(val))
		param.DataType = NUMBER
	case int8:
		param.Value = converters.EncodeInt(int(val))
		param.DataType = NUMBER
	case int:
		param.Value = converters.EncodeInt(val)
		param.DataType = NUMBER
	case float32:
		param.Value, _ = converters.EncodeDouble(float64(val))
		param.DataType = NUMBER
	case float64:
		param.Value, _ = converters.EncodeDouble(val)
		param.DataType = NUMBER
	case time.Time:
		param.Value = converters.EncodeDate(val)
		param.DataType = DATE
		param.ContFlag = 0
		param.MaxLen = 11
		param.MaxCharLen = 11
	case string:
		param.Value = stmt.connection.strConv.Encode(val)
		param.DataType = NCHAR
		param.ContFlag = 16
		param.MaxCharLen = len(val)
		if size > len(val) {
			param.MaxCharLen = size
		}
		param.MaxLen = param.MaxCharLen * converters.MaxBytePerChar(stmt.connection.strConv.LangID)
		param.CharsetForm = 1
	}
	if param.DataType == NUMBER {
		param.ContFlag = 0
		param.MaxCharLen = 22
		param.MaxLen = 22
		param.CharsetForm = 1
	}
	if direction == Output {
		param.Value = nil
	}
	stmt.Pars = append(stmt.Pars, param)
}

func (stmt *Stmt) Query(args []driver.Value) (driver.Rows, error) {
	stmt.noOfRowsToFetch = 25
	stmt.hasMoreRows = true
	for x := 0; x < len(args); x++ {
		stmt.AddParam("", args[x], 0, Input)
	}
	stmt.connection.session.ResetBuffer()
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
