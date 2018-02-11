package cmd

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh/terminal"
)

var clearCredentials bool

var rootCmd = &cobra.Command{
	Use:     "gp",
	Short:   "Git/p4 helper",
	Version: "0.0.1",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&clearCredentials, "clear-credentials", "c", false, "clear saved credentials")
}

func encrypt(plaintext string, fullKey string) (string, error) {
	key := []byte(fullKey)[:32]
	c, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	sealed := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(sealed), nil
}

func decrypt(encoded string, fullKey string) ([]byte, error) {
	key := []byte(fullKey)[:32]
	ciphertext, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, err
	}

	c, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	return gcm.Open(nil, nonce, ciphertext, nil)
}

func newCmd(name string, args ...string) *exec.Cmd {
	fmt.Println(name, strings.Join(args, " "))
	cmd := exec.Command(name, args...)
	cmd.Stderr = os.Stdout
	return cmd
}

func readSecret() (string, error) {
	key := ""
	keyCmd, err := exec.Command("git", "config", "--global", "gp.key").Output()
	if err != nil {
		b := make([]byte, 16)
		_, err = rand.Read(b)
		if err != nil {
			return "", err
		}

		dst := make([]byte, hex.EncodedLen(len(b)))
		hex.Encode(dst, b)
		key = string(dst)
		err = exec.Command("git", "config", "--global",
			"--add", "gp.key", key).Run()
		if err != nil {
			return "", err
		}
	} else {
		key = string(keyCmd)
	}
	return key, nil
}

func readSecured(n *terminal.Terminal) (string, error) {
	key, err := readSecret()
	if err != nil {
		return "", err
	}

	bytePassword, err := n.ReadPassword("")
	if err != nil {
		return "", err
	}

	return encrypt(bytePassword, key)
}

func trim(s string, err error) (string, error) {
	if err != nil {
		return "", nil
	}
	return strings.TrimSpace(s), nil
}

func readInput(isSecured bool) (string, error) {
	fd := int(os.Stdin.Fd())
	oldState, err := terminal.MakeRaw(fd)
	if err != nil {
		return "", nil
	}
	defer terminal.Restore(fd, oldState)

	n := terminal.NewTerminal(os.Stdout, "")
	if isSecured {
		return readSecured(n)
	}
	return n.ReadLine()
}

func readConfig(key string, title string, isSecured bool) (string, error) {
	gitKey := "gp." + key
	out := ""
	cmdOut, err := exec.Command("git", "config", "--global", gitKey).Output()
	if err != nil {
		fmt.Printf("Enter %s: ", title)

		out, err = trim(readInput(isSecured))
		if err != nil {
			return out, err
		}

		err = exec.Command("git", "config", "--global",
			"--add", gitKey, out).Run()
		if err != nil {
			return out, err
		}
	} else {
		out = string(cmdOut)
	}
	return out, nil
}

func login() error {
	if clearCredentials {
		err := exec.Command("git", "config", "--global",
			"--remove-section", "gp").Run()
		if err != nil {
			return err
		}
	}

	host, err := trim(readConfig("P4PORT", "Server Host", false))
	if err != nil {
		return err
	}
	os.Setenv("P4PORT", host)

	user, err := trim(readConfig("P4USER", "Username", false))
	if err != nil {
		return err
	}
	os.Setenv("P4USER", user)

	key, err := trim(readSecret())
	if err != nil {
		return err
	}

	out, err := trim(readConfig("P4PASSWD", "Password", true))
	if err != nil {
		return err
	}

	password, err := decrypt(out, key)
	if err != nil {
		return err
	}

	cmd := newCmd("p4", "login")
	cmd.Stdin = strings.NewReader(string(password))
	return cmd.Run()
}
