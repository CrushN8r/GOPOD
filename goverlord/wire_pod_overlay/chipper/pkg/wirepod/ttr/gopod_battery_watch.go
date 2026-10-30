// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_battery_watch.go
package wirepod_ttr

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/fforchino/vector-go-sdk/pkg/vector"
	"github.com/fforchino/vector-go-sdk/pkg/vectorpb"
	"github.com/kercre123/wire-pod/chipper/pkg/logger"
	"github.com/kercre123/wire-pod/chipper/pkg/vars"
)

// The battery watchdog (operator 2026-10-15: "set up a background watchdog to
// monitor battery level over time? Over the course of a session run? If let's
// say the battery level drops below a threshold (like from a lot of use), have
// the brobot comment they're feeling dizzy and need help getting home, lol").
// For the whole Wire-Pod session, every battery_watch.interval_seconds each
// Brobot's battery is read (GOPOD_BATTERY_WATCH, in the log over time). Below
// low_volts, off the charger, the Brobot says one of his battery_watch.lines,
// with his bubble - once; he says it again only after he's been charged back to
// recover_volts (or set on the charger). Never in the middle of a reply or a run:
// he waits for both to end. low_volts is golden rule 6's line, 3.7 V (Brobot 2's
// crashes came at 3.61-3.62 V).

// gopodBatteryLow: whether each Brobot has had his say for this drop.
var gopodBatteryLow = struct {
	sync.Mutex
	said map[string]bool
}{said: map[string]bool{}}

// gopodBatteryDecide: one reading -> whether to say the line now, and the new
// said state. said: he's had his say for this drop already.
func gopodBatteryDecide(said bool, volts float64, charging, onCharger bool) (sayNow, saidAfter bool) {
	bw := gopodBrobotsSettings.BatteryWatch
	if charging || onCharger || volts >= bw.RecoverVolts {
		return false, false
	}
	if volts > 0 && volts < bw.LowVolts && !said {
		return true, true
	}
	return false, said
}

func init() {
	go gopodBatteryWatch()
}

func gopodBatteryWatch() {
	for {
		bw := gopodBrobotsSettings.BatteryWatch
		every := gopodSecondsToDuration(bw.IntervalSeconds)
		if every <= 0 {
			every = time.Minute
		}
		time.Sleep(every)
		if !bw.Enabled {
			continue
		}
		for _, esn := range []string{gopodBrobot1ESN, gopodBrobot2ESN} {
			gopodBatteryCheck(esn)
		}
	}
}

func gopodBatteryCheck(esn string) {
	var robot *vector.Vector
	for _, bot := range vars.BotInfo.Robots {
		if bot.Esn == esn {
			robot, _ = vector.New(vector.WithSerialNo(esn), vector.WithToken(bot.GUID), vector.WithTarget(bot.IPAddress+":443"))
			robot = GopodRobotsOffConn(robot) // robots off (gopod_robots_off.go)
			break
		}
	}
	if robot == nil {
		return
	}
	name := gopodBrobotName(esn)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	b, err := robot.Conn.BatteryState(ctx, &vectorpb.BatteryStateRequest{})
	cancel()
	if err != nil {
		logger.Println("GOPOD_BATTERY_WATCH " + name + " unknown: " + err.Error())
		gopodBatteryRemember(esn, "unknown: "+err.Error())
		return
	}
	volts := float64(b.GetBatteryVolts())
	gopodBatteryRemember(esn, fmt.Sprintf("%.2f V, %s%s", volts, strings.TrimPrefix(b.GetBatteryLevel().String(), "BATTERY_LEVEL_"),
		map[bool]string{true: ", on the charger", false: ""}[b.GetIsOnChargerPlatform()]+map[bool]string{true: ", charging", false: ""}[b.GetIsCharging()]))
	logger.Println(fmt.Sprintf("GOPOD_BATTERY_WATCH %s volts=%.2f level=%s charging=%t on_charger=%t",
		name, volts, strings.TrimPrefix(b.GetBatteryLevel().String(), "BATTERY_LEVEL_"), b.GetIsCharging(), b.GetIsOnChargerPlatform()))
	gopodBatteryLow.Lock()
	sayNow, said := gopodBatteryDecide(gopodBatteryLow.said[esn], volts, b.GetIsCharging(), b.GetIsOnChargerPlatform())
	gopodBatteryLow.said[esn] = said
	gopodBatteryLow.Unlock()
	if sayNow {
		go gopodBatterySay(name, volts)
	}
}

// gopodBatteryLast (2026-10-17, system-status): each Brobot's last reading.
var gopodBatteryLast = struct {
	sync.Mutex
	text map[string]string
	at   map[string]time.Time
}{text: map[string]string{}, at: map[string]time.Time{}}

func gopodBatteryRemember(esn, text string) {
	gopodBatteryLast.Lock()
	gopodBatteryLast.text[esn], gopodBatteryLast.at[esn] = text, time.Now()
	gopodBatteryLast.Unlock()
}

// gopodBatterySay: once no reply and no run is going, the Brobot says one of his
// lines (gopodRecordSay: his bubble and his voice).
func gopodBatterySay(name string, volts float64) {
	lines := gopodBrobotsSettings.BatteryWatch.Lines[name]
	if len(lines) == 0 {
		return
	}
	// 2026-10-15 (operator: "This type of interjection would happen at the next
	// safe spot to interrupt * during live, not scripted runs"): live chat only,
	// between replies - never while a reply, a run or a draft is going.
	for gopodReplyBusy() || gopodRunActive() || gopodScriptPlaying.Load() {
		time.Sleep(2 * time.Second)
	}
	guard, warning, reply, line := gopodBatteryLines(name, volts, lines[rand.Intn(len(lines))])
	logger.Println(fmt.Sprintf("GOPOD_BATTERY_WATCH %s low (%.2f V): %s", name, volts, line))
	// 2026-10-26 (operator: "at the next safe spot, Intel can interject * use
	// the duo response-chain template, Intel 404 to brobot N * a warning
	// message, then what say you? * Brobot N, Nice try ..., yeah, I feel
	// dizzy"): his guard (battery_watch.guards) warns and hands over, then he
	// answers - the comeback opener, then his line (battery_watch.reply).
	if guard != "" && gopodReporterSayTo(guard, "", warning, true, true) {
		line = reply
	}
	gopodRecordSay(name, line)
}

// gopodBatteryLines: what a low battery says - his guard (none: ""), the guard's
// warning and handover, his reply after it, and his line alone (with no guard).
// The live watch and the pre-show check (gopod_preshow.go) build them the same way.
func gopodBatteryLines(name string, volts float64, line string) (guard, warning, reply, alone string) {
	bw := gopodBrobotsSettings.BatteryWatch
	guard = bw.Guards[name]
	if guard == "" {
		return "", "", line, line
	}
	warn := strings.NewReplacer("{brobot}", name, "{volts}", fmt.Sprintf("%.1f", volts)).Replace(bw.GuardWarning)
	hand := gopodNicknameFill(guard, strings.ReplaceAll(bw.GuardHandover, "{brobot}", name))
	come := gopodNicknameFill(name, gopodBrobotsSettings.Openers["comeback"])
	reply = strings.Join(strings.Fields(strings.NewReplacer("{comeback}", come, "{line}", line).Replace(bw.Reply)), " ")
	return guard, strings.TrimSpace(warn + " " + hand), reply, line
}

// ## EOF
