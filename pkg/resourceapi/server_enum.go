package resourceapi

type Server string

const (
	ServerUnknown Server = "unknown"
	ServerGlobal  Server = "global"
	ServerJapan   Server = "japan"
)

func GetServer(server string) Server {
	switch server {
	case "g", "gl", "global":
		return ServerGlobal
	case "j", "jp", "japan":
		return ServerJapan
	}

	return ServerUnknown
}

func (s Server) IsValid() bool {
	return s == ServerGlobal || s == ServerJapan
}
