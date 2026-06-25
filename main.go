package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type Aircraft struct {
	Hex        string   `json:"hex"`
	Flight     string   `json:"flight"`
	AltBaro    any      `json:"alt_baro"`
	AltGeom    float64  `json:"alt_geom"`
	GS         float64  `json:"gs"`
	IAS        float64  `json:"ias"`
	TAS        float64  `json:"tas"`
	Mach       float64  `json:"mach"`
	Track      float64  `json:"track"`
	TrackRate  float64  `json:"track_rate"`
	Roll       float64  `json:"roll"`
	MagHeading float64  `json:"mag_heading"`
	BaroRate   float64  `json:"baro_rate"`
	GeomRate   float64  `json:"geom_rate"`
	Squawk     string   `json:"squawk"`
	Emergency  string   `json:"emergency"`
	Category   string   `json:"category"`
	NavQNH     float64  `json:"nav_qnh"`
	NavAltMCP  float64  `json:"nav_altitude_mcp"`
	NavAltFMS  float64  `json:"nav_altitude_fms"`
	NavHeading float64  `json:"nav_heading"`
	NavModes   []string `json:"nav_modes"`
	Lat        float64  `json:"lat"`
	Lon        float64  `json:"lon"`
	SeenPos    float64  `json:"seen_pos"`
	Version    int      `json:"version"`
	Messages   int      `json:"messages"`
	Seen       float64  `json:"seen"`
	RSSI       float64  `json:"rssi"`
	NACp       int      `json:"nac_p"`
	NICBaro    int      `json:"nic_baro"`
	SIL        int      `json:"sil"`
	GVA        int      `json:"gva"`
	SDA        int      `json:"sda"`
}

func (a Aircraft) id() string {
	if f := strings.TrimSpace(a.Flight); f != "" {
		return f
	}
	return a.Hex
}

func (a Aircraft) altFt() (int, bool) {
	switch v := a.AltBaro.(type) {
	case float64:
		return int(v), true
	}
	return 0, false
}

type Dump struct {
	Now      float64    `json:"now"`
	Aircraft []Aircraft `json:"aircraft"`
}

type Event struct {
	Time      string `json:"time"`
	EventType string `json:"event"`
	ICAO      string `json:"icao"`
	Flight    string `json:"flight,omitempty"`
	// position
	Lat float64 `json:"lat,omitempty"`
	Lon float64 `json:"lon,omitempty"`
	// altitude
	AltBaro any     `json:"alt_baro,omitempty"`
	AltGeom float64 `json:"alt_geom,omitempty"`
	BaroRate float64 `json:"baro_rate,omitempty"`
	GeomRate float64 `json:"geom_rate,omitempty"`
	// velocity
	GS         float64 `json:"gs,omitempty"`
	IAS        float64 `json:"ias,omitempty"`
	TAS        float64 `json:"tas,omitempty"`
	Mach       float64 `json:"mach,omitempty"`
	Track      float64 `json:"track,omitempty"`
	TrackRate  float64 `json:"track_rate,omitempty"`
	Roll       float64 `json:"roll,omitempty"`
	MagHeading float64 `json:"mag_heading,omitempty"`
	// navigation
	NavQNH    float64  `json:"nav_qnh,omitempty"`
	NavAltMCP float64  `json:"nav_altitude_mcp,omitempty"`
	NavAltFMS float64  `json:"nav_altitude_fms,omitempty"`
	NavHeading float64 `json:"nav_heading,omitempty"`
	NavModes  []string `json:"nav_modes,omitempty"`
	// identity
	Squawk   string `json:"squawk,omitempty"`
	Emergency string `json:"emergency,omitempty"`
	Category string `json:"category,omitempty"`
	// signal
	RSSI     float64 `json:"rssi,omitempty"`
	Messages int     `json:"messages,omitempty"`
	// accuracy
	NACp    int    `json:"nac_p,omitempty"`
	NICBaro int    `json:"nic_baro,omitempty"`
	SIL     int    `json:"sil,omitempty"`
	GVA     int    `json:"gva,omitempty"`
	SDA     int    `json:"sda,omitempty"`
	Version int    `json:"version,omitempty"`
}

