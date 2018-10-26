// Library provides methods the load configuration variable from either.
// the command line or ENV variables. Interface is similar to flag library
//
// The name should be provided in lowercase with dashes which is typical of most
// command-line interfaces
//  e.g.
//       db-name
//       healthcheck-seconds
//
// The name of the ENV variable is computed by converting to uppercase and replacing
// dashes with underscores how environment variables are typically specified
//  e.g.
//      DB_NAME
//      HEALTHCHECK_SECONDS
//
// The ENV variable is always queried first with the command-line variable taking
// precidence if specified
//
// TODO make a configurable backend to support different types of config
// systems with a hierarchy
package flag

import (
	"flag"
	"fmt"
	"io"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"
)

func Parse() {
	flag.Parse()
}

// Pull String variable from ENV or command-line
func StringVar(p *string, name, value, usage string) {
	v, ok := Getenv(name)
	if ok {
		value = v
	}

	flag.StringVar(p, name, value, usage)
}

// Pull Int variable from ENV or command-line
func IntVar(p *int, name string, value int, usage string) {
	v, ok := Getenv(name)
	if ok {
		i, err := strconv.Atoi(v)
		if err == nil {
			value = i
		}
	}
	flag.IntVar(p, name, value, usage)
}

// Pull Float64 variable from ENV or command-line
func Float64Var(p *float64, name string, value float64, usage string) {
	v, ok := Getenv(name)
	if ok {
		f, err := strconv.ParseFloat(v, 64)
		if err == nil {
			value = f
		}
	}
	flag.Float64Var(p, name, value, usage)
}

// Pull Duration variable from ENV or command-line
func DurationVar(p *time.Duration, name string, value time.Duration, usage string) {
	v, ok := Getenv(name)
	if ok {
		i, err := strconv.Atoi(v)
		if err == nil {
			value = time.Duration(i)
		}
	}
	flag.DurationVar(p, name, value, usage)
}

// Pull Bool variable from ENV or command-line
func BoolVar(p *bool, name string, value bool, usage string) {
	v, ok := Getenv(name)
	if ok {
		b, err := strconv.ParseBool(v)
		if err == nil {
			value = b
		}
	}
	flag.BoolVar(p, name, value, usage)
}

func toEnvKey(name string) string {
	key := strings.ToUpper(strings.Replace(name, "-", "_", -1))
	return key
}

func Getenv(key string) (string, bool) {
	str := os.Getenv(toEnvKey(key))
	if str == "" {
		return "", false
	}
	return str, true
}

// Prints out the full and environment and configuration
func PrintEnv(writer io.Writer) {
	fmt.Fprintf(writer, "Processors:\n\tCPUs: %d\n\tGOMAXPROCS: %d\n", runtime.NumCPU(), runtime.GOMAXPROCS(10))

	fmt.Fprintln(writer, "Environment:")
	for _, env := range os.Environ() {
		fmt.Fprintf(writer, "\t%s\n", env)
	}

	fmt.Fprintln(writer, "Configuration:")
	flag.VisitAll(func(f *flag.Flag) {
		if f.Name == "help" || f.Name == "version" {
			return
		}
		fmt.Fprintf(writer, "\t%s: %s\n", toEnvKey(f.Name), f.Value)
	})
}
