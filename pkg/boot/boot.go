package boot

import (
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
)

const (
	url        = "https://get.k3s.io"
	script     = "k3.sh"
	kubeConfig = "/etc/rancher/k3s/k3s.yaml"
	nodeToken  = "/var/lib/rancher/k3s/server/node-token"
)

func Boot() error {
	err := fetchK3(script, url)
	if err != nil {
		log.Printf("unable to fetch k3 installation script, error %v", err)
		return err
	}
	token := "K101982131c68f7253bd3199f0b079cea5c514775ec73f7452044d3bfa35c58ab45::node:9f661b02b35db66c63a4f4231b358a28"
	server := "https://overlord-raspberrypi-agent3:6443"
	serverName := "overlord-raspberrypi-agent3"
	//server := "https://10.31.102.42:6443"
	err = installK3(script, token, server, serverName, true)
	if err != nil {
		log.Printf("unable to install k3, error: %v", err)
		return err
	}

	config, err := getFileContents(kubeConfig)
	if err != nil {
		log.Printf("unable to get kube config, error: %v", err)
		return err
	}
	log.Printf("Kubeconfig contains: %s", config)

	sToken, err := getFileContents(nodeToken)
	if err != nil {
		log.Printf("unable to get kube config, error: %v", err)
		return err
	}
	log.Printf("Kubeconfig contains: %s", sToken)
	return nil
}

func fetchK3(filename, url string) error {
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("unable to fetch k3 installation script, err %v", err)
		return err
	}
	defer resp.Body.Close()

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	return err
}

func installK3(script, token, server, serverName string, isMaster bool) error {
	permission := exec.Command("chmod", "+x", script)
	err := permission.Run()
	if err != nil {
		log.Printf("unable to provide script proper permission, error: %v", err)
		return err
	}

	log.Printf("installing k3...")
	var install *exec.Cmd
	if isMaster {
		args := []string{
			"--write-kubeconfig-mode=664",
			"--tls-san=" + serverName,
		}
		install = exec.Command("./"+script, args...)
	} else {
		args := []string{
			"--token=" + token,
			"--server=" + server,
		}
		install = exec.Command("./"+script, args...)
	}
	err = install.Run()
	if err != nil {
		log.Printf("unable to install k3, error: %v", err)
		return err
	}
	log.Printf("successfully installed K3")

	// delete extra password
	if isMaster {
		err := os.RemoveAll("/var/lib/rancher/k3s/server/cred/node-passwd")
		if err != nil {
			log.Printf("unable to delete secret file, error: %+v", err)
			return err
		}
	}
	return nil
}

func getFileContents(path string) (string, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		log.Printf("unable to read file %s, error: %+v", path, err)
		return "", err
	}
	return string(data), nil
}

/*
Kubeconfig (master only)
Node-token (master only)
Status: healthy/unhealthy
*/
