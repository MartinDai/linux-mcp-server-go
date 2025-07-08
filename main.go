package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"golang.org/x/crypto/ssh"
)

type ExecuteShellArgs struct {
	MachineIP string `json:"machine_ip"`
	Path      string `json:"path"`
	Shell     string `json:"shell"`
}

type HostConfig struct {
	IP       string `json:"ip"`
	User     string `json:"user"`
	Password string `json:"password"`
	Port     int    `json:"port"`
}

func loadHostConfigs() ([]HostConfig, error) {
	data, err := os.ReadFile("hosts.json")
	if err != nil {
		return nil, fmt.Errorf("failed to read hosts.json: %v", err)
	}

	var configs []HostConfig
	if err := json.Unmarshal(data, &configs); err != nil {
		return nil, fmt.Errorf("failed to parse hosts.json: %v", err)
	}

	return configs, nil
}

func findHostConfig(configs []HostConfig, ip string) (*HostConfig, error) {
	for _, config := range configs {
		if config.IP == ip {
			return &config, nil
		}
	}
	return nil, fmt.Errorf("host configuration not found for IP: %s", ip)
}

func executeSSHCommand(config *HostConfig, path, shell string) (string, error) {
	sshConfig := &ssh.ClientConfig{
		User: config.User,
		Auth: []ssh.AuthMethod{
			ssh.Password(config.Password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         30 * time.Second,
	}

	addr := config.IP + ":" + strconv.Itoa(config.Port)
	client, err := ssh.Dial("tcp", addr, sshConfig)
	if err != nil {
		return "", fmt.Errorf("failed to connect to SSH: %v", err)
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return "", fmt.Errorf("failed to create SSH session: %v", err)
	}
	defer session.Close()

	command := fmt.Sprintf("cd %s && %s", path, shell)
	output, err := session.CombinedOutput(command)
	if err != nil {
		return string(output), fmt.Errorf("command execution failed: %v", err)
	}

	return string(output), nil
}

func ExecuteShell(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[ExecuteShellArgs]) (*mcp.CallToolResultFor[struct{}], error) {
	configs, err := loadHostConfigs()
	if err != nil {
		return &mcp.CallToolResultFor[struct{}]{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Error loading host configurations: %v", err)},
			},
		}, nil
	}

	config, err := findHostConfig(configs, params.Arguments.MachineIP)
	if err != nil {
		return &mcp.CallToolResultFor[struct{}]{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Error finding host configuration: %v", err)},
			},
		}, nil
	}

	output, err := executeSSHCommand(config, params.Arguments.Path, params.Arguments.Shell)
	if err != nil {
		return &mcp.CallToolResultFor[struct{}]{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("SSH execution error: %v\nOutput: %s", err, output)},
			},
		}, nil
	}

	return &mcp.CallToolResultFor[struct{}]{
		Content: []mcp.Content{
			&mcp.TextContent{Text: output},
		},
	}, nil
}

func main() {
	var httpAddr = flag.String("http", "", "if set, use streamable HTTP at this address, instead of stdin/stdout")
	flag.Parse()

	server := mcp.NewServer("linux_mcp_server", "v0.0.1", nil)
	server.AddTools(mcp.NewServerTool("execute_shell", "execute shell command on remote machine via SSH", ExecuteShell, mcp.Input(
		mcp.Property("machine_ip", mcp.Description("the IP address of the target machine")),
		mcp.Property("path", mcp.Description("the working directory path on remote machine")),
		mcp.Property("shell", mcp.Description("the shell command to execute")),
	)))

	if *httpAddr != "" {
		handler := mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server {
			return server
		}, nil)
		log.Printf("MCP handler listening at %s", *httpAddr)
		http.ListenAndServe(*httpAddr, handler)
	} else {
		t := mcp.NewLoggingTransport(mcp.NewStdioTransport(), os.Stderr)
		if err := server.Run(context.Background(), t); err != nil {
			log.Printf("Server failed: %v", err)
		}
	}
}
