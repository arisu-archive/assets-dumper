package zipcrypto

const (
	FlagIsEncrypted       = 0x0001
	FlagHasDataDescriptor = 0x0008
)

func IsEncrypted(flags uint16) bool {
	return flags&FlagIsEncrypted != 0
}

func HasDataDescriptor(flags uint16) bool {
	return flags&FlagHasDataDescriptor != 0
}
