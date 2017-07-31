package utils

import (
	"errors"
	"fmt"
	"net"
	"regexp"
	"log"
	"strings"
)

func GetLocalIp(etcdip string) (string, error) {
	ifaces, err := net.Interfaces()

	if err != nil {
		fmt.Println(err)
		return "", errors.New("os error")
	}

	regex := regexp.MustCompile(`(\d+\.\d+\.\d+)\.\d+`)
	match := regex.FindStringSubmatch(etcdip)
	if match == nil {
		log.Print("# input params error!")
		return "", errors.New("input params error!")
	}

	for _, i := range ifaces {
		addrs, _ := i.Addrs()
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}

			if strings.Contains(ip.String(), match[1]) {
				//fmt.Println("ip", ip)
				return ip.String(), nil
			}
		}
	}

	return "err", errors.New("no ip found!")
}