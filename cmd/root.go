package cmd

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"os/user"
	"path"
	"regexp"
	"strings"

	"github.com/go-errors/errors"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh/terminal"
)

var clearCredentials bool

var rootCmd = &cobra.Command{
	Use:     "gp",
	Short:   "Git/p4 helper",
	Version: "0.0.2",
}

// Execute other subcommands
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err.(*errors.Error).ErrorStack())
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&clearCredentials, "clear-credentials",
		"c", false, "clear saved credentials")
}

func encrypt(plaintext string, fullKey string) (string, error) {
	if len(fullKey) < 32 {
		return "", errors.Errorf("key is too short")
	}

	key := []byte(fullKey)[:32]
	c, err := aes.NewCipher(key)
	if err != nil {
		return "", errors.Wrap(err, 1)
	}

	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return "", errors.Wrap(err, 1)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", errors.Wrap(err, 1)
	}

	sealed := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(sealed), nil
}

func decrypt(encoded string, fullKey string) ([]byte, error) {
	if len(fullKey) < 32 {
		return nil, errors.Errorf("key is too short")
	}

	key := []byte(fullKey)[:32]
	ciphertext, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, errors.Wrap(err, 1)
	}

	c, err := aes.NewCipher(key)
	if err != nil {
		return nil, errors.Wrap(err, 1)
	}

	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return nil, errors.Wrap(err, 1)
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, errors.Errorf("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	result, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, errors.Wrap(err, 1)
	}
	return result, nil
}

func newCmd(name string, args ...string) *exec.Cmd {
	fmt.Println(name, strings.Join(args, " "))

	/* #nosec */
	cmd := exec.Command(name, args...)
	cmd.Stderr = os.Stdout
	return cmd
}