func eventFrom(typ string, a Aircraft) Event {
	return Event{
		Time:       time.Now().UTC().Format(time.RFC3339Nano),
		EventType:  typ,
		ICAO:       a.Hex,
		Flight:     strings.TrimSpace(a.Flight),
		Lat:        a.Lat,
		Lon:        a.Lon,
		AltBaro:    a.AltBaro,
		AltGeom:    a.AltGeom,
		BaroRate:   a.BaroRate,
		GeomRate:   a.GeomRate,
		GS:         a.GS,
		IAS:        a.IAS,
		TAS:        a.TAS,
		Mach:       a.Mach,
		Track:      a.Track,
		TrackRate:  a.TrackRate,
		Roll:       a.Roll,
		MagHeading: a.MagHeading,
		NavQNH:     a.NavQNH,
		NavAltMCP:  a.NavAltMCP,
		NavAltFMS:  a.NavAltFMS,
		NavHeading: a.NavHeading,
		NavModes:   a.NavModes,
		Squawk:     a.Squawk,
		Emergency:  a.Emergency,
		Category:   a.Category,
		RSSI:       a.RSSI,
		Messages:   a.Messages,
		NACp:       a.NACp,
		NICBaro:    a.NICBaro,
		SIL:        a.SIL,
		GVA:        a.GVA,
		SDA:        a.SDA,
		Version:    a.Version,
	}
}

func readAircraft(dir string) (map[string]Aircraft, error) {
	data, err := os.ReadFile(filepath.Join(dir, "aircraft.json"))
	if err != nil {
		return nil, err
	}
	var d Dump
	if err := json.Unmarshal(data, &d); err != nil {
		return nil, err
	}
	m := make(map[string]Aircraft, len(d.Aircraft))
	for _, a := range d.Aircraft {
		m[a.Hex] = a
	}
	return m, nil
}

type aircraftState struct {
	lat    float64
	lon    float64
	squawk string
	flight string
}

func runDaemon(dir string, interval time.Duration, out io.Writer) error {
	enc := json.NewEncoder(out)
	prev := map[string]Aircraft{}
	state := map[string]aircraftState{}

	for {
		current, err := readAircraft(dir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "read error: %v\n", err)
			time.Sleep(interval)
			continue
		}

		// new aircraft and position/state updates
		for hex, a := range current {
			prev_a, seen := prev[hex]
			if !seen {
				enc.Encode(eventFrom("appeared", a))
				state[hex] = aircraftState{lat: a.Lat, lon: a.Lon, squawk: a.Squawk, flight: strings.TrimSpace(a.Flight)}
				continue
			}

			st := state[hex]

			// new position fix
			if a.Lat != 0 && a.Lon != 0 && (a.Lat != st.lat || a.Lon != st.lon) {
				enc.Encode(eventFrom("position", a))
				st.lat = a.Lat
				st.lon = a.Lon
			}

			// squawk changed
			if a.Squawk != "" && a.Squawk != st.squawk {
				enc.Encode(eventFrom("squawk_change", a))
				st.squawk = a.Squawk
			}

			// callsign identified
			flight := strings.TrimSpace(a.Flight)
			if flight != "" && flight != st.flight {
				enc.Encode(eventFrom("identified", a))
				st.flight = flight
			}

			_ = prev_a
			state[hex] = st
		}

		// disappeared
		for hex, a := range prev {
			if _, ok := current[hex]; !ok {
				enc.Encode(eventFrom("disappeared", a))
				delete(state, hex)
			}
		}

		prev = current
		time.Sleep(interval)
	}
}

// --- stats mode ---

type statsEntry struct {
	aircraft Aircraft
	maxAlt   int
	maxGS    float64
}

