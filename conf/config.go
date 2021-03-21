package conf

type Zconfig struct {
	Name       string
	Ip         string
	Port       int
	Type       string
	IpVer      string
	MaxClients uint32
	PoolSize   int
}
