package configurations

import (
	"errors"
	"fmt"
	"os"

	"github.com/sijms/go-ora/v2/tns"
)

// getConnParamsFromTNS gets connection parameters from TNS names file at the specified path by net service name.
//
// Net service name is an alias used in TNS names file as a key of actual connection parameters.
func getConnParamsFromTNS(
	tnsNamesPath,
	netServiceName string,
) (resolvedServiceName string, resolvedServers []ServerAddr, err error) {
	// Load TNS names
	tnsNames, err := loadTNSNames(tnsNamesPath)
	if err != nil {
		return "", nil, fmt.Errorf("failed to load TNS names file '%s': %w", tnsNamesPath, err)
	}

	// Find a connect descriptor among TNS names
	connDesc := tnsNames.GetDescriptorForService(netServiceName)
	if connDesc == nil {
		// No connect descriptor has been found in TNS names file => exit
		return "", nil, nil
	}

	// Found a connection descriptor in TNS names => use it
	var desc *tns.Description
	switch {
	case len(connDesc.DescriptionList.Descriptions) > 1:
		// We don't support multiple 'DESCRIPTION'-s at the moment
		return "", nil, errors.New("multiple 'DESCRIPTION'-s in 'DESCRIPTION_LIST' are not supported")
	case len(connDesc.DescriptionList.Descriptions) == 1:
		desc = &connDesc.DescriptionList.Descriptions[0]
	default:
		desc = &connDesc.Description
	}

	// Validate TNS connect description
	err = validateTNSConnDescription(desc)
	if err != nil {
		return "", nil, fmt.Errorf("'DESCRIPTION' is incorrect: %w", err)
	}

	// Collect server addresses
	for _, addr := range desc.AddressList.Address {
		resolvedServers = append(resolvedServers, ServerAddr{
			Protocol: addr.Protocol,
			Addr:     addr.Host,
			Port:     addr.Port,
		})
	}
	for _, addrList := range desc.AddressLists {
		for _, addr := range addrList.Address {
			resolvedServers = append(resolvedServers, ServerAddr{
				Protocol: addr.Protocol,
				Addr:     addr.Host,
				Port:     addr.Port,
			})
		}
	}

	// All done!
	return desc.ConnectData.ServiceName, resolvedServers, nil
}

// loadTNSNames loads a TNS names file at specified path.
func loadTNSNames(path string) (*tns.TNS, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read TNS names file: %w", err)
	}
	// Parse TNS names
	tnsNames, err := tns.Parse(string(data))
	if err != nil {
		return nil, fmt.Errorf("failed to parse TNS names file: %w", err)
	}
	return tnsNames, nil
}

// validateTNSConnDescription validates the specified TNS connection description for unsupported features.
func validateTNSConnDescription(desc *tns.Description) error {
	if desc.RetryCount != 0 {
		return errors.New("'RETRY_COUNT' is not supported. Please remove it")
	}
	if desc.TransportConnectTimeout != 0 {
		return errors.New("'TRANSPORT_CONNECT_TIMEOUT' is not supported. Please remove it")
	}
	if desc.ConnectTimeout != 0 {
		return errors.New("'CONNECT_TIMEOUT' is not supported. Please remove it")
	}
	if desc.Compression != "" {
		return errors.New("'COMPRESSION' is not supported. Please remove it")
	}
	if len(desc.CompressionLevels) > 0 {
		return errors.New("'COMPRESSION_LEVELS' is not supported. Please remove it")
	}
	if desc.EncryptionClient != "" {
		return errors.New("'ENCRYPTION_CLIENT' is not supported. Please remove it")
	}
	if desc.AuthenticationServices != "" {
		return errors.New("'AUTHENTICATION_SERVICES' is not supported. Please remove it")
	}
	if len(desc.AddressList.Address) > 0 && len(desc.AddressLists) > 0 {
		return errors.New("please specify either 'ADDRESS_LIST' or multiple 'ADDRESS'-es exactly under 'DESCRIPTION'")
	}
	err := validateTNSAddressList(&desc.AddressList)
	if err != nil {
		return fmt.Errorf("embedded address list 'ADDRESS' is incorrect: %w", err)
	}
	if c := len(desc.AddressLists); c > 1 {
		return fmt.Errorf("a single 'ADDRESS_LIST' is expected, got %d of them", c)
	}
	for _, addrList := range desc.AddressLists {
		err = validateTNSAddressList(&addrList)
		if err != nil {
			return fmt.Errorf("'ADDRESS_LIST' is incorrect: %w", err)
		}
	}
	err = validateTNSConnectionData(&desc.ConnectData)
	if err != nil {
		return fmt.Errorf("'CONNECT_DATA' is incorrect: %w", err)
	}
	err = validateTNSConnSecurity(&desc.Security)
	if err != nil {
		return fmt.Errorf("'SECURITY' is incorrect: %w", err)
	}
	return nil
}

