package network

import (
	"github.com/sijms/go-ora/v3/configurations"
	"github.com/sijms/go-ora/v3/network/security"
)

type SessionContext struct {
	// conn net.Conn
	// ConnOption *ConnectionOption
	// PortNo int
	// InstanceName string
	// HostName string
	// IPAddress string
	// Protocol string
	// ServiceName string
	SID                 []byte
	connConfig          *configurations.ConnectionConfig
	Version             uint16
	LoVersion           uint16
	Options             uint16
	NegotiatedOptions   uint16
	NegotiatedOptions2  uint32
	OurOne              uint16
	hisOne              uint16
	ReconAddr           string
	handshakeComplete   bool
	ACFL0               uint8
	ACFL1               uint8
	SessionDataUnit     uint32
	TransportDataUnit   uint32
	UsingAsyncReceivers bool
	IsNTConnected       bool
	OnBreakReset        bool
	GotReset            bool
	AdvancedService     struct {
		CryptAlgo  security.OracleNetworkEncryption
		HashAlgo   security.OracleNetworkDataIntegrity
		SessionKey []byte
		IV         []byte
		keyFolding bool
	}
	isRedirect       bool
	FastAuthEnabled  bool
	EODDAFlagEnabled bool
	UUID             []byte
}

func NewSessionContext(config *configurations.ConnectionConfig) *SessionContext {
	ctx := &SessionContext{
		SessionDataUnit:   config.SessionDataUnitSize,
		TransportDataUnit: config.TransportDataUnitSize,
		Version:           319,
		LoVersion:         300,
		Options:           1 | 2048, /*1024 for urgent data transport*/
		OurOne:            1,
		connConfig:        config,
		// ConnOption:        connOption,
	}
	if config.EnableOOB {
		ctx.Options |= 1024
	}
	return ctx
}

type AdvancedNegotiation interface {
	StartServices() error
	IsNewVersion(serviceType int) bool
	IsNTSAuth() bool
	IsKerberosAuth() bool
}

func (ctx *SessionContext) SetKeyFolding(key []byte, ano AdvancedNegotiation, isExternalAuth, isSys bool) error {
	ctx.AdvancedService.keyFolding = false
	if isExternalAuth {
		if ano.IsNTSAuth() && (!isSys) {
			return nil
		}
		//if ano.IsKerberosAuth() && ano.IsNewVersion(1) && ano.kkey != null {
		//	Key = ano.kkey
		//}
	}
	if key == nil {
		return nil
	}
	if ano == nil {
		return nil
	}

	length := len(key)
	if length > len(ctx.AdvancedService.SessionKey) {
		length = len(ctx.AdvancedService.SessionKey)
	}
	for i := 0; i < length; i++ {
		ctx.AdvancedService.SessionKey[i] ^= key[i]
	}
	// service type: 2 = encryption, 3 = data integrity
	if ano.IsNewVersion(2) || ano.IsNewVersion(3) {
		length = len(key)
		if length > len(ctx.AdvancedService.IV) {
			length = len(ctx.AdvancedService.IV)
		}
		for i := 0; i < length; i++ {
			ctx.AdvancedService.IV[i] ^= key[i]
		}
	}
	if ctx.AdvancedService.CryptAlgo != nil || ctx.AdvancedService.HashAlgo != nil {
		err := ano.StartServices()
		if err != nil {
			return err
		}
		ctx.AdvancedService.keyFolding = true
	}
	return nil
}
