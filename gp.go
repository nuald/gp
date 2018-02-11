package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/urfave/cli"
	"golang.org/x/crypto/ssh/terminal"
)

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

func login(c *cli.Context) error {
	if c.Bool("reset-credentials") {
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

func clone(c *cli.Context) error {
	err := login(c)
	if err != nil {
		log.Fatal(err)
		return err
	}

	args := append([]string{"p4", "clone"}, c.Args()...)
	cmd := newCmd("git", args...)
	return cmd.Run()
}

func rebase(c *cli.Context) error {
	err := login(c)
	if err != nil {
		log.Fatal(err)
		return err
	}

	cmd := newCmd("git", "p4", "rebase")
	return cmd.Run()
}

func submit(c *cli.Context) error {
	err := login(c)
	if err != nil {
		log.Fatal(err)
		return err
	}

	cmd := newCmd("git", "p4", "submit")
	return cmd.Run()
}

func shelve(c *cli.Context) error {
	err := login(c)
	if err != nil {
		log.Fatal(err)
		return err
	}

	cmd := newCmd("git", "p4", "submit", "--shelve")
	return cmd.Run()
}

func main() {
	app := cli.NewApp()
	app.Name = "gp"
	app.Version = "0.0.1"
	app.Usage = "Git/p4 helper"

	resetFlag := cli.BoolFlag{
		Name:  "reset-credentials",
		Usage: "reset saved credentials",
	}

	app.Commands = []cli.Command{
		{
			Name:      "clone",
			Usage:     "Creates a new Git directory from an existing p4 repository",
			UsageText: "gp clone <repository> [<directory>]",
			Description: `
	Creates a new Git directory from an existing p4 repository specified
	by the depot and the project (or the stream) paths:

		gp clone //depot/project
		gp clone //depot/stream destination

	To reproduce the entire p4 history in Git, please use the @all modifier
	on the depot path:

		gp clone //depot/project@all
`,
			Action: clone,
			Flags:  []cli.Flag{resetFlag},
		},
		{
			Name:      "rebase",
			Usage:     "Updates the Git repository with recent changes from p4",
			UsageText: "gp rebase",
			Action:    rebase,
			Flags:     []cli.Flag{resetFlag},
		},
		{
			Name:      "submit",
			Usage:     "Submits changes back to the p4 repository",
			UsageText: "gp submit",
			Action:    submit,
			Flags:     []cli.Flag{resetFlag},
		},
		{
			Name:      "shelve",
			Usage:     "Shelves changes back to the p4 repository",
			UsageText: "gp shelve",
			Action:    shelve,
			Flags:     []cli.Flag{resetFlag},
		},
	}

	app.Run(os.Args)
}
