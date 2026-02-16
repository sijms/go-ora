package converters

type StringCoder interface {
	GetStringCoder(charsetID, charsetForm int) (IStringConverter, error)
	GetDefaultStringCoder() (IStringConverter, error)
	GetServerStringCoder() IStringConverter
	GetServerNStringCoder() IStringConverter
	GetMaxStringLength() int64
}
