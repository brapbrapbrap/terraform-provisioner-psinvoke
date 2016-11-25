package main

import (
	"bufio"
	"fmt"
	"github.com/armon/circbuf"
	"github.com/hashicorp/terraform/terraform"
	"github.com/mitchellh/go-linereader"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"syscall"
)

type Provisioner struct {
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	Host     string `mapstructure:"host"`
	Script   string `mapstructure:"script"`
	Params   string `mapstructure:"params"`
}

func (p *Provisioner) Run(o terraform.UIOutput) error {
	// create the Powershell Invoke-Command template, plugging in our parameters
	psinvokeCommands := []string{
		// we are creating the powershell script with the commands to be run remotely
		"Set-WSManInstance -ResourceURI winrm/config/client -ValueSet @{TrustedHosts='*'}",
		"$username = '" + p.Username + "'",
		"$password = '" + p.Password + "' | ConvertTo-SecureString -asPlainText -Force",
		"$Credential = New-Object System.Management.Automation.PSCredential($username,$password)",
		"invoke-command  -ComputerName '" + p.Host + "' -Credential $Credential -scriptblock {param($command,$param) & $command $param } -argumentlist '" + p.Script + "','" + p.Params + "'",
	}

	//o.Output("params: " + p.Username + " " + p.Password + " " + p.Host + " " + p.Script + " " + p.Params)

	o.Output(fmt.Sprintf("running command: %s on host: %s with params: %s", p.Script, p.Host, p.Params))

	// create temporary file with our Powershell script, then rename it to have a ".ps1" extension
	tmpfile, filerr := ioutil.TempFile(os.TempDir(), "ps")
	defer os.Remove(tmpfile.Name() + ".ps1")

	if filerr != nil {
		o.Output(fmt.Sprintf("Error creating temporary file %q: %v", tmpfile.Name(), filerr))
		return fmt.Errorf("Error creating temporary file %q: %v", tmpfile.Name(), filerr)
	}
	tmpfile.Close()

	renerr := os.Rename(tmpfile.Name(), tmpfile.Name()+".ps1")
	if renerr != nil {
		return fmt.Errorf("Error renaming temporary file %q: %v", tmpfile.Name(), renerr)
	}

	// write our powershell script to the file
	psfile, pserr := os.OpenFile(tmpfile.Name()+".ps1", syscall.O_RDWR, 777)
	if pserr != nil {
		return fmt.Errorf("Error opening temporary file for writing %q: %v", tmpfile.Name()+".ps1", pserr)
	}
	defer psfile.Close()

	w := bufio.NewWriter(psfile)
	for _, line := range psinvokeCommands {
		fmt.Fprintln(w, line)
	}
	w.Flush()
	// remember to close after writing
	psfile.Close()

	o.Output("running temporary file as: C:\\Windows\\System32\\WindowsPowerShell\\v1.0\\powershell.exe -file " + tmpfile.Name() + ".ps1")

	// run our command in Powershell
	cmd := exec.Command("C:\\Windows\\System32\\WindowsPowerShell\\v1.0\\powershell.exe", "-file", tmpfile.Name()+".ps1")

	// Execute and capture output
	maxBufSize := 8 * 1024
	pr, pw := io.Pipe()
	copyDoneCh := make(chan struct{})
	go p.copyOutput(o, pr, copyDoneCh)
	output, _ := circbuf.NewBuffer(int64(maxBufSize))
	cmd.Stderr = io.MultiWriter(output, pw)
	cmd.Stdout = io.MultiWriter(output, pw)

	// Run the command to completion
	err := cmd.Run()

	// Close the write-end of the pipe so that the goroutine mirroring output ends properly
	pw.Close()
	<-copyDoneCh

	if err != nil {
		return fmt.Errorf("Error executing command %p :%q", cmd.Args, err)
	}

	o.Output("ran script " + p.Script + " successfully on " + p.Host)
	return nil
}

func (p *Provisioner) Validate() error {
	// validator regexes, the host one is nicked from govalidator
	var dnsregexs string = `^([a-zA-Z0-9]{1}[a-zA-Z0-9_-]{1,62}){1}(\.[a-zA-Z0-9]{1}[a-zA-Z0-9_-]{1,62})*$`
	var userregexs string = `^([a-zA-Z]{1}[a-zA-Z0-9_-]{1,24}|\.){1}\\([a-zA-Z0-9]{1}[a-zA-Z0-9_-]{1,64})*$`
	var dnsregex = regexp.MustCompile(dnsregexs)
	var userregex = regexp.MustCompile(userregexs)
	
	// host validation
	if !dnsregex.MatchString(p.Host) {
		return fmt.Errorf("Invalid host. host: %s", p.Host)
	}
	
	// username validation
	if !userregex.MatchString(p.Username) {
		return fmt.Errorf("Invalid username. username: %s", p.Username)
	}
	
	// validations successful
	return nil
}

func (p *Provisioner) copyOutput(
	o terraform.UIOutput, r io.Reader, doneCh chan<- struct{}) {
	defer close(doneCh)
	lr := linereader.New(r)
	for line := range lr.Ch {
		o.Output(line)
	}
}
