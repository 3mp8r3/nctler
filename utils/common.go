package utils

import (
	"context"
	"crypto/md5"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

// MD5File get file md5sum
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

// FileExists check file exist
func FileExists(path string) bool {
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		return true
	}
	return false
}

// GetIPAddr get local ip address by a reachable out address
func GetIPAddr(outaddr string) (net.IP, error) {
	conn, err := net.Dial("udp", outaddr)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP, nil
}

// HTTPGet http request method get
func HTTPGet(url string) ([]byte, error) {
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

// DockerExec exec container
func DockerExec(cli *client.Client, containerid string, cmd []string) ([]byte, error) {
	execConfig := types.ExecConfig{Cmd: cmd, AttachStderr: true, AttachStdout: true}
	//execStartCheck := types.ExecStartCheck{}
	ctx := context.Background()
	execResp, err := cli.ContainerExecCreate(ctx, containerid, execConfig)
	if err != nil {
		return nil, err
	} else {
		resp, err := cli.ContainerExecAttach(ctx, execResp.ID, execConfig)
		if err != nil {
			return nil, err
		}
		defer resp.Close()
		output := make([]byte, 500)
		i, err := resp.Reader.Read(output)
		if err != nil {
			return nil, err
		}
		// trim extra bytess
		return output[8:i], nil
	}
}
