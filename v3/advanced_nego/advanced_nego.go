package advanced_nego

import (
	"encoding/binary"
	"errors"
	"fmt"
	"net"

	"github.com/sijms/go-ora/v3/configurations"
	"github.com/sijms/go-ora/v3/network"
	"github.com/sijms/go-ora/v3/trace"
)

const version = uint32(0x17000000)

// KerberosAuthInterface is an alias for configurations.KerberosAuthInterface, maintained for backwards compatibility.
type KerberosAuthInterface = configurations.KerberosAuthInterface

var kerberosAuth KerberosAuthInterface = nil

// SetKerberosAuth Set Kerberos5 Authentication interface used for kerberos authentication
func SetKerberosAuth(input KerberosAuthInterface) {
	kerberosAuth = input
}

type AdvNego struct {
	version      uint32
	authKerberos bool
	authNTS      bool
	comm         *AdvancedNegoComm
	clientInfo   *configurations.ClientInfo
	negoInfo     *configurations.AdvNegoServiceInfo
	tracer       trace.Tracer
	serviceList  []AdvNegoService
}

func NewAdvNego(session *network.Session, tracer trace.Tracer, config *configurations.ConnectionConfig) (*AdvNego, error) {
	output := &AdvNego{
		version:     version,
		comm:        &AdvancedNegoComm{session: session},
		clientInfo:  &config.ClientInfo,
		negoInfo:    &config.AdvNegoServiceInfo,
		tracer:      tracer,
		serviceList: make([]AdvNegoService, 5),
	}
	var err error
	output.serviceList[1], err = newAuthService(output.comm, output.negoInfo)
	if err != nil {
		return nil, err
	}
	output.serviceList[2], err = newEncryptService(output.comm, output.negoInfo)
	if err != nil {
		return nil, err
	}
	output.serviceList[3], err = newDataIntegrityService(output.comm, output.negoInfo, output.tracer)
	if err != nil {
		return nil, err
	}
	output.serviceList[4], err = newSupervisorService(output.comm)
	if err != nil {
		return nil, err
	}
	return output, nil
}

func (nego *AdvNego) readHeader() (length, version, servCount int, err error) {
	var num int64
	num, err = nego.comm.session.GetInt64(4, false, true)
	if err != nil {
		return
	}
	if num != 0xDEADBEEF {
		err = errors.New("advanced negotiation error: during receive header")
		return
	}
	length, err = nego.comm.session.GetInt(2, false, true)
	if err != nil {
		return
	}
	version, err = nego.comm.session.GetInt(4, false, true)
	if err != nil {
		return
	}
	servCount, err = nego.comm.session.GetInt(2, false, true)
	if err != nil {
		return
	}
	var errFlags int
	errFlags, err = nego.comm.session.GetInt(1, false, true)
	if err != nil {
		return
	}
	if errFlags != 0 {
		err = network.NewOracleError(errFlags)
		//err = fmt.Errorf("advanced negotiation error: during receive ano header: network exception: ora-%d", errFlags)
	}
	return
}

func (nego *AdvNego) writeHeader(length, servCount int) {
	nego.comm.session.PutInt(uint64(0xDEADBEEF), 4, true, false)
	nego.comm.session.PutInt(length, 2, true, false)
	nego.comm.session.PutInt(nego.version, 4, true, false)
	nego.comm.session.PutInt(servCount, 2, true, false)
	var errFlags = uint8(0)
	nego.comm.session.PutBytes(errFlags)
}

func (nego *AdvNego) readServiceHeader() (serviceType, serviceSubPacket int, err error) {
	serviceType, err = nego.comm.session.GetInt(2, false, true)
	if err != nil {
		return
	}
	serviceSubPacket, err = nego.comm.session.GetInt(2, false, true)
	if err != nil {
		return
	}
	var errFlags int
	errFlags, err = nego.comm.session.GetInt(4, false, true)
	if err != nil {
		return
	}
	if errFlags != 0 {
		err = network.NewOracleError(errFlags)
	}
	return
}

func (nego *AdvNego) IsNTSAuth() bool {
	return nego.authNTS
}

func (nego *AdvNego) IsKerberosAuth() bool {
	return nego.authKerberos
}

