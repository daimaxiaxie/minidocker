package container

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	RUNNING string = "running"
	STOP    string = "stopped"
	EXIT    string = "exited"

	DefaultInfoLocation string = "/var/run/minidocker/%s/"
	ConfigName          string = "config.json"
	LogFile             string = "container.log"
	RootURL             string = "/root"
	MntURL              string = "/root/mnt/%s"
	WriteLayerURL       string = "/root/writeLayer/%s"
)

type Info struct {
	Pid         string   `json:"pid"`
	Id          string   `json:"id"`
	Name        string   `json:"name"`
	Command     string   `json:"command"`
	CreateTime  string   `json:"createTime"`
	Status      string   `json:"status"`
	Volume      string   `json:"volume"`
	PortMapping []string `json:"portMapping"`
}

var SkipList = map[string]bool{
	"network": true,
}

func recordContainerInfo(pid int, commands []string, containerName string, id string, volume string) (*Info, error) {

	containerInfo := &Info{
		Pid:        strconv.Itoa(pid),
		Id:         id,
		Name:       containerName,
		Command:    strings.Join(commands, " "),
		CreateTime: time.Now().Format("2006-01-02 15:04:05"),
		Status:     RUNNING,
		Volume:     volume,
	}
	jsonBytes, err := json.Marshal(containerInfo)
	if err != nil {
		return nil, err
	}

	pathUrl := fmt.Sprintf(DefaultInfoLocation, containerInfo.Name)
	if err := os.MkdirAll(pathUrl, 0622); err != nil {
		return nil, err
	}
	fileName := pathUrl + "/" + ConfigName
	file, err := os.Create(fileName)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	if _, err = file.Write(jsonBytes); err != nil {
		return nil, err
	}
	return containerInfo, nil
}

func deleteContainerInfo(containerId string) {
	pathUrl := fmt.Sprintf(DefaultInfoLocation, containerId)
	if err := os.RemoveAll(pathUrl); err != nil {
		logger.Errorf("Remove dir %s error %s", pathUrl, err)
	}
}

func RandID(length int) string {
	rand.Seed(time.Now().UnixNano())
	res := make([]byte, length)
	candidate := "0123456789"
	for i := range res {
		res[i] = candidate[rand.Intn(len(candidate))]
	}
	return string(res)
}

func GetContainerInfoByName(containerName string) (*Info, error) {
	pathUrl := fmt.Sprintf(DefaultInfoLocation, containerName)
	content, err := ioutil.ReadFile(pathUrl + ConfigName)
	if err != nil {
		return nil, err
	}
	var containerInfo = new(Info)
	if err = json.Unmarshal(content, containerInfo); err != nil {
		return nil, err
	}
	return containerInfo, nil
}

func GetContainerInfoByFile(file os.FileInfo) (*Info, error) {
	containerName := file.Name()
	configFile := fmt.Sprintf(DefaultInfoLocation, containerName) + ConfigName
	content, err := ioutil.ReadFile(configFile)
	if err != nil {
		return nil, err
	}
	var containerInfo = new(Info)
	if err = json.Unmarshal(content, containerInfo); err != nil {
		return nil, err
	}
	return containerInfo, nil
}

func GetAllContainer() ([]*Info, error) {
	pathUrl := fmt.Sprintf(DefaultInfoLocation, "")
	pathUrl = pathUrl[:len(pathUrl)-1]
	files, err := ioutil.ReadDir(pathUrl)
	if err != nil {
		return nil, err
	}

	var containers []*Info
	for _, file := range files {
		if _, ok := SkipList[file.Name()]; ok {
			continue
		}
		info, err := GetContainerInfoByFile(file)
		if err != nil {
			logger.Errorf("get %s container info error %s", file.Name(), err)
			continue
		}
		containers = append(containers, info)
	}
	return containers, nil
}

func GetContainerLog(containerName string) ([]byte, error) {
	pathUrl := fmt.Sprintf(DefaultInfoLocation, containerName)
	file, err := os.Open(pathUrl + LogFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	content, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}
	return content, nil
}
