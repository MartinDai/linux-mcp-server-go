package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/modelcontextprotocol/go-sdk/jsonschema"
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
	log.Println("Loading host configurations from hosts.json")
	data, err := os.ReadFile("hosts.json")
	if err != nil {
		log.Printf("Failed to read hosts.json: %v", err)
		return nil, fmt.Errorf("failed to read hosts.json: %v", err)
	}

	var configs []HostConfig
	if err := json.Unmarshal(data, &configs); err != nil {
		log.Printf("Failed to parse hosts.json: %v", err)
		return nil, fmt.Errorf("failed to parse hosts.json: %v", err)
	}

	log.Printf("Loaded %d host configurations", len(configs))
	return configs, nil
}

func findHostConfig(configs []HostConfig, ip string) (*HostConfig, error) {
	log.Printf("Looking for host configuration for IP: %s", ip)
	for _, config := range configs {
		if config.IP == ip {
			log.Printf("Found host configuration for IP: %s, user: %s, port: %d", ip, config.User, config.Port)
			return &config, nil
		}
	}
	log.Printf("Host configuration not found for IP: %s", ip)
	return nil, fmt.Errorf("host configuration not found for IP: %s", ip)
}

func executeSSHCommand(config *HostConfig, path, shell string) (string, error) {
	log.Printf("Establishing SSH connection to %s:%d as user %s", config.IP, config.Port, config.User)

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
		log.Printf("Failed to connect to SSH at %s: %v", addr, err)
		return "", fmt.Errorf("failed to connect to SSH: %v", err)
	}
	defer client.Close()

	log.Printf("SSH connection established successfully to %s", addr)

	session, err := client.NewSession()
	if err != nil {
		log.Printf("Failed to create SSH session: %v", err)
		return "", fmt.Errorf("failed to create SSH session: %v", err)
	}
	defer session.Close()

	command := fmt.Sprintf("cd %s && %s", path, shell)
	log.Printf("Executing command: %s", command)

	startTime := time.Now()
	output, err := session.CombinedOutput(command)
	duration := time.Since(startTime)

	if err != nil {
		log.Printf("Command execution failed after %v: %v", duration, err)
		return string(output), fmt.Errorf("command execution failed: %v", err)
	}

	log.Printf("Command executed successfully in %v, output length: %d bytes", duration, len(output))
	return string(output), nil
}

func ExecuteShell(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[ExecuteShellArgs]) (*mcp.CallToolResultFor[struct{}], error) {
	log.Printf("Received execute_shell tool call - IP: %s, Path: %s, Shell: %s",
		params.Arguments.MachineIP, params.Arguments.Path, params.Arguments.Shell)

	configs, err := loadHostConfigs()
	if err != nil {
		log.Printf("Tool call failed: error loading host configurations: %v", err)
		return &mcp.CallToolResultFor[struct{}]{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Error loading host configurations: %v", err)},
			},
		}, nil
	}

	config, err := findHostConfig(configs, params.Arguments.MachineIP)
	if err != nil {
		log.Printf("Tool call failed: error finding host configuration: %v", err)
		return &mcp.CallToolResultFor[struct{}]{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Error finding host configuration: %v", err)},
			},
		}, nil
	}

	output, err := executeSSHCommand(config, params.Arguments.Path, params.Arguments.Shell)
	if err != nil {
		log.Printf("Tool call failed: SSH execution error: %v", err)
		return &mcp.CallToolResultFor[struct{}]{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("SSH execution error: %v\nOutput: %s", err, output)},
			},
		}, nil
	}

	log.Printf("Tool call completed successfully for IP: %s", params.Arguments.MachineIP)
	return &mcp.CallToolResultFor[struct{}]{
		Content: []mcp.Content{
			&mcp.TextContent{Text: output},
		},
	}, nil
}

func main() {
	var httpAddr = flag.String("http", "", "if set, use streamable HTTP at this address, instead of stdin/stdout")
	flag.Parse()

	log.Println("Starting Linux MCP Server Go v0.0.1")
	log.Println("Initializing MCP server...")

	server := mcp.NewServer(&mcp.Implementation{
		Name:    "linux_mcp_server",
		Version: "v0.0.1",
	}, nil)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "execute_shell",
		Description: "execute shell command on remote machine via SSH",
		InputSchema: &jsonschema.Schema{
			Type:     "object",
			Required: []string{"machine_ip", "path", "shell"},
			Properties: map[string]*jsonschema.Schema{
				"machine_ip": {Type: "string", Description: "the IP address of the target machine"},
				"path":       {Type: "string", Description: "the working directory path on remote machine"},
				"shell":      {Type: "string", Description: "the shell command to execute"},
			},
		},
	}, ExecuteShell)

	log.Println("MCP server initialized with execute_shell tool")

	if *httpAddr != "" {
		log.Printf("Starting HTTP mode on %s", *httpAddr)
		handler := mcp.NewStreamableHTTPHandler(func(req *http.Request) *mcp.Server {
			log.Printf("HTTP request received from %s %s", req.RemoteAddr, req.URL.Path)
			return server
		}, nil)
		log.Printf("MCP handler listening at %s", *httpAddr)
		if err := http.ListenAndServe(*httpAddr, handler); err != nil {
			log.Printf("HTTP server failed: %v", err)
		}
	} else {
		log.Println("Starting stdio mode")
		t := mcp.NewLoggingTransport(mcp.NewStdioTransport(), os.Stderr)
		if err := server.Run(context.Background(), t); err != nil {
			log.Printf("Server failed: %v", err)
		}
	}
}