func (nego *AdvNego) Read() error {
	_, version, count, err := nego.readHeader()
	if err != nil {
		return err
	}
	nego.version = uint32(version)
	for i := 0; i < count; i++ {
		serviceType, subPacket, err := nego.readServiceHeader()
		if err != nil {
			return err
		}
		err = nego.serviceList[serviceType].readServiceData(subPacket)
		if err != nil {
			return err
		}
		err = nego.serviceList[serviceType].validateResponse()
		if err != nil {
			return err
		}
	}
	nego.authKerberos = false
	nego.authNTS = false
	if authServ, ok := nego.serviceList[1].(*authService); ok {
		if authServ.active {
			if authServ.serviceName == "KERBEROS5" {
				// return errors.New("advanced negotiation: KERBEROS5 authentication still not supported")
				nego.authKerberos = true
			} else if authServ.serviceName == "NTS" {
				nego.authNTS = true
			}
		}
	}
	size := 0
	numService := 0
	if dataServ, ok := nego.serviceList[3].(*dataIntegrityService); ok {
		if len(dataServ.publicKey) > 0 {
			size = size + 12 + len(dataServ.publicKey)
			numService++
		}
	}
	if nego.authKerberos {
		size += 37
		numService++
	}
	if nego.authNTS {
		size += 130
		numService++
	}
	if numService == 0 {
		return nil
	}
	nego.comm.session.ResetBuffer()
	nego.writeHeader(size+13, numService)
	if dataServ, ok := nego.serviceList[3].(*dataIntegrityService); ok {
		if len(dataServ.publicKey) > 0 {
			nego.tracer.Print("Send Client Public Key:")
			dataServ.writeHeader(1)
			nego.comm.writeBytes(dataServ.publicKey)
		}
	}
	if nego.authKerberos {
		// Validate configuration
		if kerberosAuth == nil && nego.negoInfo.Kerberos == nil {
			return fmt.Errorf("advanced negotiation error: Kerberos authenticator not set; call SetKerberosAuth to set it globally or WithKerberosAuth to set it per session")
		}

		// Prefer session-specific Kerberos auth object
		auth := nego.negoInfo.Kerberos
		if auth == nil {
			auth = kerberosAuth
		}

		if authServ, ok := nego.serviceList[1].(*authService); ok {
			authServ.writeHeader(4)
			nego.comm.writeVersion(authServ.getVersion())
			nego.comm.writeUB4(9)
			nego.comm.writeUB4(2)
			nego.comm.writeUB1(1)
			err = nego.comm.session.Write()
			if err != nil {
				return err
			}
			return nego.kerberosHandshake(auth, authServ)
		}
	}
	if nego.authNTS {
		ntsPacket, err := createNTSNegoPacket(nego.clientInfo.DomainName, nego.clientInfo.HostName)
		if err != nil {
			return err
		}
		nego.comm.session.ResetBuffer()
		nego.comm.session.PutBytes(ntsPacket...)
		err = nego.comm.session.Write()
		if err != nil {
			return err
		}
		ntsHeader, err := nego.comm.session.GetBytes(33)
		if err != nil {
			return err
		}
		sizeOffset := len(ntsHeader) - 8
		chaSize := binary.LittleEndian.Uint32(ntsHeader[sizeOffset : sizeOffset+4])
		chaData, err := nego.comm.session.GetBytes(int(chaSize))
		if err != nil {
			return err
		}
		ntsPacket, err = createNTSAuthPacket(chaData, nego.clientInfo.OSUserName,
			nego.clientInfo.OSPassword)
		if err != nil {
			return err
		}
		nego.comm.session.ResetBuffer()
		nego.comm.session.PutBytes(ntsPacket...)
		err = nego.comm.session.Write()
		if err != nil {
			return err
		}
		// fmt.Println(nego.comm.session.GetBytes(10))
		// return errors.New("interrupt")
		return nil
	}
	return nego.comm.session.Write()
}

