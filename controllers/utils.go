package controllers

import (
	"fmt"
	"math/rand"

	"github.com/seancfoley/ipaddress-go/ipaddr"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

type metaData struct {
	Namespace string
	Name      string
	Port      int32
	ProxyPort int32
}

type ServiceData map[string]metaData
type OccupiedPorts map[int32]bool

func Ptr[T any](t T) *T {
	return &t
}

func GenerateWgKeys() ([]byte, []byte, error) {
	key, err := wgtypes.GenerateKey()
	if err != nil {
		return nil, nil, err
	}

	return []byte(key.PublicKey().String()), []byte(key.String()), nil
}

func GeneratePublicKey(privateKey string) ([]byte, error) {
	key, err := wgtypes.ParseKey(privateKey)
	if err != nil {
		return nil, err
	}

	return []byte(key.PublicKey().String()), nil
}

func GetRemoteDeviceIp(deviceOffcet int64) ([]byte, error) {
	deviceRange := ipaddr.NewIPAddressString("10.13.0.0/16")

	if address, addressError := deviceRange.ToAddress(); addressError == nil {
		increment := address.Increment(deviceOffcet)
		return []byte(ipaddr.NewIPAddressString(increment.GetNetIP().String()).String()), nil
	} else {
		return nil, addressError
	}
}

func getRandomPort(existing, newData OccupiedPorts, prefix string) (int32, error) {
	// generate random port between 1024 and 65535 and return it
	result := 1024 + rand.Int31n(65535-1024)

	if len(newData) == 65535-1024 {
		return 0, fmt.Errorf("no available ports")
	}

	for existing[result] || newData[result] {
		result = 1024 + rand.Int31n(65535-1024)
	}

	return result, nil
}