func readSecret() (string, error) {
	var key string

	/* #nosec */
	keyCmd, err := exec.Command("git", "config", "--global", "gp.key").Output()
	if err != nil {
		b := make([]byte, 16)
		if _, err := rand.Read(b); err != nil {
			return "", errors.Wrap(err, 1)
		}

		dst := make([]byte, hex.EncodedLen(len(b)))
		hex.Encode(dst, b)
		key = string(dst)
		args := []string{"config", "--global", "--replace-all", "gp.key", key}

		/* #nosec */
		if err := exec.Command("git", args...).Run(); err != nil {
			return "", errors.Wrap(err, 1)
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
		return "", errors.Wrap(err, 1)
	}

	return encrypt(bytePassword, key)
}

func trim(s string, err error) (string, error) {
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(s), nil
}

func readInput(isSecured bool) (string, error) {
	fd := int(os.Stdin.Fd())
	oldState, err := terminal.MakeRaw(fd)
	if err != nil {
		return "", errors.Wrap(err, 1)
	}
	defer func() {
		if err = terminal.Restore(fd, oldState); err != nil {
			log.Fatal(err)
		}
	}()

	n := terminal.NewTerminal(os.Stdout, "")
	if isSecured {
		return readSecured(n)
	}

	result, err := n.ReadLine()
	if err != nil {
		return "", errors.Wrap(err, 1)
	}
	return result, nil
}

func readConfig(key string, title string, isSecured bool) (string, error) {
	gitKey := "gp." + key
	var out string

	/* #nosec */
	cmdOut, err := exec.Command("git", "config", "--global", gitKey).Output()
	if err != nil {
		fmt.Printf("Enter %s: ", title)

		out, err = trim(readInput(isSecured))
		if err != nil {
			return out, err
		}

		args := []string{"config", "--global", "--replace-all", gitKey, out}

		/* #nosec */
		if err := exec.Command("git", args...).Run(); err != nil {
			return out, errors.Wrap(err, 1)
		}
	} else {
		out = string(cmdOut)
	}
	return out, nil
}

func setHostEnvVar() error {
	host, err := trim(readConfig("P4PORT", "Server Host", false))
	if err != nil {
		return err
	}

	if err := os.Setenv("P4PORT", host); err != nil {
		return errors.Wrap(err, 1)
	}

	return nil
}

func getUsername() (string, error) {
	return trim(readConfig("P4USER", "Username", false))
}

func setUsernameEnvVar() error {
	user, err := getUsername()
	if err != nil {
		return err
	}

	if err := os.Setenv("P4USER", user); err != nil {
		return errors.Wrap(err, 1)
	}

	return nil
}

func login() error {
	if clearCredentials {
		args := []string{"config", "--global", "--remove-section", "gp"}

		/* #nosec */
		if err := exec.Command("git", args...).Run(); err != nil {
			return errors.Wrap(err, 1)
		}
	}

	if err := setHostEnvVar(); err != nil {
		return err
	}

	if err := setUsernameEnvVar(); err != nil {
		return err
	}

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
	if err := cmd.Run(); err != nil {
		return errors.Wrap(err, 1)
	}
	return nil
}

func getDepotPath() (string, error) {
	/* #nosec */
	msg, err := exec.Command("git", "log", "p4/HEAD", "-n1").Output()
	if err != nil {
		return "", errors.Wrap(err, 1)
	}

	gitP4Re := regexp.MustCompile(` *\[git-p4: (.*)\]`)
	m := gitP4Re.FindStringSubmatch(string(msg))
	if m == nil {
		return "", errors.Errorf("couldn't find git-p4 logs")
	}

	var depotPath string
	for _, kv := range strings.Split(m[1], ":") {
		param := strings.Split(kv, "=")
		if strings.TrimSpace(param[0]) == "depot-paths" {
			depotPath = strings.Trim(param[1], ` "`)
		}
	}
	if depotPath == "" {
		return "", errors.Errorf("couldn't find depot path")
	}

	return depotPath, nil
}

func getAllWorkspaces(username string) ([]string, error) {
	var workspaces []string

	/* #nosec */
	ztag, err := exec.Command("p4", "-ztag", "clients", "-u", username).Output()
	if err != nil {
		return nil, errors.Wrap(err, 1)
	}

	ztagClientRe := regexp.MustCompile(`(?m:^\.\.\.\s+(\w+)\s+(.*)$)`)
	mm := ztagClientRe.FindAllStringSubmatch(string(ztag), -1)
	if mm == nil {
		return nil, errors.Errorf("couldn't find p4 clients")
	}

	for _, matches := range mm {
		key := matches[1]
		if key == "client" {
			workspaces = append(workspaces, matches[2])
		}
	}
	return workspaces, nil
}

func createWorkspace() error {
	username, err := getUsername()
	if err != nil {
		return err
	}

	depotPath, err := getDepotPath()
	if err != nil {
		return err
	}

	workspace := username + strings.Replace(
		depotPath[1:len(depotPath)-1], "/", "_", -1)
	if err = os.Setenv("P4CLIENT", workspace); err != nil {
		return errors.Wrap(err, 1)
	}

	workspaces, err := getAllWorkspaces(username)
	if err != nil {
		return err
	}

	for _, ws := range workspaces {
		if ws == workspace {
			// Got created workspace
			return nil
		}
	}

	usr, err := user.Current()
	if err != nil {
		return errors.Wrap(err, 1)
	}

	root := path.Join(usr.HomeDir, ".gp", workspace)
	mapping := fmt.Sprintf("\t%s... //%s/%s...",
		depotPath, workspace, depotPath[2:])
	def := fmt.Sprintf(`
Client: %s
Owner: %s
Root: %s
Options: allwrite noclobber compress unlocked nomodtime rmdir
View:
%s
`, workspace, username, root, mapping)

	cmd := newCmd("p4", "client", "-i")
	cmd.Stdin = strings.NewReader(def)
	if err := cmd.Run(); err != nil {
		return errors.Wrap(err, 1)
	}

	return nil
}

func prepareSubmit() error {
	if err := login(); err != nil {
		return err
	}

	if err := createWorkspace(); err != nil {
		return err
	}

	args := []string{"config", "--replace-all",
		"git-p4.skipSubmitEdit", "true"}
	/* #nosec */
	if err := exec.Command("git", args...).Run(); err != nil {
		return errors.Wrap(err, 1)
	}

	return nil
}
