package go_ora

import (
	"errors"
	"fmt"
	"github.com/sijms/go-ora/v2/advanced_nego"
	"github.com/sijms/go-ora/v2/network"
	"github.com/sijms/go-ora/v2/trace"
	"io/ioutil"
	"net"
	"net/url"
	"os"
	"os/user"
	"path"
	"runtime"
	"strconv"
	"strings"
	"time"
)

type PromotableTransaction int

//const (
//	Promotable PromotableTransaction = 1
//	Local      PromotableTransaction = 0
//)

type DBAPrivilege int

const (
	NONE    DBAPrivilege = 0
	SYSDBA  DBAPrivilege = 0x20
	SYSOPER DBAPrivilege = 0x40
)

type AuthType int

const (
	Normal   AuthType = 0
	OS       AuthType = 1
	Kerberos AuthType = 2
	TCPS     AuthType = 3
)
const defaultPort int = 1521

func DBAPrivilegeFromString(s string) DBAPrivilege {
	S := strings.ToUpper(s)
	if S == "SYSDBA" {
		return SYSDBA
	} else if S == "SYSOPER" {
		return SYSOPER
	} else {
		return NONE
	}
}

func getCharsetID(charset string) (int, error) {
	charsetMap := map[string]int{
		"US7ASCII":         1,
		"WE8DEC":           2,
		"WE8HP":            3,
		"US8PC437":         4,
		"WE8EBCDIC37":      5,
		"WE8EBCDIC500":     6,
		"WE8EBCDIC1140":    7,
		"WE8EBCDIC285":     8,
		"WE8EBCDIC1146":    9,
		"WE8PC850":         10,
		"D7DEC":            11,
		"F7DEC":            12,
		"S7DEC":            13,
		"E7DEC":            14,
		"SF7ASCII":         15,
		"NDK7DEC":          16,
		"I7DEC":            17,
		"NL7DEC":           18,
		"CH7DEC":           19,
		"YUG7ASCII":        20,
		"SF7DEC":           21,
		"TR7DEC":           22,
		"IW7IS960":         23,
		"IN8ISCII":         25,
		"WE8EBCDIC1148":    27,
		"WE8PC858":         28,
		"WE8ISO8859P1":     31,
		"EE8ISO8859P2":     32,
		"SE8ISO8859P3":     33,
		"NEE8ISO8859P4":    34,
		"CL8ISO8859P5":     35,
		"AR8ISO8859P6":     36,
		"EL8ISO8859P7":     37,
		"IW8ISO8859P8":     38,
		"WE8ISO8859P9":     39,
		"NE8ISO8859P10":    40,
		"TH8TISASCII":      41,
		"TH8TISEBCDIC":     42,
		"BN8BSCII":         43,
		"VN8VN3":           44,
		"VN8MSWIN1258":     45,
		"WE8ISO8859P15":    46,
		"BLT8ISO8859P13":   47,
		"CEL8ISO8859P14":   48,
		"CL8ISOIR111":      49,
		"WE8NEXTSTEP":      50,
		"CL8KOI8U":         51,
		"AZ8ISO8859P9E":    52,
		"AR8ASMO708PLUS":   61,
		"AR8EBCDICX":       70,
		"AR8XBASIC":        72,
		"EL8DEC":           81,
		"TR8DEC":           82,
		"WE8EBCDIC37C":     90,
		"WE8EBCDIC500C":    91,
		"IW8EBCDIC424":     92,
		"TR8EBCDIC1026":    93,
		"WE8EBCDIC871":     94,
		"WE8EBCDIC284":     95,
		"WE8EBCDIC1047":    96,
		"WE8EBCDIC1140C":   97,
		"WE8EBCDIC1145":    98,
		"WE8EBCDIC1148C":   99,
		"WE8EBCDIC1047E":   100,
		"WE8EBCDIC924":     101,
		"EEC8EUROASCI":     110,
		"EEC8EUROPA3":      113,
		"LA8PASSPORT":      114,
		"BG8PC437S":        140,
		"EE8PC852":         150,
		"RU8PC866":         152,
		"RU8BESTA":         153,
		"IW8PC1507":        154,
		"RU8PC855":         155,
		"TR8PC857":         156,
		"CL8MACCYRILLIC":   158,
		"CL8MACCYRILLICS":  159,
		"WE8PC860":         160,
		"IS8PC861":         161,
		"EE8MACCES":        162,
		"EE8MACCROATIANS":  163,
		"TR8MACTURKISHS":   164,
		"IS8MACICELANDICS": 165,
		"EL8MACGREEKS":     166,
		"IW8MACHEBREWS":    167,
		"EE8MSWIN1250":     170,
		"CL8MSWIN1251":     171,
		"ET8MSWIN923":      172,
		"BG8MSWIN":         173,
		"EL8MSWIN1253":     174,
		"IW8MSWIN1255":     175,
		"LT8MSWIN921":      176,
		"TR8MSWIN1254":     177,
		"WE8MSWIN1252":     178,
		"BLT8MSWIN1257":    179,
		"D8EBCDIC273":      180,
		"I8EBCDIC280":      181,
		"DK8EBCDIC277":     182,
		"S8EBCDIC278":      183,
		"EE8EBCDIC870":     184,
		"CL8EBCDIC1025":    185,
		"F8EBCDIC297":      186,
		"IW8EBCDIC1086":    187,
		"CL8EBCDIC1025X":   188,
		"D8EBCDIC1141":     189,
		"N8PC865":          190,
		"BLT8CP921":        191,
		"LV8PC1117":        192,
		"LV8PC8LR":         193,
		"BLT8EBCDIC1112":   194,
		"LV8RST104090":     195,
		"CL8KOI8R":         196,
		"BLT8PC775":        197,
		"DK8EBCDIC1142":    198,
		"S8EBCDIC1143":     199,
		"I8EBCDIC1144":     200,
		"F7SIEMENS9780X":   201,
		"E7SIEMENS9780X":   202,
		"S7SIEMENS9780X":   203,
		"DK7SIEMENS9780X":  204,
		"N7SIEMENS9780X":   205,
		"I7SIEMENS9780X":   206,
		"D7SIEMENS9780X":   207,
		"F8EBCDIC1147":     208,
		"WE8GCOS7":         210,
		"EL8GCOS7":         211,
		"US8BS2000":        221,
		"D8BS2000":         222,
		"F8BS2000":         223,
		"E8BS2000":         224,
		"DK8BS2000":        225,
		"S8BS2000":         226,
		"WE8BS2000E":       230,
		"WE8BS2000":        231,
		"EE8BS2000":        232,
		"CE8BS2000":        233,
		"CL8BS2000":        235,
		"WE8BS2000L5":      239,
		"WE8DG":            241,
		"WE8NCR4970":       251,
		"WE8ROMAN8":        261,
		"EE8MACCE":         262,
		"EE8MACCROATIAN":   263,
		"TR8MACTURKISH":    264,
		"IS8MACICELANDIC":  265,
		"EL8MACGREEK":      266,
		"IW8MACHEBREW":     267,
		"US8ICL":           277,
		"WE8ICL":           278,
		"WE8ISOICLUK":      279,
		"EE8EBCDIC870C":    301,
		"EL8EBCDIC875S":    311,
		"TR8EBCDIC1026S":   312,
		"BLT8EBCDIC1112S":  314,
		"IW8EBCDIC424S":    315,
		"EE8EBCDIC870S":    316,
		"CL8EBCDIC1025S":   317,
		"TH8TISEBCDICS":    319,
		"AR8EBCDIC420S":    320,
		"CL8EBCDIC1025C":   322,
		"CL8EBCDIC1025R":   323,
		"EL8EBCDIC875R":    324,
		"CL8EBCDIC1158":    325,
		"CL8EBCDIC1158R":   326,
		"EL8EBCDIC423R":    327,
		"WE8MACROMAN8":     351,
		"WE8MACROMAN8S":    352,
		"TH8MACTHAI":       353,
		"TH8MACTHAIS":      354,
		"HU8CWI2":          368,
		"EL8PC437S":        380,
		"EL8EBCDIC875":     381,
		"EL8PC737":         382,
		"LT8PC772":         383,
		"LT8PC774":         384,
		"EL8PC869":         385,
		"EL8PC851":         386,
		"CDN8PC863":        390,
		"HU8ABMOD":         401,
		"AR8ASMO8X":        500,
		"AR8NAFITHA711T":   504,
		"AR8SAKHR707T":     505,
		"AR8MUSSAD768T":    506,
		"AR8ADOS710T":      507,
		"AR8ADOS720T":      508,
		"AR8APTEC715T":     509,
		"AR8NAFITHA721T":   511,
		"AR8HPARABIC8T":    514,
		"AR8NAFITHA711":    554,
		"AR8SAKHR707":      555,
		"AR8MUSSAD768":     556,
		"AR8ADOS710":       557,
		"AR8ADOS720":       558,
		"AR8APTEC715":      559,
		"AR8MSWIN1256":     560,
		"AR8NAFITHA721":    561,
		"AR8SAKHR706":      563,
		"AR8ARABICMAC":     565,
		"AR8ARABICMACS":    566,
		"AR8ARABICMACT":    567,
		"LA8ISO6937":       590,
		"JA16VMS":          829,
		"JA16EUC":          830,
		"JA16EUCYEN":       831,
		"JA16SJIS":         832,
		//"JA16DBCS" : 833,
		//"JA16SJISYEN" : 834,
		//"JA16EBCDIC930" : 835,
		//"JA16MACSJIS" : 836,
		//"JA16EUCTILDE" : 837,
		//"JA16SJISTILDE" : 838,
		//"KO16KSC5601" : 840,
		//"KO16DBCS" : 842,
		//"KO16KSCCS" : 845,
		//"KO16MSWIN949" : 846,
		"ZHS16CGB231280":    850,
		"ZHS16MACCGB231280": 851,
		"ZHS16GBK":          852,
		//"ZHS16DBCS" : 853,
		//"ZHS32GB18030" : 854,
		//"ZHT32EUC" : 860,
		//"ZHT32SOPS" : 861,
		//"ZHT16DBT" : 862,
		//"ZHT32TRIS" : 863,
		//"ZHT16DBCS" : 864,
		//"ZHT16BIG5" : 865,
		//"ZHT16CCDC" : 866,
		//"ZHT16MSWIN950" : 867,
		//"ZHT16HKSCS" : 868,
		"AL24UTFFSS": 870,
		"UTF8":       871,
		"UTFE":       872,
		"AL32UTF8":   873,
		//"ZHT16HKSCS31" : 992,
		//"JA16EUCFIXED" : 1830,
		//"JA16SJISFIXED" : 1832,
		//"JA16DBCSFIXED" : 1833,
		//"KO16KSC5601FIXED" : 1840,
		//"KO16DBCSFIXED" : 1842,
		//"ZHS16CGB231280FIXED" : 1850,
		//"ZHS16GBKFIXED" : 1852,
		//"ZHS16DBCSFIXED" : 1853,
		//"ZHT32EUCFIXED" : 1860,
		//"ZHT32TRISFIXED" : 1863,
		//"ZHT16DBCSFIXED" : 1864,
		//"ZHT16BIG5FIXED" : 1865,
		"AL16UTF16": 2000,
	}
	id, found := charsetMap[strings.ToUpper(charset)]
	if !found {
		return 0, fmt.Errorf("charset %s is not supported by the driver", charset)
	}
	return id, nil
}

