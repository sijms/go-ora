package advanced_nego

import (
	"errors"
	"fmt"
	"github.com/sijms/go-ora/network"
)

type encryptService struct {
	defaultService
	algoID int
}

func NewEncryptService(connOption *network.ConnectionOption) (*encryptService, error) {
	output := &encryptService{
		defaultService: defaultService{
			serviceType: 2,
			version:     0xB200200,
			availableServiceNames: []string{"", "RC4_40", "RC4_56", "RC4_128", "RC4_256",
				"DES40C", "DES56C", "3DES112", "3DES168", "AES128", "AES192", "AES256"},
			availableServiceIDs: []int{0, 1, 8, 10, 6, 3, 2, 11, 12, 15, 16, 17},
		},
	}
	str := ""
	level := ""
	if connOption != nil {
		snConfig := connOption.SNOConfig
		if snConfig != nil {
			var exists bool
			str, exists = snConfig["sqlnet.encryption_types_client"]
			if !exists {
				str = ""
			}
			level, exists = snConfig["sqlnet.encryption_client"]
			if !exists {
				level = ""
			}
		}
	}
	output.readAdvNegoLevel(level)
	//level := conops.Encryption != null ? conops.Encryption : snoConfig[];
	err := output.buildServiceList(str, true, true)
	//output.selectedServ, err = output.validate(strings.Split(str,","), true)
	if err != nil {
		return nil, err
	}
	return output, nil
}

func (serv *encryptService) readServiceData(session *network.Session, subPacketnum int) error {
	var err error
	serv.version, err = serv.readVersion(session)
	if err != nil {
		return err
	}
	_, err = serv.readPacketHeader(session, 2)
	if err != nil {
		return err
	}
	resp, err := session.GetByte()
	if err != nil {
		return err
	}
	serv.algoID = int(resp)
	return nil
}
func (serv *encryptService) writeServiceData(session *network.Session) error {
	serv.writeHeader(session, 3)
	err := serv.writeVersion(session)
	if err != nil {
		return err
	}
	err = serv.writePacketHeader(session, len(serv.selectedIndices), 1)
	if err != nil {
		return err
	}
	for i := 0; i < len(serv.selectedIndices); i++ {
		index := serv.selectedIndices[i]
		session.PutBytes(uint8(serv.availableServiceIDs[index]))
	}
	// send selected driver
	err = serv.writePacketHeader(session, 1, 2)
	if err != nil {
		return err
	}
	session.PutBytes(1)
	return nil
}

func (serv *encryptService) getServiceDataLength() int {
	return 17 + len(serv.selectedIndices)
}

func (serv *encryptService) activateAlgorithm() error {
	if serv.algoID == 0 {
		return nil
	} else {
		return errors.New(fmt.Sprintf("advanced negotiation error: encryption service algorithm: %d still not supported", serv.algoID))
	}
	//switch (this.m_algID)
	//{
	//case 1:
	//	this.m_sessCtx.encryptionAlg = (EncryptionAlgorithm) new RC4(true, 40);
	//	break;
	//case 6:
	//	this.m_sessCtx.encryptionAlg = (EncryptionAlgorithm) new RC4(true, 256);
	//	break;
	//case 8:
	//	this.m_sessCtx.encryptionAlg = (EncryptionAlgorithm) new RC4(true, 56);
	//	break;
	//case 10:
	//	this.m_sessCtx.encryptionAlg = (EncryptionAlgorithm) new RC4(true, 128);
	//	break;
	//case 11:
	//	this.m_sessCtx.encryptionAlg = (EncryptionAlgorithm) new DES112();
	//	break;
	//case 12:
	//	this.m_sessCtx.encryptionAlg = (EncryptionAlgorithm) new DES168();
	//	break;
	//case 15:
	//	this.m_sessCtx.encryptionAlg = (EncryptionAlgorithm) new AES(1, 1);
	//	break;
	//case 16:
	//	this.m_sessCtx.encryptionAlg = (EncryptionAlgorithm) new AES(1, 2);
	//	break;
	//case 17:
	//	this.m_sessCtx.encryptionAlg = (EncryptionAlgorithm) new AES(1, 3);
	//	break;
	//}
	//this.m_sessCtx.encryptionAlg.init(ano.skey, ano.getInitializationVector());
}
