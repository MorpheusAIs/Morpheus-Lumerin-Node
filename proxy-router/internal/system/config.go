package system

type Config struct {
	LocalPortRange   string
	TcpMaxSynBacklog string
	Somaxconn        string
	NetdevMaxBacklog string
	RlimitSoft       uint64
	RlimitHard       uint64
}
