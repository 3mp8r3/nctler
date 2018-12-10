package utils

import (
	"crypto/md5"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"

	// "github.com/docker/engine-api/client"
	// "github.com/docker/engine-api/types"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"golang.org/x/net/context"
)

// MD5
func MD5File(path string) (string, error) {
	file, err := os.Open(path)
	defer file.Close()
	if err != nil {
		return "", err
	}

	h := md5.New()
	_, err = io.Copy(h, file)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

func FileExists(path string) bool {
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		return true
	}
	return false
}

func GetIpAddr(outaddr string) (net.IP, error) {
	conn, err := net.Dial("udp", outaddr)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP, nil
}

func HttpGet(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

func DockerExec(cli *client.Client, containerid string, cmd []string) ([]byte, error) {
	execConfig := types.ExecConfig{Cmd: cmd, AttachStderr: true, AttachStdout: true}
	//execStartCheck := types.ExecStartCheck{}
	execResp, err := cli.ContainerExecCreate(context.Background(), containerid, execConfig)
	if err != nil {
		return nil, err
	} else {
		hj_resp, err := cli.ContainerExecAttach(context.Background(), execResp.ID, execConfig)
		if err != nil {
			return nil, err
		}
		defer hj_resp.Close()
		output := make([]byte, 500)
		i, err := hj_resp.Reader.Read(output)
		if err != nil {
			return nil, err
		}
		return output[8:i], nil
	}
	return nil, nil
}