//type EnList int

//const (
//	FALSE   EnList = 0
//	TRUE    EnList = 1
//	DYNAMIC EnList = 2
//)

//func EnListFromString(s string) EnList {
//	S := strings.ToUpper(s)
//	if S == "TRUE" {
//		return TRUE
//	} else if S == "DYNAMIC" {
//		return DYNAMIC
//	} else {
//		return FALSE
//	}
//}

type ConnectionString struct {
	connOption   network.ConnectionOption
	DataSource   string
	Host         string
	Port         int
	DBAPrivilege DBAPrivilege
	password     string
	Trace        string // Trace file
	WalletPath   string
	walletPass   string
	w            *wallet
	authType     AuthType
	//EnList             EnList
	//ConnectionLifeTime int
	//IncrPoolSize       int
	//DecrPoolSize       int
	//MaxPoolSize        int
	//MinPoolSize        int

	//PasswordSecurityInfo  bool
	//Pooling               bool

	//PromotableTransaction PromotableTransaction
	//ProxyUserID           string
	//ProxyPassword         string
	//ValidateConnection    bool
	//StmtCacheSize         int
	//StmtCachePurge        bool
	//HaEvent               bool
	//LoadBalance           bool
	//MetadataBooling       bool
	//ContextConnection     bool
	//SelfTuning            bool
	//ApplicationEdition    string
	//PoolRegulator         int
	//ConnectionPoolTimeout int

}

