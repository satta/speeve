package cmd

import (
	"bytes"
	"math"
	"math/rand"
	"os"
	"runtime/pprof"
	"time"

	"github.com/satta/speeve/generator"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func mainfunc(cmd *cobra.Command, args []string) {
	rand.Seed(time.Now().UTC().UnixNano())

	runDuration := viper.GetDuration("duration")
	runMaximumEvents := viper.GetInt64("total")

	if runDuration > 0 && runMaximumEvents > 0 {
		log.Fatalf("cannot use maximum duration and maximum events at the same time")
	}
	if runDuration > 0 && runDuration < 1*time.Second {
		log.Fatalf("runtime needs to be a least 1 second, got %s", runDuration)
	}

	verbose := viper.GetBool("verbose")
	if verbose {
		log.SetLevel(log.DebugLevel)
	}
	log.SetOutput(os.Stderr)

	// Start logging and profiling
	pprofFile := viper.GetString("pproffile")
	if len(pprofFile) > 0 {
		f, err := os.Create("speeve.prof")
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	profileFile := viper.GetString("profile")
	perSec := viper.GetInt64("persec")
	numGens := viper.GetInt64("parallel")

	chunkSize := int64(math.Ceil(float64(perSec) / 100.0))
	outChan := make(chan []byte, chunkSize)
	log.Debugf("chunksize is %d", chunkSize)

	if runMaximumEvents > 0 && runMaximumEvents < chunkSize {
		log.Warnf("maximum total event number requested in less than chunk "+
			"size (%d), will at least emit %d events", chunkSize, chunkSize)
	}

	for i := int64(0); i < numGens; i++ {
		go func() {
			fg, err := generator.MakeFlowGenerator(profileFile)
			if err != nil {
				log.Fatal(err)
			}
			for {
				fg.EmitFlow(outChan)
			}
		}()
	}

	ticktime := int64(1*time.Second) / (perSec / chunkSize)
	log.Debugf("ticktime is %v", time.Duration(ticktime))

	ticker := time.NewTicker(time.Duration(ticktime))
	done := make(chan bool)

	var i int64
	go func() {
		for {
			select {
			case <-done:
				done <- true
				return
			case <-ticker.C:
				var j int64
				var buf bytes.Buffer
				if runMaximumEvents > 0 && i > runMaximumEvents {
					done <- true
					return
				}
				for j = 0; j < chunkSize; j++ {
					buf.Write(<-outChan)
				}
				os.Stdout.Write(buf.Bytes())
				buf.Reset()
				i += chunkSize
			}
		}
	}()

	if runDuration > 0 {
		time.Sleep(runDuration)
		ticker.Stop()
		done <- true
	}
	<-done
}

var runCmd = &cobra.Command{
	Use:   "spew",
	Short: "generate EVE-JSON",
	Long:  `The 'spew' command starts EVE-JSON generation.`,
	Run:   mainfunc,
}

func init() {
	rootCmd.AddCommand(runCmd)

	runCmd.PersistentFlags().StringP("profile", "p", "profile.yaml", "filename of traffic profile definition file")
	viper.BindPFlag("profile", runCmd.PersistentFlags().Lookup("profile"))
	runCmd.PersistentFlags().Uint32P("parallel", "j", 2, "number of generator tasks to run in parallel")
	viper.BindPFlag("parallel", runCmd.PersistentFlags().Lookup("parallel"))
	runCmd.PersistentFlags().Uint32P("persec", "s", 1000, "number of events/s to emit")
	viper.BindPFlag("persec", runCmd.PersistentFlags().Lookup("persec"))
	runCmd.PersistentFlags().DurationP("duration", "d", 0, "duration of run")
	viper.BindPFlag("duration", runCmd.PersistentFlags().Lookup("duration"))
	runCmd.PersistentFlags().Uint64P("total", "n", 0, "total number of events to emit")
	viper.BindPFlag("total", runCmd.PersistentFlags().Lookup("total"))
	runCmd.PersistentFlags().StringP("pproffile", "", "", "filename to write pprof profiling info into")
	viper.BindPFlag("proffile", runCmd.PersistentFlags().Lookup("proffile"))

	runCmd.PersistentFlags().BoolP("verbose", "v", false, "verbose mode")
	viper.BindPFlag("verbose", runCmd.PersistentFlags().Lookup("verbose"))
}
