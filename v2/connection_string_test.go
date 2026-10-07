package go_ora

import (
	"net/url"
	"strings"
	"testing"
)

func TestBuildJDBC_NoCommaSplit(t *testing.T) {
	user := "test_user"
	password := "test_password"
	// Connection string with commas (like a DN)
	connStr := "(DESCRIPTION=(ADDRESS=(PROTOCOL=tcps)(HOST=host)(PORT=1522))(SECURITY=(SSL_SERVER_CERT_DN=CN=adwc,O=Oracle,C=US)))"

	uStr := BuildJDBC(user, password, connStr, nil)

	u, err := url.Parse(uStr)
	if err != nil {
		t.Fatalf("Failed to parse generated URL: %v", err)
	}

	q := u.Query()
	connStrVals := q["connStr"]

	if len(connStrVals) != 1 {
		t.Errorf("Expected 1 value for 'connStr' parameter, got %d. Values: %v", len(connStrVals), connStrVals)
	}

	if connStrVals[0] != connStr {
		t.Errorf("Expected connStr to be unmodified, got: %s", connStrVals[0])
	}

	// Double check that it contains the commas and they are escaped in the raw query
	if !strings.Contains(u.RawQuery, "CN%3Dadwc%2CO%3DOracle%2CC%3DUS") {
		t.Errorf("Expected escaped commas in raw query, raw query: %s", u.RawQuery)
	}
}