// BuildJDBC create url from user, password and JDBC description string
func BuildJDBC(user, password, connStr string, options map[string]string) string {
	if options == nil {
		options = make(map[string]string)
	}
	options["connStr"] = connStr
	return BuildUrl("", 0, "", user, password, options)
}

// BuildUrl create databaseURL from server, port, service, user, password, urlOptions
// this function help build a will formed databaseURL and accept any character as it
// convert special charters to corresponding values in URL
func BuildUrl(server string, port int, service, user, password string, options map[string]string) string {
	ret := fmt.Sprintf("oracle://%s:%s@%s/%s", url.PathEscape(user), url.PathEscape(password),
		net.JoinHostPort(server, strconv.Itoa(port)), url.PathEscape(service))
	if options != nil {
		ret += "?"
		for key, val := range options {
			val = strings.TrimSpace(val)
			for _, temp := range strings.Split(val, ",") {
				temp = strings.TrimSpace(temp)
				if strings.ToUpper(key) == "SERVER" {
					ret += fmt.Sprintf("%s=%s&", key, temp)
				} else {
					ret += fmt.Sprintf("%s=%s&", key, url.QueryEscape(temp))
				}
			}
		}
		ret = strings.TrimRight(ret, "&")
	}
	return ret
}

// newConnectionStringFromUrl create new connection string from databaseURL data and options
func newConnectionStringFromUrl(databaseUrl string) (*ConnectionString, error) {
	u, err := url.Parse(databaseUrl)

	if err != nil {
		return nil, err
	}
	q := u.Query()
	//p := u.Port()
	ret := &ConnectionString{
		connOption: network.ConnectionOption{
			PrefetchRows: 25,
			SessionInfo: network.SessionInfo{
				Timeout: time.Duration(15),
				//TransportDataUnitSize: 0xFFFF,
				//SessionDataUnitSize:   0xFFFF,
				TransportDataUnitSize: 0x200000,
				SessionDataUnitSize:   0x200000,
				Protocol:              "tcp",
				SSL:                   false,
				SSLVerify:             true,
			},
			DatabaseInfo: network.DatabaseInfo{
				Servers: make([]network.ServerAddr, 0, 3),
			},
			ClientInfo: network.ClientInfo{Territory: "AMERICA", Language: "AMERICAN"},
		},
		Port:         defaultPort,
		DBAPrivilege: NONE,
		//EnList:                TRUE,
		//IncrPoolSize:          5,
		//DecrPoolSize:          5,
		//MaxPoolSize:           100,
		//MinPoolSize:           1,
		//PromotableTransaction: Promotable,
		//StmtCacheSize:         20,
		//MetadataBooling:       true,
		//SelfTuning:            true,
		//PoolRegulator:         100,
		//ConnectionPoolTimeout: 15,
	}
	ret.connOption.UserID = u.User.Username()
	ret.password, _ = u.User.Password()
	if strings.ToUpper(ret.connOption.UserID) == "SYS" {
		ret.DBAPrivilege = SYSDBA
	}

	host, p, err := net.SplitHostPort(u.Host)
	if err != nil {
		return nil, err
	}
	if len(host) > 0 {
		tempAddr := network.ServerAddr{Addr: host, Port: defaultPort}
		tempAddr.Port, err = strconv.Atoi(p)
		if err != nil {
			tempAddr.Port = defaultPort
		}
		ret.connOption.Servers = append(ret.connOption.Servers, tempAddr)
	}
	ret.connOption.ServiceName = strings.Trim(u.Path, "/")
	if q != nil {
		for key, val := range q {
			switch strings.ToUpper(key) {
			case "CONNSTR":
				err = ret.connOption.UpdateDatabaseInfo(q.Get("connStr"))
				if err != nil {
					return nil, err
				}
			case "SERVER":
				for _, srv := range val {
					srv = strings.TrimSpace(srv)
					if srv != "" {
						host, p, err := net.SplitHostPort(srv)
						if err != nil {
							return nil, err
						}
						tempAddr := network.ServerAddr{Addr: host, Port: defaultPort}
						if p != "" {
							tempAddr.Port, err = strconv.Atoi(p)
							if err != nil {
								tempAddr.Port = defaultPort
							}
						}
						ret.connOption.Servers = append(ret.connOption.Servers, tempAddr)
					}
				}
			case "SERVICE NAME":
				ret.connOption.ServiceName = val[0]
			case "SID":
				ret.connOption.SID = val[0]
			case "INSTANCE NAME":
				ret.connOption.InstanceName = val[0]
			case "WALLET":
				ret.WalletPath = val[0]
			case "WALLET PASSWORD":
				ret.walletPass = val[0]
			case "AUTH TYPE":
				if strings.ToUpper(val[0]) == "OS" {
					ret.authType = OS
				} else if strings.ToUpper(val[0]) == "KERBEROS" {
					ret.authType = Kerberos
				} else if strings.ToUpper(val[0]) == "TCPS" {
					ret.authType = TCPS
				} else {
					ret.authType = Normal
				}
			case "OS USER":
				ret.connOption.ClientInfo.UserName = val[0]
			case "OS PASS":
				fallthrough
			case "OS PASSWORD":
				ret.connOption.ClientInfo.Password = val[0]
			case "OS HASH":
				fallthrough
			case "OS PASSHASH":
			case "OS PASSWORD HASH":
				ret.connOption.ClientInfo.Password = val[0]
				SetNTSAuth(&advanced_nego.NTSAuthHash{})
			case "DOMAIN":
				ret.connOption.DomainName = val[0]
			case "AUTH SERV":
				for _, tempVal := range val {
					ret.connOption.AuthService, _ = uniqueAppendString(ret.connOption.AuthService, strings.ToUpper(strings.TrimSpace(tempVal)), false)
				}
			case "SSL":
				ret.connOption.SSL = strings.ToUpper(val[0]) == "TRUE" || strings.ToUpper(val[0]) == "ENABLE"
			case "SSL VERIFY":
				ret.connOption.SSLVerify = strings.ToUpper(val[0]) == "TRUE" || strings.ToUpper(val[0]) == "ENABLE"
			case "DBA PRIVILEGE":
				ret.DBAPrivilege = DBAPrivilegeFromString(val[0])
			case "CONNECT TIMEOUT":
				fallthrough
			case "CONNECTION TIMEOUT":
				to, err := strconv.Atoi(val[0])
				if err != nil {
					return nil, errors.New("CONNECTION TIMEOUT value must be an integer")
				}
				ret.connOption.SessionInfo.Timeout = time.Duration(to)
			case "TRACE FILE":
				ret.Trace = val[0]
			case "PREFETCH_ROWS":
				ret.connOption.PrefetchRows, err = strconv.Atoi(val[0])
				if err != nil {
					ret.connOption.PrefetchRows = 25
				}
			case "UNIX SOCKET":
				ret.connOption.SessionInfo.UnixAddress = val[0]
			case "PROXY CLIENT NAME":
				ret.connOption.DatabaseInfo.ProxyClientName = val[0]
			case "FAILOVER":
				ret.connOption.Failover, err = strconv.Atoi(val[0])
				if err != nil {
					ret.connOption.Failover = 0
				}
			case "RETRYTIME":
				fallthrough
			case "RE-TRY TIME":
				fallthrough
			case "RETRY TIME":
				ret.connOption.RetryTime, err = strconv.Atoi(val[0])
				if err != nil {
					ret.connOption.RetryTime = 0
				}
			case "LOB FETCH":
				tempVal := strings.ToUpper(val[0])
				if tempVal == "PRE" {
					ret.connOption.Lob = 0
				} else if tempVal == "POST" {
					ret.connOption.Lob = 1
				} else {
					return nil, errors.New("LOB FETCH value should be: PRE(default) or POST")
				}
			case "LANGUAGE":
				ret.connOption.Language = val[0]
			case "TERRITORY":
				ret.connOption.Territory = val[0]
			case "CHARSET":
				fallthrough
			case "CLIENT CHARSET":
				ret.connOption.CharsetID, err = getCharsetID(val[0])
				if err != nil {
					return nil, err
				}
				//else if tempVal == "IMPLICIT" || tempVal == "AUTO" {
				//	ret.connOption.Lob = 1
				//} else if tempVal == "EXPLICIT" || tempVal == "MANUAL" {
				//	ret.connOption.Lob = 2
				//} else {
				//	return nil, errors.New("LOB value should be: Prefetch, Implicit(AUTO) or Explicit(manual)")
				//}
				//case "ENLIST":
				//	ret.EnList = EnListFromString(val[0])
				//case "INC POOL SIZE":
				//	ret.IncrPoolSize, err = strconv.Atoi(val[0])
				//	if err != nil {
				//		return nil, errors.New("INC POOL SIZE value must be an integer")
				//	}
				//case "DECR POOL SIZE":
				//	ret.DecrPoolSize, err = strconv.Atoi(val[0])
				//	if err != nil {
				//		return nil, errors.New("DECR POOL SIZE value must be an integer")
				//	}
				//case "MAX POOL SIZE":
				//	ret.MaxPoolSize, err = strconv.Atoi(val[0])
				//	if err != nil {
				//		return nil, errors.New("MAX POOL SIZE value must be an integer")
				//	}
				//case "MIN POOL SIZE":
				//	ret.MinPoolSize, err = strconv.Atoi(val[0])
				//	if err != nil {
				//		return nil, errors.New("MIN POOL SIZE value must be an integer")
				//	}
				//case "POOL REGULATOR":
				//	ret.PoolRegulator, err = strconv.Atoi(val[0])
				//	if err != nil {
				//		return nil, errors.New("POOL REGULATOR value must be an integer")
				//	}
				//case "STATEMENT CACHE SIZE":
				//	ret.StmtCacheSize, err = strconv.Atoi(val[0])
				//	if err != nil {
				//		return nil, errors.New("STATEMENT CACHE SIZE value must be an integer")
				//	}
				//case "CONNECTION POOL TIMEOUT":
				//	ret.ConnectionPoolTimeout, err = strconv.Atoi(val[0])
				//	if err != nil {
				//		return nil, errors.New("CONNECTION POOL TIMEOUT value must be an integer")
				//	}
				//case "CONNECTION LIFETIME":
				//	ret.ConnectionLifeTime, err = strconv.Atoi(val[0])
				//	if err != nil {
				//		return nil, errors.New("CONNECTION LIFETIME value must be an integer")
				//	}
				//case "PERSIST SECURITY INFO":
				//	ret.PasswordSecurityInfo = val[0] == "TRUE"
				//case "POOLING":
				//	ret.Pooling = val[0] == "TRUE"
				//case "VALIDATE CONNECTION":
				//	ret.ValidateConnection = val[0] == "TRUE"
				//case "STATEMENT CACHE PURGE":
				//	ret.StmtCachePurge = val[0] == "TRUE"
				//case "HA EVENTS":
				//	ret.HaEvent = val[0] == "TRUE"
				//case "LOAD BALANCING":
				//	ret.LoadBalance = val[0] == "TRUE"
				//case "METADATA POOLING":
				//	ret.MetadataBooling = val[0] == "TRUE"
				//case "SELF TUNING":
				//	ret.SelfTuning = val[0] == "TRUE"
				//case "CONTEXT CONNECTION":
				//	ret.ContextConnection = val[0] == "TRUE"
				//case "PROMOTABLE TRANSACTION":
				//	if val[0] == "PROMOTABLE" {
				//		ret.PromotableTransaction = Promotable
				//	} else {
				//		ret.PromotableTransaction = Local
				//	}
				//case "APPLICATION EDITION":
				//	ret.ApplicationEdition = val[0]
				//case "PROXY USER ID":
				//	ret.ProxyUserID = val[0]
				//case "PROXY PASSWORD":
				//	ret.ProxyPassword = val[0]
			}
		}
	}
	if len(ret.connOption.Servers) == 0 {
		return nil, errors.New("empty connection servers")
	}
	if len(ret.WalletPath) > 0 {
		if len(ret.connOption.ServiceName) == 0 {
			return nil, errors.New("you should specify server/service if you will use wallet")
		}
		if _, err = os.Stat(path.Join(ret.WalletPath, "ewallet.p12")); err == nil && len(ret.walletPass) > 0 {
			fileData, err := ioutil.ReadFile(path.Join(ret.WalletPath, "ewallet.p12"))
			if err != nil {
				return nil, err
			}
			ret.w = &wallet{password: []byte(ret.walletPass)}
			err = ret.w.readPKCS12(fileData)
			if err != nil {
				return nil, err
			}
		} else {
			ret.w, err = NewWallet(path.Join(ret.WalletPath, "cwallet.sso"))
			if err != nil {
				return nil, err
			}
		}

		if len(ret.connOption.UserID) > 0 {
			if len(ret.password) == 0 {
				serv := ret.connOption.Servers[0]
				cred, err := ret.w.getCredential(serv.Addr, serv.Port, ret.connOption.ServiceName, ret.connOption.UserID)
				if err != nil {
					return nil, err
				}
				if cred == nil {
					return nil, errors.New(
						fmt.Sprintf("cannot find credentials for server: %s:%d, service: %s,  username: %s",
							serv.Addr, serv.Port, ret.connOption.ServiceName, ret.connOption.UserID))
				}
				ret.connOption.UserID = cred.username
				ret.password = cred.password
			}
		}
	}
	return ret, ret.validate()
}