func runStats(dir string) error {
	files, err := filepath.Glob(filepath.Join(dir, "history_*.json"))
	if err != nil || len(files) == 0 {
		return fmt.Errorf("no history files found in %s", dir)
	}

	seen := map[string]*statsEntry{}
	for _, f := range files {
		data, err := os.ReadFile(f)
		if err != nil {
			continue
		}
		var d Dump
		if err := json.Unmarshal(data, &d); err != nil {
			continue
		}
		for _, a := range d.Aircraft {
			s, ok := seen[a.Hex]
			if !ok {
				s = &statsEntry{aircraft: a}
				seen[a.Hex] = s
			}
			if strings.TrimSpace(s.aircraft.Flight) == "" && strings.TrimSpace(a.Flight) != "" {
				s.aircraft.Flight = a.Flight
			}
			if alt, ok := a.altFt(); ok && alt > s.maxAlt {
				s.maxAlt = alt
			}
			if a.GS > s.maxGS {
				s.maxGS = a.GS
			}
		}
	}

	all := make([]*statsEntry, 0, len(seen))
	for _, s := range seen {
		all = append(all, s)
	}
	sort.Slice(all, func(i, j int) bool {
		return all[i].aircraft.id() < all[j].aircraft.id()
	})

	identified, unidentified := 0, 0
	for _, s := range all {
		if strings.TrimSpace(s.aircraft.Flight) != "" {
			identified++
		} else {
			unidentified++
		}
	}

	var highestAlt, fastest *statsEntry
	for _, s := range all {
		if highestAlt == nil || s.maxAlt > highestAlt.maxAlt {
			highestAlt = s
		}
		if fastest == nil || s.maxGS > fastest.maxGS {
			fastest = s
		}
	}

	fmt.Printf("Unique aircraft seen: %d (%d identified, %d hex-only)\n\n", len(all), identified, unidentified)
	if highestAlt != nil {
		fmt.Printf("Highest altitude: %s at %d ft\n", highestAlt.aircraft.id(), highestAlt.maxAlt)
	}
	if fastest != nil {
		fmt.Printf("Fastest:          %s at %.0f kts\n\n", fastest.aircraft.id(), fastest.maxGS)
	}

	fmt.Printf("%-12s  %8s  %8s\n", "Flight", "Max Alt", "Max GS")
	fmt.Println(strings.Repeat("-", 34))
	for _, s := range all {
		alt := "-"
		if s.maxAlt > 0 {
			alt = fmt.Sprintf("%d ft", s.maxAlt)
		}
		gs := "-"
		if s.maxGS > 0 {
			gs = fmt.Sprintf("%.0f kts", s.maxGS)
		}
		fmt.Printf("%-12s  %8s  %8s\n", s.aircraft.id(), alt, gs)
	}
	return nil
}

func main() {
	daemonCmd := flag.NewFlagSet("daemon", flag.ExitOnError)
	daemonDir := daemonCmd.String("input", "/tmp/adsb", "readsb JSON output directory")
	daemonOutput := daemonCmd.String("output", "", "output file path (default: stdout)")
	daemonInterval := daemonCmd.Duration("interval", time.Second, "poll interval")

	statsCmd := flag.NewFlagSet("stats", flag.ExitOnError)
	statsDir := statsCmd.String("input", "/tmp/adsb", "readsb JSON output directory")

	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: adsbstats <daemon|stats> [flags]")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "daemon":
		daemonCmd.Parse(os.Args[2:])
		var out io.Writer = os.Stdout
		if *daemonOutput != "" {
			f, err := os.OpenFile(*daemonOutput, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				fmt.Fprintf(os.Stderr, "open output: %v\n", err)
				os.Exit(1)
			}
			defer f.Close()
			out = f
		}
		if err := runDaemon(*daemonDir, *daemonInterval, out); err != nil {
			fmt.Fprintf(os.Stderr, "daemon: %v\n", err)
			os.Exit(1)
		}
	case "stats":
		statsCmd.Parse(os.Args[2:])
		if err := runStats(*statsDir); err != nil {
			fmt.Fprintf(os.Stderr, "stats: %v\n", err)
			os.Exit(1)
		}
	default:
		fmt.Fprintf(os.Stderr, "unknown command %q\n", os.Args[1])
		os.Exit(1)
	}
}