// validateTNSConnectionData validates the specified TNS connection data for unsupported features.
func validateTNSConnectionData(connData *tns.ConnectData) error {
	if connData.ColocationTag != "" {
		return errors.New("'COLOCATION_TAG' is not supported. Please remove it")
	}
	if connData.ConnectionIdPrefix != "" {
		return errors.New("'CONNECTION_ID_PREFIX' is not supported. Please remove it")
	}
	if connData.PoolConnectionClass != "" {
		return errors.New("'POOL_CONNECTION_CLASS' is not supported. Please remove it")
	}
	if connData.PoolPurity != "" {
		return errors.New("'POOL_PURITY' is not supported. Please remove it")
	}
	if connData.ShardingKey != "" {
		return errors.New("'SHARDING_KEY' is not supported. Please remove it")
	}
	if connData.SuperShardingKey != "" {
		return errors.New("'SUPER_SHARDING_KEY' is not supported. Please remove it")
	}
	if connData.TunnelServiceName != "" {
		return errors.New("'TUNNEL_SERVICE_NAME' is not supported. Please remove it")
	}
	if connData.FailoverMode != "" {
		return errors.New("'FAILOVER_MODE' is not supported. Please remove it")
	}
	if connData.GlobalName != "" {
		return errors.New("'GLOBAL_NAME' is not supported. Please remove it")
	}
	if connData.Hs != "" {
		return errors.New("'HS' is not supported. Please remove it")
	}
	if connData.InstanceName != "" {
		return errors.New("'INSTANCE_NAME' is not supported. Please remove it")
	}
	if connData.RdbDatabase != "" {
		return errors.New("'RDB_DATABASE' is not supported. Please remove it")
	}
	if connData.Server != "" {
		return errors.New("'SERVER' is not supported. Please remove it")
	}
	if connData.Sid != "" {
		return errors.New("'SID' is not supported. Please remove it")
	}
	return nil
}

// validateTNSConnSecurity validates the specified TNS connection security for unsupported features.
func validateTNSConnSecurity(sec *tns.Security) error {
	if sec.SslServerCertDN != "" {
		return errors.New("'SSL_SERVER_CERT_DN' is not supported. Please remove it")
	}
	if sec.AuthenticationService != "" {
		return errors.New("'AUTHENTICATION_SERVICE' is not supported. Please remove it")
	}
	if sec.IgnoreAnoEncryption {
		return errors.New("'IGNORE_ANO_ENCRYPTION' is not supported. Please remove it")
	}
	if sec.Kerberos5CCName != "" {
		return errors.New("'KERBEROS5_CC_NAME' is not supported. Please remove it")
	}
	if sec.SslServerDNMatch {
		return errors.New("'SSL_SERVER_DN_MATCH' is not supported. Please remove it")
	}
	if sec.SslVersion != "" {
		return errors.New("'SSL_VERSION' is not supported. Please remove it")
	}
	if sec.WalletLocation != "" {
		return errors.New("'WALLET_LOCATION' is not supported. Please remove it")
	}
	return nil
}

// validateTNSAddressList validates the specified TNS Address List for unsupported features.
func validateTNSAddressList(addrList *tns.AddressList) error {
	if addrList.Enable != "" {
		return errors.New("'ENABLE' is not supported. Please remove it")
	}
	if addrList.Failover != "" {
		return errors.New("'FAILOVER' is not supported. Please remove it")
	}
	if addrList.LoadBalance != "" {
		return errors.New("'LOAD_BALANCE' is not supported. Please remove it")
	}
	if addrList.SourceRoute != "" {
		return errors.New("'SOURCE_ROUTE' is not supported. Please remove it")
	}
	if addrList.TypeOfService != "" {
		return errors.New("'TYPE_OF_SERVICE' is not supported. Please remove it")
	}
	if addrList.RecvBufSize != 0 {
		return errors.New("'RECV_BUF_SIZE' is not supported. Please remove it")
	}
	if addrList.Sdu != 0 {
		return errors.New("'SDU' is not supported. Please remove it")
	}
	if addrList.SendBufSize != 0 {
		return errors.New("'SEND_BUF_SIZE' is not supported. Please remove it")
	}
	for _, addr := range addrList.Address {
		err := validateTNSAddress(&addr)
		if err != nil {
			return fmt.Errorf("'ADDRESS' is incorrect: %w", err)
		}
	}
	return nil
}

// validateTNSAddress validates the provided TNS Address for unsupported features.
func validateTNSAddress(addr *tns.Address) error {
	if addr.RecvBufSize != 0 {
		return errors.New("'RECV_BUF_SIZE' is not supported. Please remove it")
	}
	if addr.SendBufSize != 0 {
		return errors.New("'SEND_BUF_SIZE' is not supported. Please remove it")
	}
	if addr.HttpsProxyHost != "" {
		return errors.New("'HTTPS_PROXY' is not supported. Please remove it")
	}
	if addr.HttpsProxyPort != 0 {
		return errors.New("'HTTPS_PROXY_PORT' is not supported. Please remove it")
	}
	return nil
}
