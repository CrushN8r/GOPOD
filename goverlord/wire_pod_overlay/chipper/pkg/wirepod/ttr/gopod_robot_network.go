// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_robot_network.go
package wirepod_ttr

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/kercre123/wire-pod/chipper/pkg/logger"
	"github.com/kercre123/wire-pod/chipper/pkg/vars"
	"golang.org/x/crypto/ssh"
)

// the battery watch's readings say when a Brobot drops off the network and when he's back - how
// long, and whether he rebooted (his uptime).
var gopodNetwork = struct {
	sync.Mutex
	downAt map[string]time.Time
}{downAt: map[string]time.Time{}}

// gopodNetworkDown: a Brobot his robot couldn't be reached for (a call failed on the network) -
// down until the battery watch reads him again (gopodNetworkBack).
func gopodNetworkDown(esn string, err error) {
	if err == nil || !strings.Contains(err.Error(), "connection error") {
		return
	}
	gopodNetwork.Lock()
	if _, down := gopodNetwork.downAt[esn]; !down {
		gopodNetwork.downAt[esn] = time.Now()
		logger.Println("GOPOD_NETWORK " + gopodBrobotName(esn) + " can't be reached: " + err.Error())
	}
	gopodNetwork.Unlock()
}

// gopodNetworkUp: the Brobot isn't known to be off the network.
func gopodNetworkUp(esn string) bool {
	gopodNetwork.Lock()
	defer gopodNetwork.Unlock()
	_, down := gopodNetwork.downAt[esn]
	return !down
}

// gopodNetworkMiss: a reading that didn't work - the first one after a good reading is his drop,
// logged once.
func gopodNetworkMiss(esn string, err error) {
	gopodBatteryLow.Lock()
	seen := gopodBatteryLow.seen[esn]
	gopodBatteryLow.Unlock()
	gopodNetwork.Lock()
	_, down := gopodNetwork.downAt[esn]
	if !seen || down {
		gopodNetwork.Unlock()
		return
	}
	now := time.Now()
	gopodNetwork.downAt[esn] = now
	gopodNetwork.Unlock()
	why := err.Error()
	if i := strings.LastIndex(why, "connect: "); i >= 0 {
		why = strings.TrimSuffix(why[i+len("connect: "):], "\"")
	}
	logger.Println("GOPOD_NETWORK " + gopodBrobotName(esn) + " off the network at " + now.Format("15:04:05") + " (" + why + ")")
}

// gopodNetworkBack: a good reading - after a drop, he's back: how long, and whether he rebooted; no
// reboot, his WiFi lines go to the session log.
func gopodNetworkBack(esn string) {
	gopodNetwork.Lock()
	at, down := gopodNetwork.downAt[esn]
	delete(gopodNetwork.downAt, esn)
	gopodNetwork.Unlock()
	if !down {
		return
	}
	go func() {
		name := gopodBrobotName(esn)
		off := time.Since(at).Round(time.Second)
		head := "GOPOD_NETWORK " + name + " back at " + time.Now().Format("15:04:05") + ", off for " + off.String()
		client, err := gopodRobotSSH(esn)
		if err != nil {
			logger.Println(head + " (reboot unknown: " + err.Error() + ")")
			return
		}
		defer client.Close()
		up, err := gopodRobotRun(client, "cat /proc/uptime")
		secs, perr := strconv.ParseFloat(strings.Fields(up + " 0")[0], 64)
		if err != nil || perr != nil {
			logger.Println(head + " (reboot unknown)")
			return
		}
		if secs < time.Since(at).Seconds() {
			logger.Println(head + ", rebooted (his own log went with it)")
			return
		}
		logger.Println(head + ", no reboot")
		lines, err := gopodRobotRun(client, "grep -E 'wpa_supplicant|connmand|wlan0|dhcp' /var/log/messages | tail -n 300")
		dir := gopodSessionLogDir()
		if err != nil || dir == "" {
			return
		}
		path := filepath.Join(dir, fmt.Sprintf("network_%s_%s.txt", strings.ReplaceAll(strings.ToLower(name), " ", "_"), time.Now().Format("150405")))
		if os.WriteFile(path, []byte(lines), 0o644) == nil {
			logger.Println("GOPOD_NETWORK " + name + " WiFi lines saved: " + path)
		}
	}()
}

// gopodRobotSSH: an SSH connection to the Brobot's robot (Wire-Pod's own SSH setup's client
// settings, setup/ssh.go).
func gopodRobotSSH(esn string) (*ssh.Client, error) {
	ip := ""
	for _, bot := range vars.BotInfo.Robots {
		if bot.Esn == esn {
			ip = bot.IPAddress
		}
	}
	if ip == "" {
		return nil, fmt.Errorf("no address")
	}
	key, err := os.ReadFile(gopodHomePath(gopodBrobotsSettings.BatteryWatch.SSHKey))
	if err != nil {
		return nil, fmt.Errorf("no ssh key")
	}
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return nil, err
	}
	return ssh.Dial("tcp", ip+":22", &ssh.ClientConfig{
		User:              "root",
		Auth:              []ssh.AuthMethod{ssh.PublicKeys(signer)},
		HostKeyCallback:   ssh.InsecureIgnoreHostKey(),
		HostKeyAlgorithms: []string{"ssh-rsa", "ecdsa-sha2-nistp256"},
		Timeout:           5 * time.Second,
	})
}

func gopodRobotRun(client *ssh.Client, cmd string) (string, error) {
	s, err := client.NewSession()
	if err != nil {
		return "", err
	}
	defer s.Close()
	out, err := s.Output(cmd)
	return string(out), err
}

// ## EOF
