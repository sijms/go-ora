package go_ora

import (
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
)

//type PromotableTransaction int

//const (
//	Promotable PromotableTransaction = 1
//	Local      PromotableTransaction = 0
//)

//type DBAPrivilege int
//
//const (
//	NONE    DBAPrivilege = 0
//	SYSDBA  DBAPrivilege = 0x20
//	SYSOPER DBAPrivilege = 0x40
//)

//const defaultPort int = 1521

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

//type ConnectionString struct {
//connOption network.ConnectionOption
//DataSource string
//Host         string
//Port         int
//DBAPrivilege DBAPrivilege
//password string
//Trace        string // Trace file
//traceDir     string
//WalletPath string
//walletPass string
//w          *wallet
//authType     AuthType
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
//}

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
