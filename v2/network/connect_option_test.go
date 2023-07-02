package network

import (
	"strings"
	"testing"
)

func TestExtractServers(t *testing.T) {
	text := `(DESCRIPTION=
(ADDRESS_LIST=(LOAD_BALANCE=OFF)(FAILOVER=ON)
(ADDRESS=(PROTOCOL=tcp)(HOST=host_dguard)(PORT=1521))
(ADDRESS=(PROTOCOL=tcp)(HOST=host_active)(PORT=1521))
)
(CONNECT_DATA=(SERVICE_NAME=service)(SERVER=DEDICATED))
)`
	t.Log(extractServers(text))
	text = `(DESCRIPTION_LIST=(LOAD_BALANCE=off)(FAILOVER=on)
(DESCRIPTION=(CONNECT_TIMEOUT=5)
(ADDRESS=(PROTOCOL=TCP)(HOST=host_dguard)(PORT=1521))
(CONNECT_DATA=(SERVICE_NAME=service)(SERVER=DEDICATED))
)
(DESCRIPTION=(CONNECT_TIMEOUT=5)
(ADDRESS=(PROTOCOL=TCP)(HOST=host_active)(PORT=1521))
(CONNECT_DATA=(SERVICE_NAME=service)(SERVER=DEDICATED))
)
)`
	t.Log(extractServers(text))
}

func TestUpdateDatabaseInfo(t *testing.T) {
	text := `(DESCRIPTION_LIST=(LOAD_BALANCE=off)(FAILOVER=on)
 (DESCRIPTION=(CONNECT_TIMEOUT=5)(ADDRESS=(PROTOCOL=TCP)
  (HOST=dataguard_host)(PORT=1521))
  (CONNECT_DATA=(SERVICE_NAME=SERVICE_RO)(SERVER=DEDICATED)))
 (DESCRIPTION=(CONNECT_TIMEOUT=5)(ADDRESS=(PROTOCOL=TCP)
  (HOST=active_instance)(PORT=1521))
  (CONNECT_DATA=(SERVICE_NAME=SERVICE)(SERVER=DEDICATED))))`
	text = strings.ReplaceAll(text, "\r", "")
	text = strings.ReplaceAll(text, "\n", "")
	var op = &ConnectionOption{}
	err := op.UpdateDatabaseInfo(text)
	if err != nil {
		t.Error(err)
	}
	t.Log(op)
}
