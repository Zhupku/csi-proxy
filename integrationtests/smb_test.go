package integrationtests

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"os/exec"
	"strings"
	"time"

	"testing"
)

const letterset = "abcdefghijklmnopqrstuvwxyz" + "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var seededRand = rand.New(rand.NewSource(time.Now().UnixNano()))

func stringWithCharset(length int, charset string) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

// RandomString generates a random string with specified length
func randomString(length int) string {
	return stringWithCharset(length, letterset)
}

func setupUser(username, password string) error {
	cmdLine := fmt.Sprintf(`$PWord = ConvertTo-SecureString $Env:password -AsPlainText -Force` +
		`;New-Localuser -name $Env:username -accountneverexpires -password $PWord`)
	cmd := exec.Command("powershell", "/c", cmdLine)
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("username=%s", username),
		fmt.Sprintf("password=%s", password),
		"PSModulePath=")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("setupUser failed: %v, output: %q", err, string(output))
	}
	return nil
}

func removeUser(t *testing.T, username string) {
	cmdLine := fmt.Sprintf(`Remove-Localuser -name $Env:username`)
	cmd := exec.Command("powershell", "/c", cmdLine)
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("username=%s", username))
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("setupUser failed: %v, output: %q", err, string(output))
	}
}

func setupSmbShare(shareName, localPath, username string) error {
	if err := os.MkdirAll(localPath, 0755); err != nil {
		return fmt.Errorf("setupSmbShare failed to create local path %q: %v", localPath, err)
	}
	cmdLine := fmt.Sprintf(`New-SMBShare -Name $Env:sharename -Path $Env:path -fullaccess $Env:username`)
	cmd := exec.Command("powershell", "/c", cmdLine)
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("sharename=%s", shareName),
		fmt.Sprintf("path=%s", localPath),
		fmt.Sprintf("username=%s", username))
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("setupSmbShare failed: %v, output: %q", err, string(output))
	}

	return nil
}

func removeSmbShare(t *testing.T, shareName string) {
	cmdLine := fmt.Sprintf(`Remove-SMBShare -Name $Env:sharename -Force`)
	cmd := exec.Command("powershell", "/c", cmdLine)
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("sharename=%s", shareName))
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("setupSmbShare failed: %v, output: %q", err, string(output))
	}
	return
}

func getSmbGlobalMapping(remotePath string) error {
	// use PowerShell Environment Variables to store user input string to prevent command line injection
	// https://docs.microsoft.com/en-us/powershell/module/microsoft.powershell.core/about/about_environment_variables?view=powershell-5.1
	cmdLine := fmt.Sprintf(`(Get-SmbGlobalMapping -RemotePath $Env:smbremotepath).Status`)

	cmd := exec.Command("powershell", "/c", cmdLine)
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("smbremotepath=%s", remotePath))
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("Get-SmbGlobalMapping failed: %v, output: %q", err, string(output))
	}
	if !strings.Contains(string(output), "OK") {
		return fmt.Errorf("Get-SmbGlobalMapping return status %q instead of OK", string(output))
	}
	return nil
}

func writeReadFile(path string) error {
	fileName := path + "\\hello.txt"
	f, err := os.Create(fileName)
	if err != nil {
		return fmt.Errorf("create file %q failed: %v", fileName, err)
	}
	defer f.Close()
	fileContent := "Hello World"
	if _, err = f.WriteString(fileContent); err != nil {
		return fmt.Errorf("write to file %q failed: %v", fileName, err)
	}
	if err = f.Sync(); err != nil {
		return fmt.Errorf("sync file %q failed: %v", fileName, err)
	}
	dat, err := ioutil.ReadFile(fileName)
	if err != nil {
		return fmt.Errorf("read file %q failed: %v", fileName, err)
	}
	if fileContent != string(dat) {
		return fmt.Errorf("read content of file %q failed: expected %q, got %q", fileName, fileContent, string(dat))
	}
	return nil
}

// Todo: Enable this test once the issue is fixed
// func TestSmbAPIGroup(t *testing.T) {
// 	t.Run("v1alpha1SmbTests", func(t *testing.T) {
// 		v1alpha1SmbTests(t)
// 	})
// 	t.Run("v1beta1SmbTests", func(t *testing.T) {
// 		v1beta1SmbTests(t)
// 	})
// 	t.Run("v1beta2SmbTests", func(t *testing.T) {
// 		v1beta2SmbTests(t)
// 	})
// 	t.Run("v1SmbTests", func(t *testing.T) {
// 		v1SmbTests(t)
// 	})
// }
