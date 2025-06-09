package tns

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name     string
		tnsStr   string
		expected *TNS
		wantErr  bool
	}{
		{
			name: "Single Description",
			tnsStr: `mysuperhost = (DESCRIPTION=
				(ADDRESS=(PROTOCOL=tcp)(HOST=myhost)(PORT=1521))
				(CONNECT_DATA=
					(SERVER=DEDICATED)
					(SERVICE_NAME=orcl)
				)
			)`,
			expected: &TNS{
				Entries: []TNSMappingEntry{
					{
						NetServiceName: "mysuperhost",
						ConnectDescriptor: ConnectDescriptor{
							Description: Description{
								AddressList: AddressList{
									Address: []Address{
										{Host: "myhost", Port: 1521, Protocol: "tcp"},
									},
								},
								ConnectData: ConnectData{
									Server:      Dedicated,
									ServiceName: "orcl",
								},
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Single Description, Multiple mappings",
			tnsStr: `mysuperhost = (DESCRIPTION=
				(ADDRESS=(PROTOCOL=tcp)(HOST=myhost)(PORT=1521))
				(CONNECT_DATA=
					(SERVER=DEDICATED)
					(SERVICE_NAME=orcl)
				)
			)
				
			mysuperhost2 = (DESCRIPTION=
				(ADDRESS=(PROTOCOL=tcp)(HOST=myhost2)(PORT=1522))
				(CONNECT_DATA=
					(SERVER=DEDICATED)
					(SERVICE_NAME=orcl2)
				)
			)`,
			expected: &TNS{
				Entries: []TNSMappingEntry{
					{
						NetServiceName: "mysuperhost",
						ConnectDescriptor: ConnectDescriptor{
							Description: Description{
								AddressList: AddressList{
									Address: []Address{
										{Host: "myhost", Port: 1521, Protocol: "tcp"},
									},
								},
								ConnectData: ConnectData{
									Server:      Dedicated,
									ServiceName: "orcl",
								},
							},
						},
					},
					{
						NetServiceName: "mysuperhost2",
						ConnectDescriptor: ConnectDescriptor{
							Description: Description{
								AddressList: AddressList{
									Address: []Address{
										{Host: "myhost2", Port: 1522, Protocol: "tcp"},
									},
								},
								ConnectData: ConnectData{
									Server:      Dedicated,
									ServiceName: "orcl2",
								},
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Multiple Descriptions",
			tnsStr: `mysuperhost2 = (DESCRIPTION_LIST=
				(DESCRIPTION=
					(ADDRESS=(PROTOCOL=tcp)(HOST=myhost1)(PORT=1521))
					(CONNECT_DATA=
						(SERVER=DEDICATED)
						(SERVICE_NAME=orcl1)
					)
				)
				(DESCRIPTION=
					(ADDRESS=(PROTOCOL=tcp)(HOST=myhost2)(PORT=1522))
					(CONNECT_DATA=
						(SERVER=DEDIcATED)
						(SERVICE_NAME=orcl2)
					)
				)
			)`,
			expected: &TNS{
				Entries: []TNSMappingEntry{
					{
						NetServiceName: "mysuperhost2",
						ConnectDescriptor: ConnectDescriptor{
							DescriptionList: DescriptionList{
								Descriptions: []Description{
									{
										AddressList: AddressList{
											Address: []Address{
												{Host: "myhost1", Port: 1521, Protocol: "tcp"},
											},
										},
										ConnectData: ConnectData{
											Server:      Dedicated,
											ServiceName: "orcl1",
										},
									},
									{
										AddressList: AddressList{
											Address: []Address{
												{Host: "myhost2", Port: 1522, Protocol: "tcp"},
											},
										},
										ConnectData: ConnectData{
											Server:      Dedicated,
											ServiceName: "orcl2",
										},
									},
								},
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Multiple Descriptions, Multiple entries",
			tnsStr: `mysuperhost2 = (DESCRIPTION_LIST=
				(DESCRIPTION=
					(ADDRESS=(PROTOCOL=tcp)(HOST=myhost1)(PORT=1521))
					(CONNECT_DATA=
						(SERVER=DEDICATED)
						(SERVICE_NAME=orcl1)
					)
				)
				(DESCRIPTION=
					(ADDRESS=(PROTOCOL=tcp)(HOST=myhost2)(PORT=1522))
					(CONNECT_DATA=
						(SERVER=DEDIcATED)
						(SERVICE_NAME=orcl2)
					)
				)
			)
				
			mysuperhost3 = (DESCRIPTION=
				(ADDRESS=(PROTOCOL=tcp)(HOST=myhost3)(PORT=1523))
				(CONNECT_DATA=
					(SERVER=DEDICATED)
					(SERVICE_NAME=orcl3)
				)
			)`,
			expected: &TNS{
				Entries: []TNSMappingEntry{
					{
						NetServiceName: "mysuperhost2",
						ConnectDescriptor: ConnectDescriptor{
							DescriptionList: DescriptionList{
								Descriptions: []Description{
									{
										AddressList: AddressList{
											Address: []Address{
												{Host: "myhost1", Port: 1521, Protocol: "tcp"},
											},
										},
										ConnectData: ConnectData{
											Server:      Dedicated,
											ServiceName: "orcl1",
										},
									},
									{
										AddressList: AddressList{
											Address: []Address{
												{Host: "myhost2", Port: 1522, Protocol: "tcp"},
											},
										},
										ConnectData: ConnectData{
											Server:      Dedicated,
											ServiceName: "orcl2",
										},
									},
								},
							},
						},
					},
					{
						NetServiceName: "mysuperhost3",
						ConnectDescriptor: ConnectDescriptor{
							Description: Description{
								AddressList: AddressList{
									Address: []Address{
										{Host: "myhost3", Port: 1523, Protocol: "tcp"},
									},
								},
								ConnectData: ConnectData{
									Server:      Dedicated,
									ServiceName: "orcl3",
								},
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Net Service Name as <host>:<port>/<service>",
			tnsStr: `myhost:1523/db=(
				DESCRIPTION=
					(ADDRESS=(PROTOCOL=TCP)(HOST=1.2.3.4)(PORT=1523))
					(CONNECT_DATA=(SERVER=DEDICATED)(SERVICE_NAME=db))
			)`,
			expected: &TNS{
				Entries: []TNSMappingEntry{
					{
						NetServiceName: "myhost:1523/db",
						ConnectDescriptor: ConnectDescriptor{
							Description: Description{
								AddressList: AddressList{
									Address: []Address{
										{Host: "1.2.3.4", Port: 1523, Protocol: "TCP"},
									},
								},
								ConnectData: ConnectData{
									Server:      Dedicated,
									ServiceName: "db",
								},
							},
						},
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.tnsStr)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.DeepEqual(t, tt.expected, got)
		})
	}
}