// validate check is data in connection string is correct and fulfilled
func (connStr *ConnectionString) validate() error {
	//if !connStr.Pooling {
	//	connStr.MaxPoolSize = -1
	//	connStr.MinPoolSize = 0
	//	connStr.IncrPoolSize = -1
	//	connStr.DecrPoolSize = 0
	//	connStr.PoolRegulator = 0
	//}

	//if len(connStr.UserID) == 0 {
	//	return errors.New("empty user name")
	//}
	//if len(connStr.Password) == 0 {
	//	return errors.New("empty password")
	//}
	if len(connStr.connOption.SID) == 0 && len(connStr.connOption.ServiceName) == 0 {
		return errors.New("empty SID and service name")
	}
	if connStr.authType == Kerberos {
		connStr.connOption.AuthService = append(connStr.connOption.AuthService, "KERBEROS5")
	}
	if connStr.authType == TCPS {
		connStr.connOption.AuthService = append(connStr.connOption.AuthService, "TCPS")
	}
	if len(connStr.connOption.UserID) == 0 || len(connStr.password) == 0 && connStr.authType == Normal {
		connStr.authType = OS
	}
	if connStr.authType == OS {
		if runtime.GOOS == "windows" {
			connStr.connOption.AuthService = append(connStr.connOption.AuthService, "NTS")
		}
	}

	if connStr.connOption.SSL {
		connStr.connOption.Protocol = "tcps"
	}
	if len(connStr.Trace) > 0 {
		tf, err := os.Create(connStr.Trace)
		if err != nil {
			//noinspection GoErrorStringFormat
			return fmt.Errorf("Can't open trace file: %w", err)
		}
		connStr.connOption.Tracer = trace.NewTraceWriter(tf)
	} else {
		connStr.connOption.Tracer = trace.NilTracer()
	}

	// get client info
	var idx int
	var temp = getCurrentUser()

	if temp != nil {
		idx = strings.Index(temp.Username, "\\")
		if idx >= 0 {
			if len(connStr.connOption.DomainName) == 0 {
				connStr.connOption.DomainName = temp.Username[:idx]
			}
			if len(connStr.connOption.ClientInfo.UserName) == 0 {
				connStr.connOption.ClientInfo.UserName = temp.Username[idx+1:]
			}
		} else {
			if len(connStr.connOption.ClientInfo.UserName) == 0 {
				connStr.connOption.ClientInfo.UserName = temp.Username
			}
		}
	}
	connStr.connOption.HostName, _ = os.Hostname()
	idx = strings.LastIndex(os.Args[0], "/")
	idx++
	if idx < 0 {
		idx = 0
	}
	connStr.connOption.ProgramPath = os.Args[0]
	connStr.connOption.ProgramName = os.Args[0][idx:]
	connStr.connOption.DriverName = "OracleClientGo"
	connStr.connOption.PID = os.Getpid()
	return nil
}

func uniqueAppendString(list []string, newItem string, ignoreCase bool) ([]string, bool) {
	found := false
	for _, temp := range list {
		if ignoreCase {
			if strings.ToUpper(temp) == strings.ToUpper(newItem) {
				found = true
				break
			}
		} else {
			if temp == newItem {
				found = true
				break
			}
		}
	}
	if !found {
		list = append(list, newItem)
	}
	return list, !found
}

func getCurrentUser() *user.User {
	if userName := os.Getenv("USER"); len(userName) > 0 {
		return &user.User{
			Uid:      "",
			Gid:      "",
			Username: userName,
			Name:     userName,
			HomeDir:  "",
		}
	} else {
		temp, _ := user.Current()
		return temp
	}
}
