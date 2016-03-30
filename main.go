package main

import (
	"flag"
	"fmt"
	log "github.com/Sirupsen/logrus"
	goconf "github.com/akrennmair/goconf"
	// _ "github.com/go-sql-driver/mysql"
	"github.com/influxdata/influxdb/client/v2"
	"os"
	"time"
)

type Cfg struct {
	LogFile      string
	LogLevel     int
	FalconClient string
	Endpoint     string

	User string
	Pass string
	Host string
	Port int
}

var cfg Cfg

func init() {
	var cfgFile string
	flag.StringVar(&cfgFile, "c", "mon.cfg", "myMon configure file")
	flag.Parse()

	if _, err := os.Stat(cfgFile); err != nil {
		if os.IsNotExist(err) {
			log.WithField("cfg", cfgFile).Fatalf("myMon config file does not exists: %v", err)
		}
	}

	if err := cfg.readConf(cfgFile); err != nil {
		log.Fatalf("Read configure file failed: %v", err)
	}

	// Init log file
	log.SetFormatter(&log.JSONFormatter{})
	log.SetLevel(log.Level(cfg.LogLevel))

	if cfg.LogFile != "" {
		f, err := os.OpenFile(cfg.LogFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err == nil {
			log.SetOutput(f)
			return
		}
	}
	log.SetOutput(os.Stderr)
}

func (conf *Cfg) readConf(file string) error {
	c, err := goconf.ReadConfigFile(file)
	if err != nil {
		return err
	}

	conf.LogFile, err = c.GetString("default", "log_file")
	if err != nil {
		return err
	}

	conf.LogLevel, err = c.GetInt("default", "log_level")
	if err != nil {
		return err
	}

	conf.FalconClient, err = c.GetString("default", "falcon_client")
	if err != nil {
		return err
	}

	conf.Endpoint, err = c.GetString("default", "endpoint")
	if err != nil {
		return err
	}

	conf.User, err = c.GetString("mysql", "user")
	if err != nil {
		return err
	}

	conf.Pass, err = c.GetString("mysql", "password")
	if err != nil {
		return err
	}

	conf.Host, err = c.GetString("mysql", "host")
	if err != nil {
		return err
	}

	conf.Port, err = c.GetInt("mysql", "port")
	return err
}

func main() {
	fmt.Println(cfg.User)
	// Make client
	c, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:     "http://localhost:8086",
		Username: "admin",
		Password: "admin",
	})

	// Create a new point batch
	bp, err := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  "mysql",
		Precision: "s",
	})

	// Create a point and add to batch
	tags := map[string]string{"cpu": "cpu-total"}
	fields := map[string]interface{}{
		"idle":   10.1,
		"system": 53.3,
		"user":   46.6,
	}
	pt, err := client.NewPoint("cpu_usage", tags, fields, time.Now())
	CheckError(err)
	bp.AddPoint(pt)

	// Write the batch
	c.Write(bp)
}

func CheckError(err error) {
	if err != nil {
		panic(err)
	}
}
