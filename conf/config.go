package conf

type Zconfig struct {
	Name       string
	Ip         string
	Port       int
	IpVer      string
	MaxClients uint32
	PoolSize   int
}
