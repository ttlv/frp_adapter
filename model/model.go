package model

const (
	FrpOnline  = "online"
	FrpOffline = "offline"
)

type FrpServer struct {
	PublicIpAddress string
	Status          string
	UniqueID        string
	Port            string
}
