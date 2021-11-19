package main

import (
	"bytes"
	"flag"
	"io/ioutil"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
	"github.com/povsister/scp"
)

func main() {
	localFile  := flag.String("r", "", "Path to the local disk file")
	_           = flag.Uint("t", 4, "Not supported: remote file type")
	esxiHost   := flag.String("h", "", "IP or hostname of ESXi server")
	esxiUser   := flag.String("u", "", "ESXi user name")
	pwdFile    := flag.String("f", "", "File containing ESXi password")
	tmpFile    := flag.String("tmp-file", fmt.Sprintf("/tmp-%d.vmdk", time.Now().UTC().UnixNano()/1000000), "Temporary file on ESXi before converting (on same DS as final file)")
	flag.Parse()
	dsFile     := flag.Arg(0)

	if *localFile == "" || *esxiHost == "" || *esxiUser == "" || *pwdFile == "" {
		flag.PrintDefaults()
		Fatalf("ERROR: Following parameters are mandatory: -r -h -u -f")
	}

	if dsFile == "" {
		Fatalf("ERROR: Last argument must be datastore path on ESXi - for example '[mydatastore] /builds/blabla.vmdk'")
	}

	// read password from file
	Infof("Reading password from file \"%s\"", *pwdFile)
	f, err := os.Open(*pwdFile)
	if err != nil {
		Fatalf("ERROR: Cannot open password file %s: %s", *pwdFile, err)
	}
	esxiPassword, err := ioutil.ReadAll(f)
	if err != nil {
		Fatalf("ERROR: Cannot read password from file %s: %s", *pwdFile, err)
	}
	esxiPassword = bytes.Trim(esxiPassword, "\r\n")
	f.Close()

	// split datastore name from path on it
	dsRe := regexp.MustCompile(`^\s*\[(.*?)\]\s*(.*)$`)
	dsResult := dsRe.FindStringSubmatch(dsFile)
	if len(dsResult) != 3 {
		Fatalf("ERROR: Datastore path \"%s\" is not in format \"[datastore] path\"", dsFile)
	}
	datastore := dsResult[1]
	path := dsResult[2]
	Infof("Local file will be uploaded to datastore \"%s\" in \"%s\"", datastore, *tmpFile)
	Infof("Final disk file will be on datastore \"%s\" in \"%s\"", datastore, path)

	// connect to ESXi over SSH
	Infof("Connecting to %s as user %s", *esxiHost, *esxiUser)
	client, err := connectToESXi(*esxiHost, *esxiUser, string(esxiPassword))
	if err != nil {
		Fatalf("ERROR: Cannot SSH to %s@%s: %s", *esxiUser, *esxiHost, err)
	}
	defer client.Close()

	// copy local file to temporary file on datastore
	realPathTmp := fmt.Sprintf("/vmfs/volumes/%s/%s", datastore, strings.TrimLeft(*tmpFile, "/"))
	Infof("Copying \"%s\" to \"%s\"", *localFile, realPathTmp)
	err = copyFile(client, *localFile, realPathTmp)
	if err != nil {
		Fatalf("ERROR: Cannot copy local file \"%s\" to temporary remote file \"%s\": %s", *localFile, realPathTmp, err)
	}

	// clone the temporary to final destination and type
	realPathFinal := fmt.Sprintf("/vmfs/volumes/%s/%s", datastore, strings.TrimLeft(path, "/"))
	cloneCommand := fmt.Sprintf("vmkfstools -i \"%s\" -d thin \"%s\"", realPathTmp, realPathFinal)
	Infof("Cloning to final destination with command: %s", cloneCommand)
	returnCode, _, stderr, err := runCommand(client, cloneCommand)
	if err != nil {
		Fatalf("ERROR: Cannot run clone command: %s", err)
	}
	if returnCode != 0 {
		Fatalf("ERROR: Cannot clone to destination disk file: %s", stderr)
	}

	// delete temporary file
	deleteCommand := fmt.Sprintf("rm \"%s\"", realPathTmp)
	Infof("Deleting temporary file with command: %s", deleteCommand)
	returnCode, _, stderr, err = runCommand(client, deleteCommand)
	if err != nil {
		Fatalf("ERROR: Cannot run delete command: %s", err)
	}
	if returnCode != 0 {
		Fatalf("ERROR: Cannot delete temporary file: %s", stderr)
	}

	// all done
	Infof("Disk uploaded successfully")
}

func connectToESXi(host, user, password string) (*ssh.Client, error) {
	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
			ssh.KeyboardInteractive(func(user, instruction string, questions []string, echos []bool) (answers []string, err error) {
				for _, q := range questions {
					if strings.Contains(strings.ToLower(q), "password") {
						answers = append(answers, password)
					} else {
						answers = append(answers, "")
					}
				}
				return answers, nil
			}),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	client, err := ssh.Dial("tcp", fmt.Sprintf("%s:22", host), config)
	if err != nil {
		return nil, err
	}

	return client, nil
}

func runCommand(client *ssh.Client, cmd string) (int, string, string, error) {
	session, err := client.NewSession()
	if err != nil {
		return -1, "", "", err
	}
	defer session.Close()

	var stdout, stderr bytes.Buffer
	session.Stdout = &stdout
	session.Stderr = &stderr

	if err := session.Run(cmd); err != nil {
		switch err.(type) {
		case *ssh.ExitError:
			return err.(*ssh.ExitError).ExitStatus(), stdout.String(), stderr.String(), nil
		default:
			return -1, stdout.String(), stderr.String(), err
		}
	}

	return 0, stdout.String(), stderr.String(), nil
}

func copyFile(client *ssh.Client, local, remote string) error {
	scpClient, err := scp.NewClientFromExistingSSH(client, &scp.ClientOption{})
	if err != nil {
		return err
	}

	err = scpClient.CopyFileToRemote(local, remote, &scp.FileTransferOption{})
	if err != nil {
		return err
	}

	return nil
}
