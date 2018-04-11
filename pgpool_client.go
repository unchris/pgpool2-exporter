package main

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
)

const (
	PCPNodeCount = "/usr/sbin/pcp_node_count"
	PCPNodeInfo  = "/usr/sbin/pcp_node_info"
	PCPProcCount = "/usr/sbin/pcp_proc_count"

	NodeStatusInitialization = "Initialization"
	NodeStatusUP1            = "Node is up. No connection yet"
	NodeStatusUP2            = "Node is uo. Connections are pooled"
	NodeStatusDown           = "Node is down"
	NodeStatusUnknown        = "Unknown node status"
)

var (
	PCPValueRegExp = regexp.MustCompile(`^[^:]+: (.*)$`)

	nodeStatusToString = map[int]string{
		0: NodeStatusInitialization,
		1: NodeStatusUP1,
		2: NodeStatusUP2,
		3: NodeStatusDown,
	}
)

type PGPoolClient struct {
	Hostname         string
	Port             int
	Username         string
	Password         string
	TimeoutInSeconds int
}

func (p *PGPoolClient) Validate() error {
	if len(p.Hostname) == 0 {
		return errors.New("PgPool2 hostname must be specified")
	}
	if len(p.Username) == 0 {
		return errors.New("PgPool2 username must be specified")
	}
	if len(p.Password) == 0 {
		return errors.New("PgPool2 password must be specified")
	}
	return nil
}

func (p *PGPoolClient) execCommand(cmd string, arg ...string) (*bytes.Buffer, error) {
	stdoutBuffer := &bytes.Buffer{}
	argCommon := []string{
		strconv.Itoa(p.TimeoutInSeconds),
		p.Hostname,
		strconv.Itoa(p.Port),
		p.Username,
		p.Password,
	}
	argResult := append(argCommon, arg...)
	pgpoolExec := exec.Command(cmd, argResult...)
	pgpoolExec.Stdout = stdoutBuffer
	err := pgpoolExec.Run()
	if err != nil {
		logrus.Errorf("tried to execute %s and got error: %v", cmd, err)
		return stdoutBuffer, err
	}
	return stdoutBuffer, nil
}

func (p *PGPoolClient) ExecNodeCount() (int, error) {
	bytesBuffer, err := p.execCommand(PCPNodeCount)
	if err != nil {
		return 0, err
	}
	bytes, err := ioutil.ReadAll(bytesBuffer)
	if err != nil {
		return 0, err
	}
	resultString := strings.TrimSpace(string(bytes))
	resultInt, err := strconv.Atoi(resultString)
	if err != nil {
		return 0, err
	}
	return resultInt, nil
}

type NodeInfo struct {
	Hostname     string
	Port         int
	Status       int
	StatusString string
	Weight       float64
}

func NodeStatusToString(statusID int) string {
	status, ok := nodeStatusToString[statusID]
	if !ok {
		return NodeStatusUnknown
	}
	return status
}

func ExtractValueFromPCPString(line string) string {
	valueArr := PCPValueRegExp.FindStringSubmatch(line)
	if len(valueArr) > 0 {
		return valueArr[1]
	}
	return ""
}

func NodeInfoUnmarshal(cmdOutBuff io.Reader) (NodeInfo, error) {
	var ni NodeInfo
	reader := bufio.NewReader(cmdOutBuff)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return ni, err
			}
		}
		line = strings.TrimSpace(line)
		if strings.Contains(line, "Hostname") {
			ni.Hostname = ExtractValueFromPCPString(line)
		}
		if strings.Contains(line, "Port") {
			portRaw := ExtractValueFromPCPString(line)
			portInt, err := strconv.Atoi(portRaw)
			if err != nil {
				continue
			}
			ni.Port = portInt
		}
		if strings.Contains(line, "Status") {
			statusRaw := ExtractValueFromPCPString(line)
			statusInt, err := strconv.Atoi(statusRaw)
			if err != nil {
				continue
			}
			ni.Status = statusInt
			ni.StatusString = NodeStatusToString(statusInt)
		}
		if strings.Contains(line, "Weight") {
			weightRaw := ExtractValueFromPCPString(line)
			weightFloat, err := strconv.ParseFloat(weightRaw, 64)
			if err != nil {
				continue
			}
			ni.Weight = weightFloat
		}
	}
	return ni, nil
}

func (p *PGPoolClient) ExecNodeInfo(nodeID int) (NodeInfo, error) {
	bytesBuffer, err := p.execCommand(PCPNodeInfo, strconv.Itoa(nodeID), "-v")
	if err != nil {
		return NodeInfo{}, err
	}
	nodeInfo, err := NodeInfoUnmarshal(bytesBuffer)
	if err != nil {
		return NodeInfo{}, err
	}
	return nodeInfo, nil
}

func (p *PGPoolClient) ExecProcCount() ([]string, error) {
	bytesBuffer, err := p.execCommand(PCPProcCount)
	if err != nil {
		return []string{}, err
	}
	bytes, err := ioutil.ReadAll(bytesBuffer)
	if err != nil {
		return []string{}, err
	}
	procCountString := strings.TrimSpace(string(bytes))
	procCountArr := strings.Split(procCountString, " ")
	return procCountArr, nil
}
