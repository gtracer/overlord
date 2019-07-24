package boot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

const (
	url        = "https://get.k3s.io"
	script     = "k3.sh"
	kubeConfig = "/etc/rancher/k3s/k3s.yaml"
	nodeToken  = "/var/lib/rancher/k3s/server/node-token"
	endpoint   = "http://51.143.124.178:8080/customeroverlord/overlord/%s"
)

type Config struct {
	Kubeconfig string `json:"kubeconfig,omitempty"`
	Token      string `json:"token,omitempty"`
	Master     string `json:"master,omitempty"`
	State      string `json:"state,omitempty"`
	Message    string `json:"message,omitempty"`
}

func Boot() error {
	log.Println("Entering boot..")
	err := fetchK3(script, url)
	if err != nil {
		log.Printf("unable to fetch k3 installation script, error %v", err)
		return err
	}
	config := &Config{}
	for {
		err = postStatus(config)
		time.Sleep(10 * time.Second)
		if err != nil {
			log.Printf("retrying post status, error: %v", err)
			continue
		}
		log.Printf("current config is: %+v", config)
		err = installK3(script, config)
		if err != nil {
			log.Printf("unable to install k3, error: %v", err)
			continue
		}
	}
	return nil
}

func getNodeName() (string, error) {
	log.Println("Entering getConfig..")
	node, err := os.Hostname()
	if err != nil {
		log.Printf("unable to get hostname +%v", err)
		return "", err
	}
	node = strings.ReplaceAll(node, ".", "")
	node = strings.ReplaceAll(node, "overlord-", "")
	log.Printf("host name is: %s", node)
	return node, nil
}

func postStatus(config *Config) error {
	nodeName, err := getNodeName()
	if err != nil {
		return err
	}
	endpointName := fmt.Sprintf(endpoint, nodeName)
	sKubeConfig, err := getFileContents(kubeConfig)
	if err != nil {
		log.Printf("unable to get kube config, error: %v", err)
		return err
	}
	config.Kubeconfig = sKubeConfig
	log.Printf("Kubeconfig contains: %s", sKubeConfig)

	sToken, err := getFileContents(nodeToken)
	if err != nil {
		log.Printf("unable to get kube config, error: %v", err)
		return err
	}
	config.Token = sToken
	log.Printf("token contains: %s", sToken)

	body, err := json.Marshal(config)
	if err != nil {
		log.Printf("unable to marshal body, error %+v", err)
		return err
	}
	resp, err := http.Post(endpointName, "application/json", bytes.NewBuffer(body))
	if err != nil {
		log.Printf("unable to get config response +%v", err)
		return err
	}

	defer resp.Body.Close()
	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("unable to get config body +%v", err)
		return err
	}

	err = json.Unmarshal(body, config)
	if err != nil {
		log.Printf("unable to unmarshal config +%v", err)
		return err
	}
	return nil
}

func fetchK3(filename, url string) error {
	log.Println("Entering fetchk3..")
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

func installK3(script string, config *Config) error {
	permission := exec.Command("chmod", "+x", script)
	err := permission.Run()
	if err != nil {
		log.Printf("unable to provide script proper permission, error: %v", err)
		return err
	}

	log.Printf("installing k3...")
	var install *exec.Cmd
	nodeName, err := os.Hostname()
	if err != nil {
		log.Printf("unable to get hostname +%v", err)
		return err
	}
	nodeName = strings.ReplaceAll(nodeName, ".", "")
	var isMaster bool

	if strings.EqualFold(config.Master, nodeName) {
		isMaster = true
	}
	if isMaster {
		log.Print("bootstrapping master")
		args := []string{
			"--write-kubeconfig-mode=664",
			"--tls-san=" + config.Master,
		}
		install = exec.Command("./"+script, args...)
	} else {
		log.Print("bootstrapping agent")
		url := fmt.Sprintf("http://%s:6443", config.Master)
		args := []string{
			"--token=" + config.Token,
			"--server=" + url,
		}
		install = exec.Command("./"+script, args...)
		install.Env = append(install.Env, "K3S_TOKEN="+config.Token)
		install.Env = append(install.Env, "K3S_URL="+url)
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
	config.State = "Healthy"
	config.Message = "successfully bootstrapped node"
	return nil
}

func getFileContents(path string) (string, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return "", nil
	}

	data, err := ioutil.ReadFile(path)
	if err != nil {
		log.Printf("unable to read file %s, error: %+v", path, err)
		return "", err
	}
	return string(data), nil
}
