package main

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os/exec"
	"time"
)

func getmac(ip string) string {
	//get mac relative to the ip address which connected to the mq.
	ifaces, err := net.Interfaces()
	if err != nil {
		return ""
	}
	firstMac := ""
	for _, iface := range ifaces {
		addrs, _ := iface.Addrs()
		for _, addr := range addrs {
			if firstMac == "" {
				firstMac = iface.HardwareAddr.String()
			}
			if ipNet, ok := addr.(*net.IPNet); ok && ipNet.IP.String() == ip {
				if iface.HardwareAddr.String() != "" {
					return iface.HardwareAddr.String()
				}
				return firstMac
			}
		}
	}
	return firstMac
}

var cbcIVBlock = []byte("UHNJUSBACIJFYSQN")

var paddingArray = [][]byte{
	{0},
	{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
	{2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2},
	{3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3},
	{4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4},
	{5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5},
	{6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6},
	{7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7},
	{8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8},
	{9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9},
	{10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10},
	{11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11},
	{12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12},
	{13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13},
	{14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14},
	{15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15},
	{16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16},
}

func pkcs7Padding(plainData []byte, dataLen, blockSize int) int {
	padLen := blockSize - dataLen%blockSize
	pPadding := plainData[dataLen : dataLen+padLen]

	copy(pPadding, paddingArray[padLen][:padLen])
	return padLen
}

func pkcs7UnPadding(origData []byte, dataLen int) ([]byte, error) {
	unPadLen := int(origData[dataLen-1])
	if unPadLen <= 0 || unPadLen > 16 {
		return nil, fmt.Errorf("wrong pkcs7 padding head size:%d", unPadLen)
	}
	return origData[:(dataLen - unPadLen)], nil
}

func encryptBytes(key []byte, out, in []byte, plainLen int) ([]byte, error) {
	if len(key) == 0 {
		return in[:plainLen], nil
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	//iv := out[:aes.BlockSize]
	//if _, err := io.ReadFull(rand.Reader, iv); err != nil {
	//	return nil, err
	//}
	mode := cipher.NewCBCEncrypter(block, cbcIVBlock)
	total := pkcs7Padding(in, plainLen, aes.BlockSize) + plainLen
	mode.CryptBlocks(out[:total], in[:total])
	return out[:total], nil
}

func decryptBytes(key []byte, out, in []byte, dataLen int) ([]byte, error) {
	if len(key) == 0 {
		return in[:dataLen], nil
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	mode := cipher.NewCBCDecrypter(block, cbcIVBlock)
	mode.CryptBlocks(out[:dataLen], in[:dataLen])
	return pkcs7UnPadding(out, dataLen)
}

// {240e:3b7:622:3440:59ad:7fa1:170c:ef7f 47924975352157270363627191692449083263 China CN 0xc0000965c8 Guangdong GD 0  Guangzhou 23.1167 113.25 Asia/Shanghai AS4134 Chinanet }
func netInfo() *NetInfo {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		// DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
		// 	var d net.Dialer
		// 	return d.DialContext(ctx, "tcp6", addr)
		// },
	}
	// sometime will be failed, retry
	for i := 0; i < 2; i++ {
		client := &http.Client{Transport: tr, Timeout: time.Second * 10}
		r, err := client.Get("https://ifconfig.co/json")
		if err != nil {
			gLog.Println(LevelINFO, "netInfo error:", err)
			continue
		}
		defer r.Body.Close()
		buf := make([]byte, 1024*64)
		n, err := r.Body.Read(buf)
		if err != nil {
			gLog.Println(LevelINFO, "netInfo error:", err)
			continue
		}
		rsp := NetInfo{}
		err = json.Unmarshal(buf[:n], &rsp)
		if err != nil {
			gLog.Printf(LevelERROR, "wrong NetInfo:%s", err)
			continue
		}
		return &rsp
	}
	return nil
}

func execOutput(name string, args ...string) string {
	cmdGetOsName := exec.Command(name, args...)
	var cmdOut bytes.Buffer
	cmdGetOsName.Stdout = &cmdOut
	cmdGetOsName.Run()
	return cmdOut.String()
}
