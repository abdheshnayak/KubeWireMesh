package utils

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"

	"github.com/abdheshnayak/kubewiremesh/controllers/constants"
)

func SendBytesToReceiver(ip, message []byte) error {
	if !CheckRecieverStatus(ip) {
		return errors.New("Receiver is not available")
	}

	r, err := http.Post(fmt.Sprintf("http://%s:%d/config", ip, constants.RECEIVE_PORT), "application/json", bytes.NewBuffer(message))
	if err != nil {
		return err
	}

	if err := r.Body.Close(); err != nil {
		return err
	}

	return nil
}

func CheckRecieverStatus(ip []byte) bool {
	r, err := http.Get(fmt.Sprintf("http://%s:%d/healthy", ip, constants.RECEIVE_PORT))
	if err != nil {
		return false
	}

	if r.StatusCode != http.StatusOK {
		return false
	}

	return true
}
