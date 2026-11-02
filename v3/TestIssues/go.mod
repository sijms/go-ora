module v3_testissues

go 1.24

replace github.com/sijms/go-ora/v3 => ../

require (
	github.com/sijms/go-ora/v3 v3.0.0-00010101000000-000000000000
	xorm.io/xorm v1.4.3
)

require (
	github.com/goccy/go-json v0.10.5 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/syndtr/goleveldb v1.0.0 // indirect
	xorm.io/builder v0.3.13 // indirect
)