func (nego *AdvNego) Write() error {
	nego.comm.session.ResetBuffer()
	size := 0
	for i := 1; i < 5; i++ {
		size = size + 8 + nego.serviceList[i].getServiceDataLength()
	}
	// size += 13
	nego.writeHeader(13+size, 4)
	err := nego.serviceList[4].writeServiceData()
	if err != nil {
		return err
	}
	err = nego.serviceList[1].writeServiceData()
	if err != nil {
		return err
	}
	err = nego.serviceList[2].writeServiceData()
	if err != nil {
		return err
	}
	err = nego.serviceList[3].writeServiceData()
	if err != nil {
		return err
	}
	return nego.comm.session.Write()
}

func (nego *AdvNego) StartServices() error {
	err := nego.serviceList[3].activateAlgorithm()
	if err != nil {
		return err
	}
	err = nego.serviceList[2].activateAlgorithm()
	if err != nil {
		return err
	}
	err = nego.serviceList[1].activateAlgorithm()
	if err != nil {
		return err
	}
	err = nego.serviceList[4].activateAlgorithm()
	if err != nil {
		return err
	}
	return nil
}

func (nego *AdvNego) IsNewVersion(serviceType int) bool {
	if serviceType < len(nego.serviceList) {
		return isNew(nego.serviceList[serviceType].getVersion())
	}

	return isNew(nego.version)
}

func (nego *AdvNego) kerberosHandshake(kerberos KerberosAuthInterface, authServ *authService) error {
	_, _, count, err := nego.readHeader()
	if err != nil {
		return err
	}
	for i := 0; i < count; i++ {
		_, _, err = nego.readServiceHeader()
		if err != nil {
			return err
		}
	}
	serviceName, err := nego.comm.readString()
	if err != nil {
		return err
	}
	serverHostName, err := nego.comm.readString()
	if err != nil {
		return err
	}
	if len(serviceName) == 0 {
		return errors.New("kerberos negotiation error: Service Name not received")
	}
	if len(serverHostName) == 0 {
		return errors.New("kerberos negotiation error: Server hostname not received")
	}
	ticketData, err := kerberos.Authenticate(serverHostName, serviceName)
	if err != nil {
		return err
	}
	// get host ip address
	localAddress, err := getHostIPAddress()
	if err != nil {
		return err
	}
	// if address is ipv6 then num1 = 24 otherwise = 2
	num1 := 2
	//localAddress = net.IP{172, 17, 0, 2}
	if len(localAddress) > 4 {
		num1 = 24
	}
	nego.comm.session.ResetBuffer()
	// send ano header(length of ticket + 43 + length of address, 1 , 0)
	nego.writeHeader(len(ticketData)+43+len(localAddress), 1)
	// send header(4)
	authServ.writeHeader(4)
	// send ub2 = num1
	nego.comm.writeUB2(num1)
	// send ub4 = length of address
	nego.comm.writeUB4(len(localAddress))
	// send bytes address bytes
	nego.comm.writeBytes(localAddress)
	// send bytes ticket
	nego.comm.writeBytes(ticketData)
	// write
	err = nego.comm.session.Write()
	if err != nil {
		return err
	}
	// read ano header
	_, _, count, err = nego.readHeader()
	if err != nil {
		return err
	}
	for index := 0; index < count; index++ {
		_, _, err := nego.readServiceHeader()
		if err != nil {
			return err
		}
	}
	// get packet header (2)
	_, err = nego.comm.readPacketHeader(2)
	if err != nil {
		return err
	}
	// num2 = get ub1
	_, err = nego.comm.session.GetByte()
	if err != nil {
		return err
	}
	// receive byte array
	_, err = nego.comm.readBytes()
	if err != nil {
		return err
	}
	// send ano header (25,1, 0)
	nego.comm.session.ResetBuffer()
	nego.writeHeader(25, 1)
	// as.send header(1)
	authServ.writeHeader(1)
	// send packet header(0, 1)
	nego.comm.writePacketHeader(0, 1)
	// write
	return nego.comm.session.Write()
}

func getHostIPAddress() (net.IP, error) {
	adders, err := net.InterfaceAddrs()
	if err != nil {
		return nil, err
	}
	for _, address := range adders {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.To4(), nil
			}
			if ipnet.IP.To16() != nil {
				return ipnet.IP.To16(), nil
			}
		}
	}
	return nil, errors.New("advanced negotiation error: during get local ip address")
}
